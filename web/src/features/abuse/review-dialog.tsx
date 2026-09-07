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
import { useEffect, useState } from 'react'
import { useForm } from 'react-hook-form'
import { useTranslation } from 'react-i18next'
import { z } from 'zod'

import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Textarea } from '@/components/ui/textarea'

import {
  getReviewHistory,
  readExcerpt,
  reviewEvent,
  type AbuseEvent,
  type ReviewDecision,
} from './api'

const decisions: Record<ReviewDecision, string> = {
  confirmed: 'Confirmed violation',
  suspected_false_positive: 'Suspected false positive',
  insufficient_evidence: 'Insufficient evidence',
}
const unavailableReasons: Record<string, string> = {
  not_collected: 'This older event has no input excerpt.',
  key_unavailable:
    'No excerpt was captured because the encryption key was unavailable.',
  body_unavailable: 'The original request body was unavailable for capture.',
  body_too_large:
    'The request exceeded the safe capture size and no excerpt was stored.',
  no_user_text:
    'No eligible user text was present. Attachments were not captured.',
  unsupported_protocol: 'Input excerpts are not supported for this protocol.',
  invalid_body: 'The request could not be safely parsed for an excerpt.',
  capture_off: 'Excerpt capture was disabled when this event occurred.',
  capture_failed: 'The excerpt could not be encrypted and was not stored.',
}
const reviewSchema = z.object({
  decision: z.enum([
    'confirmed',
    'suspected_false_positive',
    'insufficient_evidence',
  ]),
  note: z
    .string()
    .trim()
    .min(1)
    .refine((s) => [...s].length <= 300),
})
type ReviewForm = z.infer<typeof reviewSchema>

