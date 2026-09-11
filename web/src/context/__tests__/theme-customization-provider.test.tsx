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
import { render, waitFor } from '@testing-library/react'
import { afterEach, describe, expect, test } from 'vitest'

import {
  DEFAULT_THEME_CUSTOMIZATION,
  THEME_COOKIE_KEYS,
} from '@/lib/theme-customization'

import { ThemeCustomizationProvider } from '../theme-customization-provider'

afterEach(() => {
  document.body.removeAttribute('data-theme-preset')
  document.cookie = `${THEME_COOKIE_KEYS.preset}=; Max-Age=0; path=/`
})

describe('ThemeCustomizationProvider preset attribute', () => {
  test('applies the station default preset to the body attribute on a fresh visit', async () => {
    // The station default is `iceberg`, a real preset — the attribute must be
    // written for its token block in theme-presets.css to activate.
    render(
      <ThemeCustomizationProvider>
        <div />
      </ThemeCustomizationProvider>
    )
    await waitFor(() =>
      expect(document.body.getAttribute('data-theme-preset')).toBe(
        DEFAULT_THEME_CUSTOMIZATION.preset
      )
    )
    expect(DEFAULT_THEME_CUSTOMIZATION.preset).toBe('iceberg')
  })

  test('an explicit upstream default selection removes the preset attribute', async () => {
    document.cookie = `${THEME_COOKIE_KEYS.preset}=default; path=/`
    render(
      <ThemeCustomizationProvider>
        <div />
      </ThemeCustomizationProvider>
    )
    await waitFor(() =>
      expect(document.body.getAttribute('data-theme-preset')).toBeNull()
    )
  })
})
