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
import {
  createMemoryHistory,
  createRootRoute,
  createRouter,
  RouterProvider,
} from '@tanstack/react-router'
import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import { afterEach, describe, expect, test, vi } from 'vitest'

import { StationLaunchpad } from '../station-launchpad'

function renderLaunchpad(isAuthenticated = false) {
  const rootRoute = createRootRoute({
    component: () => <StationLaunchpad isAuthenticated={isAuthenticated} />,
  })
  const router = createRouter({
    routeTree: rootRoute,
    history: createMemoryHistory({ initialEntries: ['/'] }),
  })
  render(<RouterProvider router={router} />)
}

afterEach(() => {
  Reflect.deleteProperty(navigator, 'clipboard')
})

describe('Homepage launchpad', () => {
  test('a visitor can copy the real station address and receives confirmation', async () => {
    const writeText = vi.fn().mockResolvedValue(undefined)
    Object.defineProperty(navigator, 'clipboard', {
      configurable: true,
      value: { writeText },
    })
    renderLaunchpad()
    fireEvent.click(
      await screen.findByRole('button', { name: 'Copy API address' })
    )
    await waitFor(() =>
      expect(screen.getByRole('status')).toHaveTextContent('Address copied.')
    )
    expect(writeText).toHaveBeenCalledWith('https://iceberg.tiktok.vip/v1')
  })

  test('clipboard failure keeps the address selectable and offers manual copy', async () => {
    Object.defineProperty(navigator, 'clipboard', {
      configurable: true,
      value: { writeText: vi.fn().mockRejectedValue(new Error('denied')) },
    })
    renderLaunchpad()
    fireEvent.click(
      await screen.findByRole('button', { name: 'Copy API address' })
    )
    await waitFor(() =>
      expect(screen.getByRole('status')).toHaveTextContent('copy it manually')
    )
    expect(screen.getByLabelText('API Base URL')).toHaveValue(
      'https://iceberg.tiktok.vip/v1'
    )
  })

  test('a signed-out visitor is directed to login with the key-management destination', async () => {
    renderLaunchpad()
    const link = await screen.findByRole('link', {
      name: /Create your API key/,
    })
    const url = new URL(link.getAttribute('href') ?? '', 'http://localhost')
    expect(url.pathname).toBe('/sign-in')
    expect(url.searchParams.get('redirect')).toBe('/keys')
  })

  test('a signed-in visitor can open key management directly', async () => {
    renderLaunchpad(true)
    expect(
      await screen.findByRole('link', { name: /Create your API key/ })
    ).toHaveAttribute('href', '/keys')
  })
})
