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
export type OpeningHoliday =
  | 'mid-autumn'
  | 'national-day'
  | 'new-year'
  | 'spring-festival'
  | 'dragon-boat'

// Spring Festival, Dragon Boat, Mid-Autumn, verified against HKO tables.
// Source: https://www.hko.gov.hk/tc/gts/time/conversion.htm
// Coverage is explicit: unknown lunar years use the ordinary opening.
const lunarDates: Record<number, readonly [string, string, string]> = {
  2026: ['2026-02-17', '2026-06-19', '2026-09-25'],
  2027: ['2027-02-06', '2027-06-09', '2027-09-15'],
  2028: ['2028-01-26', '2028-05-28', '2028-10-03'],
  2029: ['2029-02-13', '2029-06-16', '2029-09-22'],
  2030: ['2030-02-03', '2030-06-05', '2030-09-12'],
  2031: ['2031-01-23', '2031-06-24', '2031-10-01'],
  2032: ['2032-02-11', '2032-06-12', '2032-09-19'],
  2033: ['2033-01-31', '2033-06-01', '2033-09-08'],
  2034: ['2034-02-19', '2034-06-20', '2034-09-27'],
  2035: ['2035-02-08', '2035-06-10', '2035-09-16'],
}

const dayMs = 86_400_000
const beijingOffsetMs = 8 * 60 * 60 * 1000

export function getOpeningHoliday(now = new Date()): {
  holiday: OpeningHoliday | null
  year: number
} {
  const beijing = new Date(now.getTime() + beijingOffsetMs)
  const year = beijing.getUTCFullYear()
  const today = Date.UTC(year, beijing.getUTCMonth(), beijing.getUTCDate())
  const candidates: Array<{
    holiday: OpeningHoliday
    date: string
    year: number
  }> = []
  // Include next year's January 1 so the December lead-in gets the correct year.
  for (const eventYear of [year - 1, year, year + 1]) {
    const lunar = lunarDates[eventYear]
    if (lunar) {
      candidates.push(
        { holiday: 'mid-autumn', date: lunar[2], year: eventYear },
        { holiday: 'spring-festival', date: lunar[0], year: eventYear },
        { holiday: 'dragon-boat', date: lunar[1], year: eventYear }
      )
    }
    candidates.push(
      { holiday: 'national-day', date: `${eventYear}-10-01`, year: eventYear },
      { holiday: 'new-year', date: `${eventYear}-01-01`, year: eventYear }
    )
  }
  const matches = candidates.filter(({ date }) => {
    const event = Date.parse(`${date}T00:00:00Z`)
    return today >= event - 7 * dayMs && today < event + 3 * dayMs
  })
  const selected =
    matches.find(({ holiday }) => holiday === 'mid-autumn') ?? matches[0]
  return selected
    ? { holiday: selected.holiday, year: selected.year }
    : { holiday: null, year }
}
