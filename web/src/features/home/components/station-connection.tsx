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
import { Check, Copy, Terminal } from 'lucide-react'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'

import { Tabs, TabsList, TabsTrigger, TabsContent } from '@/components/ui/tabs'

const stationBaseUrl = 'https://iceberg.tiktok.vip/v1'
const requestExample = `export ICEBERG_API_KEY='YOUR_API_KEY'
curl ${stationBaseUrl}/models \\
  -H "Authorization: Bearer $ICEBERG_API_KEY"`

export function StationConnection() {
  const { t } = useTranslation()
  const [copyState, setCopyState] = useState<'idle' | 'success' | 'error'>(
    'idle'
  )

  const copyAddress = async () => {
    try {
      await navigator.clipboard.writeText(stationBaseUrl)
      setCopyState('success')
    } catch {
      setCopyState('error')
    }
  }

  return (
    <div className='station-connection'>
      <div className='station-connection-intro'>
        <Terminal size={24} strokeWidth={1.5} aria-hidden='true' />
        <h3>{t('Your tools. One starting point.')}</h3>
        <p>
          {t(
            'Use your own API key. Choose a model available to your key group.'
          )}
        </p>
        <label htmlFor='station-base-url'>API Base URL</label>
        <div className='station-address'>
          <input
            id='station-base-url'
            readOnly
            value={stationBaseUrl}
            onFocus={(event) => event.target.select()}
          />
          <button
            type='button'
            onClick={() => void copyAddress()}
            aria-label={t('Copy API address')}
          >
            {copyState === 'success' ? (
              <Check size={17} aria-hidden='true' />
            ) : (
              <Copy size={17} aria-hidden='true' />
            )}
          </button>
        </div>
        <p className='station-copy-status' role='status'>
          {copyState === 'success' && t('Address copied.')}
          {copyState === 'error' &&
            t('Could not copy. Select the address and copy it manually.')}
        </p>
      </div>
      <Tabs defaultValue='cherry' className='station-client-tabs'>
        <TabsList aria-label={t('Connection guides')}>
          <TabsTrigger value='cherry'>Cherry Studio</TabsTrigger>
          <TabsTrigger value='ccswitch'>CC Switch</TabsTrigger>
          <TabsTrigger value='curl'>cURL</TabsTrigger>
        </TabsList>
        <TabsContent value='cherry'>
          <h4>{t('Set up an OpenAI-compatible provider')}</h4>
          <p>
            {t(
              'In your client provider settings, enter the API Base URL above and a key created in the station console.'
            )}
          </p>
          <p>
            {t(
              'Fetch the model list, or copy an available model ID from the catalog. Select it to start a conversation.'
            )}
          </p>
          <a
            href='https://docs.cherry-ai.com/'
            target='_blank'
            rel='noopener noreferrer'
          >
            {t('Cherry Studio documentation')}
          </a>
        </TabsContent>
        <TabsContent value='ccswitch'>
          <h4>{t('Match the client and API protocol')}</h4>
          <p>
            {t(
              'Choose the app you want to configure in CC Switch. Check the model catalog for the interface supported by your selected model.'
            )}
          </p>
          <p>
            {t(
              'OpenAI-compatible clients use the Base URL above. Other protocols may need a different endpoint; follow the client configuration guide.'
            )}
          </p>
          <a
            href='https://github.com/farion1231/cc-switch#readme'
            target='_blank'
            rel='noopener noreferrer'
          >
            {t('CC Switch documentation')}
          </a>
        </TabsContent>
        <TabsContent value='curl'>
          <h4>{t('Check the models available to your key')}</h4>
          <p>
            {t(
              'Replace YOUR_API_KEY locally with your own key. This example requests the model list.'
            )}
          </p>
          <pre tabIndex={0} aria-label={t('Model list request example')}>
            <code>{requestExample}</code>
          </pre>
        </TabsContent>
      </Tabs>
    </div>
  )
}
