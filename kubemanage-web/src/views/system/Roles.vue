<template>
  <el-row :gutter="16">
    <el-col :span="9">
      <el-card>
        <template #header>
          <div class="toolbar">
            <b>角色</b>
            <el-button v-permission="['/api/authority','POST']" type="primary" @click="editRole()">新增</el-button>
          </div>
        </template>
        <el-table :data="roles" highlight-current-row row-key="authorityId" @current-change="selectRole">
          <el-table-column prop="authorityName" label="角色名称" min-width="110" />
          <el-table-column prop="authorityId" label="角色ID" width="78" />
          <el-table-column label="定位" min-width="130">
            <template #default="{ row }">
              <el-tag :type="roleDefinition(row).type" size="small">{{ roleDefinition(row).label }}</el-tag>
              <div v-if="row.parentId" class="role-parent">继承：{{ parentRoleName(row) }}</div>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="130">
            <template #default="{ row }">
              <el-button v-permission="['/api/authority/:authorityId','PUT']" link type="primary" @click.stop="editRole(row)">编辑</el-button>
              <el-button v-permission="['/api/authority/:authorityId','DELETE']" link type="danger" @click.stop="removeRole(row)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
        <el-alert class="role-guide" type="info" :closable="false" show-icon>
          <template #title>内置角色定位</template>
          <div>111 管理员：平台超级管理员，菜单和 API 始终全部生效。</div>
          <div>222 普通用户：业务使用基线，默认拥有非系统管理功能。</div>
          <div>2221 普通用户子角色：不重复直授，默认继承普通用户（222），可按需追加权限。</div>
        </el-alert>
      </el-card>
    </el-col>

    <el-col :span="15">
      <el-card v-if="selected" v-loading="permissionsLoading">
        <template #header>
          <div>
            <b>{{ selected.authorityName }} - 权限配置</b>
            <div class="role-description">{{ selectedRoleDescription }}</div>
          </div>
        </template>
        <el-alert
          v-if="isAdminSelected"
          type="warning"
          :closable="false"
          show-icon
          title="管理员由系统强制拥有全部菜单和 API 权限；此处仅展示，不允许取消或保存。"
        />
        <el-alert
          v-else-if="selected.parentId"
          type="info"
          :closable="false"
          show-icon
          title="下方勾选项仅代表本角色直接授予的权限；父角色权限会自动继承，即使这里未勾选也仍然生效。"
        />
        <el-tabs>
          <el-tab-pane label="菜单权限">
            <p class="permission-note">菜单名称和路径与左侧导航路由一一对应，保存的是当前角色的直授菜单。</p>
            <el-tree
              ref="menuTree"
              :data="menuTreeData"
              node-key="id"
              show-checkbox
              default-expand-all
              :props="menuTreeProps"
            >
              <template #default="{ data }">
                <span class="menu-node">
                  <span>{{ data.name }}</span>
                  <code>{{ data.path }}</code>
                </span>
              </template>
            </el-tree>
            <el-button
              v-permission="['/api/menu/add_menu_authority','POST']"
              class="save"
              type="primary"
              :disabled="isAdminSelected"
              @click="saveMenus"
            >保存菜单权限</el-button>
          </el-tab-pane>
          <el-tab-pane label="API 权限">
            <p class="permission-note">API 勾选项表示当前角色的直授策略；继承策略不在此重复勾选。</p>
            <el-table
              ref="apiTable"
              :data="apis"
              row-key="apiKey"
              @selection-change="apiSelection = $event"
            >
              <el-table-column type="selection" width="50" :selectable="apiSelectable" />
              <el-table-column prop="apiGroup" label="分组" />
              <el-table-column prop="description" label="接口" />
              <el-table-column prop="method" label="方法" width="90" />
              <el-table-column prop="path" label="路径" min-width="220" />
            </el-table>
            <el-button
              v-permission="['/api/authority/updateCasbinByAuthority','POST']"
              class="save"
              type="primary"
              :disabled="isAdminSelected"
              @click="saveApis"
            >保存 API 权限</el-button>
          </el-tab-pane>
        </el-tabs>
      </el-card>
      <el-empty v-else description="请选择角色" />
    </el-col>
  </el-row>

  <el-dialog v-model="dialog" :title="form.oldId ? '编辑角色' : '新增角色'" width="430px">
    <el-form label-width="90px">
      <el-form-item label="角色ID"><el-input-number v-model="form.authorityId" :disabled="!!form.oldId" /></el-form-item>
      <el-form-item label="角色名称"><el-input v-model="form.authorityName" /></el-form-item>
      <el-form-item label="父角色">
        <el-select v-model="form.parentId" clearable>
          <el-option v-for="role in availableParentRoles" :key="role.authorityId" :label="role.authorityName" :value="role.authorityId" />
        </el-select>
      </el-form-item>
      <el-form-item label="默认路由"><el-input v-model="form.defaultRouter" /></el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="dialog = false">取消</el-button>
      <el-button type="primary" @click="saveRole">保存</el-button>
    </template>
  </el-dialog>
