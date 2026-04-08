<template>
  <div class="icon-view">
    <ProTable
      ref="proTableRef"
      api="/asset/icon/list"
      statApi="/asset/icon/stat"
      batchDeleteApi="/asset/icon/batchDelete"
      rowKey="id"
      :columns="iconColumns"
      :searchItems="searchItems"
      :statLabels="statLabels"
      selection
      @data-changed="$emit('data-changed')"
      :searchKeys="['icon_hash']"
      :searchPlaceholder="t('asset.iconView.filters.iconHash')"
    >
      <template #toolbar-left>
        <el-dropdown @command="handleExport">
          <el-button type="success" size="default">
            {{ $t('common.export') || '导出' }}<el-icon class="el-icon--right"><ArrowDown /></el-icon>
          </el-button>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="selected" :disabled="selectedRows.length === 0">
                {{ t('asset.iconView.exportSelected', { count: selectedRows.length }) }}
              </el-dropdown-item>
              <el-dropdown-item divided command="all">{{ t('asset.iconView.exportAll') }}</el-dropdown-item>
              <el-dropdown-item command="csv">{{ t('asset.iconView.exportCsv') }}</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </template>

      <template #toolbar-right>
        <el-button type="danger" plain @click="handleClear">{{ $t('asset.clearData') || '清空数据' }}</el-button>
      </template>

      <template #iconImage="{ row }">
        <div class="icon-image-cell">
          <el-popover
            v-if="row.iconData && getIconDataUrl(row.iconData)"
            placement="right"
            trigger="hover"
            width="auto"
            :show-after="200"
          >
            <template #reference>
              <img
                :src="getIconDataUrl(row.iconData)"
                class="icon-image"
                :alt="row.iconHash"
                @error="handleImageError($event)"
              />
            </template>
            <img :src="getIconDataUrl(row.iconData)" style="max-width: 400px; max-height: 400px; object-fit: contain;" />
          </el-popover>
          <span v-else class="text-gray-400">-</span>
        </div>
      </template>

      <template #iconHash="{ row }">
        <span class="hash-text">{{ row.iconHash }}</span>
      </template>

      <template #assets="{ row }">
        <div v-if="row.assets && row.assets.length > 0">
          <el-tag v-for="asset in row.assets.slice(0, 3)" :key="asset" size="small" type="info" class="mr-1">{{ asset }}</el-tag>
          <span v-if="row.assets.length > 3" class="text-xs text-gray-500">+{{ row.assets.length - 3 }}</span>
        </div>
        <span v-else class="text-gray-400">-</span>
      </template>

      <template #screenshot="{ row }">
        <div class="screenshot-cell">
          <el-popover
            v-if="row.screenshot && getScreenshotUrl(row.screenshot)"
            placement="left"
            trigger="hover"
            width="auto"
            :show-after="200"
          >
            <template #reference>
              <img
                :src="getScreenshotUrl(row.screenshot)"
                class="screenshot-image"
                :alt="row.iconHash"
                @error="handleImageError($event)"
              />
            </template>
            <img :src="getScreenshotUrl(row.screenshot)" style="max-width: 600px; max-height: 600px; object-fit: contain;" />
          </el-popover>
          <span v-else class="text-gray-400">-</span>
        </div>
      </template>

      <template #operation="{ row }">
        <el-button type="primary" link size="small" @click="viewAssets(row)">{{ t('asset.iconView.viewAssets') }}</el-button>
        <el-button type="danger" link size="small" @click="handleDelete(row)">{{ $t('common.delete') || '删除' }}</el-button>
      </template>
    </ProTable>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ArrowDown } from '@element-plus/icons-vue'
import request from '@/api/request'
import ProTable from '@/components/common/ProTable.vue'
import { formatScreenshotUrl } from '@/utils/screenshot'

const { t } = useI18n()
const emit = defineEmits(['data-changed'])
const proTableRef = ref(null)

const selectedRows = computed(() => proTableRef.value?.selectedRows || [])

const statLabels = computed(() => ({
  total: t('asset.iconView.total'),
  newCount: t('asset.iconView.newCount')
}))

const iconColumns = computed(() => [
  { label: t('asset.iconView.columns.iconImage'), prop: 'iconData', slot: 'iconImage', width: 90 },
  { label: t('asset.iconView.columns.iconHash'), prop: 'iconHash', slot: 'iconHash', minWidth: 240 },
  { label: t('asset.iconView.columns.assets'), prop: 'assets', slot: 'assets', minWidth: 250 },
  { label: t('asset.iconView.columns.screenshot'), prop: 'screenshot', slot: 'screenshot', width: 90 },
  { label: t('asset.iconView.columns.createTime'), prop: 'createTime', width: 160 },
  { label: t('asset.iconView.columns.updateTime'), prop: 'updateTime', width: 160 },
  { label: t('asset.iconView.columns.operation'), slot: 'operation', width: 140, fixed: 'right' }
])

const searchItems = computed(() => [
  { label: t('asset.iconView.filters.iconHash'), prop: 'icon_hash', type: 'input' }
])

async function handleDelete(row) {
  try {
    await ElMessageBox.confirm(t('asset.iconView.confirmDelete'), t('common.tip'), { type: 'warning' })
    const res = await request.post('/asset/icon/delete', { id: row.id })
    if (res.code === 0) {
      ElMessage.success(t('common.deleteSuccess'))
      proTableRef.value?.loadData()
      emit('data-changed')
    }
  } catch (e) {
    // cancelled
  }
}

async function handleClear() {
  try {
    await ElMessageBox.confirm(t('asset.iconView.confirmClear'), t('common.warning'), { type: 'error' })
    const res = await request.post('/asset/icon/clear')
    if (res.code === 0) {
      ElMessage.success(t('asset.clearSuccess'))
      proTableRef.value?.loadData()
      emit('data-changed')
    }
  } catch (e) {
    // cancelled
  }
}

