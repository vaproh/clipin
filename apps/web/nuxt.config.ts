// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  compatibilityDate: '2024-11-01',
  devtools: { enabled: true },

  modules: [
    '@clerk/nuxt',
    '@nuxtjs/tailwindcss',
    '@vueuse/nuxt',
  ],

  runtimeConfig: {
    clerkSecretKey: process.env.NUXT_CLERK_SECRET_KEY || process.env.CLERK_SECRET_KEY,
    public: {
      clerkPublishableKey: process.env.NUXT_PUBLIC_CLERK_PUBLISHABLE_KEY,
      apiBase: process.env.NUXT_PUBLIC_API_BASE || 'http://localhost:8080',
      verifierBase: process.env.NUXT_PUBLIC_VERIFIER_BASE || 'http://localhost:8081',
    },
  },

  clerk: {
    appearance: {
      variables: {
        colorPrimary: '#5e6ad2',
        colorBackground: '#0f1011',
        colorInputBackground: '#141516',
        colorInputText: '#f7f8f8',
        colorText: '#f7f8f8',
        colorTextSecondary: '#8a8f98',
      },
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
