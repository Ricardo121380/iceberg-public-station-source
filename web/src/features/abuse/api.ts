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
import { api } from '@/lib/api'

export type SafetyRule = {
  id: string
  verified: boolean
  category: string
  evidence: string
}
export type AbuseSettings = {
  evidence_capture_ready?: boolean
  mode: 'off' | 'observe' | 'enforce'
  limit_10m: number
  limit_24h: number
  freeze_minutes: number
  enabled_rules: string[]
  rules: SafetyRule[]
}
export type AbuseEvent = {
  id: number
  evidence_status?: string
  evidence_expires_at?: number
  review_status?: ReviewDecision | ''
  review_version?: number
  reviewed_at?: number
  reviewed_by?: number
  user_id: number
  token_id: number
  channel_id: number
  request_id: string
  rule_id: string
  category: string
  action: string
  signal: string
  summary: string
  model: string
  protocol: string
  created_at: number
  count_10m: number
  count_24h: number
}
export type AbuseState = { user_id: number; blocked_until: number }
export type SelfAbuse = { suspended: boolean; blocked_until: number }
export type Page<T> = { items: T[]; total: number }
export type EventFilters = {
  user_id: string
  category: string
  rule_id: string
  action: string
  start: string
  end: string
}

async function read<T>(
  path: string,
  params?: Record<string, string | number>
): Promise<T> {
  const response = await api.get<{
    success: boolean
    data: T
    message?: string
  }>(path, { params })
  if (!response.data.success) {
    throw new Error(
      response.data.message || 'Safety service temporarily unavailable'
    )
  }
  return response.data.data
}
export const getSettings = () => read<AbuseSettings>('/api/abuse/settings')
export async function saveSettings(values: Omit<AbuseSettings, 'rules'>) {
  const response = await api.put('/api/abuse/settings', values)
  if (!response.data.success) throw new Error(response.data.message)
}
export const getEvents = (page: number, filters: EventFilters) => {
  const params: Record<string, string | number> = { p: page, page_size: 20 }
  for (const [key, value] of Object.entries(filters)) {
    if (!value) continue
    params[key] =
      key === 'start' || key === 'end'
        ? Math.floor(new Date(value).getTime() / 1000)
        : value
  }
  return read<Page<AbuseEvent>>('/api/abuse/events', params)
}
export const getSuspensions = (page: number) =>
  read<Page<AbuseState>>('/api/abuse/users', { p: page, page_size: 20 })
export const getSelfAbuse = () => read<SelfAbuse>('/api/user/self/abuse')
export async function unfreezeUser(id: number, reason: string) {
  const response = await api.post(`/api/abuse/users/${id}/unfreeze`, { reason })
  if (!response.data.success) throw new Error(response.data.message)
}

export type ReviewDecision =
  | 'confirmed'
  | 'suspected_false_positive'
  | 'insufficient_evidence'
export type SafetyExcerpt = {
  status: string
  expires_at: number
  excerpt: {
    text: string
    source: string
    truncated: boolean
    omitted_parts: boolean
  }
}
export type ReviewAudit = {
  id: number
  operator_id: number
  kind: string
  decision: ReviewDecision | ''
  note: string
  version: number
  created_at: number
}
export async function readExcerpt(id: number): Promise<SafetyExcerpt> {
  const response = await api.post<{
    success: boolean
    data: SafetyExcerpt
    message?: string
  }>(`/api/abuse/events/${id}/excerpt`)
  if (!response.data.success) throw new Error(response.data.message)
  return response.data.data
}
export async function reviewEvent(
  id: number,
  values: { decision: ReviewDecision; note: string; version: number }
): Promise<void> {
  const response = await api.post(`/api/abuse/events/${id}/review`, values)
  if (!response.data.success) throw new Error(response.data.message)
}
export const getReviewHistory = (id: number, page: number) =>
  read<Page<ReviewAudit>>(`/api/abuse/events/${id}/history`, {
    p: page,
    page_size: 20,
  })
