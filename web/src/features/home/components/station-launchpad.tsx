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
import {
  ArrowRight,
  Check,
  Copy,
  KeyRound,
  Layers3,
  PlugZap,
} from 'lucide-react'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'

export function StationLaunchpad(props: { isAuthenticated: boolean }) {
  const { t } = useTranslation()
  const [copyState, setCopyState] = useState<'idle' | 'copied' | 'error'>(
    'idle'
  )
  const address = 'https://iceberg.tiktok.vip/v1'

  const copyAddress = async () => {
    try {
      await navigator.clipboard.writeText(address)
      setCopyState('copied')
    } catch {
      setCopyState('error')
    }
  }

  return (
    <aside className='station-launchpad' aria-labelledby='launchpad-title'>
      <div className='launchpad-heading'>
        <span className='launchpad-plug'>
          <PlugZap size={23} strokeWidth={1.8} aria-hidden='true' />
        </span>
        <h2 id='launchpad-title'>{t('One address. A new connection.')}</h2>
      </div>
      <div className='launchpad-endpoint'>
        <label htmlFor='launchpad-address'>API Base URL</label>
        <input
          id='launchpad-address'
          value={address}
          readOnly
          onFocus={(event) => event.target.select()}
        />
        <button type='button' onClick={() => void copyAddress()}>
          {copyState === 'copied' ? (
            <Check size={16} aria-hidden='true' />
          ) : (
            <Copy size={16} aria-hidden='true' />
          )}
          {t(copyState === 'copied' ? 'Address copied.' : 'Copy API address')}
        </button>
        <span className='launchpad-copy-status' role='status'>
          {copyState === 'copied' && t('Address copied.')}
          {copyState === 'error' &&
            t('Could not copy. Select the address and copy it manually.')}
        </span>
      </div>
      <div className='launchpad-links'>
        <Link to='/pricing'>
          <Layers3 size={19} aria-hidden='true' />
          <span>
            <strong>{t('Explore models and quota')}</strong>
            <small>{t('Find the model that fits your idea.')}</small>
          </span>
          <ArrowRight size={16} aria-hidden='true' />
        </Link>
        <Link
          to={props.isAuthenticated ? '/keys' : '/sign-in'}
          search={props.isAuthenticated ? undefined : { redirect: '/keys' }}
        >
          <KeyRound size={19} aria-hidden='true' />
          <span>
            <strong>{t('Create your API key')}</strong>
            <small>{t('Your key stays in your own hands.')}</small>
          </span>
          <ArrowRight size={16} aria-hidden='true' />
        </Link>
      </div>
    </aside>
  )
}