function ReviewContent(props: { event: AbuseEvent; close: () => void }) {
  const { t } = useTranslation()
  const client = useQueryClient()
  const [page, setPage] = useState(1)
  const [expired, setExpired] = useState(false)
  const form = useForm<ReviewForm>({
    resolver: zodResolver(reviewSchema),
    defaultValues: { decision: 'insufficient_evidence', note: '' },
  })
  const history = useQuery({
    queryKey: ['abuse-review-history', props.event.id, page],
    queryFn: () => getReviewHistory(props.event.id, page),
    gcTime: 0,
  })
  const excerpt = useMutation({
    mutationFn: () => readExcerpt(props.event.id),
    gcTime: 0,
    retry: false,
    onSuccess: () => {
      void client.invalidateQueries({
        queryKey: ['abuse-review-history', props.event.id],
      })
    },
  })
  const save = useMutation({
    mutationFn: (values: ReviewForm) =>
      reviewEvent(props.event.id, {
        ...values,
        version: props.event.review_version ?? 0,
      }),
    onSuccess: () => {
      void client.invalidateQueries({ queryKey: ['abuse', 'events'] })
      props.close()
    },
  })
  const resetExcerpt = excerpt.reset
  useEffect(() => {
    if (!excerpt.data || excerpt.data.status !== 'available') return
    const delay = Math.max(0, excerpt.data.expires_at * 1000 - Date.now())
    const timer = setTimeout(() => {
      resetExcerpt()
      setExpired(true)
    }, delay)
    return () => clearTimeout(timer)
  }, [excerpt.data, resetExcerpt])
  const visible =
    excerpt.data?.status === 'available' &&
    excerpt.data.expires_at * 1000 > Date.now()
  return (
    <div className='space-y-5'>
      <p className='text-muted-foreground text-sm'>
        {t(
          'This is only the last user message excerpt. It may not be the cause of the upstream block. History, attachments and output may also trigger a block.'
        )}
      </p>
      <div className='space-y-2 rounded-md border p-3'>
        {!visible && (
          <Button
            variant='outline'
            onClick={() => {
              setExpired(false)
              excerpt.mutate()
            }}
            disabled={excerpt.isPending}
          >
            {t('View redacted excerpt (access is audited)')}
          </Button>
        )}
        {visible && excerpt.data && (
          <>
            <p className='text-sm'>
              {t('Source')}:{' '}
              {excerpt.data.excerpt.source === 'input'
                ? t('Current input text')
                : t('Last user message text')}
            </p>
            <p className='text-muted-foreground text-sm'>
              {t('Excerpt expires at')}:{' '}
              {new Date(excerpt.data.expires_at * 1000).toLocaleString()}
            </p>
            <p
              className='bg-muted max-h-52 overflow-auto rounded p-3 text-sm break-words whitespace-pre-wrap'
              aria-label={t('Redacted input excerpt')}
            >
              {excerpt.data.excerpt.text}
            </p>
            {excerpt.data.excerpt.truncated && (
              <p>{t('Only the first 300 redacted characters are shown.')}</p>
            )}
            {excerpt.data.excerpt.omitted_parts && (
              <p>{t('Non-text parts were omitted.')}</p>
            )}
            <Button variant='outline' onClick={() => excerpt.reset()}>
              {t('Hide excerpt')}
            </Button>
          </>
        )}
        {(expired || excerpt.data?.status === 'expired') && (
          <p role='status'>
            {t('The excerpt has expired and is no longer available.')}
          </p>
        )}
        {excerpt.data && !visible && excerpt.data.status !== 'expired' && (
          <p role='status'>
            {t(
              unavailableReasons[excerpt.data.status] ??
                'No excerpt is available for this event.'
            )}
          </p>
        )}
        {excerpt.isError && (
          <p role='alert'>
            {t(
              'Unable to read the excerpt. Access audit or decryption may be unavailable; retry later.'
            )}
          </p>
        )}
      </div>
      <form
        className='space-y-3'
        onSubmit={form.handleSubmit((values) => save.mutate(values))}
      >
        <p className='text-sm'>
          {t('Current review')}:{' '}
          {props.event.review_status
            ? t(decisions[props.event.review_status])
            : t('Not reviewed')}
        </p>
        <label className='block space-y-2 text-sm'>
          {t('Review decision')}
          <select
            className='border-input bg-background h-9 w-full rounded-md border px-3'
            {...form.register('decision')}
          >
            {Object.entries(decisions).map(([value, label]) => (
              <option key={value} value={value}>
                {t(label)}
              </option>
            ))}
          </select>
        </label>
        <label className='block space-y-2 text-sm'>
          {t('Review note')}
          <Textarea
            {...form.register('note')}
            aria-invalid={Boolean(form.formState.errors.note)}
            rows={3}
          />
        </label>
        <p className='text-muted-foreground text-sm'>
          {t(
            'Use 1 to 300 characters to explain your decision. Do not paste prompts or personal information into review notes.'
          )}
        </p>
        {form.formState.errors.note && (
          <p role='alert'>{t('Enter a review note of 1 to 300 characters.')}</p>
        )}
        <p className='text-sm'>
          {t(
            'Saving this review does not freeze an account or add punishment counts. Account actions remain separate.'
          )}
        </p>
        {save.isError && (
          <p role='alert'>
            {t(
              'Unable to save the review. Close and refresh the event before retrying if another reviewer has updated it.'
            )}
          </p>
        )}
        <Button type='submit' disabled={save.isPending}>
          {t('Save review only')}
        </Button>
      </form>
      <section
        className='space-y-2 border-t pt-3'
        aria-label={t('Review and access history')}
      >
        <h3 className='font-medium'>{t('Review and access history')}</h3>
        {history.isPending && (
          <p role='status'>{t('Loading review history')}</p>
        )}
        {history.isError && (
          <p role='alert'>{t('Unable to load review history')}</p>
        )}
        {history.data?.items.length === 0 && (
          <p>{t('No review history yet')}</p>
        )}
        {history.data?.items.map((audit) => (
          <div key={audit.id} className='border-b py-2 text-sm break-words'>
            <p>
              {new Date(audit.created_at * 1000).toLocaleString()} ·{' '}
              {t('Operator ID')} {audit.operator_id} ·{' '}
              {audit.kind === 'excerpt_access'
                ? t('Excerpt access recorded')
                : t(
                    decisions[audit.decision as ReviewDecision] ??
                      'Not reviewed'
                  )}
            </p>
            {audit.note && <p className='whitespace-pre-wrap'>{audit.note}</p>}
          </div>
        ))}
        <div className='flex items-center gap-2'>
          <Button
            variant='outline'
            disabled={page === 1}
            onClick={() => setPage((p) => p - 1)}
          >
            {t('Previous reviews')}
          </Button>
          <span>{page}</span>
          <Button
            variant='outline'
            disabled={!history.data || page * 20 >= history.data.total}
            onClick={() => setPage((p) => p + 1)}
          >
            {t('Next reviews')}
          </Button>
        </div>
      </section>
    </div>
  )
}

export function SafetyReviewDialog(props: { event: AbuseEvent }) {
  const { t } = useTranslation()
  const [open, setOpen] = useState(false)
  return (
    <>
      <Button variant='outline' className='mt-3' onClick={() => setOpen(true)}>
        {t('Review safety event')}
      </Button>
      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent className='max-h-[85vh] overflow-y-auto sm:max-w-2xl'>
          <DialogHeader>
            <DialogTitle>
              {t('Review safety event')} · {props.event.id}
            </DialogTitle>
            <DialogDescription>
              {t('User ID')} {props.event.user_id} · {props.event.model}
            </DialogDescription>
          </DialogHeader>
          {open && (
            <ReviewContent
              key={props.event.id}
              event={props.event}
              close={() => setOpen(false)}
            />
          )}
        </DialogContent>
      </Dialog>
    </>
  )
}
