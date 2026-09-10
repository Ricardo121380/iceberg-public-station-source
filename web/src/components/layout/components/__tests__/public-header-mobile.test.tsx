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
import { render, screen, within } from '@testing-library/react'
import { describe, expect, test, vi } from 'vitest'

import { PublicHeader } from '../public-header'

vi.mock('@/components/language-switcher', () => ({
  LanguageSwitcher: () => (
    <button type='button' aria-label='Change language stub' />
  ),
}))
vi.mock('@/components/notification-popover', () => ({
  NotificationPopover: () => null,
}))
vi.mock('@/components/profile-dropdown', () => ({
  ProfileDropdown: () => null,
}))
vi.mock('@/components/theme-switch', () => ({
  ThemeSwitch: () => <button type='button' aria-label='Theme switch stub' />,
}))
vi.mock('@/hooks/use-system-config', () => ({
  useSystemConfig: () => ({
    systemName: 'Iceberg',
    logo: '/logo.png',
    loading: false,
    logoLoaded: true,
  }),
}))
vi.mock('@/hooks/use-top-nav-links', () => ({
  useTopNavLinks: () => [],
}))
vi.mock('@/hooks/use-notifications', () => ({
  useNotifications: () => ({
    popoverOpen: false,
    setPopoverOpen: vi.fn(),
    unreadCount: 0,
    activeTab: 'notice',
    setActiveTab: vi.fn(),
    notice: '',
    announcements: [],
    loading: false,
  }),
}))

function renderHeader() {
  const rootRoute = createRootRoute({
    component: () => <PublicHeader />,
  })
  const router = createRouter({
    routeTree: rootRoute,
    history: createMemoryHistory({ initialEntries: ['/'] }),
  })
  render(<RouterProvider router={router} />)
}

describe('PublicHeader mobile drawer', () => {
  test('the mobile navigation overlay offers language switching and the sign-in action', async () => {
    renderHeader()

    // The full-screen overlay is the mobile drawer (hidden on >=sm via CSS,
    // always present in the DOM). Before this change it had no language path.
    expect(
      await screen.findByRole('link', { name: /Iceberg/ })
    ).toBeInTheDocument()
    const drawer = document.querySelector('.fixed.inset-0') as HTMLElement
    expect(drawer).not.toBeNull()
    expect(
      within(drawer).getByRole('button', { name: 'Change language stub' })
    ).toBeInTheDocument()
    expect(
      within(drawer).getByRole('link', { name: 'Sign in' })
    ).toBeInTheDocument()

    // The desktop toolbar keeps its own switcher.
    expect(
      screen.getAllByRole('button', { name: 'Change language stub' })
    ).toHaveLength(2)
  })
})
