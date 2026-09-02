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
import { act, fireEvent, render, waitFor } from '@testing-library/react'
import { afterEach, describe, expect, test, vi } from 'vitest'

import { Turnstile } from './turnstile'

afterEach(() => {
  delete window.turnstile
  document.querySelector('#cf-turnstile')?.remove()
})

describe('Turnstile', () => {
  test('passes the configured action and removes its widget on unmount', async () => {
    const renderWidget = vi.fn().mockReturnValue('widget-1')
    const removeWidget = vi.fn()
    const onVerify = vi.fn()
    const onExpire = vi.fn()
    const onError = vi.fn()
    window.turnstile = {
      render: renderWidget,
      remove: removeWidget,
    }

    const { unmount } = render(
      <Turnstile
        siteKey='site-key'
        action='linuxdo_login'
        onVerify={onVerify}
        onExpire={onExpire}
        onError={onError}
      />
    )

    await waitFor(() => expect(renderWidget).toHaveBeenCalledTimes(1))
    const options = renderWidget.mock.calls[0][1] as Record<string, unknown>
    expect(options).toMatchObject({
      sitekey: 'site-key',
      action: 'linuxdo_login',
    })

    act(() => {
      ;(options.callback as (token: string) => void)('verified-token')
      ;(options['error-callback'] as () => void)()
    })

    expect(onVerify).toHaveBeenCalledWith('verified-token')
    expect(onExpire).toHaveBeenCalledTimes(1)
    expect(onError).toHaveBeenCalledTimes(1)

    unmount()
    expect(removeWidget).toHaveBeenCalledWith('widget-1')
  })

  test('reports a script loading failure instead of silently continuing', async () => {
    const onError = vi.fn()
    render(
      <Turnstile
        siteKey='site-key'
        onVerify={() => undefined}
        onError={onError}
      />
    )

    const script = document.querySelector('#cf-turnstile')
    expect(script).toHaveAttribute(
      'src',
      'https://challenges.cloudflare.com/turnstile/v0/api.js?render=explicit'
    )
    fireEvent.error(script as HTMLScriptElement)

    await waitFor(() => expect(onError).toHaveBeenCalledTimes(1))
    expect(document.querySelector('#cf-turnstile')).not.toBeInTheDocument()
  })
})
