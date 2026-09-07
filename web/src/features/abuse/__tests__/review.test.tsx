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
import { beforeEach, afterEach, describe, expect, it, vi } from 'vitest'

import { api } from '@/lib/api'

import type { AbuseEvent } from '../api'
import { SafetyReviewDialog } from '../review-dialog'

const event: AbuseEvent = {
  id: 9,
  user_id: 59,
  token_id: 1,
  channel_id: 1,
  request_id: 'test-request',
  rule_id: 'upstream_unclassified',
  category: 'unclassified',
  action: 'recorded',
  signal: 'content_policy_violation',
  summary: '',
  model: 'test-model',
  protocol: 'error',
  created_at: 1000,
  count_10m: 0,
  count_24h: 0,
  review_version: 0,
}
function mount() {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  })
  return render(
    <QueryClientProvider client={client}>
      <SafetyReviewDialog event={event} />
    </QueryClientProvider>
  )
}
beforeEach(() => {
  vi.spyOn(api, 'get').mockResolvedValue({
    data: { success: true, data: { items: [], total: 0 } },
  })
})
afterEach(() => {
  vi.restoreAllMocks()
})

describe('Safety review', () => {
  it('does not fetch the excerpt until explicitly requested and clears it when closed', async () => {
    const post = vi
      .spyOn(api, 'post')
      .mockResolvedValue({
        data: {
          success: true,
          data: {
            status: 'available',
            expires_at: Math.floor(Date.now() / 1000) + 3600,
            excerpt: {
              text: 'redacted question',
              source: 'input',
              truncated: true,
              omitted_parts: true,
            },
          },
        },
      })
    mount()
    await userEvent.click(
      screen.getByRole('button', { name: 'Review safety event' })
    )
    expect(post).not.toHaveBeenCalled()
    await userEvent.click(
      screen.getByRole('button', {
        name: 'View redacted excerpt (access is audited)',
      })
    )
    expect(
      await screen.findByLabelText('Redacted input excerpt')
    ).toHaveTextContent('redacted question')
    expect(
      screen.getByText('Only the first 300 redacted characters are shown.')
    ).toBeVisible()
    expect(screen.getByText('Non-text parts were omitted.')).toBeVisible()
    await userEvent.keyboard('{Escape}')
    await waitFor(() =>
      expect(screen.queryByText('redacted question')).not.toBeInTheDocument()
    )
    await userEvent.click(
      screen.getByRole('button', { name: 'Review safety event' })
    )
    expect(screen.queryByText('redacted question')).not.toBeInTheDocument()
    expect(post).toHaveBeenCalledTimes(1)
  })
  it('shows expiry without exposing returned stale plaintext', async () => {
    vi.spyOn(api, 'post').mockResolvedValue({
      data: {
        success: true,
        data: { status: 'expired', expires_at: 1, excerpt: { text: '' } },
      },
    })
    mount()
    await userEvent.click(
      screen.getByRole('button', { name: 'Review safety event' })
    )
    await userEvent.click(
      screen.getByRole('button', {
        name: 'View redacted excerpt (access is audited)',
      })
    )
    expect(
      await screen.findByText(
        'The excerpt has expired and is no longer available.'
      )
    ).toBeVisible()
    expect(
      screen.queryByLabelText('Redacted input excerpt')
    ).not.toBeInTheDocument()
  })
  it('does not expose content when the read audit fails', async () => {
    vi.spyOn(api, 'post').mockRejectedValue(new Error('unavailable'))
    mount()
    await userEvent.click(
      screen.getByRole('button', { name: 'Review safety event' })
    )
    await userEvent.click(
      screen.getByRole('button', {
        name: 'View redacted excerpt (access is audited)',
      })
    )
    expect(await screen.findByRole('alert')).toHaveTextContent(
      'Unable to read the excerpt'
    )
  })
  it('validates the note and submits a review without any account action', async () => {
    const post = vi
      .spyOn(api, 'post')
      .mockResolvedValue({ data: { success: true } })
    mount()
    await userEvent.click(
      screen.getByRole('button', { name: 'Review safety event' })
    )
    await userEvent.click(
      screen.getByRole('button', { name: 'Save review only' })
    )
    expect(await screen.findByRole('alert')).toHaveTextContent(
      'Enter a review note'
    )
    expect(post).not.toHaveBeenCalled()
    await userEvent.selectOptions(
      screen.getByLabelText('Review decision'),
      'confirmed'
    )
    await userEvent.type(
      screen.getByLabelText('Review note'),
      'Verified upstream category'
    )
    await userEvent.click(
      screen.getByRole('button', { name: 'Save review only' })
    )
    await waitFor(() =>
      expect(post).toHaveBeenCalledWith('/api/abuse/events/9/review', {
        decision: 'confirmed',
        note: 'Verified upstream category',
        version: 0,
      })
    )
    expect(post).toHaveBeenCalledTimes(1)
  })
  it('keeps the note and shows a failure when a concurrent review wins', async () => {
    vi.spyOn(api, 'post').mockRejectedValue(new Error('conflict'))
    mount()
    await userEvent.click(
      screen.getByRole('button', { name: 'Review safety event' })
    )
    await userEvent.type(
      screen.getByLabelText('Review note'),
      'Unclear classification'
    )
    await userEvent.click(
      screen.getByRole('button', { name: 'Save review only' })
    )
    expect(await screen.findByRole('alert')).toHaveTextContent(
      'Unable to save the review'
    )
    expect(screen.getByLabelText('Review note')).toHaveValue(
      'Unclear classification'
    )
  })
})
