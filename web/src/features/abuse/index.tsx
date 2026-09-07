import { zodResolver } from '@hookform/resolvers/zod'
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
import { useForm } from 'react-hook-form'
import { useTranslation } from 'react-i18next'
import { z } from 'zod'

import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { ROLE } from '@/lib/roles'
import { useAuthStore } from '@/stores/auth-store'

import {
  getSettings,
  saveSettings,
  getEvents,
  getSuspensions,
  unfreezeUser,
  type AbuseSettings,
  type EventFilters,
} from './api'
import { SafetyReviewDialog } from './review-dialog'

const settingsSchema = z.object({
  mode: z.enum(['off', 'observe', 'enforce']),
  limit_10m: z.number().int().min(1).max(1000),
  limit_24h: z.number().int().min(1).max(10000),
  freeze_minutes: z.number().int().min(1).max(1440),
  enabled_rules: z.array(z.string()),
})
type SettingsForm = z.infer<typeof settingsSchema>
const emptyFilters: EventFilters = {
  user_id: '',
  category: '',
  rule_id: '',
  action: '',
  start: '',
  end: '',
}
const selectStyle =
  'border-input bg-background h-9 w-full rounded-md border px-3 text-sm'

function SafetySettingsForm(props: { settings: AbuseSettings }) {
  const { t } = useTranslation()
  const client = useQueryClient()
  const form = useForm<SettingsForm>({
    resolver: zodResolver(settingsSchema),
    defaultValues: props.settings,
  })
  const [confirmed, setConfirmed] = useState(false)
  const save = useMutation({
    mutationFn: saveSettings,
    onSuccess: () => {
      void client.invalidateQueries({ queryKey: ['abuse'] })
      setConfirmed(false)
    },
  })
  return (
    <form
      className='space-y-4'
      onSubmit={form.handleSubmit((values) => save.mutate(values))}
    >
      <p className='text-muted-foreground max-w-prose text-sm'>
        {t(
          'Observe for seven days before manually enabling enforcement. Unclassified refusals never suspend accounts.'
        )}
      </p>
      <div className='grid gap-4 sm:grid-cols-2'>
        <label className='space-y-2 text-sm'>
          {t('Safety mode')}
          <select
            className={selectStyle}
            {...form.register('mode')}
            onChange={(event) => {
              void form.setValue(
                'mode',
                event.target.value as SettingsForm['mode']
              )
              setConfirmed(false)
            }}
          >
            <option value='off'>{t('Off')}</option>
            <option value='observe'>{t('Observe only')}</option>
            <option value='enforce'>{t('Enforce verified rules')}</option>
          </select>
        </label>
        <label className='space-y-2 text-sm'>
          {t('Suspend for minutes')}
          <Input
            type='number'
            min={1}
            max={1440}
            {...form.register('freeze_minutes', { valueAsNumber: true })}
          />
        </label>
        <label className='space-y-2 text-sm'>
          {t('Violations in ten minutes')}
          <Input
            type='number'
            min={1}
            max={1000}
            {...form.register('limit_10m', { valueAsNumber: true })}
          />
        </label>
        <label className='space-y-2 text-sm'>
          {t('Violations in twenty-four hours')}
          <Input
            type='number'
            min={1}
            max={10000}
            {...form.register('limit_24h', { valueAsNumber: true })}
          />
        </label>
      </div>
      <fieldset className='space-y-2'>
        <legend className='mb-2 text-sm font-medium'>
          {t('Verified input rules')}
        </legend>
        {props.settings.rules.length === 0 && (
          <p className='text-muted-foreground text-sm'>
            {t(
              'No verified category evidence is available. This channel remains observation only.'
            )}
          </p>
        )}
        {props.settings.rules.map((rule) => (
          <label key={rule.id} className='flex items-center gap-2 text-sm'>
            <input
              type='checkbox'
              value={rule.id}
              disabled={!rule.verified}
              {...form.register('enabled_rules')}
            />
            {rule.id}
          </label>
        ))}
      </fieldset>
      {form.watch('mode') === 'enforce' && (
        <label className='flex items-start gap-2 text-sm'>
          <input
            type='checkbox'
            checked={confirmed}
            onChange={(event) => setConfirmed(event.target.checked)}
          />
          {t(
            'I reviewed the evidence and understand that verified rules can suspend model access.'
          )}
        </label>
      )}
      {Object.keys(form.formState.errors).length > 0 && (
        <p role='alert' className='text-destructive text-sm'>
          {t('Enter valid positive limits within the displayed ranges.')}
        </p>
      )}
      {save.isError && (
        <p role='alert' className='text-destructive text-sm'>
          {t(
            'Safety settings could not be saved. Retry after checking the service.'
          )}
        </p>
      )}
      {save.isSuccess && (
        <p role='status' className='text-sm'>
          {t('Safety settings saved')}
        </p>
      )}
      <Button
        type='submit'
        disabled={
          save.isPending || (form.watch('mode') === 'enforce' && !confirmed)
        }
      >
        {t('Save safety settings')}
      </Button>
      <p className='text-muted-foreground text-xs'>
        {t(
          'Changing mode does not release existing suspensions. Historical observation events are not punished retroactively.'
        )}
      </p>
    </form>
  )
}

