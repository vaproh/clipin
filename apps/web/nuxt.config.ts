// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  compatibilityDate: '2024-11-01',
  devtools: { enabled: true },

  experimental: {
    appManifest: false,
  },

  modules: [
    [
      '@clerk/nuxt',
      {
        publishableKey: process.env.CLERK_PUBLISHABLE_KEY || process.env.NUXT_PUBLIC_CLERK_PUBLISHABLE_KEY || 'pk_test_cGlja2VkLWdhcmZpc2gtMTc0NS5jbGVyay5hY2NvdW50cy5kZXYk',
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
      },
    ],
    '@nuxtjs/tailwindcss',
    '@vueuse/nuxt',
  ],

  runtimeConfig: {
    clerkSecretKey: process.env.CLERK_SECRET_KEY || process.env.NUXT_CLERK_SECRET_KEY || 'sk_test_itnjoan2YJ9vAB0qqkBZrFSbizhGSFN5zfNKX3mqpr',
    public: {
      clerkPublishableKey: process.env.CLERK_PUBLISHABLE_KEY || process.env.NUXT_PUBLIC_CLERK_PUBLISHABLE_KEY || 'pk_test_cGlja2VkLWdhcmZpc2gtMTc0NS5jbGVyay5hY2NvdW50cy5kZXYk',
      apiBase: process.env.NUXT_PUBLIC_API_BASE || 'http://localhost:8080',
      verifierBase: process.env.NUXT_PUBLIC_VERIFIER_BASE || 'http://localhost:8081',
    },
  },

  css: ['~/assets/css/main.css'],

  app: {
    head: {
      title: 'ClipIN — India\'s Performance Clipping Marketplace',
      meta: [
        { charset: 'utf-8' },
        { name: 'viewport', content: 'width=device-width, initial-scale=1' },
        {
          name: 'description',
          content: 'ClipIN connects content owners with clippers to turn long-form content into distributed short-form reach. Clip. Post. Get Paid.',
        },
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
