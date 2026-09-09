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
import { renderHook } from '@testing-library/react'
import { expect, it, vi } from 'vitest'

import { SECURITY_SECTION_IDS } from '@/features/system-settings/security/section-registry'
import { useAuthStore } from '@/stores/auth-store'

import { useSidebarView } from '../use-sidebar-view'

vi.mock('@tanstack/react-router', () => ({ useLocation: () => '/anomalies' }))
vi.mock('../use-sidebar-config', () => ({
  useSidebarConfig: (groups: unknown) => groups,
}))
it.each([1, 10, 100])(
  'role %s sees only authorized management entries',
  (role) => {
    const state = useAuthStore.getState()
    useAuthStore.setState({
      auth: {
        ...state.auth,
        user: { id: 1, role } as NonNullable<typeof state.auth.user>,
      },
    })
    const { result } = renderHook(() => useSidebarView())
    const entries = new Set(
      result.current.navGroups
        .flatMap((group) => group.items)
        .filter((item) => 'url' in item)
        .map((item) => item.url)
    )
    expect(entries.has('/anomalies')).toBe(role >= 10)
    expect(entries.has('/safety')).toBe(role === 100)
  }
)
it('safety management is no longer a system settings section', () => {
  expect(SECURITY_SECTION_IDS).not.toContain('abuse')
  expect(SECURITY_SECTION_IDS).toContain('rate-limit')
})
