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
	"clipin/apps/api/internal/redis"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
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

	// Middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// CORS: echo back the request origin only if it is allowlisted.
	// Non-matching origins get no header, so the browser blocks the response.
	// An empty allowlist rejects all cross-origin requests (safer than *).
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

	// Huma OpenAPI router
	humaConfig := huma.DefaultConfig("ClipIN API", "1.0.0")
	humaConfig.OpenAPIPath = "/openapi.json"
	humaConfig.DocsPath = "/docs"
	humaConfig.Info.Description = "ClipIN Performance Clipping Marketplace API"

	api := humachi.New(r, humaConfig)

	// Handlers (public: /health, /openapi.json, /docs stay on the root router)
	handlers.RegisterHealthHandler(api, deps, deps.Config.Env)

	// Authenticated routes live in this group, behind Clerk JWT verification
	// and session loading. The middleware no-ops when CLERK_JWKS_URL is unset
	// (development).
	//
	// The group binds its own huma.API so protected handlers stay behind the
	// middleware chain; their OpenAPI schemas are not served on the root API.
	r.Group(func(r chi.Router) {
		api := humachi.New(r, humaConfig)

		var userStore auth.UserStore
		if deps.DB != nil {
			userStore = deps.DB.Queries
		}

		r.Use(auth.AuthMiddleware(auth.NewJWKSProvider(deps.Config.ClerkJWKSURL)))
		r.Use(auth.SessionMiddleware(userStore))

		handlers.RegisterUserHandlers(api, userStore)
	})

	return r
}
