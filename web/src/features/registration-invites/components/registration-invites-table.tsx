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
import type { PaginationState } from '@tanstack/react-table'
import { Filter, RotateCcw } from 'lucide-react'
import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'

import { DataTablePage, useDataTable } from '@/components/data-table'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'

import { getRegistrationInviteStats, getRegistrationInvites } from '../api'
import { REGISTRATION_INVITE_STATUS_CONFIG } from '../constants'
import type {
  RegistrationInvite,
  RegistrationInviteListParams,
  RegistrationInviteStatus,
} from '../types'
import { useRegistrationInviteColumns } from './registration-invite-columns'

type RegistrationInvitesTableProps = {
  refreshKey: number
  onCreate: () => void
  onRevoke: (invite: RegistrationInvite) => void
}

type RegistrationInviteFilterDraft = {
  status: RegistrationInviteStatus | 'all'
  note: string
  createdAfter: string
  createdBefore: string
}

const INITIAL_FILTER_DRAFT: RegistrationInviteFilterDraft = {
  status: 'all',
  note: '',
  createdAfter: '',
  createdBefore: '',
}

function toTimestamp(value: string): number | undefined {
  if (!value) return undefined

  const timestamp = new Date(value).getTime()
  return Number.isFinite(timestamp) ? Math.floor(timestamp / 1000) : undefined
}

