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
import { render, screen } from '@testing-library/react'
import { describe, expect, test, vi } from 'vitest'

import { StationErrorShell } from '../components/station-error-shell'

vi.mock('@/hooks/use-system-config', () => ({
  useSystemConfig: () => ({
    systemName: 'Iceberg',
    logo: '/logo.png',
    loading: false,
    logoLoaded: true,
  }),
}))

function renderShell() {
  const rootRoute = createRootRoute({
    component: () => (
      <StationErrorShell
        code='404'
        title='This page is not on the map.'
        description='It may have sailed away or never docked here.'
      >
        <button type='button'>Back to Home</button>
      </StationErrorShell>
    ),
  })
  const router = createRouter({
    routeTree: rootRoute,
    history: createMemoryHistory({ initialEntries: ['/'] }),
  })
  render(<RouterProvider router={router} />)
}

describe('StationErrorShell', () => {
  test('presents the status code, station-voiced copy, actions, and a way home', async () => {
    renderShell()

    // The numeric code is supporting text; the title carries the heading role.
    expect(await screen.findByText('404')).toBeVisible()
    expect(
      screen.getByRole('heading', { name: 'This page is not on the map.' })
    ).toBeVisible()
    expect(screen.getByText(/sailed away/)).toBeVisible()
    expect(
      screen.getByRole('button', { name: 'Back to Home' })
    ).toBeInTheDocument()

    // The brand mark links home so the page is never a dead end.
    expect(screen.getByRole('link', { name: /Iceberg/ })).toHaveAttribute(
      'href',
      '/'
    )

    // The decorative boat and landscape stay out of the accessibility tree.
    expect(screen.queryByRole('img', { name: /boat/i })).not.toBeInTheDocument()
  })
})
