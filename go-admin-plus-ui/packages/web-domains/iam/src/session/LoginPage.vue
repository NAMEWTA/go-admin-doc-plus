<script setup lang="ts">
import { reactive, ref } from 'vue'
import type { SessionController } from '@go-admin/domain-iam/session'

const props = defineProps<{ controller: SessionController }>()
const emit = defineEmits<{ authenticated: [] }>()
const credentials = reactive({ username: '', password: '' })
const submitting = ref(false)
const failed = ref(false)

const submit = async () => {
  if (submitting.value) return
  submitting.value = true
  failed.value = false
  try { await props.controller.login(credentials) } finally {
    credentials.password = ''
    submitting.value = false
  }
  if (props.controller.state().status === 'authenticated') emit('authenticated')
  else failed.value = true
}
</script>

<template>
  <main class="login-page">
    <form aria-label="Sign in" @submit.prevent="submit">
      <h1>Sign in</h1>
      <label>Username<input v-model.trim="credentials.username" autocomplete="username" required minlength="3" maxlength="64"></label>
      <label>Password<input v-model="credentials.password" autocomplete="current-password" type="password" required minlength="12" maxlength="128"></label>
      <p v-if="failed" role="alert">The credentials could not be verified.</p>
      <button type="submit" :disabled="submitting">Sign in</button>
    </form>
  </main>
</template>

<style scoped>
.login-page { display: grid; min-height: 100%; place-items: center; padding: 24px; }
form { display: grid; width: min(100%, 360px); gap: 16px; }
label { display: grid; gap: 6px; }
input, button { min-height: 40px; font: inherit; }
</style>
