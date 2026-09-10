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
import { Link } from '@tanstack/react-router'
import { useEffect, type ReactNode } from 'react'

import { StationLandscape } from '@/components/station-landscape'
import { useSystemConfig } from '@/hooks/use-system-config'

import '@/styles/iceberg.css'

/**
 * Branded full-viewport shell for the public error pages (401/403/404/500/503).
 * The little boat leaves the hero and meets the visitor again where things
 * went wrong — drifting next to the status code over the iceberg landscape.
 */
export function StationErrorShell(props: {
  code: string
  title: string
  description: string
  children?: ReactNode
}) {
  const { systemName, logo, loading } = useSystemConfig()

  useEffect(() => {
    if (loading) return
    document.title = `${props.code} · ${props.title} · ${systemName}`
  }, [props.code, props.title, systemName, loading])

  return (
    <div className='station-site station-error'>
      <StationLandscape />
      <header className='station-error-brand'>
        <Link to='/' className='station-brand'>
          <img src={logo} alt='' width='34' height='34' />
          <span>{systemName}</span>
        </Link>
      </header>
      <main className='station-error-body'>
        <img
          className='station-error-boat'
          src='/iceberg/intro/vector-v1/boat.svg'
          alt=''
          width='132'
          height='132'
        />
        <p className='station-error-code'>{props.code}</p>
        <h1 className='station-error-title'>{props.title}</h1>
        <p className='station-error-desc'>{props.description}</p>
        {props.children && (
          <div className='station-error-actions'>{props.children}</div>
        )}
      </main>
    </div>
  )
}