</template>

<script>
import {
  createAuthority,
  deleteAuthority,
  getApis,
  getAuthorities,
  getAuthorityPolicies,
  getBaseMenus,
  getDirectAuthorityMenus,
  saveAuthorityMenus,
  saveAuthorityPolicies,
  updateAuthority
} from '@/api/system'
import { ADMIN_AUTHORITY_ID } from '@/utils/auth'

const flattenIds = (nodes, out = []) => {
  nodes.forEach(node => {
    out.push(node.id)
    flattenIds(node.children || [], out)
  })
  return out
}

const builtInRoles = {
  111: { label: '超级管理员', type: 'danger', description: '平台超级管理员；菜单和 API 由系统强制全部生效，不依赖策略表是否逐条列出。' },
  222: { label: '业务基线', type: 'success', description: '普通用户基线角色；默认拥有工作流和 Kubernetes 业务功能，不包含系统管理。' },
  2221: { label: '继承角色', type: 'warning', description: '普通用户子角色；默认不重复直授权限，通过父角色 222 继承业务基线，并可叠加自身权限。' }
}

export default {
  name: 'SystemRoles',
  data() {
    return {
      roles: [],
      menus: [],
      apis: [],
      selected: null,
      apiSelection: [],
      dialog: false,
      form: {},
      permissionsLoading: false,
      menuTreeProps: { label: 'name', children: 'children', disabled: 'permissionDisabled' }
    }
  },
  computed: {
    isAdminSelected() {
      return Number(this.selected && this.selected.authorityId) === ADMIN_AUTHORITY_ID
    },
    selectedRoleDescription() {
      if (!this.selected) return ''
      const definition = this.roleDefinition(this.selected)
      if (builtInRoles[this.selected.authorityId]) return definition.description
      return this.selected.parentId
        ? `自定义子角色；继承 ${this.parentRoleName(this.selected)}，并叠加当前角色直接授予的权限。`
        : '自定义独立角色；仅使用当前角色直接授予的权限。'
    },
    menuTreeData() {
      const markDisabled = nodes => nodes.map(node => ({
        ...node,
        permissionDisabled: this.isAdminSelected,
        children: markDisabled(node.children || [])
      }))
      return markDisabled(this.menus)
    },
    availableParentRoles() {
      return this.roles.filter(role => role.authorityId !== this.form.authorityId)
    }
  },
  mounted() {
    this.init()
  },
  methods: {
    async init() {
      const [roleResponse, menuResponse, apiResponse] = await Promise.all([
        getAuthorities({ page: 1, pageSize: 100 }),
        getBaseMenus(),
        getApis({ page: 1, pageSize: 1000 })
      ])
      this.roles = (roleResponse.data && roleResponse.data.list) || []
      this.menus = (menuResponse.data && menuResponse.data.menus) || menuResponse.data || []
      const apiList = (apiResponse.data && apiResponse.data.list) || apiResponse.data || []
      this.apis = apiList.map(api => ({ ...api, apiKey: `${api.path},${api.method}` }))
    },
    roleDefinition(role) {
      return builtInRoles[role.authorityId] || {
        label: role.parentId ? '自定义子角色' : '自定义角色',
        type: 'info',
        description: ''
      }
    },
    parentRoleName(role) {
      const parent = this.roles.find(item => item.authorityId === role.parentId)
      return parent ? `${parent.authorityName}（${parent.authorityId}）` : `角色 ${role.parentId}`
    },
    apiSelectable() {
      return !this.isAdminSelected
    },
    async selectRole(row) {
      if (!row) return
      this.selected = row
      this.permissionsLoading = true
      try {
        const [menuResponse, policyResponse] = await Promise.all([
          getDirectAuthorityMenus(row.authorityId),
          getAuthorityPolicies(row.authorityId)
        ])
        const directMenus = (menuResponse.data && menuResponse.data.menus) || []
        const checkedMenuIds = this.isAdminSelected ? flattenIds(this.menus) : flattenIds(directMenus)
        const directRules = policyResponse.data || []
        this.$nextTick(() => {
          this.$refs.menuTree.setCheckedKeys(checkedMenuIds)
          this.$refs.apiTable.clearSelection()
          this.apis.forEach(api => {
            const checked = this.isAdminSelected || directRules.some(rule => (
              rule.path === api.path && String(rule.method).toUpperCase() === String(api.method).toUpperCase()
            ))
            if (checked) this.$refs.apiTable.toggleRowSelection(api, true)
          })
          if (this.isAdminSelected) this.apiSelection = [...this.apis]
        })
      } finally {
        this.permissionsLoading = false
      }
    },
    editRole(row) {
      this.form = row
        ? { ...row, oldId: row.authorityId }
        : { authorityId: null, authorityName: '', parentId: 0, defaultRouter: '/home' }
      this.dialog = true
    },
    async saveRole() {
      if (this.form.oldId) await updateAuthority(this.form.oldId, this.form)
      else await createAuthority(this.form)
      this.dialog = false
      await this.init()
    },
    async removeRole(row) {
      await this.$confirm(`确认删除角色 ${row.authorityName}？`, '警告', { type: 'warning' })
      await deleteAuthority(row.authorityId)
      if (this.selected && this.selected.authorityId === row.authorityId) this.selected = null
      await this.init()
    },
    async saveMenus() {
      if (this.isAdminSelected) return
      const ids = this.$refs.menuTree.getCheckedKeys().concat(this.$refs.menuTree.getHalfCheckedKeys())
      const menuMap = new Map()
      const walk = nodes => nodes.forEach(node => {
        menuMap.set(node.id, node)
        walk(node.children || [])
      })
      walk(this.menus)
      await saveAuthorityMenus(this.selected.authorityId, ids.map(id => menuMap.get(id)).filter(Boolean))
      this.$message.success('菜单权限已保存')
    },
    async saveApis() {
      if (this.isAdminSelected) return
      await saveAuthorityPolicies(this.selected.authorityId, this.apiSelection.map(api => ({ path: api.path, method: api.method })))
      this.$message.success('API 权限已保存')
    }
  }
}
</script>

<style scoped>
.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.role-parent,
.role-description,
.permission-note {
  color: var(--el-text-color-secondary);
  font-size: 12px;
}

.role-parent {
  margin-top: 4px;
}

.role-description {
  margin-top: 6px;
  font-weight: normal;
}

.role-guide {
  margin-top: 16px;
}

.role-guide div + div {
  margin-top: 4px;
}

.permission-note {
  margin: 0 0 12px;
}

.menu-node {
  display: inline-flex;
  align-items: center;
  gap: 10px;
}

.menu-node code {
  color: var(--el-text-color-secondary);
  font-size: 12px;
}

.save {
  margin-top: 16px;
}
</style>
