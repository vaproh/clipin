<script setup lang="ts">
import Navbar from '~/components/landing/Navbar.vue'
import Hero from '~/components/landing/Hero.vue'
import HowItWorks from '~/components/landing/HowItWorks.vue'
import ForClippers from '~/components/landing/ForClippers.vue'
import ForCampaignOwners from '~/components/landing/ForCampaignOwners.vue'
import WhyClipin from '~/components/landing/WhyClipin.vue'
import Faq from '~/components/landing/Faq.vue'
import FinalCta from '~/components/landing/FinalCta.vue'
import Footer from '~/components/landing/Footer.vue'

definePageMeta({
  layout: false,
})

useSeoMeta({
  title: 'ClipIN - Performance Clipping Marketplace',
  description: 'India\'s open marketplace for performance clipping. Fund campaigns, create clips, earn from verified views.',
  ogTitle: 'ClipIN - Performance Clipping Marketplace',
  ogDescription: 'India\'s open marketplace for performance clipping. Fund campaigns, create clips, earn from verified views.',
  ogImage: '/og-image.png',
  twitterCard: 'summary_large_image',
})
</script>

<template>
  <div class="min-h-screen bg-black text-white font-sans selection:bg-white selection:text-black">
    <Navbar />
    <main>
      <Hero />
      <HowItWorks />
      <ForClippers />
      <ForCampaignOwners />
      <WhyClipin />
      <Faq />
      <FinalCta />
    </main>
    <Footer />
  </div>
</template>
