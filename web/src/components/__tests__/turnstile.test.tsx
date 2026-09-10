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
import { render, waitFor } from '@testing-library/react'
import { afterEach, describe, expect, test, vi } from 'vitest'

import { Turnstile } from '../turnstile'

const renderMock = vi.fn().mockReturnValue('widget-1')
const removeMock = vi.fn()

function stubTurnstileGlobal() {
  window.turnstile = { render: renderMock, remove: removeMock }
}

afterEach(() => {
  renderMock.mockClear()
  removeMock.mockClear()
  delete window.turnstile
})

describe('Turnstile', () => {
  test('renders the widget with the active theme and interface language', async () => {
    stubTurnstileGlobal()
    render(<Turnstile siteKey='site-key' onVerify={vi.fn()} />)

    await waitFor(() => expect(renderMock).toHaveBeenCalledTimes(1))
    const options = renderMock.mock.calls[0][1]
    expect(options.sitekey).toBe('site-key')
    // Test environment defaults: light theme, English UI.
    expect(options.theme).toBe('light')
    expect(options.language).toBe('en')
  })

  test('maps the Traditional Chinese UI code to the Turnstile locale', async () => {
    stubTurnstileGlobal()
    const { default: i18n } = await import('i18next')
    await i18n.changeLanguage('zhTW')
    render(<Turnstile siteKey='site-key' onVerify={vi.fn()} />)

    await waitFor(() => expect(renderMock).toHaveBeenCalledTimes(1))
    expect(renderMock.mock.calls[0][1].language).toBe('zh-TW')
    await i18n.changeLanguage('en')
  })

  test('removes the widget on unmount', async () => {
    stubTurnstileGlobal()
    const view = render(<Turnstile siteKey='site-key' onVerify={vi.fn()} />)
    await waitFor(() => expect(renderMock).toHaveBeenCalledTimes(1))

    view.unmount()
    expect(removeMock).toHaveBeenCalledWith('widget-1')
  })
})
