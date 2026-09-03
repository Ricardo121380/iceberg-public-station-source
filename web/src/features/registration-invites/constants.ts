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
import type { ComponentProps } from 'react'

import type { Badge } from '@/components/ui/badge'

import type { RegistrationInviteStatus } from './types'

export const REGISTRATION_INVITE_VALIDATION = {
  COUNT_MIN: 1,
  COUNT_MAX: 100,
  VALID_DAYS_MIN: 1,
  VALID_DAYS_MAX: 365,
  NOTE_MAX_BYTES: 255,
} as const

export const REGISTRATION_INVITE_DEFAULT_VALUES = {
  count: 1,
  valid_days: 7,
  note: '',
} as const

export const REGISTRATION_INVITE_STATUS_CONFIG: Record<
  RegistrationInviteStatus,
  {
    labelKey: string
    variant: ComponentProps<typeof Badge>['variant']
  }
> = {
  active: { labelKey: 'Available', variant: 'default' },
  expired: { labelKey: 'Expired', variant: 'warning' },
  used: { labelKey: 'Used', variant: 'secondary' },
  revoked: { labelKey: 'Revoked invitation', variant: 'destructive' },
}
