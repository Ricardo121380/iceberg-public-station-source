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
import { describe, expect, test, vi } from 'vitest'

import tokens from '../holiday-chart-tokens.json'
import {
  icebergSeriesRamp,
  registerHolidayChartTheme,
  resolveChartThemeName,
} from '../iceberg-chart-theme'

describe('Seasonal chart palette', () => {
  test('holiday overrides a stored preset and ordinary dates restore it', () => {
    expect(resolveChartThemeName('rose-garden', 'dark', 'mid-autumn')).toBe(
      'iceberg-mid-autumn-dark'
    )
    expect(resolveChartThemeName('rose-garden', 'dark', null)).toBe('dark')
    expect(resolveChartThemeName('iceberg', 'light', null)).toBe(
      'iceberg-light'
    )
  })
  test('series repeats approved HEX colors without changing data domains', () => {
    const colors = icebergSeriesRamp('dark', 8, 'dragon-boat')
    expect(colors).toHaveLength(8)
    expect(colors[0]).toBe(tokens['dragon-boat'].dark['chart-1'])
    expect(colors[5]).toBe(colors[0])
    expect(colors.every((c) => /^#[0-9a-f]{6}$/i.test(c))).toBe(true)
  })
  test('registers seasonal axes and tooltip colors and skips an existing theme', () => {
    const manager = {
      themeExist: vi.fn(() => false),
      getTheme: vi.fn(() => ({})),
      registerTheme: vi.fn(),
    }
    registerHolidayChartTheme(manager, 'mid-autumn', 'light')
    expect(manager.registerTheme).toHaveBeenCalledWith(
      'iceberg-mid-autumn-light',
      expect.objectContaining({
        background: 'transparent',
        colorScheme: { default: icebergSeriesRamp('light', 5, 'mid-autumn') },
      })
    )
    manager.themeExist.mockReturnValue(true)
    registerHolidayChartTheme(manager, 'mid-autumn', 'light')
    expect(manager.registerTheme).toHaveBeenCalledTimes(1)
  })
})
