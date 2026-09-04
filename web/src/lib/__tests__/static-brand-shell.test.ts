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
import { existsSync, readFileSync } from 'node:fs'
import path from 'node:path'

import { describe, expect, test } from 'vitest'

const shell = readFileSync(path.resolve(process.cwd(), 'index.html'), 'utf8')
const document = new DOMParser().parseFromString(shell, 'text/html')

describe('static HTML brand shell', () => {
  test('identifies the station before JavaScript loads while preserving New API attribution', () => {
    const title = document.title
    const metaTitle =
      document.querySelector<HTMLMetaElement>('meta[name="title"]')?.content
    const description = document.querySelector<HTMLMetaElement>(
      'meta[name="description"]'
    )?.content

    expect(title).toContain('冰山公益站')
    expect(title).toContain('New API')
    expect(metaTitle).toBe(title)
    expect(description).toContain('冰山公益站')
    expect(description).toContain('New API')
    expect(
      document.querySelector<HTMLMetaElement>('meta[name="application-name"]')
        ?.content
    ).toBe('冰山公益站')
    expect(
      document.querySelector<HTMLMetaElement>('meta[property="og:title"]')
        ?.content
    ).toBe(title)
    expect(
      document.querySelector<HTMLMetaElement>('meta[property="og:site_name"]')
        ?.content
    ).toBe('冰山公益站')
  })

  test('provides a no-script fallback with station and upstream attribution', () => {
    const fallback = document.querySelector('noscript')?.textContent

    expect(fallback).toContain('冰山公益站')
    expect(fallback).toContain('New API')
  })

  test('uses the station mark before JavaScript loads', () => {
    const favicon = document.querySelector<HTMLLinkElement>('link[rel="icon"]')
    const touchIcon = document.querySelector<HTMLLinkElement>(
      'link[rel="apple-touch-icon"]'
    )
    const faviconHref = favicon?.getAttribute('href')
    const touchIconHref = touchIcon?.getAttribute('href')

    expect(faviconHref).toBe('/iceberg-station-favicon-v2.png')
    expect(favicon?.getAttribute('sizes')).toBe('64x64')
    expect(touchIconHref).toBe('/apple-touch-icon-v2.png')
    expect(touchIcon?.getAttribute('sizes')).toBe('180x180')

    for (const href of [faviconHref, touchIconHref]) {
      expect(href).toBeTruthy()
      expect(
        existsSync(path.resolve(process.cwd(), 'public', href!.slice(1)))
      ).toBe(true)
    }
  })
})
