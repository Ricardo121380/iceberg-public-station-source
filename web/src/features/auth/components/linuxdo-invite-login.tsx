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
import { Loader2 } from 'lucide-react'
import { useState, type KeyboardEvent } from 'react'
import { useTranslation } from 'react-i18next'

import { IconLinuxDo } from '@/assets/brand-icons'
import { Turnstile } from '@/components/turnstile'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'

type LinuxDOInviteLoginProps = {
  siteKey: string
  action?: string
  disabled?: boolean
  loading?: boolean
  onStart: (input: {
    inviteCode?: string
    turnstileToken?: string
  }) => Promise<boolean>
}

export function LinuxDOInviteLogin({
  siteKey,
  action,
  disabled = false,
  loading = false,
  onStart,
}: LinuxDOInviteLoginProps) {
  const { t } = useTranslation()
  const [inviteCode, setInviteCode] = useState('')
  const [turnstileToken, setTurnstileToken] = useState('')
  const [widgetKey, setWidgetKey] = useState(0)
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [message, setMessage] = useState('')
  const [widgetFailed, setWidgetFailed] = useState(false)

  const blocked = disabled || loading || isSubmitting
  const canRegister = Boolean(
    !blocked && siteKey && inviteCode.trim() && turnstileToken
  )

  const resetWidget = () => {
    setTurnstileToken('')
    setWidgetFailed(false)
    setWidgetKey((current) => current + 1)
  }

  const handleDirectLogin = async () => {
    if (blocked) return

    setMessage('')
    setIsSubmitting(true)
    try {
      await onStart({})
    } finally {
      setIsSubmitting(false)
    }
  }

  const handleRegistration = async () => {
    if (blocked) return

    const trimmedInviteCode = inviteCode.trim()
    if (!trimmedInviteCode) {
      setMessage(t('Please enter an invitation code.'))
      return
    }
    if (!siteKey) {
      setMessage(t('Verification could not be loaded. Please retry.'))
      return
    }
    if (!turnstileToken) {
      setMessage(t('Please complete the security check to continue.'))
      return
    }

    const submittedToken = turnstileToken
    setMessage('')
    setIsSubmitting(true)
    resetWidget()

    try {
      const started = await onStart({
        inviteCode: trimmedInviteCode,
        turnstileToken: submittedToken,
      })
      if (started) setInviteCode('')
    } finally {
      setIsSubmitting(false)
    }
  }

  const handleInviteKeyDown = (event: KeyboardEvent<HTMLInputElement>) => {
    if (event.key !== 'Enter') return
    event.preventDefault()
    void handleRegistration()
  }

  return (
    <div
      className='space-y-3'
      role='group'
      aria-labelledby='linuxdo-login-title'
    >
      <div className='space-y-1'>
        <p id='linuxdo-login-title' className='text-sm font-medium'>
          {t('Continue with LinuxDO')}
        </p>
        <p className='text-muted-foreground text-xs'>
          {t(
            'Existing users can sign in directly. An invitation code is only required for first-time registration.'
          )}
        </p>
      </div>

      <Button
        type='button'
        disabled={blocked}
        onClick={() => void handleDirectLogin()}
        className='h-11 w-full justify-center gap-2 rounded-lg'
      >
        {isSubmitting || loading ? (
          <Loader2 className='h-4 w-4 animate-spin' />
        ) : (
          <IconLinuxDo className='h-4 w-4' aria-hidden='true' />
        )}
        {t('Continue with LinuxDO')}
      </Button>

      <div className='border-border border-t pt-3'>
        <p className='text-muted-foreground text-xs'>
          {t('First time here? Register with an invitation code.')}
        </p>
      </div>

      <div className='grid gap-2'>
        <Label htmlFor='linuxdo-invite-code'>{t('Invitation Code')}</Label>
        <Input
          id='linuxdo-invite-code'
          value={inviteCode}
          onChange={(event) => {
            setInviteCode(event.target.value)
            if (message) setMessage('')
          }}
          onKeyDown={handleInviteKeyDown}
          autoComplete='off'
          spellCheck={false}
          disabled={blocked}
          aria-describedby='linuxdo-invite-description'
          aria-invalid={Boolean(message)}
        />
        <p id='linuxdo-invite-description' className='sr-only'>
          {t(
            'Existing users can sign in directly. An invitation code is only required for first-time registration.'
          )}
        </p>
      </div>

      {siteKey ? (
        <div className='min-h-[65px]'>
          <Turnstile
            key={widgetKey}
            siteKey={siteKey}
            action={action}
            onVerify={(token) => {
              setTurnstileToken(token)
              setWidgetFailed(false)
              setMessage('')
            }}
            onExpire={() => {
              setTurnstileToken('')
              setMessage(t('Please complete the security check to continue.'))
            }}
            onError={() => {
              setTurnstileToken('')
              setWidgetFailed(true)
              setMessage(t('Verification could not be loaded. Please retry.'))
            }}
          />
        </div>
      ) : (
        <p className='text-destructive text-sm' role='alert'>
          {t('Verification could not be loaded. Please retry.')}
        </p>
      )}

      {message && (
        <p className='text-destructive text-sm' role='alert'>
          {message}
        </p>
      )}

      {widgetFailed && siteKey && (
        <Button
          type='button'
          variant='outline'
          size='sm'
          disabled={blocked}
          onClick={() => {
            setMessage('')
            resetWidget()
          }}
        >
          {t('Retry')}
        </Button>
      )}

      <Button
        type='button'
        variant='outline'
        disabled={!canRegister}
        onClick={() => void handleRegistration()}
        className='h-11 w-full justify-center gap-2 rounded-lg'
      >
        {isSubmitting || loading ? (
          <Loader2 className='h-4 w-4 animate-spin' />
        ) : (
          <IconLinuxDo className='h-4 w-4' aria-hidden='true' />
        )}
        {t('Register with LinuxDO')}
      </Button>
    </div>
  )
}
