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
import { useTranslation } from 'react-i18next'

import { getSelfAbuse } from './api'

export function AbuseStatusBanner() {
  const { t } = useTranslation()
  const state = useQuery({
    queryKey: ['abuse', 'self'],
    queryFn: getSelfAbuse,
    refetchInterval: 30000,
    retry: 1,
  })
  if (state.isError) {
    return (
      <p role='status' className='bg-muted px-4 py-2 text-sm'>
        {t('Model access status is temporarily unavailable.')}
      </p>
    )
  }
  if (!state.data?.suspended) return null
  return (
    <p role='alert' className='bg-muted px-4 py-3 text-sm'>
      {t(
        'Model access is temporarily suspended. You can still sign in. Resumes at:'
      )}{' '}
      {new Date(state.data.blocked_until * 1000).toLocaleString()}
    </p>
  )
}
