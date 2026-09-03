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
import { z } from 'zod'

export const REGISTRATION_INVITE_STATUS_VALUES = [
  'active',
  'expired',
  'used',
  'revoked',
] as const

export const registrationInviteStatusSchema = z.enum(
  REGISTRATION_INVITE_STATUS_VALUES
)

export type RegistrationInviteStatus = z.infer<
  typeof registrationInviteStatusSchema
>

export const registrationInviteSchema = z.object({
  id: z.number(),
  code_prefix: z.string(),
  note: z.string(),
  created_by: z.number(),
  created_at: z.number(),
  expires_at: z.number(),
  used_by: z.number().optional(),
  used_at: z.number().optional(),
  revoked_by: z.number().optional(),
  revoked_at: z.number().optional(),
  status: registrationInviteStatusSchema,
})

export type RegistrationInvite = z.infer<typeof registrationInviteSchema>

export type CreatedRegistrationInvite = {
  id: number
  code: string
  code_prefix: string
  expires_at: number
}

export type RegistrationInviteListParams = {
  p?: number
  page_size?: number
  status?: RegistrationInviteStatus
  created_after?: number
  created_before?: number
  note?: string
}

export type RegistrationInviteCreateParams = {
  count: number
  valid_days: number
  note: string
}

export type ApiResponse<T = unknown> = {
  success: boolean
  message?: string
  data?: T
}

export type RegistrationInviteListData = {
  items: RegistrationInvite[]
  total: number
  page: number
  page_size: number
}

export type RegistrationInviteStats = Record<RegistrationInviteStatus, number>
