import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
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
import { useState } from 'react'
import { useTranslation } from 'react-i18next'

import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'

import { acknowledge, getEvents, type EventFilters } from './api'
import { dateLabel } from './format'

const anomalyLabels: Record<string, string> = {
  invalid_schema: 'Invalid output schema',
  rate_limit: 'Local rate limit',
  upstream: 'Upstream failure',
  quota: 'Insufficient quota',
  unsupported_model: 'Unsupported model',
  invalid_request: 'Invalid request',
}
const emptyFilters: EventFilters = { user_id: '', kind: '', state: 'pending' }
export function AnomalyEvents() {
  const { t } = useTranslation()
  const client = useQueryClient()
  const [page, setPage] = useState(1)
  const [draft, setDraft] = useState(emptyFilters)
  const [filters, setFilters] = useState(emptyFilters)
  const query = useQuery({
    queryKey: ['anomalies', 'events', page, filters],
    queryFn: () => getEvents(page, filters),
    refetchInterval: 30000,
  })
  const ack = useMutation({
    mutationFn: acknowledge,
    onSuccess: () => {
      void client.invalidateQueries({ queryKey: ['anomalies'] })
    },
  })
  return (
    <div className='space-y-5'>
      <form
        className='flex flex-wrap items-end gap-3'
        onSubmit={(e) => {
          e.preventDefault()
          setPage(1)
          setFilters({ ...draft })
        }}
      >
        <div className='space-y-2'>
          <label htmlFor='anomaly-user' className='block text-sm'>
            {t('User ID')}
          </label>
          <Input
            id='anomaly-user'
            type='number'
            min={1}
            step={1}
            className='w-36'
            value={draft.user_id}
            onChange={(e) => setDraft({ ...draft, user_id: e.target.value })}
            placeholder={t('All users')}
          />
        </div>
        <div className='space-y-2'>
          <label htmlFor='anomaly-kind' className='block text-sm'>
            {t('Failure type')}
          </label>
          <select
            id='anomaly-kind'
            className='border-input bg-background flex h-9 max-w-full rounded-md border px-3 text-sm'
            value={draft.kind}
            onChange={(e) => setDraft({ ...draft, kind: e.target.value })}
          >
            <option value=''>{t('All types')}</option>
            {Object.entries(anomalyLabels).map(([value, label]) => (
              <option key={value} value={value}>
                {t(label)}
              </option>
            ))}
          </select>
        </div>
        <div className='space-y-2'>
          <label htmlFor='anomaly-state' className='block text-sm'>
            {t('Review status')}
          </label>
          <select
            id='anomaly-state'
            className='border-input bg-background flex h-9 rounded-md border px-3 text-sm'
            value={draft.state}
            onChange={(e) => setDraft({ ...draft, state: e.target.value })}
          >
            <option value='pending'>{t('Needs attention')}</option>
            <option value=''>{t('All events')}</option>
            <option value='acknowledged'>{t('Acknowledged')}</option>
          </select>
        </div>
        <Button type='submit' variant='outline'>
          {t('Apply filters')}
        </Button>
        <Button
          type='button'
          variant='ghost'
          disabled={query.isFetching}
          onClick={() => void query.refetch()}
        >
          {t('Refresh')}
        </Button>
      </form>
      <p className='text-muted-foreground text-sm'>
        {t(
          'Acknowledging an alert records your review. New failures can make it pending again; no account access is changed.'
        )}
      </p>
      {query.isPending && <p role='status'>{t('Loading anomaly events')}</p>}
      {query.isError && (
        <div role='alert' className='space-y-2'>
          <p>{t('Unable to load anomaly data. Check Redis and retry.')}</p>
          <Button variant='outline' onClick={() => void query.refetch()}>
            {t('Retry')}
          </Button>
        </div>
      )}
      {ack.isError && (
        <p role='alert' className='text-destructive text-sm'>
          {t(
            'The event may have changed. Refresh and acknowledge the latest count.'
          )}
        </p>
      )}
      {query.data && (
        <>
          <p role='status' className='text-sm'>
            {t('Pending alerts: {{count}}', { count: query.data.pending })} ·{' '}
            {t('Matching groups: {{count}}', { count: query.data.total })}
          </p>
          {query.data.items.length === 0 && (
            <div className='border-border rounded-lg border border-dashed px-5 py-10 text-center'>
              <h2 className='font-medium'>{t('No matching anomalies')}</h2>
              <p className='text-muted-foreground mt-2 text-sm'>
                {t(
                  'Try All events or adjust the filters. Collection starts after this feature is enabled; historical logs are not imported.'
                )}
              </p>
            </div>
          )}
          <div className='divide-border divide-y'>
            {query.data.items.map((event) => {
              const pending =
                event.alert && event.count > event.acknowledged_count
              return (
                <article
                  key={event.id}
                  className='space-y-3 py-5 first:pt-0'
                  aria-label={`${t(anomalyLabels[event.kind] || 'Invalid request')} · ${t('User ID')} ${event.user_id}`}
                >
                  <div className='flex flex-wrap items-start justify-between gap-3'>
                    <div className='min-w-0 space-y-2'>
                      <div className='flex flex-wrap items-center gap-2'>
                        <h3 className='font-medium'>
                          {t(anomalyLabels[event.kind] || 'Invalid request')}
                        </h3>
                        <Badge variant={pending ? 'destructive' : 'secondary'}>
                          {pending ? t('Needs attention') : t('Observed')}
                        </Badge>
                      </div>
                      <p className='text-muted-foreground text-sm break-all'>
                        {t('User ID')} {event.user_id} ·{' '}
                        {event.model || t('Model unavailable')} · {t('Channel')}{' '}
                        {event.channel_id || '—'}
                      </p>
                    </div>
                    <Button
                      variant='outline'
                      size='sm'
                      disabled={
                        ack.isPending || event.count <= event.acknowledged_count
                      }
                      onClick={() => ack.mutate(event)}
                    >
                      {event.count <= event.acknowledged_count
                        ? t('Acknowledged')
                        : t('Acknowledge')}
                    </Button>
                  </div>
                  <dl className='grid grid-cols-2 gap-x-6 gap-y-3 text-sm lg:grid-cols-4'>
                    <div>
                      <dt className='text-muted-foreground'>
                        {t('Requests in window')}
                      </dt>
                      <dd className='mt-1 font-medium tabular-nums'>
                        {event.count}
                      </dd>
                    </div>
                    <div>
                      <dt className='text-muted-foreground'>
                        {t('Alert threshold')}
                      </dt>
                      <dd className='mt-1 tabular-nums'>
                        {event.threshold || t('Observation only')}
                      </dd>
                    </div>
                    <div>
                      <dt className='text-muted-foreground'>
                        {t('First seen')}
                      </dt>
                      <dd className='mt-1'>{dateLabel(event.first_seen)}</dd>
                    </div>
                    <div>
                      <dt className='text-muted-foreground'>
                        {t('Last seen')}
                      </dt>
                      <dd className='mt-1'>{dateLabel(event.last_seen)}</dd>
                    </div>
                  </dl>
                  {event.acknowledged_at > 0 && (
                    <p className='text-muted-foreground text-xs'>
                      {t(
                        'Acknowledged by {{user}} at {{time}} ({{count}} requests)',
                        {
                          user: event.acknowledged_by,
                          time: dateLabel(event.acknowledged_at),
                          count: event.acknowledged_count,
                        }
                      )}
                    </p>
                  )}
                </article>
              )
            })}
          </div>
          <div className='flex items-center gap-3 border-t pt-4'>
            <Button
              variant='outline'
              disabled={page === 1}
              onClick={() => setPage(page - 1)}
            >
              {t('Previous')}
            </Button>
            <span className='text-sm'>
              {page} / {Math.max(1, Math.ceil(query.data.total / 20))}
            </span>
            <Button
              variant='outline'
              disabled={page * 20 >= query.data.total}
              onClick={() => setPage(page + 1)}
            >
              {t('Next')}
            </Button>
          </div>
        </>
      )}
    </div>
  )
}
