import { onBeforeUnmount, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { formatMoney } from '@/lib/portfolio-format'

export function useWealthPage() {
  const { t, locale } = useI18n()
  const busy = ref(false), error = ref(''), notice = ref('')
  let active = true
  onBeforeUnmount(() => { active = false })
  async function run(action, saved = false) {
    if (busy.value) return
    busy.value = true; error.value = ''; notice.value = ''
    try { await action(); if (saved && active) notice.value = t('wealth.saved') }
    catch (e) { if (active) error.value = e.message }
    finally { if (active) busy.value = false }
  }
  const money = (value, currency) => formatMoney(value, currency, locale.value)
  const today = new Date().toISOString().slice(0, 10)
  return { t, locale, busy, error, notice, run, money, today }
}
