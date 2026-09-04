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
import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import { describe, expect, test, vi } from 'vitest'

import { LinuxDOInviteLogin } from '../linuxdo-invite-login'

vi.mock('@/components/turnstile', () => ({
  Turnstile: ({ onVerify }: { onVerify: (token: string) => void }) => (
    <button type='button' onClick={() => onVerify('verified-token')}>
      Complete human check
    </button>
  ),
}))

describe('LinuxDOInviteLogin', () => {
  test('shows only direct LinuxDO authorization in sign-in mode', async () => {
    const onStart = vi.fn().mockResolvedValue(true)
    render(<LinuxDOInviteLogin mode='sign-in' onStart={onStart} />)

    expect(screen.queryByLabelText('Invitation Code')).not.toBeInTheDocument()
    expect(
      screen.queryByRole('button', { name: 'Register with LinuxDO' })
    ).not.toBeInTheDocument()

    const continueButton = screen.getByRole('button', {
      name: 'Continue with LinuxDO',
    })
    expect(continueButton).toBeEnabled()
    fireEvent.click(continueButton)

    await waitFor(() => expect(onStart).toHaveBeenCalledWith({}))
  })

  test('shows invite verification only in sign-up mode', async () => {
    const onStart = vi.fn().mockResolvedValue(true)
    render(
      <LinuxDOInviteLogin mode='sign-up' siteKey='site-key' onStart={onStart} />
    )

    expect(
      screen.queryByRole('button', { name: 'Continue with LinuxDO' })
    ).not.toBeInTheDocument()

    const inviteCode = screen.getByLabelText('Invitation Code')
    const registerButton = screen.getByRole('button', {
      name: 'Register with LinuxDO',
    })

    expect(registerButton).toBeDisabled()
    fireEvent.change(inviteCode, { target: { value: ' invite-code ' } })
    expect(registerButton).toBeDisabled()

    fireEvent.click(
      screen.getByRole('button', { name: 'Complete human check' })
    )
    await waitFor(() => expect(registerButton).toBeEnabled())

    fireEvent.keyDown(inviteCode, { key: 'Enter' })
    await waitFor(() =>
      expect(onStart).toHaveBeenCalledWith({
        inviteCode: 'invite-code',
        turnstileToken: 'verified-token',
      })
    )
    expect(inviteCode).toHaveValue('')
  })

  test('keeps the invitation code when OAuth state creation does not start', async () => {
    const onStart = vi.fn().mockResolvedValue(false)
    render(
      <LinuxDOInviteLogin mode='sign-up' siteKey='site-key' onStart={onStart} />
    )

    const inviteCode = screen.getByLabelText('Invitation Code')
    fireEvent.change(inviteCode, { target: { value: 'keep-this-code' } })
    fireEvent.click(
      screen.getByRole('button', { name: 'Complete human check' })
    )
    fireEvent.click(
      screen.getByRole('button', { name: 'Register with LinuxDO' })
    )

    await waitFor(() => expect(onStart).toHaveBeenCalledTimes(1))
    expect(inviteCode).toHaveValue('keep-this-code')
  })

  test('prevents repeated direct login while OAuth state creation is pending', async () => {
    let resolveStart: (started: boolean) => void = () => undefined
    const onStart = vi.fn(
      () =>
        new Promise<boolean>((resolve) => {
          resolveStart = resolve
        })
    )
    render(<LinuxDOInviteLogin mode='sign-in' onStart={onStart} />)

    const continueButton = screen.getByRole('button', {
      name: 'Continue with LinuxDO',
    })
    fireEvent.click(continueButton)
    fireEvent.click(continueButton)

    expect(onStart).toHaveBeenCalledTimes(1)
    resolveStart(true)
    await waitFor(() => expect(continueButton).toBeEnabled())
  })
})
