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
import { act, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, test, vi } from 'vitest'

import { CountUpNumber } from '../count-up-number'

afterEach(() => {
  vi.useRealTimers()
  vi.restoreAllMocks()
})

describe('CountUpNumber', () => {
  test('starts at zero and settles on the formatted final value', () => {
    vi.useFakeTimers()
    render(
      <CountUpNumber
        value={9992}
        format={(n) => '$' + Math.round(n).toLocaleString()}
      />
    )
    expect(screen.getByText('$0')).toBeInTheDocument()
    act(() => {
      vi.advanceTimersByTime(2000)
    })
    expect(screen.getByText('$9,992')).toBeInTheDocument()
  })

  test('a value change tweens from the settled value, not from zero', () => {
    vi.useFakeTimers()
    const view = render(<CountUpNumber value={100} />)
    act(() => {
      vi.advanceTimersByTime(2000)
    })
    expect(screen.getByText('100')).toBeInTheDocument()

    view.rerender(<CountUpNumber value={150} />)
    // Mid-tween the display must sit between the two endpoints.
    act(() => {
      vi.advanceTimersByTime(300)
    })
    const mid = Number(screen.getByText(/^\d+$/).textContent)
    expect(mid).toBeGreaterThan(100)
    expect(mid).toBeLessThan(150)
    act(() => {
      vi.advanceTimersByTime(2000)
    })
    expect(screen.getByText('150')).toBeInTheDocument()
  })
})
