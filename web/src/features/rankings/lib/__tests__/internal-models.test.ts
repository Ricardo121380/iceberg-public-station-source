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
import { describe, expect, it } from 'vitest'

import type { RankingsSnapshot } from '../../types'
import { dropInternalModels, isInternalModelName } from '../internal-models'

function makeSnapshot(): RankingsSnapshot {
  return {
    models: [
      {
        rank: 1,
        model_name: 'gpt-6-astra',
        vendor: 'OpenAI',
        category: 'all',
        total_tokens: 9000,
        share: 0.6,
        growth_pct: 12.5,
      },
      {
        rank: 2,
        model_name: 'iceberg-abuse-canary-0ea5fdbf',
        vendor: 'unknown',
        category: 'all',
        total_tokens: 1,
        share: 0,
        growth_pct: 100,
      },
      {
        rank: 3,
        model_name: 'claude-opus',
        vendor: 'Anthropic',
        category: 'all',
        total_tokens: 5000,
        share: 0.4,
        growth_pct: -3.2,
      },
    ],
    vendors: [],
    top_movers: [
      {
        model_name: 'review-canary-f0c1b73c',
        vendor: 'unknown',
        rank_delta: 9,
        current_rank: 2,
        growth_pct: 100,
      },
      {
        model_name: 'gpt-6-astra',
        vendor: 'OpenAI',
        rank_delta: 2,
        current_rank: 1,
        growth_pct: 12.5,
      },
    ],
    top_droppers: [
      {
        model_name: 'claude-opus',
        vendor: 'Anthropic',
        rank_delta: -1,
        current_rank: 3,
        growth_pct: -3.2,
      },
    ],
    models_history: {
      points: [
        {
          ts: '2026-09-01T00:00:00Z',
          label: 'Sep 1',
          model: 'gpt-6-astra',
          vendor: 'OpenAI',
          tokens: 100,
        },
        {
          ts: '2026-09-01T00:00:00Z',
          label: 'Sep 1',
          model: 'iceberg-abuse-canary-0ea5fdbf',
          vendor: 'unknown',
          tokens: 1,
        },
      ],
      models: [
        { name: 'gpt-6-astra', vendor: 'OpenAI', total: 100 },
        { name: 'iceberg-abuse-canary-0ea5fdbf', vendor: 'unknown', total: 1 },
      ],
      buckets: 1,
    },
    vendor_share_history: { points: [], vendors: [], buckets: 0 },
  }
}

describe('isInternalModelName', () => {
  it.each([
    ['iceberg-abuse-canary-0ea5fdbf', true],
    ['review-canary-f0c1b73c', true],
    ['gpt-6-astra', false],
    ['claude-opus', false],
  ])('classifies %s as internal=%s', (name, expected) => {
    expect(isInternalModelName(name)).toBe(expected)
  })
})

describe('dropInternalModels', () => {
  it('removes canary rows from models, movers, and history, and re-ranks densely', () => {
    const result = dropInternalModels(makeSnapshot())

    expect(result.models.map((row) => row.model_name)).toEqual([
      'gpt-6-astra',
      'claude-opus',
    ])
    // No rank gaps after removal: public rows re-rank to 1..n.
    expect(result.models.map((row) => row.rank)).toEqual([1, 2])
    // Fields other than the rank are preserved.
    expect(result.models[1].total_tokens).toBe(5000)

    expect(result.top_movers.map((row) => row.model_name)).toEqual([
      'gpt-6-astra',
    ])
    expect(result.top_droppers).toHaveLength(1)

    expect(result.models_history.points.map((point) => point.model)).toEqual([
      'gpt-6-astra',
    ])
    expect(result.models_history.models.map((entry) => entry.name)).toEqual([
      'gpt-6-astra',
    ])
    expect(result.models_history.buckets).toBe(1)
  })
})
