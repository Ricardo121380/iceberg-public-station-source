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
import { ArrowDown, ArrowRight, ArrowUpRight, Check } from 'lucide-react'
import { useTranslation } from 'react-i18next'

import { StationLandscape } from '@/components/station-landscape'
import { useStatus } from '@/hooks/use-status'

import { StationConnection } from './station-connection'
import { StationLaunchpad } from './station-launchpad'

export function StationHome(props: { isAuthenticated: boolean }) {
  const { t } = useTranslation()
  const { status } = useStatus()
  const registrationOpen =
    status?.register_enabled !== false && !status?.self_use_mode_enabled

  return (
    <main id='station-main' className='station-main'>
      <div className='station-hero-scene'>
        <StationLandscape />
        <section
          className='station-hero station-container'
          aria-labelledby='station-title'
        >
          <div className='station-hero-copy'>
            <h1 id='station-title' className='station-display'>
              <span>{t('Explore AI.')}</span>
              <span>{t('Find a new beginning.')}</span>
            </h1>
            <p className='station-intro'>
              {t(
                'A community AI API station for LinuxDO. Discover available models, connect your favorite tools, and keep exploring.'
              )}
            </p>
            <div className='station-actions'>
              <Link
                to={props.isAuthenticated ? '/dashboard' : '/sign-in'}
                className='station-button'
              >
                {t(
                  props.isAuthenticated
                    ? 'Go to Dashboard'
                    : 'Sign in with LinuxDO'
                )}
                <ArrowRight size={17} aria-hidden='true' />
              </Link>
              {!props.isAuthenticated && registrationOpen && (
                <Link
                  to='/sign-up'
                  className='station-button station-button-secondary'
                >
                  {t('Register with an invitation')}
                </Link>
              )}
            </div>
            <p className='station-eligibility'>
              <Check size={15} aria-hidden='true' />
              {t('New here? Bring an invitation and a LinuxDO TL1+ account.')}
            </p>
            <a href='#connect' className='station-scroll-link'>
              <ArrowDown size={17} aria-hidden='true' />
              {t('See how to connect')}
            </a>
          </div>
          <StationLaunchpad isAuthenticated={props.isAuthenticated} />
        </section>
        <div className='station-motto station-container'>
          <span>ICEBERG</span>
          <p>{t('A little boat meets a big iceberg.')}</p>
          <span className='station-motto-rule' aria-hidden='true' />
        </div>
      </div>

      <section
        className='station-models station-container'
        aria-labelledby='station-models-title'
      >
        <div>
          <h2 id='station-models-title'>
            {t('The right model for your next idea.')}
          </h2>
          <p>
            {t(
              'Browse model IDs, supported interfaces, and quota rates before connecting.'
            )}
          </p>
        </div>
        <Link to='/pricing' className='station-text-link'>
          {t('Explore models and quota')}
          <ArrowUpRight size={18} aria-hidden='true' />
        </Link>
      </section>

      <section
        id='connect'
        className='station-start'
        aria-labelledby='station-start-title'
      >
        <div className='station-container'>
          <div className='station-section-heading'>
            <h2 id='station-start-title' className='station-display'>
              {t('Your first connection starts here.')}
            </h2>
            <p>
              {t(
                'Three steps from joining the community to your first API request.'
              )}
            </p>
          </div>
          <ol className='station-steps'>
            <li>
              <span className='station-step-number'>1</span>
              <div>
                <h3>{t('Join with LinuxDO')}</h3>
                <p>
                  {t(
                    'Already registered? Sign in directly. New members need an invitation and a LinuxDO TL1+ account.'
                  )}
                </p>
              </div>
            </li>
            <li>
              <span className='station-step-number'>2</span>
              <div>
                <h3>{t('Create your API key')}</h3>
                <p>
                  {t(
                    'Check your available quota in the console, then create a key. Keep it private.'
                  )}
                </p>
              </div>
            </li>
            <li>
              <span className='station-step-number'>3</span>
              <div>
                <h3>{t('Connect your tools')}</h3>
                <p>
                  {t(
                    'Add the station address and your key to a compatible client, then choose an available model.'
                  )}
                </p>
              </div>
            </li>
          </ol>
          <StationConnection />
        </div>
      </section>

      <section
        className='station-faq station-container'
        aria-labelledby='station-faq-title'
      >
        <div>
          <h2 id='station-faq-title' className='station-display'>
            {t('Before you begin.')}
          </h2>
          <p>
            {t(
              'A few things worth knowing, so you can explore with confidence.'
            )}
          </p>
        </div>
        <div className='station-questions'>
          <details>
            <summary>{t('Do I need an invitation every time?')}</summary>
            <p>
              {t(
                'No. Invitations are only required for first-time registration. Once your LinuxDO identity is linked, sign in directly with the same account.'
              )}
            </p>
          </details>
          <details>
            <summary>{t('How much quota can I use?')}</summary>
            <p>
              {t(
                'Your console shows your available quota. Registration does not automatically grant quota or create an API key. Follow station announcements for allocation details.'
              )}
            </p>
          </details>
          <details>
            <summary>{t('Why is a model unavailable to my key?')}</summary>
            <p>
              {t(
                'Model access depends on your account and key group. Check the model catalog and key settings, and confirm that your quota is sufficient.'
              )}
            </p>
          </details>
          <details>
            <summary>{t('What if my invitation does not work?')}</summary>
            <p>
              {t(
                'An invitation may be expired, already used, or revoked. Ask the person who invited you for a valid code. If you already registered, return to sign-in.'
              )}
            </p>
          </details>
        </div>
      </section>
      <section className='station-closing station-container'>
        <h2 className='station-display'>
          {t('Keep a little room for your next idea.')}
        </h2>
        <Link
          to={props.isAuthenticated ? '/dashboard' : '/sign-in'}
          className='station-text-link'
        >
          {t(
            props.isAuthenticated ? 'Go to Dashboard' : 'Sign in with LinuxDO'
          )}
          <ArrowRight size={18} aria-hidden='true' />
        </Link>
      </section>
    </main>
  )
}
