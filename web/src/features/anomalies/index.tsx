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
import { useTranslation } from 'react-i18next'

import { SectionPageLayout } from '@/components/layout'
import { Button } from '@/components/ui/button'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { ROLE } from '@/lib/roles'
import { useAuthStore } from '@/stores/auth-store'

import { getAudit, getSettings } from './api'
import { AnomalyEvents } from './events'
import { dateLabel } from './format'
import { AnomalySettings } from './settings'

function AuditHistory() {
  const { t } = useTranslation()
  const query = useQuery({
    queryKey: ['anomalies', 'audit'],
    queryFn: getAudit,
  })
  return (
    <div className='space-y-4'>
      <h2 className='text-lg font-semibold'>{t('Handling history')}</h2>
      <p className='text-muted-foreground text-sm'>
        {t(
          'Latest 200 actions within seven days. Acknowledgements do not suspend accounts.'
        )}
      </p>
      {query.isPending && <p role='status'>{t('Loading...')}</p>}
      {query.isError && (
        <div role='alert'>
          {t('Unable to load anomaly data. Check Redis and retry.')}{' '}
          <Button variant='outline' onClick={() => void query.refetch()}>
            {t('Retry')}
          </Button>
        </div>
      )}
      {query.data?.length === 0 && (
        <p className='text-muted-foreground'>{t('No handling records yet')}</p>
      )}
      <ol className='divide-border divide-y'>
        {query.data?.map((audit) => (
          <li key={audit.id} className='space-y-2 py-4 text-sm'>
            <p className='font-medium'>
              {t(
                (
                  {
                    acknowledged: 'Alert acknowledged',
                    settings_updated: 'Alert rules updated',
                    tg_pause: 'Telegram temporary suspension',
                    tg_release: 'Telegram suspension released',
                  } as Record<string, string>
                )[audit.action] || 'Handling history'
              )}
            </p>
            <p className='text-muted-foreground'>
              {t('Operator')}{' '}
              {audit.telegram_operator_id
                ? `TG ${audit.telegram_operator_id}`
                : audit.operator_id}{' '}
              · {dateLabel(audit.created_at)}
            </p>
            {audit.event_id && (
              <p className='text-muted-foreground break-all'>
                {t('Event ID')}: {audit.event_id}
              </p>
            )}
            {audit.policy && (
              <p>
                {audit.policy.enabled
                  ? t('Collection enabled')
                  : t('Collection disabled')}{' '}
                · {audit.policy.window_minutes} {t('minutes')} ·{' '}
                {t('Schema / rate limit / upstream')}:{' '}
                {audit.policy.schema_threshold} / {audit.policy.rate_threshold}{' '}
                / {audit.policy.upstream_threshold}
              </p>
            )}
          </li>
        ))}
      </ol>
    </div>
  )
}
function ConsoleContent() {
  const { t } = useTranslation()
  const role = useAuthStore((state) => state.auth.user?.role)
  const settings = useQuery({
    queryKey: ['anomalies', 'settings'],
    queryFn: getSettings,
  })
  return (
    <SectionPageLayout>
      <SectionPageLayout.Title>{t('Anomaly alerts')}</SectionPageLayout.Title>
      <SectionPageLayout.Content>
        <div className='space-y-6'>
          <p className='text-muted-foreground max-w-3xl text-sm leading-relaxed'>
            {t(
              'Review repeated request failures and tune reminders. Events are grouped by user, channel, model and failure type in fixed time windows; no prompts or keys are stored.'
            )}
          </p>
          <div className='bg-muted rounded-lg px-4 py-3 text-sm' role='status'>
            {t('Observation only. Automatic suspension is not enabled.')}
            {settings.data && (
              <>
                {' '}
                {settings.data.enabled
                  ? t('Collection enabled')
                  : t('Collection disabled')}{' '}
                ·{' '}
                {t('Window: {{count}} minutes', {
                  count: settings.data.window_minutes,
                })}
              </>
            )}
          </div>
          <details className='text-muted-foreground text-xs'>
            <summary className='cursor-pointer text-sm'>
              {t('Collection coverage and retention')}
            </summary>
            <p className='mt-2 max-w-3xl leading-relaxed'>
              {t(
                'Covers failed POST requests on /v1 and /pg. Retains up to 2,000 groups for seven days in Redis; clearing Redis removes records and resets rules. Unauthenticated requests and HTTP 200 stream interruptions are excluded.'
              )}
            </p>
          </details>
          <Tabs defaultValue='events'>
            <TabsList className='mb-5 max-w-full'>
              <TabsTrigger value='events'>{t('Anomaly events')}</TabsTrigger>
              <TabsTrigger value='rules'>{t('Alert rules')}</TabsTrigger>
              <TabsTrigger value='history'>{t('Handling history')}</TabsTrigger>
            </TabsList>
            <TabsContent value='events'>
              <AnomalyEvents />
            </TabsContent>
            <TabsContent value='rules'>
              {settings.isPending && <p role='status'>{t('Loading...')}</p>}
              {settings.isError && (
                <div role='alert'>
                  {t('Unable to load anomaly data. Check Redis and retry.')}{' '}
                  <Button
                    variant='outline'
                    onClick={() => void settings.refetch()}
                  >
                    {t('Retry')}
                  </Button>
                </div>
              )}
              {settings.data && (
                <AnomalySettings
                  key={settings.data.version}
                  policy={settings.data}
                  canEdit={role === ROLE.SUPER_ADMIN}
                />
              )}
            </TabsContent>
            <TabsContent value='history'>
              <AuditHistory />
            </TabsContent>
          </Tabs>
        </div>
      </SectionPageLayout.Content>
    </SectionPageLayout>
  )
}
export function AnomalyConsole() {
  const { t } = useTranslation()
  const role = useAuthStore((state) => state.auth.user?.role ?? 0)
  if (role < ROLE.ADMIN) {
    return <p role='alert'>{t('Administrator access is required')}</p>
  }
  return <ConsoleContent />
}
