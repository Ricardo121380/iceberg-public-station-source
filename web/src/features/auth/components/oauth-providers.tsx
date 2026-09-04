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
import type { ReactNode } from 'react'
import { useTranslation } from 'react-i18next'

import {
  IconDiscord,
  IconGithub,
  IconLinuxDo,
  IconTelegram,
  IconWeChat,
} from '@/assets/brand-icons'
import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'

import { useOAuthLogin } from '../hooks/use-oauth-login'
import type { SystemStatus } from '../types'
import { LinuxDOInviteLogin } from './linuxdo-invite-login'
import { TelegramLoginDialog } from './telegram-login-dialog'

type OAuthProvidersProps = {
  status: SystemStatus | null
  flow?: 'sign-in' | 'sign-up'
  disabled?: boolean
  className?: string
  onWeChatLogin?: () => void
  isWeChatLoading?: boolean
  redirectTo?: string
}

type ProviderButton = {
  key: string
  label: string
  onClick: () => void
  icon?: ReactNode
  disabled?: boolean
}

export function OAuthProviders({
  status,
  flow = 'sign-in',
  disabled = false,
  className,
  onWeChatLogin,
  isWeChatLoading = false,
  redirectTo,
}: OAuthProvidersProps) {
  const { t } = useTranslation()
  const {
    isLoading,
    githubButtonText,
    githubButtonDisabled,
    handleGitHubLogin,
    handleDiscordLogin,
    handleOIDCLogin,
    handleLinuxDOLogin,
    handleTelegramLogin,
    handleCustomOAuthLogin,
    isTelegramDialogOpen,
    isTelegramPending,
    handleTelegramAuthorization,
    setIsTelegramDialogOpen,
  } = useOAuthLogin(status, redirectTo)

  const registrationInviteRequired = Boolean(
    status?.registration_invite_required ??
    status?.data?.registration_invite_required
  )
  const registrationTurnstileRequired = Boolean(
    status?.registration_turnstile_required ??
    status?.data?.registration_turnstile_required
  )
  const registrationTurnstileSiteKey = (
    status?.registration_turnstile_site_key ??
    status?.data?.registration_turnstile_site_key ??
    ''
  ).trim()
  const registrationTurnstileAction =
    status?.registration_turnstile_action ??
    status?.data?.registration_turnstile_action
  const providerButtons: ProviderButton[] = []

  if (!registrationInviteRequired && status?.wechat_login && onWeChatLogin) {
    providerButtons.push({
      key: 'wechat',
      label: t('Continue with WeChat'),
      onClick: onWeChatLogin,
      icon: <IconWeChat className='h-4 w-4' />,
      disabled: isWeChatLoading,
    })
  }

  if (!registrationInviteRequired && status?.github_oauth) {
    providerButtons.push({
      key: 'github',
      label: githubButtonText || t('Continue with GitHub'),
      onClick: handleGitHubLogin,
      icon: <IconGithub className='h-4 w-4' />,
      disabled: githubButtonDisabled,
    })
  }

  if (!registrationInviteRequired && status?.discord_oauth) {
    providerButtons.push({
      key: 'discord',
      label: t('Continue with Discord'),
      onClick: handleDiscordLogin,
      icon: <IconDiscord className='h-4 w-4' />,
    })
  }

  if (!registrationInviteRequired && status?.oidc_enabled) {
    const oidcDisplayName = status.oidc_display_name?.trim() || 'OIDC'
    providerButtons.push({
      key: 'oidc',
      label: t('Continue with {{name}}', {
        name: oidcDisplayName,
      }),
      onClick: handleOIDCLogin,
    })
  }

  if (!registrationInviteRequired && status?.linuxdo_oauth) {
    providerButtons.push({
      key: 'linuxdo',
      label: t('Continue with LinuxDO'),
      onClick: handleLinuxDOLogin,
      icon: <IconLinuxDo className='h-4 w-4' />,
    })
  }

  if (!registrationInviteRequired && status?.telegram_oauth) {
    providerButtons.push({
      key: 'telegram',
      label: t('Continue with Telegram'),
      onClick: handleTelegramLogin,
      icon: <IconTelegram data-icon='inline-start' />,
    })
  }

  // Custom OAuth providers
  const customProviders = status?.custom_oauth_providers
  if (
    !registrationInviteRequired &&
    customProviders &&
    customProviders.length > 0
  ) {
    for (const provider of customProviders) {
      providerButtons.push({
        key: `custom-${provider.slug}`,
        label: t('Continue with {{name}}', { name: provider.name }),
        onClick: () => handleCustomOAuthLogin(provider),
      })
    }
  }

  const linuxDOFlow =
    registrationInviteRequired && status?.linuxdo_oauth ? (
      <LinuxDOInviteLogin
        mode={flow}
        siteKey={
          registrationTurnstileRequired ? registrationTurnstileSiteKey : ''
        }
        action={registrationTurnstileAction}
        disabled={disabled}
        loading={isLoading}
        onStart={handleLinuxDOLogin}
      />
    ) : null

  if (providerButtons.length === 0 && !linuxDOFlow) return null

  return (
    <>
      <div className={cn('space-y-3', className)}>
        {linuxDOFlow}

        {providerButtons.length > 0 && (
          <>
            <div className='relative'>
              <div className='absolute inset-0 flex items-center'>
                <span className='w-full border-t' />
              </div>
              <div className='relative flex justify-center text-xs uppercase'>
                <span className='bg-background text-muted-foreground px-2'>
                  {t('Or continue with')}
                </span>
              </div>
            </div>

            <div className='flex flex-col gap-2'>
              {providerButtons.map(
                ({ key, label, onClick, icon, disabled: extraDisabled }) => (
                  <Button
                    key={key}
                    variant='outline'
                    type='button'
                    disabled={disabled || isLoading || extraDisabled}
                    onClick={onClick}
                    className='h-11 w-full justify-center gap-2 rounded-lg'
                  >
                    {icon}
                    {label}
                  </Button>
                )
              )}
            </div>
          </>
        )}
      </div>

      <TelegramLoginDialog
        open={isTelegramDialogOpen}
        botName={status?.telegram_bot_name ?? ''}
        pending={isTelegramPending}
        onOpenChange={setIsTelegramDialogOpen}
        onAuthorization={handleTelegramAuthorization}
      />
    </>
  )
}
