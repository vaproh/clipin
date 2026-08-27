package http

import (
	"context"
	"net/http"
	"slices"
	"time"

	"clipin/apps/api/internal/auth"
	"clipin/apps/api/internal/config"
	"clipin/apps/api/internal/db"
	"clipin/apps/api/internal/http/handlers"
	"clipin/apps/api/internal/middleware"
	"clipin/apps/api/internal/payout"
	"clipin/apps/api/internal/redis"
	"clipin/apps/api/internal/service"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
)

type AppDependencies struct {
	Config *config.Config
	DB     *db.Database
	Redis  *redis.Client
}

func (a *AppDependencies) CheckDB(ctx context.Context) bool {
	if a.DB == nil {
		return false
	}
	return a.DB.Ping(ctx) == nil
}

func (a *AppDependencies) CheckRedis(ctx context.Context) bool {
	if a.Redis == nil {
		return false
	}
	return a.Redis.Ping(ctx) == nil
}

func NewRouter(deps *AppDependencies) http.Handler {
	r := chi.NewRouter()

	// --- Middleware (all Use() calls must precede any Route/Group/Handle) ---
	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(chimw.Timeout(60 * time.Second))

	// CORS: echo back the request origin only if it is allowlisted.
	allowedOrigins := deps.Config.AllowedOrigins
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin != "" && slices.Contains(allowedOrigins, origin) {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}
			next.ServeHTTP(w, r)
		})
	})

	// Rate limiter (in-memory, suitable for single-instance).
	rl := middleware.NewRateLimiter()
	r.Use(rl.Middleware(100, middleware.IPKey))

	// --- Create sub-routers BEFORE humachi so chi doesn't panic ---
	internal := r.Group(func(r chi.Router) {
		r.Use(auth.InternalAuthMiddleware(deps.Config.VerifierAPIKey))
	})

	authenticated := r.Group(func(r chi.Router) {
		var userStore auth.UserStore
		if deps.DB != nil {
			userStore = deps.DB.Queries
		}
		r.Use(auth.AuthMiddleware(auth.NewJWKSProvider(deps.Config.ClerkJWKSURL)))
		r.Use(auth.SessionMiddleware(userStore, deps.Redis))
	})

	// --- Huma OpenAPI config ---
	humaConfig := huma.DefaultConfig("ClipIN API", "1.0.0")
	humaConfig.OpenAPIPath = "/openapi.json"
	humaConfig.DocsPath = "/docs"
	humaConfig.Info.Description = "ClipIN Performance Clipping Marketplace API"

	// --- Public API (root) ---
	api := humachi.New(r, humaConfig)
	handlers.RegisterHealthHandler(api, deps, deps.Config.Env)
	handlers.RegisterWebhookHandlers(api)

	// Campaign marketplace service (public endpoints)
	var campaignSvc *service.CampaignService
	if deps.DB != nil {
		campaignCache := service.NewCampaignCache(deps.Redis)
		campaignSvc = service.NewCampaignService(deps.DB.Queries, campaignCache)
		if ledgerSvc := service.NewLedgerService(deps.DB.Queries); ledgerSvc != nil {
			campaignSvc.WithLedger(ledgerSvc)
		}
	}
	handlers.RegisterCampaignHandlers(api, campaignSvc)

	// Clipper public profile (public endpoints)
	if deps.DB != nil {
		clipperCache := service.NewClipperProfileCache(deps.Redis)
		handlers.RegisterClipperHandlers(api, deps.DB.Queries, clipperCache)
	}

	// Public leaderboard (no auth)
	if deps.DB != nil {
		leaderboardCache := service.NewLeaderboardCache(deps.Redis)
		handlers.RegisterLeaderboardHandlers(api, deps.DB.Queries, leaderboardCache)
	}

	// --- Internal API (verifier key auth) ---
	intCfg := humaConfig
	intCfg.OpenAPIPath = ""
	intCfg.DocsPath = ""
	internalAPI := humachi.New(internal, intCfg)
	if deps.DB != nil {
		verificationSvc := service.NewVerificationService(deps.DB.Queries)
		handlers.RegisterInternalVerificationHandlers(internalAPI, verificationSvc)
	}

	// --- Authenticated API (Clerk JWT + session) ---
	authCfg := humaConfig
	authCfg.OpenAPIPath = ""
	authCfg.DocsPath = ""
	authenticatedAPI := humachi.New(authenticated, authCfg)

	var userStore auth.UserStore
	if deps.DB != nil {
		userStore = deps.DB.Queries
	}
	handlers.RegisterUserHandlers(authenticatedAPI, userStore)

	if campaignSvc != nil {
		handlers.RegisterCampaignOwnerHandlers(authenticatedAPI, campaignSvc)
	}

	if deps.DB != nil {
		analyticsCache := service.NewCampaignAnalyticsCache(deps.Redis)
		handlers.RegisterCampaignAnalyticsHandlers(authenticatedAPI, deps.DB.Queries, analyticsCache)
	}

	if deps.DB != nil {
		submissionSvc := service.NewSubmissionService(deps.DB.Queries)
		submissionSvc.WithLedger(service.NewLedgerService(deps.DB.Queries))
		notifSvc := service.NewNotificationServiceWithCache(deps.DB.Queries, deps.Redis)
		submissionSvc.WithNotifications(notifSvc)
		handlers.RegisterSubmissionHandlers(authenticatedAPI, submissionSvc)
	}

	if deps.DB != nil {
		verificationSvc := service.NewVerificationService(deps.DB.Queries)
		handlers.RegisterVerificationStatusHandlers(authenticatedAPI, verificationSvc)
	}

	if deps.DB != nil {
		ledgerSvc := service.NewLedgerService(deps.DB.Queries)
		handlers.RegisterLedgerHandlers(authenticatedAPI, ledgerSvc, campaignSvc)
	}

	if deps.DB != nil {
		rp := payout.NewRazorpayProviderOrStub(
			deps.Config.RazorpayKeyID,
			deps.Config.RazorpayKeySecret,
			deps.Config.RazorpayAccountNumber,
		)
		payoutSvc := service.NewPayoutService(deps.DB.Queries, rp)
		payoutSvc.WithNotifications(service.NewNotificationServiceWithCache(deps.DB.Queries, deps.Redis))
		handlers.RegisterPayoutHandlers(authenticatedAPI, payoutSvc)
	}

	if deps.DB != nil {
		auditSvc := service.NewAuditService(deps.DB.Queries)
		fraudSvc := service.NewFraudService(deps.DB.Queries)
		adminGroup := authenticated.Group(nil)
		adminGroup.Use(auth.AdminOnly)
		adminAPI := humachi.New(adminGroup, authCfg)
		handlers.RegisterAdminHandlers(adminAPI, deps.DB.Queries, auditSvc, fraudSvc)
	}

	// --- Notification handlers (authenticated) ---
	if deps.DB != nil {
		notifSvc := service.NewNotificationServiceWithCache(deps.DB.Queries, deps.Redis)
		handlers.RegisterNotificationHandlers(authenticatedAPI, notifSvc)
	}

	// --- Social account handlers (authenticated) ---
	if deps.DB != nil {
		socialAccountSvc := service.NewSocialAccountService(deps.DB.Queries)
		handlers.RegisterSocialAccountHandlers(authenticatedAPI, socialAccountSvc)
	}

	// --- Template handlers (public GET + admin POST) ---
	if deps.DB != nil {
		templateSvc := service.NewTemplateServiceWithCache(deps.DB.Queries, deps.Redis)
		handlers.RegisterTemplateHandlers(api, templateSvc)
	}

	return r
}
