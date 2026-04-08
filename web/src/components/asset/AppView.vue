<template>
  <div class="app-view">
    <ProTable
      ref="proTableRef"
      api="/asset/app/list"
      statApi="/asset/app/stat"
      batchDeleteApi="/asset/app/batchDelete"
      rowKey="id"
      :columns="appColumns"
      :searchItems="searchItems"
      :statLabels="statLabels"
      selection
      :searchPlaceholder="t('asset.appView.searchPlaceholder')"
      @data-changed="$emit('data-changed')"
      :searchKeys="['app']"
    >
      <template #toolbar-left>
        <el-dropdown @command="handleExport">
          <el-button type="success" size="default">
            {{ $t('common.export') || '导出' }}<el-icon class="el-icon--right"><ArrowDown /></el-icon>
          </el-button>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="selected" :disabled="selectedRows.length === 0">
                {{ t('asset.appView.exportSelected', { count: selectedRows.length }) }}
              </el-dropdown-item>
              <el-dropdown-item divided command="all">{{ t('asset.appView.exportAll') }}</el-dropdown-item>
              <el-dropdown-item command="csv">{{ t('asset.appView.exportCsv') }}</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </template>

      <template #toolbar-right>
        <el-button type="danger" plain @click="handleClear">{{ $t('asset.clearData') || '清空数据' }}</el-button>
      </template>

      <template #app="{ row }">
        <div class="app-cell font-bold">
          <span>{{ row.app }}</span>
        </div>
      </template>

      <template #assets="{ row }">
        <div v-if="row.assets && row.assets.length > 0">
          <el-tag v-for="ast in row.assets.slice(0, 3)" :key="ast" size="small" type="info" class="mr-1">{{ ast }}</el-tag>
          <span v-if="row.assets.length > 3" class="text-xs text-gray-500">+{{ row.assets.length - 3 }}</span>
        </div>
        <span v-else class="text-gray-400">-</span>
      </template>

      <template #org="{ row }">
        {{ row.orgName || $t('common.defaultOrganization') || '默认组织' }}
      </template>

      <template #operation="{ row }">
        <el-button type="primary" link size="small" @click="viewAssets(row)">{{ t('asset.appView.viewAssets') }}</el-button>
        <el-button type="danger" link size="small" @click="handleDelete(row)">{{ $t('common.delete') || '删除' }}</el-button>
      </template>
    </ProTable>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ArrowDown } from '@element-plus/icons-vue'
import request from '@/api/request'
import ProTable from '@/components/common/ProTable.vue'

const { t } = useI18n()
const emit = defineEmits(['data-changed'])
const proTableRef = ref(null)
const organizations = ref([])

const selectedRows = computed(() => proTableRef.value?.selectedRows || [])

const statLabels = computed(() => ({
  total: t('asset.appView.total'),
  newCount: t('asset.appView.newCount')
}))

const appColumns = computed(() => [
  { label: t('asset.appView.columns.app'), prop: 'app', slot: 'app', minWidth: 200 },
  { label: t('asset.appView.columns.assets'), prop: 'assets', slot: 'assets', minWidth: 250 },
  { label: t('asset.appView.columns.organization'), prop: 'orgName', slot: 'org', width: 120 },
  { label: t('asset.appView.columns.createTime'), prop: 'createTime', width: 160 },
  { label: t('asset.appView.columns.operation'), slot: 'operation', width: 140, fixed: 'right' }
])

const searchItems = computed(() => [
  { label: '关联资产', prop: 'assets', type: 'input' },
  {
    label: t('asset.appView.filters.organization'), prop: 'orgId', type: 'select',
    options: [{ label: t('asset.appView.filters.allOrganizations'), value: '' }, ...organizations.value.map(org => ({ label: org.name, value: org.id }))]
  }
])

async function loadOrganizations() {
  try {
    const res = await request.post('/organization/list', { page: 1, pageSize: 100 })
    if (res.code === 0) organizations.value = res.list || []
  } catch (e) {
    console.error(e)
  }
}

async function handleDelete(row) {
  try {
    await ElMessageBox.confirm(t('asset.appView.confirmDelete'), t('common.tip'), { type: 'warning' })
    const res = await request.post('/asset/app/delete', { id: row.id })
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
    await ElMessageBox.confirm(t('asset.appView.confirmClear'), t('common.warning'), { type: 'error' })
    const res = await request.post('/asset/app/clear')
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
  window.location.href = `/asset-management?tab=inventory&app=${encodeURIComponent(row.app)}`
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
      const res = await request.post('/asset/app/list', { ...proTableRef.value?.searchForm, page: 1, pageSize: 10000 })
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
    const headers = ['App', 'Category', 'Assets', 'Organization', 'CreateTime']
    const csvRows = [headers.join(',')]
    for (const row of data) {
      csvRows.push([
        escapeCsvField(row.app || ''),
        escapeCsvField(row.category || ''),
        escapeCsvField((row.assets || []).join(';')),
        escapeCsvField(row.orgName || ''),
        escapeCsvField(row.createTime || '')
      ].join(','))
    }
    downloadBlob(`apps_${new Date().toISOString().slice(0, 10)}.csv`, new Blob(['\uFEFF' + csvRows.join('\n')], { type: 'text/csv;charset=utf-8' }))
    ElMessage.success(t('asset.exportSuccess', { count: data.length }))
    return
  }

  const lines = data.map(row => row.app).filter(Boolean)
  if (lines.length === 0) {
    ElMessage.warning(t('asset.noDataToExport'))
    return
  }
  downloadBlob(command === 'selected' ? 'apps_selected.txt' : 'apps_all.txt', new Blob([lines.join('\n')], { type: 'text/plain;charset=utf-8' }))
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

onMounted(() => {
  loadOrganizations()
})

defineExpose({
  refresh: () => proTableRef.value?.loadData()
})
</script>

<style scoped>
.app-view {
  height: 100%;
}
.app-cell {
  display: flex;
  align-items: center;
}
.mr-1 {
  margin-right: 4px;
}
</style>
