<script setup lang="ts">
import { nextTick, reactive, ref } from 'vue'
import type { SessionController } from '@go-admin-plus/domain-iam/session'
import GopherMark from './GopherMark.vue'

const props = defineProps<{ controller: SessionController }>()
const emit = defineEmits<{ authenticated: [] }>()
const credentials = reactive({ username: '', password: '' })
const submitting = ref(false)
const failed = ref(false)
const passwordVisible = ref(false)
const passwordInput = ref<HTMLInputElement | null>(null)

const togglePassword = async () => {
  passwordVisible.value = !passwordVisible.value
  await nextTick()
  passwordInput.value?.focus()
}

const submit = async () => {
  if (submitting.value) return
  submitting.value = true
  failed.value = false
  try { await props.controller.login(credentials) } finally {
    credentials.password = ''
    passwordVisible.value = false
    submitting.value = false
  }
  if (props.controller.state().status === 'authenticated') emit('authenticated')
  else failed.value = true
}
</script>

<template>
  <main class="login-page">
    <section class="login-page__stage" aria-label="Go Admin Plus development runtime">
      <header class="login-page__stage-header"><span>go<span>-</span>admin-plus</span><small>developer edition</small></header>
      <div class="login-page__scene">
        <div class="login-page__terminal" role="img" aria-label="Go Admin Plus 启动过程示意">
          <div class="login-page__terminal-bar"><i /><i /><i /><span>~/go-admin-plus</span></div>
          <pre><code><span class="login-page__line login-page__line--1"><b>$</b> task dev <em>TARGET=server PROFILE=server-sqlite</em></span>
<span class="login-page__line login-page__line--2 success">✓ server ready on http://localhost:8000</span>
<span class="login-page__line login-page__line--3"><b>$</b> task dev <em>TARGET=web</em></span>
<span class="login-page__line login-page__line--4 success">✓ web ready on http://localhost:5173</span>
<span class="login-page__line login-page__line--5"><b>$</b> task dev <em>TARGET=desktop</em></span>
<span class="login-page__line login-page__line--6 success">✓ Tauri 2 desktop host ready</span>
<span class="login-page__line login-page__line--7"><b>$</b> <i class="login-page__caret" /></span></code></pre>
        </div>
        <GopherMark class="login-page__mascot" />
      </div>
      <footer><span>Go</span><span>Vue 3</span><span>TypeScript</span><span>Tauri 2</span><span>SQLite / PostgreSQL</span></footer>
    </section>

    <section class="login-page__panel">
      <form aria-label="登录" @submit.prevent="submit">
        <h1>Go Admin Plus</h1>
        <p class="login-page__subtitle">使用管理员账号登录控制台</p>
        <label>账号<input v-model.trim="credentials.username" autocomplete="username" placeholder="请输入账号" autofocus required minlength="3" maxlength="64"></label>
        <label>密码
          <span class="login-page__password-field">
            <input ref="passwordInput" v-model="credentials.password" autocomplete="current-password" :type="passwordVisible ? 'text' : 'password'" placeholder="请输入密码" required minlength="12" maxlength="128">
            <button
              class="login-page__password-toggle"
              type="button"
              :aria-label="passwordVisible ? '隐藏密码' : '显示密码'"
              :aria-pressed="passwordVisible"
              :title="passwordVisible ? '隐藏密码' : '显示密码'"
              @click="togglePassword"
            >{{ passwordVisible ? '隐藏' : '显示' }}</button>
          </span>
        </label>
        <p v-if="failed" role="alert">账号或密码无法验证，请重试。</p>
        <button type="submit" :disabled="submitting">{{ submitting ? '登录中' : '登录' }}</button>
        <p class="login-page__tip">忘记密码请联系系统管理员重置</p>
      </form>
    </section>
  </main>
</template>

