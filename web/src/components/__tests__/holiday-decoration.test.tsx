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
import { toast } from 'sonner'
import { afterEach, beforeEach, describe, expect, test, vi } from 'vitest'

import { ThemeProvider } from '@/context/theme-provider'
import { holidaySuccess } from '@/lib/holiday-success'

import { HolidayDecoration } from '../holiday-decoration'
import { Toaster } from '../ui/sonner'

beforeEach(() => {
  vi.useFakeTimers()
  vi.setSystemTime(new Date('2026-09-19T04:00:00Z'))
})
afterEach(() => {
  toast.dismiss()
  cleanup()
  vi.useRealTimers()
})
describe('Seasonal decorations', () => {
  test('holiday decoration is hidden from assistive technology and vanishes outside the window', () => {
    const { container } = render(
      <ThemeProvider>
        <HolidayDecoration kind='scene' />
        <button type='button'>Action</button>
      </ThemeProvider>
    )
    expect(container.querySelector('.holo-scene')).toHaveAttribute(
      'aria-hidden',
      'true'
    )
    vi.setSystemTime(new Date('2026-10-04T04:00:00Z'))
    fireEvent.focus(window)
    expect(container.querySelector('.holo')).toBeNull()
    expect(screen.getByRole('button', { name: 'Action' })).toBeEnabled()
  })
  test('document visibility explicitly pauses ambient animation and cleans up', () => {
    const view = render(
      <ThemeProvider>
        <HolidayDecoration kind='mark' ambient />
      </ThemeProvider>
    )
    const hidden = vi.spyOn(document, 'hidden', 'get').mockReturnValue(true)
    fireEvent(document, new Event('visibilitychange'))
    expect(document.body).toHaveAttribute('data-holiday-paused')
    hidden.mockReturnValue(false)
    fireEvent(document, new Event('visibilitychange'))
    expect(document.body).not.toHaveAttribute('data-holiday-paused')
    view.unmount()
    hidden.mockRestore()
  })
  test('only explicitly opted-in success receives a celebration; regular deletion success and errors do not', async () => {
    render(
      <ThemeProvider>
        <Toaster />
      </ThemeProvider>
    )
    await act(async () => {
      holidaySuccess('Created')
      toast.success('Deleted')
      toast.error('Failed')
      vi.advanceTimersByTime(100)
    })
    expect(document.querySelectorAll('.holo-celebrate')).toHaveLength(1)
    expect(screen.getByText('Created')).toBeInTheDocument()
    expect(screen.getByText('Deleted')).toBeInTheDocument()
    expect(screen.getByText('Failed')).toBeInTheDocument()
  })
  test('ordinary dates use standard success feedback without seasonal icon', async () => {
    vi.setSystemTime(new Date('2026-07-01T04:00:00Z'))
    render(
      <ThemeProvider>
        <Toaster />
      </ThemeProvider>
    )
    await act(async () => {
      holidaySuccess('Copied')
      vi.advanceTimersByTime(100)
    })
    expect(document.querySelector('.holo-celebrate')).toBeNull()
    expect(screen.getByText('Copied')).toBeInTheDocument()
  })
})
