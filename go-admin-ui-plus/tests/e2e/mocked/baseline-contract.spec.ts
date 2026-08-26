import { test, expect } from '@playwright/test'
import { authenticate, installApiMocks, userInfo } from './fixtures'
import { captureBodies, json } from './support/crud'

const CAPTCHA_IMAGE = 'data:image/gif;base64,R0lGODlhAQABAIAAAAAAAP///ywAAAAAAQABAAACAUwAOw=='

async function installLoginMocks(page: Parameters<typeof installApiMocks>[0]) {
  const calls = { login: 0, captcha: 0 }

  await installApiMocks(page)
  await page.route('**/api/v1/captcha*', async route => {
    calls.captcha++
    await route.fulfill(json({ code: 200, id: 'baseline-captcha', data: CAPTCHA_IMAGE }))
  })
  await page.route('**/api/v1/login', async route => {
    calls.login++
    await route.fulfill(json({ code: 200, token: 'baseline-session-token', msg: '登录成功' }))
  })

  return calls
}

test.describe('phase-1 behavior baseline', () => {
  test('logs in through the top-level token envelope and loads the dynamic menu', async({ page, context }) => {
    const calls = await installLoginMocks(page)

    await page.goto('/#/login')
    await page.getByPlaceholder('请输入验证码').fill('12345')
    await page.getByRole('button', { name: '登录' }).click()

    await expect(page).not.toHaveURL(/#\/login/)
    await expect(page.locator('.el-menu').first()).toBeVisible()
    await expect(page.locator('.el-sub-menu__title', { hasText: 'Demo' })).toBeVisible()
    expect(calls.login).toBe(1)
    expect(calls.captcha).toBe(1)

    const token = (await context.cookies()).find(cookie => cookie.name === 'Admin-Token')
    expect(token?.value).toBe('baseline-session-token')
  })

  test('keeps the login usable after rejected credentials', async({ page }) => {
    const calls = await installLoginMocks(page)
    await page.route('**/api/v1/login', async route => {
      calls.login++
      await route.fulfill(json({ code: 400, msg: 'incorrect Username or Password' }))
    })

    await page.goto('/#/login')
    await page.getByPlaceholder('请输入验证码').fill('12345')
    await page.getByRole('button', { name: '登录' }).click()

    await expect(page).toHaveURL(/#\/login/)
    await expect(page.locator('.el-message--error')).toHaveCount(1)
    await expect(page.locator('.el-message--error')).toContainText('incorrect Username or Password')
    await expect(page.getByRole('button', { name: '登录' })).toBeEnabled()
    expect(calls.login).toBe(1)
    await expect.poll(() => calls.captcha).toBe(2)
  })

  test('keeps the login usable when the backend is unreachable', async({ page }) => {
    const calls = await installLoginMocks(page)
    await page.route('**/api/v1/login', async route => {
      calls.login++
      await route.abort('connectionrefused')
    })

    await page.goto('/#/login')
    await page.getByPlaceholder('请输入验证码').fill('12345')
    await page.getByRole('button', { name: '登录' }).click()

    await expect(page).toHaveURL(/#\/login/)
    await expect(page.locator('.el-message--error')).toHaveCount(1)
    await expect(page.getByRole('button', { name: '登录' })).toBeEnabled()
    expect(calls.login).toBe(1)
    await expect.poll(() => calls.captcha).toBe(2)
  })

  test('sends the representative product create, update, and delete contracts', async({ page, context }) => {
    await authenticate(context)
    const { calls } = await installApiMocks(page)
    await page.route('**/api/v1/getinfo*', async route => {
      await route.fulfill(json({
        ...userInfo,
        data: {
          ...userInfo.data,
          permissions: [...userInfo.data.permissions, 'demo:product:delete']
        }
      }))
    })
    const creates = await captureBodies(page, /\/api\/v1\/demo-product(\?|$)/, 'POST')
    const updates = await captureBodies(page, /\/api\/v1\/demo-product\/\d+(\?|$)/, 'PUT')
    const deletes = await captureBodies(page, /\/api\/v1\/demo-product(\?|$)/, 'DELETE')

    await page.goto('/#/demo/product')
    await page.waitForSelector('.el-table')

    await page.locator('.pro-table__toolbar').getByRole('button', { name: '新增' }).click()
    const createDialog = page.getByRole('dialog', { name: '新增' })
    await createDialog.getByPlaceholder('请输入名称').fill('Baseline Product')
    await createDialog.getByPlaceholder('请输入编码').fill('BASELINE-UI-001')
    await createDialog.getByRole('button', { name: '确 定' }).click()
    await expect.poll(() => calls.product.create).toBe(1)
    expect(JSON.parse(creates.at(-1) ?? '{}')).toMatchObject({
      name: 'Baseline Product',
      code: 'BASELINE-UI-001',
      status: '1'
    })

    await page.getByRole('row', { name: /Alpha/ }).getByRole('button', { name: '修改' }).click()
    const updateDialog = page.getByRole('dialog', { name: '修改' })
    await updateDialog.getByPlaceholder('请输入名称').fill('Alpha Updated')
    await updateDialog.getByRole('button', { name: '确 定' }).click()
    await expect.poll(() => calls.product.update).toBe(1)
    expect(JSON.parse(updates.at(-1) ?? '{}')).toMatchObject({ id: 1, name: 'Alpha Updated' })

    await page.getByRole('row', { name: /Beta/ }).getByRole('button', { name: '删除' }).click()
    await page.locator('.el-message-box').getByRole('button', { name: '确定' }).click()
    await expect.poll(() => calls.product.remove).toBe(1)
    expect(JSON.parse(deletes.at(-1) ?? '{}')).toEqual({ ids: [2] })
  })
})