function SafetyConsole() {
  const { t } = useTranslation()
  const client = useQueryClient()
  const settings = useQuery({
    queryKey: ['abuse', 'settings'],
    queryFn: getSettings,
  })
  const [page, setPage] = useState(1)
  const [userPage, setUserPage] = useState(1)
  const [filters, setFilters] = useState<EventFilters>(emptyFilters)
  const [draftFilters, setDraftFilters] = useState<EventFilters>(emptyFilters)
  const [selectedUser, setSelectedUser] = useState<number | null>(null)
  const [reason, setReason] = useState('')
  const events = useQuery({
    queryKey: ['abuse', 'events', page, filters],
    queryFn: () => getEvents(page, filters),
  })
  const users = useQuery({
    queryKey: ['abuse', 'users', userPage],
    queryFn: () => getSuspensions(userPage),
    refetchInterval: 60000,
  })
  const release = useMutation({
    mutationFn: (id: number) => unfreezeUser(id, reason.trim()),
    onSuccess: () => {
      setSelectedUser(null)
      setReason('')
      void client.invalidateQueries({ queryKey: ['abuse'] })
    },
  })
  return (
    <div className='space-y-8'>
      <section className='space-y-4' aria-label={t('Safety settings')}>
        <h2 className='text-lg font-semibold'>{t('Safety settings')}</h2>
        {settings.isPending && (
          <p role='status'>{t('Loading safety settings')}</p>
        )}
        {settings.isError && (
          <div role='alert'>
            {t('Unable to load safety settings')}{' '}
            <Button variant='outline' onClick={() => void settings.refetch()}>
              {t('Retry')}
            </Button>
          </div>
        )}
        {settings.data && (
          <SafetySettingsForm
            key={JSON.stringify(settings.data)}
            settings={settings.data}
          />
        )}
      </section>
      <section
        className='space-y-4 border-t pt-6'
        aria-label={t('Suspended accounts')}
      >
        <h2 className='text-lg font-semibold'>{t('Suspended accounts')}</h2>
        {users.isPending && (
          <p role='status'>{t('Loading suspended accounts')}</p>
        )}
        {users.isError && (
          <div role='alert'>
            {t('Unable to load suspended accounts')}{' '}
            <Button variant='outline' onClick={() => void users.refetch()}>
              {t('Retry')}
            </Button>
          </div>
        )}
        {users.data?.items.length === 0 && (
          <p className='text-muted-foreground text-sm'>
            {t('No accounts are currently suspended.')}
          </p>
        )}
        {users.data?.items.map((user) => (
          <div
            key={user.user_id}
            className='flex flex-wrap items-center justify-between gap-3 border-b pb-3 text-sm'
          >
            <span>
              {t('User ID')} {user.user_id} · {t('Resumes at')}{' '}
              {new Date(user.blocked_until * 1000).toLocaleString()}
            </span>
            <Button
              variant='outline'
              onClick={() => {
                setSelectedUser(user.user_id)
                setReason('')
                release.reset()
              }}
            >
              {t('Release suspension')}
            </Button>
          </div>
        ))}
        {selectedUser !== null && (
          <form
            className='space-y-3 rounded-md border p-4'
            onSubmit={(event) => {
              event.preventDefault()
              release.mutate(selectedUser)
            }}
          >
            <label className='block space-y-2 text-sm'>
              {t('Release reason for user')} {selectedUser}
              <Input
                required
                maxLength={300}
                value={reason}
                onChange={(event) => setReason(event.target.value)}
              />
            </label>
            {release.isError && (
              <p role='alert' className='text-destructive text-sm'>
                {t('Release failed. Refresh the account status and retry.')}
              </p>
            )}
            <div className='flex gap-2'>
              <Button
                type='submit'
                disabled={release.isPending || !reason.trim()}
              >
                {t('Confirm release')}
              </Button>
              <Button
                type='button'
                variant='outline'
                onClick={() => setSelectedUser(null)}
              >
                {t('Cancel')}
              </Button>
            </div>
          </form>
        )}
        {users.data && (
          <div className='flex items-center gap-3'>
            <Button
              variant='outline'
              disabled={userPage === 1}
              onClick={() => setUserPage((p) => p - 1)}
            >
              {t('Previous accounts')}
            </Button>
            <span>{userPage}</span>
            <Button
              variant='outline'
              disabled={userPage * 20 >= users.data.total}
              onClick={() => setUserPage((p) => p + 1)}
            >
              {t('Next accounts')}
            </Button>
          </div>
        )}
      </section>
      <section
        className='space-y-4 border-t pt-6'
        aria-label={t('Safety events')}
      >
        <h2 className='text-lg font-semibold'>{t('Safety events')}</h2>
        {settings.data?.evidence_capture_ready === false && (
          <p role='status' className='text-muted-foreground text-sm'>
            {t(
              'Excerpt encryption is not configured. Events will be recorded without input excerpts.'
            )}
          </p>
        )}
        <p className='text-muted-foreground text-sm'>
          {t(
            'Event metadata is retained for thirty days. Redacted input excerpts are encrypted for seven days and require an audited Root view.'
          )}
        </p>
        <form
          className='grid items-end gap-3 sm:grid-cols-3'
          onSubmit={(event) => {
            event.preventDefault()
            setFilters(draftFilters)
            setPage(1)
          }}
        >
          {(
            [
              'user_id',
              'category',
              'rule_id',
              'action',
              'start',
              'end',
            ] as const
          ).map((field) => (
            <label key={field} className='space-y-2 text-sm'>
              {t(`Safety filter ${field}`)}
              <Input
                type={
                  field === 'start' || field === 'end'
                    ? 'datetime-local'
                    : 'text'
                }
                value={draftFilters[field]}
                onChange={(event) =>
                  setDraftFilters({
                    ...draftFilters,
                    [field]: event.target.value,
                  })
                }
              />
            </label>
          ))}
          <Button type='submit'>{t('Filter safety events')}</Button>
        </form>
        {events.isPending && <p role='status'>{t('Loading safety events')}</p>}
        {events.isError && (
          <div role='alert'>
            {t('Unable to load safety events')}{' '}
            <Button variant='outline' onClick={() => void events.refetch()}>
              {t('Retry')}
            </Button>
          </div>
        )}
        {events.data?.items.length === 0 && (
          <p className='text-muted-foreground text-sm'>
            {t('No safety events match these filters.')}
          </p>
        )}
        <div className='space-y-3'>
          {events.data?.items.map((event) => (
            <details key={event.id} className='rounded-md border p-3 text-sm'>
              <summary className='cursor-pointer font-medium break-words'>
                {new Date(event.created_at * 1000).toLocaleString()} ·{' '}
                {t('User ID')} {event.user_id} ·{' '}
                {t(`Safety action ${event.action}`)}
              </summary>
              <dl className='mt-3 grid gap-2 break-words sm:grid-cols-2'>
                <div>
                  <dt className='text-muted-foreground'>
                    {t('Safety category')}
                  </dt>
                  <dd>{t(`Safety category ${event.category}`)}</dd>
                </div>
                <div>
                  <dt className='text-muted-foreground'>{t('Model')}</dt>
                  <dd>{event.model || '—'}</dd>
                </div>
                <div>
                  <dt className='text-muted-foreground'>{t('Safety rule')}</dt>
                  <dd>{event.rule_id || '—'}</dd>
                </div>
                <div>
                  <dt className='text-muted-foreground'>
                    {t('Safety signal')}
                  </dt>
                  <dd>{event.signal || '—'}</dd>
                </div>
                <div>
                  <dt className='text-muted-foreground'>{t('Request ID')}</dt>
                  <dd>{event.request_id}</dd>
                </div>
                <div>
                  <dt className='text-muted-foreground'>
                    {t('Ten minutes / twenty-four hours')}
                  </dt>
                  <dd>
                    {event.count_10m} / {event.count_24h}
                  </dd>
                </div>
                <div className='sm:col-span-2'>
                  <dt className='text-muted-foreground'>
                    {t('Safety summary')}
                  </dt>
                  <dd>{t(event.summary)}</dd>
                </div>
              </dl>
              <SafetyReviewDialog event={event} />
            </details>
          ))}
        </div>
        {events.data && (
          <div className='flex items-center gap-3'>
            <Button
              variant='outline'
              disabled={page === 1}
              onClick={() => setPage((p) => p - 1)}
            >
              {t('Previous events')}
            </Button>
            <span>{page}</span>
            <Button
              variant='outline'
              disabled={page * 20 >= events.data.total}
              onClick={() => setPage((p) => p + 1)}
            >
              {t('Next events')}
            </Button>
          </div>
        )}
      </section>
    </div>
  )
}

export function AbuseConsole() {
  const { t } = useTranslation()
  const role = useAuthStore((state) => state.auth.user?.role)
  if (role !== ROLE.SUPER_ADMIN) {
    return <p role='alert'>{t('Root access is required')}</p>
  }
  return <SafetyConsole />
}
