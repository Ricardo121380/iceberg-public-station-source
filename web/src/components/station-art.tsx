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
import { useTranslation } from 'react-i18next'

/** The existing station illustration is the visual authority for this surface. */
export function StationArt() {
  const { t } = useTranslation()

  return (
    <div className='station-art'>
      <div className='station-art-lens' aria-hidden='true' />
      <div className='station-art-ripple' aria-hidden='true' />
      <img
        src='/iceberg-station-mark-v2.png'
        width='512'
        height='512'
        fetchPriority='high'
        alt={t('Iceberg station mascot sailing between blue icebergs')}
      />
      <span
        className='station-art-droplet station-art-droplet-blue'
        aria-hidden='true'
      />
      <span
        className='station-art-droplet station-art-droplet-pink'
        aria-hidden='true'
      />
    </div>
  )
}