function viewAssets(row) {
  window.location.href = `/asset-management?tab=inventory&subTab=port&iconHash=${encodeURIComponent(row.iconHash)}`
}

async function handleExport(command) {
  let data = []

  if (command === 'selected') {
    if (selectedRows.value.length === 0) {
      ElMessage.warning(t('common.pleaseSelect'))
      return
    }
    data = selectedRows.value
  } else {
    ElMessage.info(t('asset.gettingAllData'))
    try {
      const res = await request.post('/asset/icon/list', { ...proTableRef.value?.searchForm, page: 1, pageSize: 10000 })
      if (res.code === 0) {
        data = res.list || []
      } else {
        ElMessage.error(t('asset.getDataFailed'))
        return
      }
    } catch (e) {
      ElMessage.error(t('asset.getDataFailed'))
      return
    }
  }

  if (data.length === 0) {
    ElMessage.warning(t('asset.noDataToExport'))
    return
  }

  if (command === 'csv') {
    const headers = ['IconHash', 'IconHashFile', 'Assets', 'CreateTime']
    const csvRows = [headers.join(',')]
    for (const row of data) {
      csvRows.push([
        escapeCsvField(row.iconHash || ''),
        escapeCsvField(row.iconHashFile || ''),
        escapeCsvField((row.assets || []).join(';')),
        escapeCsvField(row.createTime || '')
      ].join(','))
    }
    downloadBlob(`icons_${new Date().toISOString().slice(0, 10)}.csv`, new Blob(['\uFEFF' + csvRows.join('\n')], { type: 'text/csv;charset=utf-8' }))
    ElMessage.success(t('asset.exportSuccess', { count: data.length }))
    return
  }

  const lines = data.map(row => row.iconHash).filter(Boolean)
  if (lines.length === 0) {
    ElMessage.warning(t('asset.noDataToExport'))
    return
  }
  downloadBlob(command === 'selected' ? 'icons_selected.txt' : 'icons_all.txt', new Blob([lines.join('\n')], { type: 'text/plain;charset=utf-8' }))
  ElMessage.success(t('asset.exportSuccess', { count: lines.length }))
}

function downloadBlob(filename, blob) {
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = filename
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
  URL.revokeObjectURL(url)
}

function escapeCsvField(field) {
  if (field == null) return ''
  const str = String(field)
  if (str.includes(',') || str.includes('"') || str.includes('\n') || str.includes('\r')) {
    return '"' + str.replace(/"/g, '""') + '"'
  }
  return str
}

function getIconDataUrl(iconData) {
  if (!iconData || iconData.length === 0) return ''
  if (typeof iconData === 'string' && iconData.startsWith('data:')) return iconData
  const base64Str = typeof iconData === 'string' ? iconData : ''
  if (!base64Str) return ''
  try {
    const binaryStr = atob(base64Str)
    if (binaryStr.length < 4) return ''
    let start = 0
    while (start < binaryStr.length && (binaryStr[start] === ' ' || binaryStr[start] === '\t' || binaryStr[start] === '\n' || binaryStr[start] === '\r')) { start++ }
    if (binaryStr[start] === '<') {
      const header = binaryStr.substring(start, start + 100).toLowerCase()
      if (header.startsWith('<!doctype') || header.startsWith('<html') || header.startsWith('<?xml')) return ''
      if (header.startsWith('<svg')) return `data:image/svg+xml;base64,${base64Str}`
      return ''
    }
    const bytes = new Uint8Array(binaryStr.length)
    for (let i = 0; i < binaryStr.length; i++) { bytes[i] = binaryStr.charCodeAt(i) }
    if (bytes[0] === 0x89 && bytes[1] === 0x50 && bytes[2] === 0x4E && bytes[3] === 0x47) return `data:image/png;base64,${base64Str}`
    if (bytes[0] === 0xFF && bytes[1] === 0xD8 && bytes[2] === 0xFF) return `data:image/jpeg;base64,${base64Str}`
    if (bytes[0] === 0x47 && bytes[1] === 0x49 && bytes[2] === 0x46 && bytes[3] === 0x38) return `data:image/gif;base64,${base64Str}`
    if (bytes[0] === 0x00 && bytes[1] === 0x00 && (bytes[2] === 0x01 || bytes[2] === 0x02) && bytes[3] === 0x00) return `data:image/*;base64,${base64Str}`
    if (bytes[0] === 0x42 && bytes[1] === 0x4D) return `data:image/bmp;base64,${base64Str}`
    if (bytes.length >= 12 && bytes[0] === 0x52 && bytes[1] === 0x49 && bytes[2] === 0x46 && bytes[3] === 0x46 &&
        bytes[8] === 0x57 && bytes[9] === 0x45 && bytes[10] === 0x42 && bytes[11] === 0x50) return `data:image/webp;base64,${base64Str}`
    return ''
  } catch (e) { return '' }
}

function getScreenshotUrl(screenshot) {
  return formatScreenshotUrl(screenshot)
}

function handleImageError(event) {
  event.target.style.display = 'none'
}

defineExpose({
  refresh: () => proTableRef.value?.loadData()
})
</script>

<style scoped>
.icon-view {
  height: 100%;
}
.icon-image-cell,
.screenshot-cell {
  display: flex;
  align-items: center;
  justify-content: center;
}
.hash-text {
  font-family: monospace;
}
.icon-image,
.screenshot-image {
  width: 40px;
  height: 40px;
  object-fit: cover;
  border-radius: 4px;
  border: 1px solid var(--el-border-color-lighter);
  background: var(--el-fill-color-lighter);
}
.mr-1 {
  margin-right: 4px;
}
</style>
