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
import { api, type ApiRequestConfig } from '@/lib/api'

import {
  REGISTRATION_INVITE_STATUS_VALUES,
  type ApiResponse,
  type CreatedRegistrationInvite,
  type RegistrationInviteCreateParams,
  type RegistrationInviteListData,
  type RegistrationInviteListParams,
  type RegistrationInviteStats,
} from './types'

export async function getRegistrationInvites(
  params: RegistrationInviteListParams = {},
  requestConfig: ApiRequestConfig = {}
): Promise<ApiResponse<RegistrationInviteListData>> {
  const search = new URLSearchParams()

  if (params.p !== undefined) search.set('p', String(params.p))
  if (params.page_size !== undefined) {
    search.set('page_size', String(params.page_size))
  }
  if (params.status) search.set('status', params.status)
  if (params.created_after !== undefined) {
    search.set('created_after', String(params.created_after))
  }
  if (params.created_before !== undefined) {
    search.set('created_before', String(params.created_before))
  }
  if (params.note) search.set('note', params.note)

  const suffix = search.size > 0 ? `?${search.toString()}` : ''
  const response = await api.get(
    `/api/registration-invites${suffix}`,
    requestConfig
  )
  return response.data
}

export async function createRegistrationInvites(
  params: RegistrationInviteCreateParams
): Promise<ApiResponse<{ invites: CreatedRegistrationInvite[] }>> {
  const response = await api.post('/api/registration-invites', params)
  return response.data
}

export async function revokeRegistrationInvite(
  id: number
): Promise<ApiResponse<{ id: number; revoked: boolean }>> {
  const response = await api.post(`/api/registration-invites/${id}/revoke`)
  return response.data
}

export async function getRegistrationInviteStats(): Promise<
  ApiResponse<RegistrationInviteStats>
> {
  const responses = await Promise.all(
    REGISTRATION_INVITE_STATUS_VALUES.map(async (status) => {
      const response = await getRegistrationInvites(
        {
          p: 1,
          page_size: 1,
          status,
        },
        { skipBusinessError: true }
      )
      return { response, status }
    })
  )

  const failedResponse = responses.find(
    ({ response }) => !response.success || !response.data
  )
  if (failedResponse) {
    return {
      success: false,
      message: failedResponse.response.message,
    }
  }

  return {
    success: true,
    data: responses.reduce<RegistrationInviteStats>(
      (stats, { response, status }) => {
        stats[status] = response.data?.total ?? 0
        return stats
      },
      {
        active: 0,
        expired: 0,
        used: 0,
        revoked: 0,
      }
    ),
  }
}
