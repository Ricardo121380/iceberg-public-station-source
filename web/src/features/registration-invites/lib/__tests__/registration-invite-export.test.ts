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

import type { CreatedRegistrationInvite } from '../../types'
import {
  getRegistrationInviteCodesCsv,
  getRegistrationInviteCodesText,
} from '../registration-invite-export'

const invites: CreatedRegistrationInvite[] = [
  {
    id: 1,
    code: 'test-invitation-code-1',
    code_prefix: 'test',
    expires_at: 1_800_000_000,
  },
  {
    id: 2,
    code: 'test-invitation-code-2',
    code_prefix: 'test',
    expires_at: 1_800_000_001,
  },
]

describe('registration invitation exports', () => {
  test('exports only the generated codes as UTF-8 text rows', () => {
    expect(getRegistrationInviteCodesText(invites)).toBe(
      'test-invitation-code-1\ntest-invitation-code-2'
    )
  })

  test('exports generated codes as CSV without note or user context', () => {
    expect(getRegistrationInviteCodesCsv(invites)).toBe(
      '"code","code_prefix","expires_at"\r\n"test-invitation-code-1","test","1800000000"\r\n"test-invitation-code-2","test","1800000001"\r\n'
    )
  })
})
