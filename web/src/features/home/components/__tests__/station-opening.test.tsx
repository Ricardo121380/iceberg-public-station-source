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
import { act, fireEvent, render, screen } from '@testing-library/react'
import { afterEach, beforeEach, describe, expect, test, vi } from 'vitest'

import { StationOpening } from '../station-opening'

const images: Array<{
  onload: (() => void) | null
  onerror: (() => void) | null
}> = []

beforeEach(() => {
  localStorage.clear()
  images.length = 0
  vi.useFakeTimers()
  vi.stubGlobal(
    'Image',
    class {
      onload: (() => void) | null = null
      onerror: (() => void) | null = null
      src = ''
      constructor() {
        images.push(this)
      }
    }
  )
})
afterEach(() => {
  vi.useRealTimers()
  vi.unstubAllGlobals()
})

async function finishLoading() {
  await act(async () => {
    images.forEach((image) => image.onload?.())
  })
}

describe('Homepage opening', () => {
  test('plays on every homepage mount even with a legacy seen marker', async () => {
    localStorage.setItem('iceberg-opening-v1-seen', '1')
    const view = render(<StationOpening />)
    expect(screen.queryByTestId('station-opening')).not.toBeInTheDocument()
    await finishLoading()
    expect(screen.getByTestId('station-opening')).toBeVisible()
    act(() => vi.advanceTimersByTime(2400))
    expect(screen.getByTestId('station-opening')).toBeVisible()
    act(() => vi.advanceTimersByTime(1000))
    expect(screen.queryByTestId('station-opening')).not.toBeInTheDocument()
    view.unmount()
    render(<StationOpening />)
    await finishLoading()
    expect(screen.getByTestId('station-opening')).toBeVisible()
  })

  test('clicking a real page action dismisses the intro without preventing the action', async () => {
    const onAction = vi.fn()
    render(
      <>
        <StationOpening />
        <button type='button' onClick={onAction}>
          Open console
        </button>
      </>
    )
    await finishLoading()
    const action = screen.getByRole('button', { name: 'Open console' })
    fireEvent.pointerDown(action)
    fireEvent.click(action)
    expect(onAction).toHaveBeenCalledOnce()
    expect(screen.queryByTestId('station-opening')).not.toBeInTheDocument()
  })

  test('a keyboard action dismisses the opening without cancelling the event', async () => {
    render(<StationOpening />)
    await finishLoading()
    expect(fireEvent.keyDown(window, { key: 'Tab' })).toBe(true)
    expect(screen.queryByTestId('station-opening')).not.toBeInTheDocument()
  })

  test('never starts the opening when reduced motion is requested', async () => {
    const original = window.matchMedia
    vi.spyOn(window, 'matchMedia').mockImplementation((query) => ({
      ...original(query),
      matches: true,
    }))
    render(<StationOpening />)
    // The reduced-motion path returns before preloading any artwork.
    expect(images).toHaveLength(0)
    await finishLoading()
    act(() => vi.advanceTimersByTime(5000))
    expect(screen.queryByTestId('station-opening')).not.toBeInTheDocument()
  })

  test('failed artwork leaves the page unobstructed', async () => {
    render(<StationOpening />)
    await act(async () => {
      images[0].onerror?.()
    })
    expect(screen.queryByTestId('station-opening')).not.toBeInTheDocument()
  })

  test('slow artwork never opens an animation late over an already-used page', async () => {
    render(<StationOpening />)
    act(() => vi.advanceTimersByTime(1500))
    await finishLoading()
    expect(screen.queryByTestId('station-opening')).not.toBeInTheDocument()
  })

  test('storage restrictions do not prevent playback', async () => {
    vi.spyOn(localStorage, 'getItem').mockImplementation(() => {
      throw new Error('blocked')
    })
    render(<StationOpening />)
    await finishLoading()
    expect(screen.getByTestId('station-opening')).toBeVisible()
  })
})
