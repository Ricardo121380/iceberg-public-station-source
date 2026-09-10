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

import type { OAuthStartResult } from '../types'

type LinuxDOInviteLoginProps = {
  mode?: 'sign-in' | 'sign-up'
  siteKey?: string
  action?: string
  disabled?: boolean
  loading?: boolean
  onStart: (input: {
    inviteCode?: string
    turnstileToken?: string
  }) => Promise<OAuthStartResult>
}

export function LinuxDOInviteLogin({
  mode = 'sign-in',
  siteKey = '',
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

  // The disabled register button explains what is still missing.
  let missingRequirement = ''
  if (!inviteCode.trim()) {
    missingRequirement = t('Enter your invitation code to continue.')
  } else if (siteKey && !turnstileToken) {
    missingRequirement = t('Complete the security check above to continue.')
  }

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

    try {
      const result = await onStart({
        inviteCode: trimmedInviteCode,
        turnstileToken: submittedToken,
      })
      if (result.started) {
        setInviteCode('')
      } else if (!result.preserveVerification) {
        resetWidget()
      }
    } finally {
      setIsSubmitting(false)
    }
  }

  const handleInviteKeyDown = (event: KeyboardEvent<HTMLInputElement>) => {
    if (event.key !== 'Enter') return
    event.preventDefault()
    void handleRegistration()
  }

  if (mode === 'sign-in') {
    return (
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
    )
  }

  return (
    <div
      className='space-y-3'
      role='group'
      aria-labelledby='linuxdo-registration-title'
    >
      <div className='space-y-1'>
        <p id='linuxdo-registration-title' className='text-sm font-medium'>
          {t('Register with LinuxDO')}
        </p>
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
          placeholder={t('Paste the invitation code here')}
          aria-describedby='linuxdo-invite-description linuxdo-invite-guidance'
          aria-invalid={Boolean(message)}
        />
        <p
          id='linuxdo-invite-guidance'
          className='text-muted-foreground text-xs'
        >
          {t(
            'Invitation-only station — ask a friend who is already aboard for a code.'
          )}
        </p>
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
        aria-describedby={
          !canRegister && missingRequirement
            ? 'linuxdo-register-requirement'
            : undefined
        }
      >
        {isSubmitting || loading ? (
          <Loader2 className='h-4 w-4 animate-spin' />
        ) : (
          <IconLinuxDo className='h-4 w-4' aria-hidden='true' />
        )}
        {t('Register with LinuxDO')}
      </Button>
      {!canRegister && !blocked && missingRequirement && (
        <p
          id='linuxdo-register-requirement'
          className='text-muted-foreground text-center text-xs'
        >
          {missingRequirement}
        </p>
      )}
    </div>
  )
}
