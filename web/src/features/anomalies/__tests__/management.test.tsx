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
import { beforeEach, expect, it, vi } from 'vitest'

import { api } from '@/lib/api'
import { useAuthStore } from '@/stores/auth-store'

import { AnomalyConsole } from '../index'

const policy = {
  enabled: true,
  window_minutes: 5,
  schema_threshold: 10,
  rate_threshold: 100,
  upstream_threshold: 5,
  version: 2,
}
const event = {
  id: 'a'.repeat(32),
  user_id: 305,
  channel_id: 1,
  model: 'model-'.repeat(35),
  kind: 'invalid_schema',
  count: 12,
  threshold: 10,
  first_seen: 1000,
  last_seen: 1100,
  alert: true,
  acknowledged_count: 0,
  acknowledged_at: 0,
  acknowledged_by: 0,
}
function setRole(role: number) {
  const state = useAuthStore.getState()
  useAuthStore.setState({
    auth: {
      ...state.auth,
      user: { id: 1, username: 'test', role } as NonNullable<
        typeof state.auth.user
      >,
    },
  })
}
function mount() {
  return render(
    <QueryClientProvider
      client={
        new QueryClient({
          defaultOptions: {
            queries: { retry: false },
            mutations: { retry: false },
          },
        })
      }
    >
      <AnomalyConsole />
    </QueryClientProvider>
  )
}
beforeEach(() => {
  setRole(100)
  vi.spyOn(api, 'get').mockImplementation(async (url) => {
    let data: unknown = { items: [event], total: 1, pending: 1, retained: 1 }
    if (url.endsWith('/settings')) data = policy
    if (url.endsWith('/audit')) data = []
    return { data: { success: true, data } }
  })
})
it('ordinary users cannot load the console or request its data', () => {
  setRole(1)
  mount()
  expect(screen.getByRole('alert')).toHaveTextContent(
    'Administrator access is required'
  )
  expect(api.get).not.toHaveBeenCalled()
})
it('an administrator can acknowledge an event but cannot edit rules', async () => {
  setRole(10)
  const post = vi
    .spyOn(api, 'post')
    .mockResolvedValue({ data: { success: true } })
  mount()
  await userEvent.click(
    await screen.findByRole('button', { name: 'Acknowledge' })
  )
  expect(post).toHaveBeenCalledWith(
    `/api/anomalies/events/${event.id}/acknowledge`,
    { count: 12 }
  )
  await userEvent.click(screen.getByRole('tab', { name: 'Alert rules' }))
  expect(
    await screen.findByLabelText('Collect request anomalies')
  ).toBeDisabled()
  expect(
    screen.queryByRole('button', { name: 'Save alert rules' })
  ).not.toBeInTheDocument()
})
it('root can save validated thresholds with the loaded version', async () => {
  const put = vi
    .spyOn(api, 'put')
    .mockResolvedValue({ data: { success: true } })
  mount()
  await userEvent.click(screen.getByRole('tab', { name: 'Alert rules' }))
  const input = await screen.findByLabelText('Schema error alert threshold')
  await userEvent.clear(input)
  await userEvent.type(input, '20')
  await userEvent.click(
    screen.getByRole('button', { name: 'Save alert rules' })
  )
  await waitFor(() =>
    expect(put).toHaveBeenCalledWith('/api/anomalies/settings', {
      ...policy,
      schema_threshold: 20,
    })
  )
})
it('filter submission resets pagination and requests the selected user and category', async () => {
  mount()
  await screen.findByRole('button', { name: 'Acknowledge' })
  await userEvent.type(screen.getByLabelText('User ID'), '305')
  await userEvent.selectOptions(
    screen.getByLabelText('Failure type'),
    'invalid_schema'
  )
  await userEvent.click(screen.getByRole('button', { name: 'Apply filters' }))
  await waitFor(() =>
    expect(api.get).toHaveBeenCalledWith('/api/anomalies/events', {
      params: {
        p: 1,
        user_id: '305',
        kind: 'invalid_schema',
        state: 'pending',
      },
    })
  )
  expect(screen.getByRole('button', { name: 'Next' })).toBeDisabled()
})
it('empty results explain collection coverage instead of claiming there are no failures', async () => {
  vi.mocked(api.get).mockImplementation(async (url) => ({
    data: {
      success: true,
      data: url.endsWith('/settings')
        ? policy
        : { items: [], total: 0, pending: 0, retained: 0 },
    },
  }))
  mount()
  expect(await screen.findByText('No matching anomalies')).toBeInTheDocument()
  expect(
    screen.getByText(/historical logs are not imported/)
  ).toBeInTheDocument()
})
it('API failure provides a retry and never renders an empty success', async () => {
  vi.mocked(api.get).mockRejectedValue(new Error('offline'))
  mount()
  expect(await screen.findByRole('alert')).toHaveTextContent(
    'Unable to load anomaly data'
  )
  expect(screen.getByRole('button', { name: 'Retry' })).toBeEnabled()
  expect(screen.queryByText('No matching anomalies')).not.toBeInTheDocument()
})
it('acknowledgement conflict prompts refresh and does not claim success', async () => {
  vi.spyOn(api, 'post').mockRejectedValue(new Error('conflict'))
  mount()
  await userEvent.click(
    await screen.findByRole('button', { name: 'Acknowledge' })
  )
  expect(await screen.findByRole('alert')).toHaveTextContent(
    'Refresh and acknowledge the latest count'
  )
})
it('handling history has a useful empty state', async () => {
  mount()
  await userEvent.click(screen.getByRole('tab', { name: 'Handling history' }))
  expect(await screen.findByText('No handling records yet')).toBeInTheDocument()
})
