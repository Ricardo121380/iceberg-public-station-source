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
import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'

import '@/styles/station-opening.css'

import { getOpeningHoliday } from '../lib/opening-holidays'

const parts = ['boat', 'ice-left', 'ice-right', 'splash', 'chip'] as const

export function StationOpening() {
  const { t } = useTranslation()
  const [playing, setPlaying] = useState(false)
  const [theme] = useState(() => getOpeningHoliday())
  const [useHoliday, setUseHoliday] = useState(Boolean(theme.holiday))
  const assetRoot = useHoliday
    ? `/iceberg/intro/holidays-v1/${theme.holiday}`
    : '/iceberg/intro/vector-v1'

  useEffect(() => {
    let cancelled = false
    let finishTimer: ReturnType<typeof setTimeout> | undefined
    const images: HTMLImageElement[] = []
    const dismiss = () => {
      cancelled = true
      setPlaying(false)
      clearTimeout(finishTimer)
    }
    const loadTimer = setTimeout(() => {
      cancelled = true
    }, 1500)
    const preload = (root: string, names: readonly string[]) =>
      Promise.all(
        names.map(
          (part) =>
            new Promise<void>((resolve, reject) => {
              const image = new Image()
              images.push(image)
              image.onload = () => resolve()
              image.onerror = () =>
                reject(new Error('Opening artwork unavailable'))
              image.src = `${root}/${part}.svg`
            })
        )
      )
    const ordinaryRoot = '/iceberg/intro/vector-v1'
    const ready = theme.holiday
      ? preload(`/iceberg/intro/holidays-v1/${theme.holiday}`, [
          ...parts,
          'backdrop',
        ]).catch(() => {
          if (cancelled) return
          setUseHoliday(false)
          return preload(ordinaryRoot, parts)
        })
      : preload(ordinaryRoot, parts)

    // Input is never intercepted: the same event continues to the real control.
    window.addEventListener('pointerdown', dismiss, {
      capture: true,
      once: true,
    })
    window.addEventListener('keydown', dismiss, { capture: true, once: true })
    ready
      .then(() => {
        if (cancelled) return
        clearTimeout(loadTimer)
        setPlaying(true)
        finishTimer = setTimeout(dismiss, 3400)
      })
      .catch(() => {
        cancelled = true
      })

    return () => {
      cancelled = true
      clearTimeout(loadTimer)
      clearTimeout(finishTimer)
      images.forEach((image) => {
        image.onload = null
        image.onerror = null
      })
      window.removeEventListener('pointerdown', dismiss, true)
      window.removeEventListener('keydown', dismiss, true)
    }
  }, [theme.holiday])

  if (!playing) return null

  return (
    <div
      className='station-opening'
      data-testid='station-opening'
      data-holiday={useHoliday ? theme.holiday : undefined}
    >
      <div className='opening-stage' aria-hidden='true'>
        {useHoliday && (
          <img
            className='opening-holiday-backdrop'
            src={`${assetRoot}/backdrop.svg`}
            alt=''
          />
        )}
        {useHoliday && theme.holiday === 'new-year' && (
          <span className='opening-holiday-year'>{theme.year}</span>
        )}
        <div className='opening-water' />
        <img
          className='opening-ice opening-ice-left'
          src={`${assetRoot}/ice-left.svg`}
          alt=''
        />
        <img
          className='opening-ice opening-ice-right'
          src={`${assetRoot}/ice-right.svg`}
          alt=''
        />
        <img className='opening-boat' src={`${assetRoot}/boat.svg`} alt='' />
        <img
          className='opening-splash'
          src={`${assetRoot}/splash.svg`}
          alt=''
        />
        <img
          className='opening-chip opening-chip-one'
          src={`${assetRoot}/chip.svg`}
          alt=''
        />
        <img
          className='opening-chip opening-chip-two'
          src={`${assetRoot}/chip.svg`}
          alt=''
        />
        <p className='opening-caption'>
          {t('A little boat meets a big iceberg.')}
        </p>
      </div>
    </div>
  )
}
