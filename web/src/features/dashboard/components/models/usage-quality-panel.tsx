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
import { useQuery } from '@tanstack/react-query'
import { useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { CartesianGrid, Line, LineChart, XAxis, YAxis } from 'recharts'

import { Button } from '@/components/ui/button'
import {
  ChartContainer,
  ChartTooltip,
  ChartTooltipContent,
} from '@/components/ui/chart'
import { Skeleton } from '@/components/ui/skeleton'
import { getUsageQuality } from '@/features/dashboard/api'
import { buildQueryParams, getDefaultDays } from '@/features/dashboard/lib'
import {
  buildUsageQualitySeries,
  summarizeUsageQuality,
} from '@/features/dashboard/lib/usage-quality'
import type { DashboardFilters } from '@/features/dashboard/types'
import { computeTimeRange, formatChartTime } from '@/lib/time'
import { useAuthStore } from '@/stores/auth-store'

export function UsageQualityPanel(props: { filters?: DashboardFilters }) {
  const { t } = useTranslation()
  const userId = useAuthStore((state) => state.auth.user?.id)
  const role = useAuthStore((state) => state.auth.user?.role)
  const isAdmin = (role ?? 0) >= 10
  const [selectedModel, setSelectedModel] = useState('')
  const params = useMemo(
    () =>
      buildQueryParams(
        computeTimeRange(
          getDefaultDays(props.filters?.time_granularity),
          props.filters?.start_timestamp,
          props.filters?.end_timestamp
        ),
        props.filters
      ),
    [props.filters]
  )
  const query = useQuery({
    queryKey: ['usage-quality', userId, isAdmin, params],
    queryFn: ({ signal }) => getUsageQuality(params, isAdmin, signal),
    staleTime: 60_000,
    retry: false,
  })
  const models = useMemo(
    () => [...new Set(query.data?.map((row) => row.model_name) ?? [])].sort(),
    [query.data]
  )
  const activeModel = models.includes(selectedModel) ? selectedModel : ''
  const rows = useMemo(
    () =>
      (query.data ?? []).filter(
        (row) => !activeModel || row.model_name === activeModel
      ),
    [query.data, activeModel]
  )
  const summary = useMemo(() => summarizeUsageQuality(rows), [rows])
  const granularity = props.filters?.time_granularity ?? 'day'
  const series = useMemo(
    () => buildUsageQualitySeries(rows, granularity),
    [rows, granularity]
  )
  const metrics = [
    {
      key: 'avgTtftSeconds' as const,
      label: t('Average TTFT'),
      value: summary.avgTtftSeconds,
      unit: 's',
      count: summary.ttftCount,
      description: t(
        'Time to first response for valid streaming requests, weighted by request count.'
      ),
    },
    {
      key: 'cacheHitPercent' as const,
      label: t('Cache hit rate'),
      value: summary.cacheHitPercent,
      unit: '%',
      count: summary.cacheCount,
      description: t(
        'Cached input tokens divided by total input tokens. Cache writes are not hits.'
      ),
    },
  ]

  return (
    <section
      className='min-w-0 overflow-hidden rounded-lg border'
      aria-label={t('Response and cache statistics')}
    >
      <div className='flex flex-wrap items-center justify-between gap-3 border-b px-4 py-3 sm:px-5'>
        <h2 className='text-sm font-semibold'>
          {t('Response and cache statistics')}
        </h2>
        <label className='flex max-w-full min-w-0 items-center gap-2 text-xs'>
          {t('Model')}
          <select
            aria-label={t('Statistics model')}
            className='bg-background focus-visible:ring-ring max-w-52 min-w-0 rounded-md border px-2 py-1.5 focus-visible:ring-2'
            value={activeModel}
            disabled={query.isPending || query.isError || models.length === 0}
            onChange={(event) => setSelectedModel(event.target.value)}
          >
            <option value=''>{t('All models')}</option>
            {models.map((model) => (
              <option key={model} value={model}>
                {model}
              </option>
            ))}
          </select>
        </label>
      </div>
      {query.isError ? (
        <div
          role='alert'
          className='flex flex-wrap items-center gap-3 px-4 py-5 text-sm'
        >
          <span>{t('Failed to load response and cache statistics')}</span>
          <Button
            variant='outline'
            size='sm'
            onClick={() => void query.refetch()}
          >
            {t('Retry')}
          </Button>
        </div>
      ) : (
        <div className='grid min-w-0 grid-cols-1 divide-y md:grid-cols-2 md:divide-x md:divide-y-0'>
          {metrics.map((metric) => (
            <div key={metric.key} className='min-w-0 p-4 sm:p-5'>
              <h3 className='text-muted-foreground text-xs font-medium'>
                {metric.label}
              </h3>
              {query.isPending ? (
                <Skeleton className='my-2 h-8 w-28' />
              ) : (
                <p
                  className='my-2 font-mono text-2xl font-semibold tabular-nums'
                  aria-label={metric.label}
                >
                  {metric.value === null
                    ? t('No data available')
                    : `${metric.value.toFixed(2)} ${metric.unit}`}
                </p>
              )}
              <p className='text-muted-foreground text-xs leading-relaxed'>
                {metric.description}
              </p>
              {!query.isPending && (
                <p className='text-muted-foreground mt-1 text-xs'>
                  {t('Valid samples: {{count}}', { count: metric.count })}
                </p>
              )}
              {!query.isPending && metric.value !== null && (
                <ChartContainer
                  className='mt-4 h-44 w-full'
                  aria-label={t('{{metric}} trend', { metric: metric.label })}
                  config={{
                    [metric.key]: {
                      label: metric.label,
                      color: 'var(--chart-1)',
                    },
                  }}
                >
                  <LineChart data={series} accessibilityLayer>
                    <CartesianGrid vertical={false} />
                    <XAxis
                      dataKey='ts'
                      tickFormatter={(ts: number) =>
                        formatChartTime(ts, granularity)
                      }
                      minTickGap={28}
                      tickLine={false}
                      axisLine={false}
                    />
                    <YAxis
                      width={40}
                      domain={metric.unit === '%' ? [0, 100] : [0, 'auto']}
                      tickLine={false}
                      axisLine={false}
                    />
                    <ChartTooltip
                      content={
                        <ChartTooltipContent
                          labelFormatter={(ts) =>
                            formatChartTime(Number(ts), granularity)
                          }
                          formatter={(value) =>
                            `${Number(value).toFixed(2)} ${metric.unit}`
                          }
                        />
                      }
                    />
                    <Line
                      type='linear'
                      dataKey={metric.key}
                      stroke='var(--chart-1)'
                      strokeWidth={2}
                      dot={{ r: 2 }}
                      connectNulls={false}
                      isAnimationActive={false}
                    />
                  </LineChart>
                </ChartContainer>
              )}
            </div>
          ))}
        </div>
      )}
      <p className='text-muted-foreground border-t px-4 py-3 text-xs leading-relaxed sm:px-5'>
        {t(
          'Based on retained consumption logs in the selected period (up to 30 days). Missing fields are excluded; coverage may differ from usage totals.'
        )}
      </p>
    </section>
  )
}
