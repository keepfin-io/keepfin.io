<template>
  <span class="flex w-full min-w-0 items-center gap-3">
    <AssetIcon :asset="asset" :compact="compact" />
    <span class="min-w-0 flex-1 space-y-1 text-left">
      <span class="flex flex-wrap items-center gap-x-2 gap-y-1">
        <span class="text-sm font-semibold">{{ asset.Code }}</span>
        <span class="text-[11px] font-normal text-muted-foreground">{{ [asset.Exchange, asset.Currency].filter(Boolean).join(' · ') }}</span>
      </span>
      <span v-if="asset.Name" class="block truncate text-xs font-normal text-muted-foreground" :title="asset.Name">{{ asset.Name }}</span>
      <span v-if="detailed" class="block text-[11px] font-normal text-muted-foreground">{{ $t(`holdings.assetTypes.${assetCategory(asset)}`) }}<template v-if="asset.Country"> · {{ asset.Country }}</template></span>
    </span>
  </span>
</template>
<script setup>
import AssetIcon from './AssetIcon.vue'
import { assetCategory } from '@/lib/asset-search'
defineProps({ asset: { type: Object, required: true }, detailed: Boolean, compact: Boolean })
</script>
