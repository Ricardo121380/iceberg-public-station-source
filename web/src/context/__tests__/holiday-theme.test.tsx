/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/
import { act, cleanup, fireEvent, render, screen } from '@testing-library/react'
import { createPortal } from 'react-dom'
import { afterEach, beforeEach, describe, expect, test, vi } from 'vitest'

import { ThemeCustomizationProvider } from '../theme-customization-provider'
import { ThemeProvider, useTheme } from '../theme-provider'

function Consumer() {
  const { holiday, resolvedTheme, setTheme } = useTheme()
  return (
    <>
      <output>
        {holiday ?? 'ordinary'}:{resolvedTheme}
      </output>
      <button type='button' onClick={() => setTheme('light')}>
        Light
      </button>
      {createPortal(<div role='dialog'>Portal</div>, document.body)}
    </>
  )
}
function mount() {
  return render(
    <ThemeProvider>
      <ThemeCustomizationProvider>
        <Consumer />
      </ThemeCustomizationProvider>
    </ThemeProvider>
  )
}
beforeEach(() => {
  vi.useFakeTimers()
  vi.setSystemTime(new Date('2026-09-19T04:00:00Z'))
})
afterEach(() => {
  cleanup()
  vi.useRealTimers()
  document.body.removeAttribute('data-theme-preset')
  for (const cookie of document.cookie.split(';')) {
    document.cookie = `${cookie.split('=')[0].trim()}=;Max-Age=0;path=/`
  }
})
describe('Global holiday theme', () => {
  test('direct console mount applies the calendar theme to portal ancestor without changing stored preferences', () => {
    document.cookie = 'vite-ui-theme=dark;path=/'
    document.cookie = 'theme_preset=rose-garden;path=/'
    const before = document.cookie
    mount()
    expect(document.body).toHaveAttribute('data-holiday', 'mid-autumn')
    expect(screen.getByRole('dialog').parentElement).toBe(document.body)
    expect(screen.getByText('mid-autumn:dark')).toBeInTheDocument()
    expect(document.cookie).toBe(before)
  })
  test('manual light mode keeps the seasonal theme', () => {
    mount()
    fireEvent.click(screen.getByRole('button', { name: 'Light' }))
    expect(screen.getByText('mid-autumn:light')).toBeInTheDocument()
    expect(document.body).toHaveAttribute('data-holiday', 'mid-autumn')
  })
  test('Beijing midnight changes Mid-Autumn to National Day without remounting', () => {
    vi.setSystemTime(new Date('2026-09-27T15:59:59.900Z'))
    mount()
    act(() => vi.advanceTimersByTime(200))
    expect(document.body).toHaveAttribute('data-holiday', 'national-day')
  })
  test('returning from a suspended tab removes the seasonal overlay after the window', () => {
    mount()
    vi.setSystemTime(new Date('2026-10-04T04:00:00Z'))
    fireEvent(document, new Event('visibilitychange'))
    expect(document.body).not.toHaveAttribute('data-holiday')
    expect(screen.getByRole('status')).toHaveTextContent('ordinary')
  })
  test('focus rechecks a changed device date and cleanup removes listeners and overlay', () => {
    const view = mount()
    vi.setSystemTime(new Date('2027-06-09T04:00:00Z'))
    fireEvent.focus(window)
    expect(document.body).toHaveAttribute('data-holiday', 'dragon-boat')
    view.unmount()
    expect(document.body).not.toHaveAttribute('data-holiday')
    expect(vi.getTimerCount()).toBe(0)
  })
})
