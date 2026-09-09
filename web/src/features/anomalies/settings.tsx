import { zodResolver } from '@hookform/resolvers/zod'
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
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useForm } from 'react-hook-form'
import { useTranslation } from 'react-i18next'
import { z } from 'zod'

import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'

import { saveSettings, type AnomalyPolicy } from './api'

const schema = z.object({
  enabled: z.boolean(),
  window_minutes: z.number().int().min(1).max(60),
  schema_threshold: z.number().int().min(2).max(1000),
  rate_threshold: z.number().int().min(2).max(10000),
  upstream_threshold: z.number().int().min(2).max(1000),
  version: z.number().int().min(0),
})
export function AnomalySettings(props: {
  policy: AnomalyPolicy
  canEdit: boolean
}) {
  const { t } = useTranslation()
  const client = useQueryClient()
  const form = useForm<AnomalyPolicy>({
    resolver: zodResolver(schema),
    defaultValues: props.policy,
  })
  const save = useMutation({
    mutationFn: saveSettings,
    onSuccess: () => {
      void client.invalidateQueries({ queryKey: ['anomalies'] })
    },
  })
  const fields = [
    {
      key: 'window_minutes' as const,
      label: t('Aggregation window (minutes)'),
      max: 60,
      min: 1,
    },
    {
      key: 'schema_threshold' as const,
      label: t('Schema error alert threshold'),
      max: 1000,
      min: 2,
    },
    {
      key: 'rate_threshold' as const,
      label: t('Rate limit alert threshold'),
      max: 10000,
      min: 2,
    },
    {
      key: 'upstream_threshold' as const,
      label: t('Upstream failure alert threshold'),
      max: 1000,
      min: 2,
    },
  ]
  return (
    <form
      onSubmit={form.handleSubmit((values) => save.mutate(values))}
      className='max-w-2xl space-y-6'
    >
      <div className='space-y-2'>
        <h2 className='text-lg font-semibold'>{t('Alert rules')}</h2>
        <p className='text-muted-foreground text-sm'>
          {t(
            'Rules create dashboard reminders only. They never suspend accounts or change request limits.'
          )}
        </p>
      </div>
      {!props.canEdit && (
        <p role='status' className='text-muted-foreground text-sm'>
          {t(
            'Only Root can change alert rules. Administrators can review and acknowledge events.'
          )}
        </p>
      )}
      <fieldset
        disabled={!props.canEdit || save.isPending}
        className='space-y-5'
      >
        <label className='flex items-center gap-3 text-sm'>
          <input
            type='checkbox'
            {...form.register('enabled')}
            className='accent-primary size-4'
          />
          {t('Collect request anomalies')}
        </label>
        <div className='grid gap-5 sm:grid-cols-2'>
          {fields.map((field) => (
            <div key={field.key} className='space-y-2'>
              <label htmlFor={field.key} className='text-sm font-medium'>
                {field.label}
              </label>
              <Input
                id={field.key}
                type='number'
                min={field.min}
                max={field.max}
                {...form.register(field.key, { valueAsNumber: true })}
                aria-invalid={!!form.formState.errors[field.key]}
                aria-describedby={`${field.key}-range`}
              />
              <p
                id={`${field.key}-range`}
                className='text-muted-foreground text-xs'
              >
                {field.min}–{field.max}
              </p>
              {form.formState.errors[field.key] && (
                <p role='alert' className='text-destructive text-sm'>
                  {t('Enter an integer within the allowed range.')}
                </p>
              )}
            </div>
          ))}
        </div>
        {props.canEdit && (
          <Button
            type='submit'
            disabled={save.isPending || !form.formState.isDirty}
          >
            {save.isPending ? t('Saving...') : t('Save alert rules')}
          </Button>
        )}
      </fieldset>
      {save.isError && (
        <p role='alert' className='text-destructive text-sm'>
          {t('Could not save. Refresh to get the latest rules and retry.')}
        </p>
      )}
      {save.isSuccess && (
        <p role='status' className='text-sm'>
          {t('Alert rules saved')}
        </p>
      )}
    </form>
  )
}
