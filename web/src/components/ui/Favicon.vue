<template>
  <img
    :src="faviconUrl"
    :alt="domain"
    :width="size"
    :height="size"
    class="inline-block shrink-0 rounded-sm transition-opacity"
    @error="handleError"
    :style="{ opacity: loaded ? 1 : 0 }"
    @load="loaded = true"
  />
</template>

<script>
export default {
  name: 'Favicon',
  props: {
    domain: {
      type: String,
      required: true,
    },
    size: {
      type: Number,
      default: 16,
    },
    provider: {
      type: String,
      default: 'google',
      validator: (v) => ['google', 'duckduckgo'].indexOf(v) !== -1,
    },
  },
  data() {
    return {
      loaded: false,
      hasError: false,
    }
  },
  computed: {
    faviconUrl() {
      if (this.hasError) {
        // Fallback to a generic icon
        return `data:image/svg+xml,<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="%2394a3b8"><rect width="24" height="24" rx="4"/></svg>`
      }

      const cleanDomain = this.domain
        .replace(/^https?:\/\//, '')
        .replace(/\/.*$/, '')

      if (this.provider === 'google') {
        return `https://www.google.com/s2/favicons?domain=${cleanDomain}&sz=${
          this.size * 2
        }`
      }
      return `https://icons.duckduckgo.com/ip3/${cleanDomain}.ico`
    },
  },
  methods: {
    handleError() {
      this.hasError = true
      this.loaded = true
    },
  },
}
</script>
