<template>
  <footer data-testid="app-footer" class="app-footer">
    <span class="footer-brand">keepfin.io</span>
    <p data-testid="footer-financial-notice" class="footer-notice">{{ $t('footer.summary') }}</p>
    <nav class="footer-links" :aria-label="$t('footer.navigation')">
        <Dialog v-model:open="legalOpen">
          <DialogTrigger as-child>
            <button data-testid="footer-legal-trigger" type="button" class="footer-link">{{ $t('footer.legalNotice') }}</button>
          </DialogTrigger>
          <DialogContent data-testid="footer-legal-dialog" class="rounded-2xl sm:max-w-xl">
            <DialogHeader class="pr-6 text-left">
              <DialogTitle>{{ $t('footer.legalNotice') }}</DialogTitle>
              <DialogDescription>{{ $t('footer.legalDescription') }}</DialogDescription>
            </DialogHeader>
            <div class="legal-sections">
              <section v-for="section in legalSections" :key="section">
                <h3>{{ $t(`footer.sections.${section}.title`) }}</h3>
                <p>{{ $t(`footer.sections.${section}.body`) }}</p>
              </section>
            </div>
            <DialogClose as-child>
              <Button data-testid="footer-legal-close" variant="outline" class="min-h-11 justify-self-end rounded-xl">{{ $t('common.dismiss') }}</Button>
            </DialogClose>
          </DialogContent>
        </Dialog>
        <a data-testid="footer-source-link" class="footer-link" href="https://github.com/keepfin-io/keepfin.io" target="_blank" rel="noopener noreferrer" :aria-label="$t('footer.sourceLinkLabel')">
          <Github class="size-3.5" aria-hidden="true" />
          <span>{{ $t('footer.openSource') }}</span>
        </a>
    </nav>
  </footer>
</template>

<script setup>
import { ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { Github } from 'lucide-vue-next'
import { Button, Dialog, DialogClose, DialogContent, DialogDescription, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui'

const route = useRoute()
const legalOpen = ref(false)
const legalSections = ['purpose', 'data', 'risk', 'decisions', 'rights']
watch(() => route.fullPath, () => { legalOpen.value = false })
</script>

<style scoped>
.app-footer { display: grid; grid-template-columns: minmax(0, 1fr) auto; grid-template-areas: 'brand links' 'notice notice'; align-items: center; column-gap: 16px; flex-shrink: 0; width: 100%; padding: 0 16px max(12px, env(safe-area-inset-bottom, 0px)); color: var(--muted-foreground); font-size: 11px; }
.footer-brand { grid-area: brand; font-size: inherit; font-weight: 400; }
.footer-links { grid-area: links; display: flex; align-items: center; flex-wrap: wrap; justify-content: flex-end; gap: 0 16px; }
.footer-link { display: inline-flex; align-items: center; justify-content: center; min-height: 44px; gap: 6px; border-radius: 4px; color: inherit; background: transparent; font-size: inherit; font-weight: 400; text-decoration: none; text-underline-offset: 4px; cursor: pointer; transition: color 180ms ease; }
.footer-link:hover { color: var(--foreground); text-decoration: underline; }
.footer-link:focus-visible { outline: 2px solid var(--ring); outline-offset: 3px; }
.footer-notice { grid-area: notice; min-width: 0; font-size: inherit; line-height: 1.65; overflow-wrap: anywhere; text-wrap: pretty; }
.legal-sections { display: grid; gap: 18px; }
.legal-sections h3 { margin-bottom: 5px; font-size: 13px; font-weight: 500; }
.legal-sections p { color: var(--muted-foreground); font-size: 12px; line-height: 1.8; overflow-wrap: anywhere; }
@media (min-width: 640px) { .app-footer { padding-inline: 20px; } }
@media (min-width: 768px) {
  .app-footer { grid-template-columns: auto minmax(0, 1fr) auto; grid-template-areas: 'brand notice links'; column-gap: 24px; padding-top: 4px; padding-bottom: max(4px, env(safe-area-inset-bottom, 0px)); }
  .footer-notice { text-align: center; }
}
@media (min-width: 1024px) { .app-footer { padding-inline: 32px; } }
@media (prefers-reduced-motion: reduce) { .footer-link { transition: none; } }
</style>
