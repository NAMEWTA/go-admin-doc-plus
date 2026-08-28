<script setup lang="ts">
import { reactive, ref } from 'vue'
import type { AccountProfile, SessionController } from '@go-admin/domain-iam/session'

const props = defineProps<{ controller: SessionController; profile: AccountProfile }>()
const emit = defineEmits<{ signedOut: [] }>()
const form = reactive({ displayName: props.profile.displayName, email: props.profile.email, avatarRef: props.profile.avatarRef })
const password = reactive({ currentPassword: '', newPassword: '' })
const busy = ref(false)
const save = async () => { if (busy.value) return; busy.value = true; try { await props.controller.updateProfile(form) } finally { busy.value = false } }
const changePassword = async () => {
  if (busy.value) return
  busy.value = true
  try {
    await props.controller.changePassword(password.currentPassword, password.newPassword)
    if (props.controller.state().status === 'unauthenticated') emit('signedOut')
  } finally {
    password.currentPassword = ''
    password.newPassword = ''
    busy.value = false
  }
}
const logout = async () => {
  if (busy.value) return
  busy.value = true
  try { await props.controller.logout() } finally {
    busy.value = false
    if (props.controller.state().status === 'unauthenticated') emit('signedOut')
  }
}
</script>

<template>
  <main class="account-page">
    <h1>个人中心</h1>
    <p v-if="controller.state().status === 'error'" role="alert">请求未能完成，请稍后重试。</p>
    <form @submit.prevent="save">
      <h2>基本资料</h2>
      <label>用户昵称<input v-model.trim="form.displayName" required maxlength="80"></label>
      <label>邮箱<input v-model.trim="form.email" type="email" required></label>
      <label>头像引用<input v-model.trim="form.avatarRef" pattern="files/.+"></label>
      <button type="submit" :disabled="busy">保存资料</button>
    </form>
    <form @submit.prevent="changePassword">
      <h2>修改密码</h2>
      <label>当前密码<input v-model="password.currentPassword" type="password" autocomplete="current-password" required minlength="12"></label>
      <label>新密码<input v-model="password.newPassword" type="password" autocomplete="new-password" required minlength="12"></label>
      <button type="submit" :disabled="busy">修改密码</button>
    </form>
    <button type="button" :disabled="busy" @click="logout">退出登录</button>
  </main>
</template>

<style scoped>
.account-page, form { display: grid; gap: 16px; }
.account-page > form { width: min(100%, 720px); }
label { display: grid; gap: 6px; }
</style>
