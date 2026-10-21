import { expect, test } from '@playwright/test'

test('production shell mounts without runtime errors or placeholder pages', async ({ page }) => {
  const errors: string[] = []
  page.on('pageerror', (error) => errors.push(error.message))
  page.on('console', (message) => {
    if (message.type() === 'error') errors.push(message.text())
  })

  await page.goto('/')
  await expect(page.locator('#app')).toHaveAttribute('data-v-app', '')
  await expect(page.locator('main')).toHaveText('')
  await page.reload()
  await expect(page.locator('#app')).toHaveAttribute('data-v-app', '')
  expect(errors).toEqual([])
})
