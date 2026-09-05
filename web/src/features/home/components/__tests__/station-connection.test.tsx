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
import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import { afterEach, describe, expect, test, vi } from 'vitest'

import { StationConnection } from '../station-connection'

afterEach(() => {
  Reflect.deleteProperty(navigator, 'clipboard')
})

describe('Station connection guide', () => {
  test('copying the address writes the public station URL and announces success', async () => {
    const writeText = vi.fn().mockResolvedValue(undefined)
    Object.defineProperty(navigator, 'clipboard', {
      configurable: true,
      value: { writeText },
    })
    render(<StationConnection />)
    fireEvent.click(screen.getByRole('button', { name: 'Copy API address' }))
    await waitFor(() =>
      expect(screen.getByRole('status')).toHaveTextContent('Address copied.')
    )
    expect(writeText).toHaveBeenCalledWith('https://iceberg.tiktok.vip/v1')
  })

  test('a denied clipboard permission leaves a selectable address and manual copy instructions', async () => {
    Object.defineProperty(navigator, 'clipboard', {
      configurable: true,
      value: { writeText: vi.fn().mockRejectedValue(new Error('denied')) },
    })
    render(<StationConnection />)
    fireEvent.click(screen.getByRole('button', { name: 'Copy API address' }))
    await waitFor(() =>
      expect(screen.getByRole('status')).toHaveTextContent('copy it manually')
    )
    expect(screen.getByRole('textbox', { name: 'API Base URL' })).toHaveValue(
      'https://iceberg.tiktok.vip/v1'
    )
  })

  test('selecting cURL exposes the model-list request and marks the selected tab', async () => {
    render(<StationConnection />)
    const curl = screen.getByRole('tab', { name: 'cURL' })
    fireEvent.click(curl)
    await waitFor(() => expect(curl).toHaveAttribute('aria-selected', 'true'))
    expect(screen.getByRole('tabpanel')).toHaveTextContent(
      'https://iceberg.tiktok.vip/v1/models'
    )
    expect(screen.getByRole('tabpanel')).toHaveTextContent('YOUR_API_KEY')
    expect(screen.getByRole('tab', { name: 'Cherry Studio' })).toHaveAttribute(
      'aria-selected',
      'false'
    )
  })

  test('keyboard navigation changes the focused client tab', async () => {
    render(<StationConnection />)
    const cherry = screen.getByRole('tab', { name: 'Cherry Studio' })
    cherry.focus()
    fireEvent.keyDown(cherry, { key: 'ArrowRight' })
    await waitFor(() =>
      expect(screen.getByRole('tab', { name: 'CC Switch' })).toHaveFocus()
    )
  })
})
