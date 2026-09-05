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
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import { api } from '@/lib/api'
import { ROLE } from '@/lib/roles'
import { useAuthStore } from '@/stores/auth-store'

import { AbuseConsole } from '../index'
import { AbuseStatusBanner } from '../status-banner'

function mount(component: React.ReactNode) {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  })
  return render(
    <QueryClientProvider client={client}>{component}</QueryClientProvider>
  )
}
beforeEach(() => {
  const state = useAuthStore.getState()
  useAuthStore.setState({
    auth: {
      ...state.auth,
      user: { id: 1, username: 'root', role: ROLE.SUPER_ADMIN } as NonNullable<
        typeof state.auth.user
      >,
    },
  })
})
const settings = {
  mode: 'observe',
  limit_10m: 3,
  limit_24h: 8,
  freeze_minutes: 60,
  enabled_rules: [],
  rules: [],
}

describe('Safety management', () => {
  it('requires explicit confirmation before saving enforcement mode', async () => {
    vi.spyOn(api, 'get').mockImplementation(async (url) => ({
      data: {
        success: true,
        data:
          url === '/api/abuse/settings' ? settings : { items: [], total: 0 },
      },
    }))
    const put = vi
      .spyOn(api, 'put')
      .mockResolvedValue({ data: { success: true } })
    mount(<AbuseConsole />)
    await screen.findByText(
      'No verified category evidence is available. This channel remains observation only.'
    )
    await userEvent.selectOptions(
      screen.getByLabelText('Safety mode'),
      'enforce'
    )
    expect(
      screen.getByRole('button', { name: 'Save safety settings' })
    ).toBeDisabled()
    await userEvent.click(screen.getByRole('checkbox'))
    await userEvent.click(
      screen.getByRole('button', { name: 'Save safety settings' })
    )
    await waitFor(() =>
      expect(put).toHaveBeenCalledWith(
        '/api/abuse/settings',
        expect.objectContaining({ mode: 'enforce' })
      )
    )
  })
  it('shows service errors and permits retry without inventing empty state', async () => {
    vi.spyOn(api, 'get').mockRejectedValue(new Error('offline'))
    mount(<AbuseConsole />)
    expect(
      await screen.findByText('Unable to load safety events')
    ).toBeVisible()
    expect(
      screen.queryByText('No safety events match these filters.')
    ).not.toBeInTheDocument()
    expect(screen.getAllByRole('button', { name: 'Retry' })).toHaveLength(3)
  })
  it('requires an explanation before releasing an account', async () => {
    vi.spyOn(api, 'get').mockImplementation(async (url) => ({
      data: {
        success: true,
        data:
          url === '/api/abuse/settings'
            ? settings
            : {
                items:
                  url === '/api/abuse/users'
                    ? [{ user_id: 17, blocked_until: 2000000000 }]
                    : [],
                total: url === '/api/abuse/users' ? 1 : 0,
              },
      },
    }))
    const post = vi
      .spyOn(api, 'post')
      .mockResolvedValue({ data: { success: true } })
    mount(<AbuseConsole />)
    await userEvent.click(
      await screen.findByRole('button', { name: 'Release suspension' })
    )
    expect(
      screen.getByRole('button', { name: 'Confirm release' })
    ).toBeDisabled()
    await userEvent.type(
      screen.getByLabelText('Release reason for user 17'),
      'Reviewed upstream evidence'
    )
    await userEvent.click(
      screen.getByRole('button', { name: 'Confirm release' })
    )
    await waitFor(() =>
      expect(post).toHaveBeenCalledWith('/api/abuse/users/17/unfreeze', {
        reason: 'Reviewed upstream evidence',
      })
    )
  })
  it('prevents non-root users from querying administrative evidence', async () => {
    const state = useAuthStore.getState()
    useAuthStore.setState({
      auth: {
        ...state.auth,
        user: { ...state.auth.user, id: 1, username: 'user', role: ROLE.USER },
      },
    })
    const get = vi.spyOn(api, 'get')
    mount(<AbuseConsole />)
    expect(screen.getByRole('alert')).toHaveTextContent(
      'Root access is required'
    )
    expect(get).not.toHaveBeenCalled()
  })
  it('shows only own suspension and restoration time in the user banner', async () => {
    vi.spyOn(api, 'get').mockResolvedValue({
      data: {
        success: true,
        data: { suspended: true, blocked_until: 2000000000 },
      },
    })
    mount(<AbuseStatusBanner />)
    expect(await screen.findByRole('alert')).toHaveTextContent(
      'Model access is temporarily suspended. You can still sign in.'
    )
  })
})
