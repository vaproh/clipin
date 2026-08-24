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
</script>

<template>
  <div class="min-h-screen bg-[#010102] text-[#f7f8f8] font-sans selection:bg-[#5e6ad2]/30 selection:text-white">
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
