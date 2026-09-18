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
import { describe, expect, test } from 'vitest'

import { getOpeningHoliday } from '../opening-holidays'

describe('Opening holiday calendar in Beijing time', () => {
  test.each([
    ['2026-09-17T15:59:59.999Z', null],
    ['2026-09-17T16:00:00Z', 'mid-autumn'],
    ['2026-09-27T15:59:59.999Z', 'mid-autumn'],
    ['2026-09-27T16:00:00Z', 'national-day'],
    ['2026-10-03T15:59:59.999Z', 'national-day'],
    ['2026-10-03T16:00:00Z', null],
    ['2027-02-06T04:00:00Z', 'spring-festival'],
    ['2027-06-09T04:00:00Z', 'dragon-boat'],
    ['2028-10-01T04:00:00Z', 'mid-autumn'],
    ['2026-07-01T04:00:00Z', null],
  ] as const)('%s selects %s', (date, holiday) => {
    expect(getOpeningHoliday(new Date(date)).holiday).toBe(holiday)
  })

  test('December uses the upcoming New Year on the decoration', () => {
    expect(getOpeningHoliday(new Date('2026-12-24T16:00:00Z'))).toEqual({
      holiday: 'new-year',
      year: 2027,
    })
    expect(
      getOpeningHoliday(new Date('2027-01-03T16:00:00Z')).holiday
    ).toBeNull()
  })

  test('calendar choice is independent of the visitor timezone', () => {
    expect(getOpeningHoliday(new Date('2026-09-18T00:00:00+08:00'))).toEqual(
      getOpeningHoliday(new Date('2026-09-17T09:00:00-07:00'))
    )
  })

  test('unsupported lunar years safely keep the ordinary opening', () => {
    expect(
      getOpeningHoliday(new Date('2036-02-01T04:00:00Z')).holiday
    ).toBeNull()
  })
})
