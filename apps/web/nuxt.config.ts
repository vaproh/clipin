import { resolve } from 'path'

const e2e = !!process.env.E2E_TEST

// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  compatibilityDate: '2024-11-01',
  devtools: { enabled: true },

  experimental: {
    appManifest: false,
  },

  alias: e2e
    ? {
        '@clerk/nuxt': resolve(__dirname, 'e2e/clerk-stub.ts'),
        '@clerk/vue': resolve(__dirname, 'e2e/clerk-stub.ts'),
      }
    : {},

  imports: e2e
    ? {
        presets: [
          {
            from: resolve(__dirname, 'e2e/clerk-stub.ts'),
            imports: ['useAuth', 'useUser', 'useClerk', 'useSignIn', 'useSignUp'],
          },
        ],
      }
    : undefined,

  components: e2e
    ? {
        dirs: [{ path: resolve(__dirname, 'e2e/components'), pathPrefix: false }],
      }
    : undefined,

  modules: [
    ...(process.env.E2E_TEST
      ? []
      : [['@clerk/nuxt', {
          publishableKey: process.env.CLERK_PUBLISHABLE_KEY || process.env.NUXT_PUBLIC_CLERK_PUBLISHABLE_KEY,
          secretKey: process.env.CLERK_SECRET_KEY || process.env.NUXT_CLERK_SECRET_KEY,
          appearance: {
            variables: {
              colorPrimary: '#ffffff',
              colorBackground: '#0a0a0a',
              colorInputBackground: '#121212',
              colorInputText: '#ffffff',
              colorText: '#ffffff',
              colorTextSecondary: '#a3a3a3',
              colorBorder: '#262626',
              borderRadius: '8px',
            },
            elements: {
              card: 'bg-neutral-950 border border-neutral-800 shadow-2xl shadow-black rounded-lg',
              headerTitle: 'text-white font-semibold tracking-tight',
              headerSubtitle: 'text-neutral-400 text-xs font-normal',
              socialButtonsBlockButton: 'bg-neutral-900 border border-neutral-800 text-white hover:bg-neutral-800 transition-colors',
              formButtonPrimary: 'bg-white text-black hover:bg-neutral-200 transition-colors font-medium text-xs rounded h-9',
              formFieldInput: 'bg-neutral-900 border border-neutral-800 text-white focus:border-neutral-500 rounded text-xs h-9',
              footerActionLink: 'text-white hover:underline font-medium',
            },
          },
        }] as const]),
    '@nuxtjs/tailwindcss',
    '@vueuse/nuxt',
    ...(process.env.E2E_TEST ? [] : ['@vite-pwa/nuxt']),
  ] as any,

  pwa: {
    registerType: 'autoUpdate',
    manifest: {
      name: 'ClipIN - Performance Clipping Marketplace',
      short_name: 'ClipIN',
      description: 'Create clips, publish them, and earn from verified views.',
      start_url: '/app',
      scope: '/',
      display: 'standalone',
      background_color: '#010102',
      theme_color: '#010102',
      icons: [
        { src: '/icon.svg', sizes: 'any', type: 'image/svg+xml', purpose: 'any' },
        { src: '/maskable-icon.svg', sizes: 'any', type: 'image/svg+xml', purpose: 'maskable' },
      ],
    },
    workbox: {
      cleanupOutdatedCaches: true,
      navigateFallbackDenylist: [/^\/api\//],
    },
    devOptions: {
      enabled: false,
    },
  },

  runtimeConfig: {
    clerkSecretKey: process.env.CLERK_SECRET_KEY || process.env.NUXT_CLERK_SECRET_KEY,
    clerk: {
      secretKey: process.env.CLERK_SECRET_KEY || process.env.NUXT_CLERK_SECRET_KEY,
    },
    public: {
      clerkPublishableKey: process.env.CLERK_PUBLISHABLE_KEY || process.env.NUXT_PUBLIC_CLERK_PUBLISHABLE_KEY,
      clerk: {
        publishableKey: process.env.CLERK_PUBLISHABLE_KEY || process.env.NUXT_PUBLIC_CLERK_PUBLISHABLE_KEY,
      },
      apiBase: process.env.NUXT_PUBLIC_API_BASE || 'http://localhost:8080',
      verifierBase: process.env.NUXT_PUBLIC_VERIFIER_BASE || 'http://localhost:8081',
    },
  },

  css: ['~/assets/css/main.css'],

  app: {
    head: {
      title: 'ClipIN - Performance Clipping Marketplace',
      htmlAttrs: { lang: 'en' },
      meta: [
        { charset: 'utf-8' },
        { name: 'viewport', content: 'width=device-width, initial-scale=1' },
        {
          name: 'description',
          content: 'India\'s open marketplace for performance clipping. Fund campaigns, create clips, earn from verified views.',
        },
        { name: 'theme-color', content: '#000000' },
        { property: 'og:type', content: 'website' },
        { property: 'og:title', content: 'ClipIN - Performance Clipping Marketplace' },
        { property: 'og:description', content: 'India\'s open marketplace for performance clipping. Fund campaigns, create clips, earn from verified views.' },
        { property: 'og:site_name', content: 'ClipIN' },
        { name: 'twitter:card', content: 'summary_large_image' },
        { name: 'twitter:title', content: 'ClipIN - Performance Clipping Marketplace' },
        { name: 'twitter:description', content: 'India\'s open marketplace for performance clipping. Fund campaigns, create clips, earn from verified views.' },
      ],
      link: [
        {
          rel: 'preconnect',
          href: 'https://fonts.googleapis.com',
        },
        {
          rel: 'preconnect',
          href: 'https://fonts.gstatic.com',
          crossorigin: '',
        },
        {
          rel: 'stylesheet',
          href: 'https://fonts.googleapis.com/css2?family=Plus+Jakarta+Sans:ital,wght@0,300..800;1,300..800&family=JetBrains+Mono:wght@400;500;600&display=swap',
        },
      ],
    },
  },

  typescript: {
    strict: true,
    typeCheck: false,
  },
})
