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
import { render, screen } from '@testing-library/react'
import { describe, expect, test, vi } from 'vitest'

import { CountUpNumber } from '../count-up-number'

// Separate file on purpose: motion/react caches the reduced-motion read at
// module scope, so mocking it in the shared animation file would poison those
// tests. Each test file gets its own module registry, so this suite is free
// to pin "reduce" for the whole run.
describe('CountUpNumber (reduced motion)', () => {
  test('renders the final value immediately when reduced motion is requested', () => {
    const original = window.matchMedia
    vi.spyOn(window, 'matchMedia').mockImplementation((query) => ({
      ...original(query),
      matches: true,
    }))
    render(<CountUpNumber value={1992} />)
    expect(screen.getByText('1992')).toBeVisible()
  })
})
