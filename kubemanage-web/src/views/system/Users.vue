<template>
  <el-card>
    <template #header><div class="toolbar"><b>用户管理</b><el-button v-permission="['/api/user','POST']" type="primary" @click="openCreate">新增用户</el-button></div></template>
    <div class="toolbar search"><el-input v-model="query.keyword" placeholder="用户名/昵称" clearable @keyup.enter="load"/><el-button @click="load">查询</el-button></div>
    <el-table :data="rows" v-loading="loading" border>
      <el-table-column prop="id" label="ID" width="80"/><el-table-column prop="userName" label="用户名"/><el-table-column prop="nickName" label="昵称"/>
      <el-table-column label="角色"><template #default="{row}">{{ roleName(row.authorityId) }}</template></el-table-column>
      <el-table-column label="状态" width="90"><template #default="{row}"><el-switch v-permission="['/api/user/:id/enable','PUT']" :model-value="row.enable === 1" @change="value => toggle(row,value)"/></template></el-table-column>
      <el-table-column label="操作" width="270"><template #default="{row}"><el-button v-permission="['/api/user/:id','PUT']" size="small" @click="openEdit(row)">编辑</el-button><el-button v-permission="['/api/user/:id/reset_pwd','PUT']" size="small" @click="reset(row)">重置密码</el-button><el-button v-permission="['/api/user/:id/delete_user','DELETE']" size="small" type="danger" @click="remove(row)">删除</el-button></template></el-table-column>
    </el-table>
    <el-pagination class="pager" layout="total, prev, pager, next" :total="total" :page-size="query.pageSize" v-model:current-page="query.page" @current-change="load"/>
  </el-card>
  <el-dialog v-model="dialog" :title="form.ID ? '编辑用户' : '新增用户'" width="480px"><el-form :model="form" label-width="90px">
    <el-form-item label="用户名"><el-input v-model="form.username" :disabled="!!form.ID"/></el-form-item><el-form-item v-if="!form.ID" label="初始密码"><el-input v-model="form.password" type="password" show-password/></el-form-item>
    <el-form-item label="昵称"><el-input v-model="form.nickName"/></el-form-item><el-form-item label="邮箱"><el-input v-model="form.email"/></el-form-item><el-form-item label="手机"><el-input v-model="form.phone"/></el-form-item>
    <el-form-item label="角色"><el-select v-model="form.authorityId"><el-option v-for="r in roles" :key="r.authorityId" :label="r.authorityName" :value="r.authorityId"/></el-select></el-form-item>
  </el-form><template #footer><el-button @click="dialog=false">取消</el-button><el-button type="primary" @click="save">保存</el-button></template></el-dialog>
</template>
<script>
import { createUser, deleteUser, getAuthorities, getUsers, resetPassword, setUserEnable, updateUser } from '@/api/system'
export default { name:'SystemUsers', data: () => ({ loading:false, rows:[], roles:[], total:0, dialog:false, query:{page:1,pageSize:10,keyword:''}, form:{} }), mounted(){this.loadRoles();this.load()}, methods:{
  async load(){this.loading=true;try{const r=await getUsers(this.query);this.rows=(r.data&&r.data.list)||[];this.total=(r.data&&r.data.total)||0}finally{this.loading=false}},
  async loadRoles(){const r=await getAuthorities({page:1,pageSize:100});this.roles=(r.data&&r.data.list)||[]}, roleName(id){const r=this.roles.find(x=>Number(x.authorityId)===Number(id));return r?r.authorityName:id},
  openCreate(){this.form={username:'',password:'',nickName:'',email:'',phone:'',authorityId:this.roles[0]&&this.roles[0].authorityId,enable:1};this.dialog=true},
  openEdit(row){this.form={...row,ID:row.ID||row.id,username:row.userName||row.username};this.dialog=true},
  async save(){if(this.form.ID) await updateUser(this.form.ID,this.form);else await createUser(this.form);this.dialog=false;this.$message.success('保存成功');this.load()},
  async toggle(row,value){await setUserEnable(row.ID||row.id,value?1:2);row.enable=value?1:2},
  async reset(row){const {value}=await this.$prompt(`请输入 ${row.userName} 的新密码`,'重置密码',{inputType:'password',inputPattern:/^.{6,72}$/,inputErrorMessage:'密码长度须为 6-72 位'});await resetPassword(row.ID||row.id,value);this.$message.success('密码已重置')},
  async remove(row){await this.$confirm(`确认删除用户 ${row.userName}？`,'警告',{type:'warning'});await deleteUser(row.ID||row.id);this.load()}
}}
</script>
<style scoped>.toolbar{display:flex;align-items:center;justify-content:space-between;gap:10px}.search{justify-content:flex-start;margin-bottom:14px}.search .el-input{width:260px}.pager{margin-top:16px;justify-content:flex-end}</style>