<style scoped>
.login-page { display: flex; width: 100vw; min-height: 100vh; overflow: hidden; background: var(--ga-bg-body); }
.login-page__stage { position: relative; display: flex; width: 72%; min-height: 100vh; flex-direction: column; padding: 44px 56px 40px; overflow: hidden; color: #c9d1d9; background: linear-gradient(160deg, #0f1f3a 0%, #16305a 55%, #112544 100%); }
.login-page__stage::before { position: absolute; inset: 0; background: radial-gradient(50% 46% at 84% 62%, rgb(0 200 255 / 24%) 0%, transparent 70%), radial-gradient(48% 44% at 16% 24%, rgb(124 92 255 / 14%) 0%, transparent 68%); content: ''; pointer-events: none; }
.login-page__stage-header, .login-page__scene, .login-page__stage footer { position: relative; }
.login-page__stage-header { display: flex; align-items: baseline; gap: 12px; font: 600 17px/1 ui-monospace, SFMono-Regular, Menlo, Consolas, monospace; color: #fff; }
.login-page__stage-header > span > span { color: #00c8ff; }
.login-page__stage-header small { color: #8ba0bd; font-size: 11px; font-weight: 400; }
.login-page__scene { display: flex; align-items: flex-end; justify-content: space-between; gap: 32px; margin: auto 0; }
.login-page__terminal { flex: 0 1 620px; min-width: 0; overflow: hidden; background: rgb(8 20 40 / 62%); border: 1px solid #2a4574; border-radius: 10px; box-shadow: 0 24px 60px rgb(0 0 0 / 45%); }
.login-page__terminal-bar { display: flex; align-items: center; gap: 7px; padding: 11px 14px; background: rgb(255 255 255 / 2%); border-bottom: 1px solid #2a4574; }
.login-page__terminal-bar i { width: 11px; height: 11px; border-radius: 50%; background: #ff5f57; }
.login-page__terminal-bar i:nth-child(2) { background: #febc2e; }
.login-page__terminal-bar i:nth-child(3) { background: #28c840; }
.login-page__terminal-bar span { margin-left: 8px; color: #8ba0bd; font: 12px/1 ui-monospace, SFMono-Regular, Menlo, Consolas, monospace; }
.login-page__terminal pre { margin: 0; padding: 20px 22px 24px; overflow: hidden; color: #c9d1d9; font: 12.5px/1.85 ui-monospace, SFMono-Regular, Menlo, Consolas, monospace; white-space: pre-wrap; word-break: break-all; }
.login-page__line { display: block; opacity: 0; animation: line-in .32s ease forwards; }
.login-page__line--1 { animation-delay: .15s; }
.login-page__line--2 { animation-delay: .5s; }
.login-page__line--3 { animation-delay: .8s; }
.login-page__line--4 { animation-delay: 1.1s; }
.login-page__line--5 { animation-delay: 1.28s; }
.login-page__line--6 { animation-delay: 1.46s; }
.login-page__line--7 { animation-delay: 1.64s; }
.login-page__terminal b { margin-right: 6px; color: #00c8ff; }
.login-page__terminal em { color: #8ba0bd; font-style: normal; }
.login-page__terminal .success { color: #5ce68b; }
.login-page__caret { display: inline-block; width: 8px; height: 15px; vertical-align: -2px; background: #00c8ff; animation: caret-blink 1.1s step-end infinite; }
.login-page__mascot { flex: 0 0 clamp(130px, 14vw, 194px); margin-bottom: -6px; filter: drop-shadow(0 16px 24px rgb(0 0 0 / 32%)); }
.login-page__stage footer { display: flex; flex-wrap: wrap; gap: 18px; color: #8ba0bd; font: 12px/1 ui-monospace, SFMono-Regular, Menlo, Consolas, monospace; }
.login-page__panel { display: flex; flex: 1; align-items: center; justify-content: center; padding: 40px; background: var(--ga-bg-body); }
form { display: block; width: min(100%, 300px); }
h1 { margin: 0 0 6px; color: var(--ga-text-1); font-size: 22px; font-weight: 650; letter-spacing: 0; }
.login-page__subtitle { margin: 0 0 30px; color: var(--ga-text-2); font-size: 13px; }
.login-page__tip { margin: 22px 0 0; color: var(--ga-text-3); font-size: 12px; }
label { display: grid; gap: 7px; margin-bottom: 18px; color: var(--ga-text-2); font-size: 13px; font-weight: 500; }
.login-page__password-field { position: relative; display: block; }
input { width: 100%; min-height: 42px; padding: 0 12px; color: var(--ga-text-1); background: var(--ga-bg-container); border: 1px solid var(--ga-border-light); border-radius: 8px; outline: none; transition: border-color .16s, box-shadow .16s; }
.login-page__password-field input { padding-right: 54px; }
input:hover { border-color: var(--ga-border); }
input:focus { border-color: #00add8; box-shadow: 0 0 0 3px rgb(0 173 216 / 16%); }
.login-page__password-toggle { position: absolute; top: 1px; right: 1px; bottom: 1px; width: 50px; padding: 0; color: var(--ga-text-3); background: transparent; border: 0; border-radius: 0 8px 8px 0; cursor: pointer; font-size: 12px; }
.login-page__password-toggle:hover, .login-page__password-toggle:focus-visible { color: #00add8; background: var(--ga-bg-hover); }
.login-page__password-toggle:focus-visible { outline: 2px solid #00add8; outline-offset: -3px; }
form > button { width: 100%; min-height: 42px; margin-top: 4px; color: #04121f; background: linear-gradient(135deg, #00c8ff, #00add8); border: 0; border-radius: 8px; box-shadow: 0 4px 14px rgb(0 173 216 / 28%); cursor: pointer; font-weight: 600; letter-spacing: 3px; }
form > button:disabled { cursor: not-allowed; opacity: .6; }
[role="alert"] { margin: 0 0 18px; padding: 9px 10px; color: #b42318; background: #fff1f0; border-left: 3px solid #b42318; font-size: 12px; }
@keyframes line-in { from { opacity: 0; transform: translateY(4px); } to { opacity: 1; transform: none; } }
@keyframes caret-blink { 50% { opacity: 0; } }
@media (max-width: 1120px) { .login-page__mascot { display: none; } .login-page__terminal { flex: 1 1 auto; } }
@media (max-width: 1100px) { .login-page__stage { width: 60%; padding: 32px 36px; } }
@media (max-width: 860px) {
  .login-page { flex-direction: column; overflow: auto; }
  .login-page__stage { width: 100%; min-height: auto; padding: 24px; }
  .login-page__scene { margin: 20px 0; }
  .login-page__stage footer { display: none; }
  .login-page__panel { min-height: 390px; padding: 32px 24px; }
  form { width: min(100%, 340px); }
}
@media (max-width: 520px) { .login-page__terminal pre { font-size: 10.5px; } }
@media (prefers-reduced-motion: reduce) { .login-page__line { opacity: 1; animation: none; } .login-page__caret { animation: none; } }
</style>
