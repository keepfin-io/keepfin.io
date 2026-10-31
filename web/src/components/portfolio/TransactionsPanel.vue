<template>
  <div class="space-y-4">
    <div class="flex flex-wrap gap-2">
      <Button data-testid="activities-add" :disabled="!account.is_active" @click="edit(null)">{{ $t('portfolio.addActivity') }}</Button>
      <Button variant="outline" :disabled="!account.is_active" @click="transferOpen = true">{{ $t('portfolio.transfer') }}</Button>
      <Button variant="outline" @click="showImport = !showImport">{{ $t('portfolio.importCSV') }}</Button>
      <Button variant="outline" :disabled="busy" @click="exportCSV">{{ $t('portfolio.exportCSV') }}</Button>
    </div>
    <Alert v-if="error" variant="error" :description="error" />
    <p v-if="notice" role="status" class="text-sm">{{ notice }}</p>
    <Card v-if="showImport"><CardHeader><CardTitle>{{ $t('portfolio.importCSV') }}</CardTitle><CardDescription>{{ $t('portfolio.csvHint') }}</CardDescription></CardHeader><CardContent class="space-y-3">
      <Label for="csv-file">{{ $t('portfolio.csvFile') }}</Label><Input id="csv-file" type="file" accept=".csv,text/csv" :disabled="busy" @change="readFile" />
      <Label for="csv-text">{{ $t('portfolio.csvText') }}</Label><Textarea id="csv-text" data-testid="csv-text" v-model="csvText" rows="5" :disabled="busy" @input="preview = null" />
      <Button variant="outline" data-testid="csv-preview" :disabled="busy || !csvText" @click="importCSV(false)">{{ $t('portfolio.preview') }}</Button>
      <div v-if="preview" data-testid="csv-result" class="space-y-2"><p>{{ $t('portfolio.csvCounts', { rows: preview.rows.length, duplicates: preview.duplicates }) }}</p><p v-if="preview.error" class="text-destructive">{{ preview.error }}</p><div class="max-h-64 overflow-auto"><Table><TableHeader><TableRow><TableHead>{{ $t('portfolio.line') }}</TableHead><TableHead>{{ $t('portfolio.type') }}</TableHead><TableHead>{{ $t('portfolio.symbol') }}</TableHead><TableHead>{{ $t('portfolio.status') }}</TableHead></TableRow></TableHeader><TableBody><TableRow v-for="row in preview.rows" :key="row.line"><TableCell>{{ row.line }}</TableCell><TableCell>{{ row.activity.type }}</TableCell><TableCell>{{ row.activity.symbol }}</TableCell><TableCell>{{ row.error || $t(`portfolio.${row.status}`) }}</TableCell></TableRow></TableBody></Table></div><Button data-testid="csv-commit" :disabled="busy || !preview.valid || previewCSV !== csvText" @click="importCSV(true)">{{ $t('portfolio.applyImport') }}</Button></div>
    </CardContent></Card>
    <p v-if="!activities.length" class="py-5 text-center text-muted-foreground">{{ $t('accountDetail.noActivitiesRecorded') }}</p>
    <Card v-else class="overflow-hidden"><CardContent class="overflow-x-auto p-0"><Table data-testid="activities-table"><TableHeader><TableRow><TableHead v-for="key in ['date', 'type', 'symbol', 'quantity', 'unitPrice', 'fees', 'currency', 'actions']" :key="key">{{ $t(`portfolio.${key}`) }}</TableHead></TableRow></TableHeader><TableBody><TableRow v-for="a in activities" :key="a.id" :data-testid="`activity-row-${a.id}`"><TableCell>{{ a.date.slice(0, 10) }}</TableCell><TableCell>{{ $t(`portfolio.types.${a.type}`) }}<Badge v-if="!a.ledger" variant="outline">{{ $t('portfolio.archived') }}</Badge></TableCell><TableCell>{{ a.symbol }}</TableCell><TableCell>{{ a.quantity }}</TableCell><TableCell>{{ a.unit_price }}</TableCell><TableCell>{{ a.fee }}</TableCell><TableCell>{{ a.currency }}</TableCell><TableCell><div class="flex gap-2"><Button size="sm" variant="outline" :disabled="!a.ledger || !!a.transfer_id || busy" @click="edit(a)">{{ $t('common.edit') }}</Button><Button size="sm" variant="outline" :disabled="!a.ledger || busy" @click="deleting = a">{{ $t('common.delete') }}</Button></div></TableCell></TableRow></TableBody></Table></CardContent></Card>
    <ActivityDialog :open="dialogOpen" :account="account" :activity="selected" @close="dialogOpen = false" @saved="$emit('changed')" />
    <TransferDialog :open="transferOpen" :account="account" @close="transferOpen = false" @saved="$emit('changed')" />
    <ConfirmModal :visible="!!deleting" variant="danger" :title="$t('common.delete')" :message="$t(deleting?.transfer_id ? 'portfolio.deleteTransfer' : 'portfolio.deleteActivity')" :confirm-text="$t('common.delete')" :cancel-text="$t('common.cancel')" :loading="busy" @cancel="deleting = null" @confirm="remove" />
  </div>
</template>
<script setup>
import { ref } from 'vue'
import { activityAPI } from '@/services/activityApi'
import { portfolioAPI } from '@/services/portfolioApi'
import { useI18n } from 'vue-i18n'
import ActivityDialog from './ActivityDialog.vue'
import TransferDialog from './TransferDialog.vue'
import { Button, Card, CardHeader, CardTitle, CardDescription, CardContent, Input, Textarea, Label, Badge, ConfirmModal, FeedbackAlert as Alert, Table, TableHeader, TableHead, TableRow, TableBody, TableCell } from '@/components/ui'
const props = defineProps({ account: { type: Object, required: true }, activities: { type: Array, default: () => [] } })
const emit = defineEmits(['changed']); const { t } = useI18n()
const dialogOpen = ref(false), selected = ref(null), transferOpen = ref(false), deleting = ref(null), busy = ref(false), error = ref(''), notice = ref('')
const showImport = ref(false), csvText = ref(''), preview = ref(null), previewCSV = ref('')
function edit(a) { selected.value = a; dialogOpen.value = true }
async function remove() { if (busy.value || !deleting.value) return; busy.value = true; error.value = ''; try { await activityAPI.deleteActivity(props.account.id, deleting.value.id); deleting.value = null; emit('changed') } catch (e) { error.value = e.message; deleting.value = null } finally { busy.value = false } }
async function exportCSV() { busy.value = true; error.value = ''; try { await portfolioAPI.exportCSV(props.account.id) } catch (e) { error.value = e.message } finally { busy.value = false } }
async function readFile(event) { preview.value = null; error.value = ''; const file = event.target.files?.[0]; if (!file) return; if (file.size > 2 * 1024 * 1024) { error.value = t('portfolio.csvSize'); return } try { csvText.value = await file.text() } catch { error.value = t('portfolio.csvReadError') } }
async function importCSV(commit) {
  if (busy.value) return
  busy.value = true; error.value = ''; notice.value = ''
  try { const data = await portfolioAPI.importCSV(props.account.id, csvText.value, commit); preview.value = data.result; previewCSV.value = csvText.value; if (commit && data.result.valid) { notice.value = t('portfolio.imported', { count: data.result.imported }); preview.value = null; csvText.value = ''; showImport.value = false; emit('changed') } }
  catch (e) { error.value = e.message } finally { busy.value = false }
}
</script>
