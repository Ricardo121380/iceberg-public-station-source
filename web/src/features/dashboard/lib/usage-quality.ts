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
import dayjs from 'dayjs'

import type { TimeGranularity } from '@/lib/time'

import type { UsageQualityBucket } from '../types'

export function summarizeUsageQuality(rows: UsageQualityBucket[]) {
  let ttftSum = 0
  let ttftCount = 0
  let cacheRead = 0
  let input = 0
  let cacheCount = 0
  for (const row of rows) {
    ttftSum += row.ttft_sum_ms
    ttftCount += row.ttft_count
    cacheRead += row.cache_read_tokens
    input += row.input_tokens
    cacheCount += row.cache_count
  }
  return {
    avgTtftSeconds: ttftCount > 0 ? ttftSum / ttftCount / 1000 : null,
    cacheHitPercent: input > 0 ? (cacheRead / input) * 100 : null,
    ttftCount,
    cacheCount,
  }
}

export function buildUsageQualitySeries(
  rows: UsageQualityBucket[],
  granularity: TimeGranularity
) {
  const buckets = new Map<number, UsageQualityBucket[]>()
  for (const row of rows) {
    const ts = dayjs.unix(row.created_at).startOf(granularity).unix()
    const bucket = buckets.get(ts) ?? []
    bucket.push(row)
    buckets.set(ts, bucket)
  }
  return [...buckets]
    .sort(([a], [b]) => a - b)
    .map(([ts, items]) => ({
      ts,
      ...summarizeUsageQuality(items),
    }))
}
