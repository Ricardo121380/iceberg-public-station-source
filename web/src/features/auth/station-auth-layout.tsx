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
import { ArrowLeft } from 'lucide-react'
import type { ReactNode } from 'react'
import { useTranslation } from 'react-i18next'

import { LanguageSwitcher } from '@/components/language-switcher'
import { StationArt } from '@/components/station-art'
import { ThemeSwitch } from '@/components/theme-switch'
import { useSystemConfig } from '@/hooks/use-system-config'

import '@/styles/iceberg.css'

export function StationAuthLayout(props: { children: ReactNode }) {
  const { t } = useTranslation()
  const { systemName, logo } = useSystemConfig()

  return (
    <div className='station-site station-auth'>
      <a href='#station-auth-form' className='station-skip'>
        {t('Skip to sign-in or registration')}
      </a>
      <header className='station-auth-header'>
        <Link to='/' className='station-brand'>
          <img src={logo} alt='' width='34' height='34' />
          <span>{systemName}</span>
        </Link>
        <div className='station-auth-tools'>
          <LanguageSwitcher />
          <ThemeSwitch />
          <Link to='/' className='station-back'>
            <ArrowLeft size={16} aria-hidden='true' />
            {t('Back to home')}
          </Link>
        </div>
      </header>
      <main className='station-auth-main'>
        <aside className='station-auth-art'>
          <div className='station-auth-art-copy'>
            <h1 className='station-display'>
              {t('A new connection.')}
              <br />
              {t('A new possibility.')}
            </h1>
            <p>{t('A community AI API station for LinuxDO.')}</p>
          </div>
          <StationArt />
          <div className='station-auth-art-footer'>
            <span>ICEBERG</span>
            <span>{t('A little space for curiosity.')}</span>
          </div>
        </aside>
        <div
          className='station-auth-panel'
          id='station-auth-form'
          tabIndex={-1}
        >
          <div className='station-auth-form'>{props.children}</div>
          <p className='station-auth-attribution'>
            Powered by{' '}
            <a
              href='https://github.com/QuantumNous/new-api'
              target='_blank'
              rel='noopener noreferrer'
            >
              New API
            </a>{' '}
            · QuantumNous
          </p>
        </div>
      </main>
    </div>
  )
}
