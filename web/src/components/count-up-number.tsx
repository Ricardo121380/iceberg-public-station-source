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
import { useReducedMotion } from 'motion/react'
import { useEffect, useRef, useState } from 'react'

/**
 * Renders a number with an exponential ease-out count-up: from zero on first
 * mount, and from the currently visible value on later changes. Honors
 * `prefers-reduced-motion` by rendering the final value immediately. The
 * caller owns formatting, so currency/quota strings stay exact.
 */
export function CountUpNumber(props: {
  value: number
  format?: (value: number) => string
  durationMs?: number
  className?: string
}) {
  const shouldReduce = useReducedMotion()
  const format = props.format ?? ((n: number) => String(Math.round(n)))
  const [display, setDisplay] = useState(() => (shouldReduce ? props.value : 0))
  const currentRef = useRef(shouldReduce ? props.value : 0)

  useEffect(() => {
    if (shouldReduce) {
      currentRef.current = props.value
      setDisplay(props.value)
      return
    }
    const from = currentRef.current
    const to = props.value
    if (from === to) return
    const duration = props.durationMs ?? 800
    // Date.now() (not the rAF timestamp) keeps the tween under one clock in
    // every environment, including test rigs with faked timers.
    const start = Date.now()
    let raf = 0
    const tick = () => {
      const t = Math.min((Date.now() - start) / duration, 1)
      const eased = t >= 1 ? 1 : 1 - Math.pow(2, -10 * t)
      currentRef.current = from + (to - from) * eased
      setDisplay(currentRef.current)
      if (t < 1) {
        raf = requestAnimationFrame(tick)
      } else {
        currentRef.current = to
      }
    }
    raf = requestAnimationFrame(tick)
    return () => cancelAnimationFrame(raf)
  }, [props.value, props.durationMs, shouldReduce])

  return <span className={props.className}>{format(display)}</span>
}
