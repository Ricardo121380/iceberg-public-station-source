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
import { useRouterState } from '@tanstack/react-router'
import { useEffect } from 'react'
import { useTranslation } from 'react-i18next'

import { useSystemConfig } from '@/hooks/use-system-config'
import { toIntlLocale } from '@/i18n/languages'

/**
 * Per-route document titles for the public surfaces and `<html lang>`
 * synchronization with the active interface language. The app bootstrap only
 * sets the site name once, so every route shared one static title and the
 * served `lang` attribute could contradict the rendered language.
 */
const ROUTE_TITLE_KEYS: Record<string, string> = {
  '/': 'Home',
  '/pricing': 'Model Square',
  '/rankings': 'Rankings',
  '/sign-in': 'Sign in',
  '/sign-up': 'Sign up',
  '/privacy-policy': 'Privacy Policy',
  '/user-agreement': 'User Agreement',
  '/about': 'About',
}

export function RouteDocumentMeta() {
  const { i18n, t } = useTranslation()
  const { systemName, loading } = useSystemConfig()
  const pathname = useRouterState({ select: (state) => state.location.pathname })

  useEffect(() => {
    // While the station config is still loading, the static index.html title
    // holds; writing now would flash the upstream default name.
    if (loading) return
    const titleKey = ROUTE_TITLE_KEYS[pathname]
    document.title = titleKey
      ? `${t(titleKey)} · ${systemName}`
      : systemName
  }, [pathname, systemName, loading, t])

  useEffect(() => {
    const applyLanguage = (language: string) => {
      document.documentElement.lang = toIntlLocale(language) ?? 'en'
    }
    applyLanguage(i18n.language)
    i18n.on('languageChanged', applyLanguage)
    return () => {
      i18n.off('languageChanged', applyLanguage)
    }
  }, [i18n])

  return null
}
