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
import { useEffect, useRef } from 'react'

declare global {
  interface Window {
    turnstile?: {
      render: (element: HTMLElement, options: Record<string, unknown>) => string
      remove?: (widgetId: string) => void
    }
  }
}

const turnstileScriptId = 'cf-turnstile'
const turnstileScriptSource =
  'https://challenges.cloudflare.com/turnstile/v0/api.js?render=explicit'

let turnstileScriptPromise: Promise<void> | null = null

function loadTurnstileScript(): Promise<void> {
  if (window.turnstile) return Promise.resolve()
  if (turnstileScriptPromise) return turnstileScriptPromise

  let script = document.querySelector<HTMLScriptElement>(
    `#${turnstileScriptId}`
  )
  if (!script) {
    script = document.createElement('script')
    script.id = turnstileScriptId
    script.src = turnstileScriptSource
    script.async = true
    script.defer = true
    document.head.appendChild(script)
  }

  turnstileScriptPromise = new Promise((resolve, reject) => {
    const onLoad = () => {
      if (!window.turnstile) {
        turnstileScriptPromise = null
        script.remove()
        reject(new Error('Turnstile did not initialize'))
        return
      }
      resolve()
    }
    const onError = () => {
      turnstileScriptPromise = null
      script.remove()
      reject(new Error('Turnstile failed to load'))
    }

    script.addEventListener('load', onLoad, { once: true })
    script.addEventListener('error', onError, { once: true })
  })

  return turnstileScriptPromise
}

interface TurnstileProps {
  siteKey: string
  onVerify: (token: string) => void
  onExpire?: () => void
  onError?: () => void
  action?: string
  className?: string
}

export function Turnstile({
  siteKey,
  onVerify,
  onExpire,
  onError,
  action,
  className,
}: TurnstileProps) {
  const ref = useRef<HTMLDivElement | null>(null)
  const widgetIdRef = useRef<string | null>(null)
  const onVerifyRef = useRef(onVerify)
  const onExpireRef = useRef(onExpire)
  const onErrorRef = useRef(onError)

  useEffect(() => {
    onVerifyRef.current = onVerify
    onExpireRef.current = onExpire
    onErrorRef.current = onError
  }, [onError, onExpire, onVerify])

  useEffect(() => {
    let disposed = false

    const reportError = () => {
      onExpireRef.current?.()
      onErrorRef.current?.()
    }

    const render = () => {
      if (disposed || !ref.current || !window.turnstile) {
        if (!disposed) reportError()
        return
      }
      try {
        widgetIdRef.current = window.turnstile.render(ref.current, {
          sitekey: siteKey,
          ...(action ? { action } : {}),
          callback: (token: string) => onVerifyRef.current(token),
          'error-callback': reportError,
          'expired-callback': () => onExpireRef.current?.(),
        })
      } catch {
        reportError()
      }
    }

    void loadTurnstileScript().then(render).catch(reportError)

    return () => {
      disposed = true
      if (widgetIdRef.current) {
        window.turnstile?.remove?.(widgetIdRef.current)
        widgetIdRef.current = null
      }
    }
  }, [action, siteKey])

  return <div ref={ref} className={className} />
}
