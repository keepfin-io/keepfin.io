<template>
  <Popover v-model:open="open">
    <PopoverTrigger as-child>
      <Button
        v-bind="$attrs"
        :id="id"
        type="button"
        variant="outline"
        role="combobox"
        :aria-expanded="open"
        class="w-full min-w-0 justify-between font-normal"
      >
        <span class="truncate">
          {{ modelValue || $t('accounts.selectPlatform') }}
        </span>
        <ChevronsUpDown class="size-4 shrink-0 opacity-50" />
      </Button>
    </PopoverTrigger>
    <PopoverContent
      :data-testid="$attrs['data-testid'] && `${$attrs['data-testid']}-options`"
      class="w-[var(--reka-popover-trigger-width)] p-0"
      align="start"
    >
      <Command>
        <CommandInput
          :data-testid="
            $attrs['data-testid'] && `${$attrs['data-testid']}-search`
          "
          :placeholder="$t('accounts.searchPlatform')"
        />
        <CommandList>
          <CommandEmpty
            :data-testid="
              $attrs['data-testid'] && `${$attrs['data-testid']}-empty`
            "
          >
            {{ $t('common.noResults') }}
          </CommandEmpty>
          <CommandGroup>
            <CommandItem
              :data-testid="
                $attrs['data-testid'] && `${$attrs['data-testid']}-option-none`
              "
              value="__none__"
              @select="select('')"
            >
              {{ $t('accounts.general') }}
            </CommandItem>
          </CommandGroup>
          <CommandGroup
            v-for="group in platforms"
            :key="group.label"
            :heading="group.label"
          >
            <CommandItem
              :data-testid="
                $attrs['data-testid'] &&
                `${$attrs['data-testid']}-option-${item.domain}`
              "
              v-for="item in group.options"
              :key="item.name"
              :value="item.name"
              @select="select(item.name)"
            >
              <Favicon :domain="item.domain" :size="16" />
              <span class="truncate">{{ item.name }}</span>
              <Check v-if="modelValue === item.name" class="ml-auto size-4" />
            </CommandItem>
          </CommandGroup>
        </CommandList>
      </Command>
    </PopoverContent>
  </Popover>
</template>
<script setup>
defineOptions({ inheritAttrs: false })
import { ref } from 'vue'
import { Button } from '@/components/ui/button'
import {
  Popover,
  PopoverTrigger,
  PopoverContent,
} from '@/components/ui/popover'
import {
  Command,
  CommandInput,
  CommandList,
  CommandEmpty,
  CommandGroup,
  CommandItem,
} from '@/components/ui/command'
import { Check, ChevronsUpDown } from 'lucide-vue-next'
import Favicon from './Favicon.vue'
defineProps({ modelValue: String, id: String })
const emit = defineEmits(['update:modelValue'])
const open = ref(false)
function select(value) {
  emit('update:modelValue', value)
  open.value = false
}
const platforms = [
  {
    label: 'Banks',
    options: [
      { name: 'Bank of America', domain: 'bankofamerica.com' },
      { name: 'Chase', domain: 'chase.com' },
      { name: 'Wells Fargo', domain: 'wellsfargo.com' },
      { name: 'Citibank', domain: 'citibank.com' },
      { name: 'HSBC', domain: 'hsbc.com' },
      { name: 'Barclays', domain: 'barclays.co.uk' },
      { name: 'Deutsche Bank', domain: 'db.com' },
      { name: 'ING', domain: 'ing.com' },
      { name: 'N26', domain: 'n26.com' },
      { name: 'Revolut', domain: 'revolut.com' },
    ],
  },
  {
    label: 'Brokerages',
    options: [
      { name: 'Interactive Brokers', domain: 'interactivebrokers.com' },
      { name: 'Fidelity', domain: 'fidelity.com' },
      { name: 'Charles Schwab', domain: 'schwab.com' },
      { name: 'TD Ameritrade', domain: 'tdameritrade.com' },
      { name: 'E*TRADE', domain: 'etrade.com' },
      { name: 'Robinhood', domain: 'robinhood.com' },
      { name: 'Webull', domain: 'webull.com' },
      { name: 'Trading 212', domain: 'trading212.com' },
      { name: 'eToro', domain: 'etoro.com' },
      { name: 'Degiro', domain: 'degiro.com' },
    ],
  },
  {
    label: 'Crypto',
    options: [
      { name: 'Coinbase', domain: 'coinbase.com' },
      { name: 'Binance', domain: 'binance.com' },
      { name: 'Kraken', domain: 'kraken.com' },
      { name: 'Crypto.com', domain: 'crypto.com' },
      { name: 'Gemini', domain: 'gemini.com' },
      { name: 'Bybit', domain: 'bybit.com' },
      { name: 'OKX', domain: 'okx.com' },
    ],
  },
  {
    label: 'Payment Services',
    options: [
      { name: 'PayPal', domain: 'paypal.com' },
      { name: 'Wise', domain: 'wise.com' },
      { name: 'Stripe', domain: 'stripe.com' },
      { name: 'Venmo', domain: 'venmo.com' },
      { name: 'Cash App', domain: 'cash.app' },
    ],
  },
  {
    label: 'Turkish Banks',
    options: [
      { name: 'Garanti BBVA', domain: 'garantibbva.com.tr' },
      { name: 'İş Bankası', domain: 'isbank.com.tr' },
      { name: 'Yapı Kredi', domain: 'yapikredi.com.tr' },
      { name: 'Akbank', domain: 'akbank.com' },
      { name: 'QNB Finansbank', domain: 'qnbfinansbank.com' },
      { name: 'Ziraat Bankası', domain: 'ziraatbank.com.tr' },
      { name: 'Midas', domain: 'getmidas.com' },
    ],
  },
]
</script>
