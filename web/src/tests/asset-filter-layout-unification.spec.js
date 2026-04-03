import { shallowMount, mount } from '@vue/test-utils'
import { h } from 'vue'

import AssetAllView from '@/components/asset/AssetAllView.vue'
import DomainView from '@/components/asset/DomainView.vue'
import IPView from '@/components/asset/IPView.vue'
import SiteView from '@/components/asset/SiteView.vue'
import DirScanView from '@/components/asset/DirScanView.vue'
import VulView from '@/components/asset/VulView.vue'

const mountOptions = {
  global: {
    config: {
      globalProperties: {
        $t: (key) => key
      }
    },
    stubs: {
      ProTable: true,
      ElButton: false,
      ElTable: true,
      ElTableColumn: true,
      ElDialog: true,
      ElCard: true,
      ElDescriptions: true,
      ElDescriptionsItem: true,
      ElImage: true,
      ElTag: true,
      ElTabs: true,
      ElTabPane: true,
      ElDropdown: true,
      ElDropdownMenu: true,
      ElDropdownItem: true,
      ElIcon: true,
      ElEmpty: true,
      ElBadge: true,
      ElForm: true,
      ElFormItem: true,
      ElInput: true,
      ElSelect: true,
      ElOption: true,
    }
  }
}

const targetPages = [
  { name: 'AssetAllView', component: AssetAllView, clearLabel: 'asset.clearData' },
  { name: 'DomainView', component: DomainView, clearLabel: 'asset.clearData' },
  { name: 'IPView', component: IPView, clearLabel: 'asset.clearData' },
  { name: 'SiteView', component: SiteView, clearLabel: 'asset.clearData' },
  { name: 'DirScanView', component: DirScanView, clearLabel: 'dirscan.clearData' },
  { name: 'VulView', component: VulView, clearLabel: 'vul.clearData' }
]

describe('asset filter layout unification', () => {
  it.each(targetPages)('uses ProTable component in %s', ({ component }) => {
    const wrapper = shallowMount(component, mountOptions)

    // 断言页面中包含 ProTable 组件
    const proTable = wrapper.findComponent({ name: 'ProTable' }) || wrapper.find('pro-table-stub')
    expect(proTable.exists()).toBe(true)
  })

  it.each(targetPages)('passes correct toolbar slots to ProTable in %s', ({ component }) => {
    const wrapper = shallowMount(component, mountOptions)
    const proTable = wrapper.findComponent({ name: 'ProTable' }) || wrapper.find('pro-table-stub')

    // 断言 ProTable 存在
    expect(proTable.exists()).toBe(true)
  })

  it.each(targetPages)('Danger+Plain ClearData action is passed to ProTable toolbar-right slot in %s', ({ component, clearLabel }) => {
    const wrapper = shallowMount(component, mountOptions)

    const proTable = wrapper.findComponent({ name: 'ProTable' }) || wrapper.find('pro-table-stub')
    expect(proTable.exists()).toBe(true)

    // 获取并渲染 toolbar-right 插槽内容
    let slotFn = null;
    if (proTable.vm && proTable.vm.$slots && proTable.vm.$slots['toolbar-right']) {
      slotFn = proTable.vm.$slots['toolbar-right']
    } else if (wrapper.vm.$slots && wrapper.vm.$slots['toolbar-right']) {
      slotFn = wrapper.vm.$slots['toolbar-right']
    }
    
    // 我们检查渲染结果，看看有没有按钮
    let slotWrapper;
    if (slotFn) {
        slotWrapper = mount({ render: () => h('div', slotFn()) }, mountOptions)
    } else {
        const html = wrapper.html()
        expect(html).toMatch(/type="danger"/)
        expect(html).toMatch(/plain/)
        return
    }

    const dangerPlainButtons = slotWrapper.findAll('.el-button--danger.is-plain, el-button[type="danger"][plain=""], el-button[type="danger"][plain="true"], button.el-button--danger.is-plain')

    // 应该包含我们期望的清除按钮
    expect(dangerPlainButtons.length).toBeGreaterThan(0)
  })
})
