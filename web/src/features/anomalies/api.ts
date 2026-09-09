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

export type AnomalyPolicy = {
  enabled: boolean
  window_minutes: number
  schema_threshold: number
  rate_threshold: number
  upstream_threshold: number
  version: number
}
export type AnomalyEvent = {
  id: string
  user_id: number
  channel_id: number
  model: string
  kind: string
  count: number
  threshold: number
  first_seen: number
  last_seen: number
  alert: boolean
  acknowledged_by: number
  acknowledged_at: number
  acknowledged_count: number
}
export type AnomalyAudit = {
  telegram_operator_id?: number
  user_id?: number
  id: string
  operator_id: number
  action: string
  event_id?: string
  created_at: number
  policy?: AnomalyPolicy
}
export type EventFilters = { user_id: string; kind: string; state: string }
export type EventPage = {
  items: AnomalyEvent[]
  total: number
  pending: number
  retained: number
}
async function read<T>(
  path: string,
  params?: Record<string, string | number>
): Promise<T> {
  const res = await api.get<{ success: boolean; data: T; message?: string }>(
    path,
    { params }
  )
  if (!res.data.success) {
    throw new Error(res.data.message || 'Unable to load anomaly data')
  }
  return res.data.data
}
export const getSettings = () => read<AnomalyPolicy>('/api/anomalies/settings')
export const getEvents = (p: number, filters: EventFilters) =>
  read<EventPage>('/api/anomalies/events', { p, ...filters })
export const getAudit = () => read<AnomalyAudit[]>('/api/anomalies/audit')
export async function saveSettings(values: AnomalyPolicy) {
  const res = await api.put('/api/anomalies/settings', values)
  if (!res.data.success) throw new Error(res.data.message)
}
export async function acknowledge(event: AnomalyEvent) {
  const res = await api.post(`/api/anomalies/events/${event.id}/acknowledge`, {
    count: event.count,
  })
  if (!res.data.success) throw new Error(res.data.message)
}
