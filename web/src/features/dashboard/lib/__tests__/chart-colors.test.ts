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
import { afterEach, describe, expect, test } from 'vitest'

import { getDashboardChartColors } from '../charts'

afterEach(() => {
  document.body.removeAttribute('data-theme-preset')
  document.documentElement.classList.remove('dark')
})

describe('getDashboardChartColors', () => {
  test('returns the ice-sea ramp under the iceberg preset in light mode', () => {
    document.body.setAttribute('data-theme-preset', 'iceberg')
    const colors = getDashboardChartColors(3)
    expect(colors).toEqual(['#0877ac', '#3d9dc7', '#79c3e2'])
  })

  test('returns the dark-cabin ramp under the iceberg preset in dark mode', () => {
    document.body.setAttribute('data-theme-preset', 'iceberg')
    document.documentElement.classList.add('dark')
    const colors = getDashboardChartColors(2)
    expect(colors).toEqual(['#85d6ff', '#4a9bc4'])
  })

  test('repeats the ramp for domains larger than the base palette', () => {
    document.body.setAttribute('data-theme-preset', 'iceberg')
    const colors = getDashboardChartColors(9)
    expect(colors).toHaveLength(9)
    expect(colors[8]).toBe(colors[0])
  })

  test('keeps the upstream default scheme for other presets', () => {
    document.body.setAttribute('data-theme-preset', 'rose-garden')
    const colors = getDashboardChartColors(3)
    expect(colors[0]).toBe('#1664FF')
    expect(colors).not.toContain('#0877ac')
  })

  test('keeps the upstream default scheme with no preset attribute', () => {
    const colors = getDashboardChartColors(3)
    expect(colors[0]).toBe('#1664FF')
    expect(colors).not.toContain('#0877ac')
  })
})