export function RegistrationInvitesTable(props: RegistrationInvitesTableProps) {
  const { t } = useTranslation()
  const [draftFilters, setDraftFilters] =
    useState<RegistrationInviteFilterDraft>(INITIAL_FILTER_DRAFT)
  const [filters, setFilters] = useState<RegistrationInviteListParams>({})
  const [filterError, setFilterError] = useState<string | null>(null)
  const [pagination, setPagination] = useState<PaginationState>({
    pageIndex: 0,
    pageSize: 20,
  })
  const columns = useRegistrationInviteColumns({ onRevoke: props.onRevoke })

  const {
    data: listData,
    isError: isListError,
    isFetching,
    isLoading,
    refetch,
  } = useQuery({
    queryKey: [
      'registration-invites',
      pagination.pageIndex,
      pagination.pageSize,
      filters,
      props.refreshKey,
    ],
    queryFn: async () => {
      const response = await getRegistrationInvites({
        ...filters,
        p: pagination.pageIndex + 1,
        page_size: pagination.pageSize,
      })

      if (!response.success || !response.data) {
        throw new Error('registration invitation list request failed')
      }

      return response.data
    },
    retry: false,
  })

  const { data: statsData } = useQuery({
    queryKey: ['registration-invite-stats', props.refreshKey],
    queryFn: async () => {
      const response = await getRegistrationInviteStats()
      if (!response.success || !response.data) {
        throw new Error('registration invitation statistics request failed')
      }
      return response.data
    },
    retry: false,
  })

  useEffect(() => {
    if (listData === undefined) return

    const lastPageIndex = Math.max(
      0,
      Math.ceil(listData.total / pagination.pageSize) - 1
    )
    if (pagination.pageIndex > lastPageIndex) {
      setPagination((current) => ({
        ...current,
        pageIndex: lastPageIndex,
      }))
    }
  }, [listData, pagination.pageIndex, pagination.pageSize])

  const { table } = useDataTable({
    data: listData?.items || [],
    columns,
    manualFiltering: true,
    manualPagination: true,
    pagination,
    onPaginationChange: setPagination,
    totalCount: listData?.total || 0,
  })

  const applyFilters = (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault()

    const createdAfter = toTimestamp(draftFilters.createdAfter)
    const createdBefore = toTimestamp(draftFilters.createdBefore)
    if (createdAfter && createdBefore && createdAfter > createdBefore) {
      setFilterError(t('The start time must be before the end time.'))
      return
    }

    setFilterError(null)
    setFilters({
      status: draftFilters.status === 'all' ? undefined : draftFilters.status,
      note: draftFilters.note.trim() || undefined,
      created_after: createdAfter,
      created_before: createdBefore,
    })
    setPagination((current) => ({ ...current, pageIndex: 0 }))
  }

  const resetFilters = () => {
    setDraftFilters(INITIAL_FILTER_DRAFT)
    setFilters({})
    setFilterError(null)
    setPagination((current) => ({ ...current, pageIndex: 0 }))
  }

  return (
    <DataTablePage
      table={table}
      columns={columns}
      isLoading={isLoading}
      isFetching={isFetching}
      emptyTitle={t('No invitation codes found')}
      emptyDescription={t(
        'Create invitation codes to allow approved users to join the station.'
      )}
      emptyAction={
        <Button type='button' onClick={props.onCreate}>
          {t('Create invitation codes')}
        </Button>
      }
      skeletonKeyPrefix='registration-invites-skeleton'
      applyHeaderSize
      toolbar={
        <div className='space-y-3'>
          <div className='grid grid-cols-2 gap-2 sm:grid-cols-4'>
            {(['active', 'used', 'expired', 'revoked'] as const).map(
              (status) => (
                <div
                  key={status}
                  className='bg-card rounded-lg border px-3 py-2.5'
                >
                  <p className='text-muted-foreground text-xs font-medium'>
                    {t(REGISTRATION_INVITE_STATUS_CONFIG[status].labelKey)}
                  </p>
                  <p className='mt-1 text-xl font-semibold tabular-nums'>
                    {statsData?.[status] ?? '—'}
                  </p>
                </div>
              )
            )}
          </div>

          <form
            onSubmit={applyFilters}
            className='bg-card grid gap-3 rounded-lg border p-3 sm:grid-cols-2 lg:grid-cols-[minmax(0,1.25fr)_minmax(0,0.8fr)_minmax(0,1fr)_minmax(0,1fr)_auto] lg:items-end'
          >
            <div className='space-y-1.5'>
              <Label htmlFor='registration-invite-note-filter'>
                {t('Note')}
              </Label>
              <Input
                id='registration-invite-note-filter'
                value={draftFilters.note}
                maxLength={255}
                placeholder={t('Filter by note')}
                onChange={(event) => {
                  const note = event.currentTarget.value
                  setDraftFilters((current) => ({
                    ...current,
                    note,
                  }))
                }}
              />
            </div>
            <div className='space-y-1.5'>
              <Label>{t('Status')}</Label>
              <Select
                value={draftFilters.status}
                onValueChange={(value) =>
                  setDraftFilters((current) => ({
                    ...current,
                    status: value as RegistrationInviteStatus | 'all',
                  }))
                }
              >
                <SelectTrigger className='w-full'>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value='all'>{t('All statuses')}</SelectItem>
                  {(['active', 'used', 'expired', 'revoked'] as const).map(
                    (status) => (
                      <SelectItem key={status} value={status}>
                        {t(REGISTRATION_INVITE_STATUS_CONFIG[status].labelKey)}
                      </SelectItem>
                    )
                  )}
                </SelectContent>
              </Select>
            </div>
            <div className='space-y-1.5'>
              <Label htmlFor='registration-invite-created-after'>
                {t('Created after')}
              </Label>
              <Input
                id='registration-invite-created-after'
                type='datetime-local'
                value={draftFilters.createdAfter}
                onChange={(event) => {
                  const createdAfter = event.currentTarget.value
                  setDraftFilters((current) => ({
                    ...current,
                    createdAfter,
                  }))
                }}
              />
            </div>
            <div className='space-y-1.5'>
              <Label htmlFor='registration-invite-created-before'>
                {t('Created before')}
              </Label>
              <Input
                id='registration-invite-created-before'
                type='datetime-local'
                value={draftFilters.createdBefore}
                onChange={(event) => {
                  const createdBefore = event.currentTarget.value
                  setDraftFilters((current) => ({
                    ...current,
                    createdBefore,
                  }))
                }}
              />
            </div>
            <div className='flex gap-2 lg:pb-0.5'>
              <Button type='submit' size='sm' className='flex-1 lg:flex-none'>
                <Filter />
                {t('Apply')}
              </Button>
              <Button
                type='button'
                size='sm'
                variant='outline'
                onClick={resetFilters}
                aria-label={t('Reset filters')}
              >
                <RotateCcw />
                <span className='sr-only sm:not-sr-only'>{t('Reset')}</span>
              </Button>
            </div>
          </form>

          {filterError ? (
            <Alert variant='destructive'>
              <AlertTitle>{t('Invalid date range')}</AlertTitle>
              <AlertDescription>{filterError}</AlertDescription>
            </Alert>
          ) : null}

          {isListError ? (
            <Alert variant='destructive'>
              <AlertTitle>{t('Failed to load invitation codes')}</AlertTitle>
              <AlertDescription>
                <div className='flex flex-wrap items-center gap-2'>
                  <span>{t('Check the connection and try again.')}</span>
                  <Button
                    type='button'
                    size='sm'
                    variant='outline'
                    onClick={() => refetch()}
                  >
                    {t('Retry')}
                  </Button>
                </div>
              </AlertDescription>
            </Alert>
          ) : null}
        </div>
      }
    />
  )
}
