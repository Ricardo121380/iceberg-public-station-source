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
import type { ColumnDef } from '@tanstack/react-table'
import dayjs from 'dayjs'
import { Ban } from 'lucide-react'
import { useTranslation } from 'react-i18next'

import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'

import { REGISTRATION_INVITE_STATUS_CONFIG } from '../constants'
import type { RegistrationInvite } from '../types'

type RegistrationInviteColumnsProps = {
  onRevoke: (invite: RegistrationInvite) => void
}

function formatTimestamp(timestamp?: number): string {
  return timestamp ? dayjs.unix(timestamp).format('YYYY-MM-DD HH:mm') : '—'
}

export function useRegistrationInviteColumns(
  props: RegistrationInviteColumnsProps
): ColumnDef<RegistrationInvite>[] {
  const { t } = useTranslation()

  return [
    {
      accessorKey: 'id',
      header: t('ID'),
      meta: { mobileHidden: true },
      size: 72,
    },
    {
      accessorKey: 'code_prefix',
      header: t('Prefix'),
      meta: { mobileTitle: true },
      cell: ({ row }) => (
        <span className='font-mono text-sm font-medium'>
          {row.original.code_prefix}
        </span>
      ),
      size: 120,
    },
    {
      accessorKey: 'note',
      header: t('Note'),
      cell: ({ row }) => (
        <span className='block max-w-64 truncate'>
          {row.original.note || '—'}
        </span>
      ),
      size: 220,
    },
    {
      accessorKey: 'status',
      header: t('Status'),
      meta: { mobileBadge: true },
      cell: ({ row }) => {
        const config = REGISTRATION_INVITE_STATUS_CONFIG[row.original.status]
        return <Badge variant={config.variant}>{t(config.labelKey)}</Badge>
      },
      size: 110,
    },
    {
      accessorKey: 'created_by',
      header: t('Created by'),
      meta: { mobileHidden: true },
      cell: ({ row }) => `#${row.original.created_by}`,
      size: 110,
    },
    {
      accessorKey: 'created_at',
      header: t('Created at'),
      meta: { mobileHidden: true },
      cell: ({ row }) => formatTimestamp(row.original.created_at),
      size: 160,
    },
    {
      accessorKey: 'expires_at',
      header: t('Expires at'),
      meta: { mobileHidden: true },
      cell: ({ row }) => formatTimestamp(row.original.expires_at),
      size: 160,
    },
    {
      id: 'used',
      header: t('Used by'),
      meta: { mobileHidden: true },
      cell: ({ row }) => {
        const invite = row.original
        if (!invite.used_by) return '—'
        return `#${invite.used_by} · ${formatTimestamp(invite.used_at)}`
      },
      size: 180,
    },
    {
      id: 'actions',
      header: t('Actions'),
      cell: ({ row }) => {
        const invite = row.original
        const canRevoke =
          invite.status === 'active' || invite.status === 'expired'

        if (!canRevoke) return '—'

        return (
          <Button
            type='button'
            variant='outline'
            size='sm'
            onClick={() => props.onRevoke(invite)}
            aria-label={t('Revoke invitation {{prefix}}', {
              prefix: invite.code_prefix,
            })}
          >
            <Ban />
            {t('Revoke')}
          </Button>
        )
      },
      size: 126,
    },
  ]
}
