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
    <h1>Account</h1>
    <p v-if="controller.state().status === 'error'" role="alert">The request could not be completed.</p>
    <form @submit.prevent="save">
      <label>Display name<input v-model.trim="form.displayName" required maxlength="80"></label>
      <label>Email<input v-model.trim="form.email" type="email" required></label>
      <label>Avatar reference<input v-model.trim="form.avatarRef" pattern="files/.+"></label>
      <button type="submit" :disabled="busy">Save profile</button>
    </form>
    <form @submit.prevent="changePassword">
      <h2>Change password</h2>
      <label>Current password<input v-model="password.currentPassword" type="password" autocomplete="current-password" required minlength="12"></label>
      <label>New password<input v-model="password.newPassword" type="password" autocomplete="new-password" required minlength="12"></label>
      <button type="submit" :disabled="busy">Change password</button>
    </form>
    <button type="button" :disabled="busy" @click="logout">Sign out</button>
  </main>
</template>

<style scoped>
.account-page, form { display: grid; gap: 16px; }
.account-page { width: min(100%, 720px); padding: 24px; }
label { display: grid; gap: 6px; }
input, button { min-height: 40px; font: inherit; }
</style>
