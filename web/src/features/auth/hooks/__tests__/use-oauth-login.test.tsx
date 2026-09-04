/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published
by the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/
import { act, renderHook } from '@testing-library/react'
import { afterEach, describe, expect, test, vi } from 'vitest'

import { useAuthStore } from '@/stores/auth-store'

import type { OAuthStartResult } from '../../types'
import { useOAuthLogin } from '../use-oauth-login'

const mocks = vi.hoisted(() => ({
  createOAuthFlow: vi.fn(),
  logout: vi.fn(),
  toastError: vi.fn(),
}))

vi.mock('../../api', () => ({
  createOAuthFlow: mocks.createOAuthFlow,
  logout: mocks.logout,
  telegramLogin: vi.fn(),
}))

vi.mock('sonner', () => ({
  toast: { error: mocks.toastError },
}))

vi.mock('../use-auth-redirect', () => ({
  useAuthRedirect: () => ({ handleLoginSuccess: vi.fn() }),
}))

afterEach(() => {
  vi.clearAllMocks()
  useAuthStore.getState().auth.reset('complete')
})

describe('LinuxDO OAuth initialization', () => {
  test('keeps verification and reports Retry-After when OAuth state is rate limited', async () => {
    useAuthStore.getState().auth.reset('complete')
    mocks.createOAuthFlow.mockRejectedValue({
      isAxiosError: true,
      response: { status: 429, headers: { 'retry-after': '42' } },
    })
    const { result } = renderHook(() =>
      useOAuthLogin({
        linuxdo_client_id: 'linuxdo-client',
        registration_invite_required: true,
      })
    )

    let outcome: OAuthStartResult | undefined
    await act(async () => {
      outcome = await result.current.handleLinuxDOLogin({
        inviteCode: 'invite-code',
        turnstileToken: 'turnstile-token',
      })
    })

    expect(mocks.logout).not.toHaveBeenCalled()
    expect(outcome).toEqual({ started: false, preserveVerification: true })
    expect(mocks.toastError).toHaveBeenCalledWith(
      'Too many requests. Please try again in 42 seconds.'
    )
  })
})
