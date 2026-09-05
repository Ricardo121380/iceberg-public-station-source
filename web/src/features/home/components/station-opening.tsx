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

const seenKey = 'iceberg-opening-v1-seen'
const parts = ['boat', 'ice-left', 'ice-right', 'splash', 'chip'] as const

export function StationOpening() {
  const { t } = useTranslation()
  const [playing, setPlaying] = useState(false)

  useEffect(() => {
    const motion = window.matchMedia('(prefers-reduced-motion: reduce)')
    try {
      if (motion.matches || localStorage.getItem(seenKey)) return
    } catch {
      // An optional introduction must not interfere with restricted browsers.
      return
    }

    let cancelled = false
    let finishTimer: ReturnType<typeof setTimeout> | undefined
    const images: HTMLImageElement[] = []
    const dismiss = () => {
      cancelled = true
      setPlaying(false)
      clearTimeout(finishTimer)
      try {
        localStorage.setItem(seenKey, '1')
      } catch {
        /* Storage may become unavailable. */
      }
    }
    const loadTimer = setTimeout(() => {
      cancelled = true
    }, 1500)
    const ready = parts.map(
      (part) =>
        new Promise<void>((resolve, reject) => {
          const image = new Image()
          images.push(image)
          image.onload = () => resolve()
          image.onerror = () => reject(new Error('Opening artwork unavailable'))
          image.src = `/iceberg/intro/${part}.png`
        })
    )

    // Input is never intercepted: the same event continues to the real control.
    window.addEventListener('pointerdown', dismiss, {
      capture: true,
      once: true,
    })
    window.addEventListener('keydown', dismiss, { capture: true, once: true })
    motion.addEventListener('change', dismiss)
    Promise.all(ready)
      .then(() => {
        if (cancelled || motion.matches) return
        clearTimeout(loadTimer)
        try {
          localStorage.setItem(seenKey, '1')
        } catch {
          return
        }
        setPlaying(true)
        finishTimer = setTimeout(dismiss, 2400)
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
      motion.removeEventListener('change', dismiss)
    }
  }, [])

  if (!playing) return null

  return (
    <div className='station-opening' data-testid='station-opening'>
      <div className='opening-stage' aria-hidden='true'>
        <div className='opening-water' />
        <img
          className='opening-ice opening-ice-left'
          src='/iceberg/intro/ice-left.png'
          alt=''
        />
        <img
          className='opening-ice opening-ice-right'
          src='/iceberg/intro/ice-right.png'
          alt=''
        />
        <img className='opening-boat' src='/iceberg/intro/boat.png' alt='' />
        <img
          className='opening-splash'
          src='/iceberg/intro/splash.png'
          alt=''
        />
        <img
          className='opening-chip opening-chip-one'
          src='/iceberg/intro/chip.png'
          alt=''
        />
        <img
          className='opening-chip opening-chip-two'
          src='/iceberg/intro/chip.png'
          alt=''
        />
        <p className='opening-caption'>
          {t('A little boat meets a big iceberg.')}
        </p>
      </div>
      <button
        type='button'
        className='opening-skip'
        onClick={() => setPlaying(false)}
      >
        {t('Skip opening animation')}
      </button>
    </div>
  )
}
