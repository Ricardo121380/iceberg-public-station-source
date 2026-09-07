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
import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import { api } from '@/lib/api'

import { UsageQualityPanel } from '../components/models/usage-quality-panel'
import {
  buildUsageQualitySeries,
  summarizeUsageQuality,
} from '../lib/usage-quality'
import type { UsageQualityBucket } from '../types'

const rows: UsageQualityBucket[] = [
  {
    model_name: 'a',
    created_at: 172800,
    ttft_sum_ms: 1000,
    ttft_count: 1,
    cache_read_tokens: 80,
    input_tokens: 100,
    cache_count: 1,
  },
  {
    model_name: 'b',
    created_at: 172800,
    ttft_sum_ms: 9000,
    ttft_count: 3,
    cache_read_tokens: 90,
    input_tokens: 900,
    cache_count: 3,
  },
]

function renderPanel() {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return render(
    <QueryClientProvider client={client}>
      <UsageQualityPanel
        filters={{
          start_timestamp: new Date(172800000),
          end_timestamp: new Date(259200000),
        }}
      />
    </QueryClientProvider>
  )
}

beforeEach(() => {
  vi.spyOn(api, 'get').mockResolvedValue({
    data: { success: true, data: rows },
  })
})

describe('usage quality aggregation', () => {
  it('weights TTFT by samples and cache hits by input tokens', () => {
    expect(summarizeUsageQuality(rows)).toEqual({
      avgTtftSeconds: 2.5,
      cacheHitPercent: 17,
      ttftCount: 4,
      cacheCount: 4,
    })
    const series = buildUsageQualitySeries(rows, 'day')
    expect(series).toHaveLength(1)
    expect(series[0].avgTtftSeconds).toBe(2.5)
    expect(series[0].cacheHitPercent).toBe(17)
  })
  it('distinguishes absent metrics from a measured zero hit rate', () => {
    expect(summarizeUsageQuality([]).cacheHitPercent).toBeNull()
    expect(summarizeUsageQuality([]).avgTtftSeconds).toBeNull()
    expect(
      summarizeUsageQuality([{ ...rows[0], cache_read_tokens: 0 }])
        .cacheHitPercent
    ).toBe(0)
  })
})

it('renders weighted values and updates both metrics when a model is selected', async () => {
  renderPanel()
  expect(await screen.findByText('2.50 s')).toBeVisible()
  expect(screen.getByText('17.00 %')).toBeVisible()
  fireEvent.change(screen.getByRole('combobox', { name: 'Statistics model' }), {
    target: { value: 'a' },
  })
  expect(screen.getByText('1.00 s')).toBeVisible()
  expect(screen.getByText('80.00 %')).toBeVisible()
  expect(api.get).toHaveBeenCalledWith(
    '/api/data/quality/self',
    expect.objectContaining({
      params: expect.objectContaining({
        start_timestamp: 172800,
        end_timestamp: 259200,
      }),
    })
  )
})

it('shows no data and disables selection when the period has no eligible logs', async () => {
  vi.mocked(api.get).mockResolvedValue({ data: { success: true, data: [] } })
  renderPanel()
  expect(await screen.findAllByText('No data available')).toHaveLength(2)
  expect(screen.getByRole('combobox')).toBeDisabled()
  expect(screen.queryByText('0.00 %')).not.toBeInTheDocument()
})

it('shows a recoverable error instead of zero on a failed API response', async () => {
  vi.mocked(api.get).mockResolvedValueOnce({
    data: { success: false, message: 'query failed' },
  })
  renderPanel()
  expect(await screen.findByRole('alert')).toHaveTextContent(
    'Failed to load response and cache statistics'
  )
  expect(screen.queryByText('0.00 %')).not.toBeInTheDocument()
  fireEvent.click(screen.getByRole('button', { name: 'Retry' }))
  expect(await screen.findByText('2.50 s')).toBeVisible()
})

it('keeps values hidden and model selection disabled until the request resolves', async () => {
  let finish: (value: unknown) => void = () => undefined
  vi.mocked(api.get).mockImplementationOnce(
    () =>
      new Promise((resolve) => {
        finish = resolve
      })
  )
  renderPanel()
  expect(screen.getByRole('combobox')).toBeDisabled()
  expect(screen.queryByText('No data available')).not.toBeInTheDocument()
  finish({ data: { success: true, data: rows } })
  await waitFor(() => expect(screen.getByRole('combobox')).toBeEnabled())
})

it('reloads the selected period and falls back to all models when a model disappears', async () => {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  const start = new Date(172800000)
  const renderPeriod = (end: Date) => (
    <QueryClientProvider client={client}>
      <UsageQualityPanel
        filters={{ start_timestamp: start, end_timestamp: end }}
      />
    </QueryClientProvider>
  )
  const view = render(renderPeriod(new Date(259200000)))
  await screen.findByText('2.50 s')
  fireEvent.change(screen.getByRole('combobox'), { target: { value: 'a' } })
  expect(screen.getByText('1.00 s')).toBeVisible()
  vi.mocked(api.get).mockResolvedValueOnce({
    data: { success: true, data: [rows[1]] },
  })
  view.rerender(renderPeriod(new Date(345600000)))
  expect(await screen.findByText('3.00 s')).toBeVisible()
  expect(screen.getByText('10.00 %')).toBeVisible()
  expect(screen.getByRole('combobox')).toHaveValue('')
  expect(api.get).toHaveBeenLastCalledWith(
    '/api/data/quality/self',
    expect.objectContaining({
      params: expect.objectContaining({ end_timestamp: 345600 }),
    })
  )
})
