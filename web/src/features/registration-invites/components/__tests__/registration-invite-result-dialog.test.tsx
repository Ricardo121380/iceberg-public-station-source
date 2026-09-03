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
import { fireEvent, render, screen } from '@testing-library/react'
import { useState } from 'react'
import { describe, expect, test } from 'vitest'

import type { CreatedRegistrationInvite } from '../../types'
import { RegistrationInviteResultDialog } from '../registration-invite-result-dialog'

const generatedInvite: CreatedRegistrationInvite = {
  id: 1,
  code: 'test-invitation-code-one-time',
  code_prefix: 'test',
  expires_at: 1_800_000_000,
}

function ResultDialogHarness() {
  const [invites, setInvites] = useState<CreatedRegistrationInvite[] | null>([
    generatedInvite,
  ])

  return invites ? (
    <RegistrationInviteResultDialog
      invites={invites}
      onDiscard={() => setInvites(null)}
    />
  ) : (
    <p>Codes cleared</p>
  )
}

describe('registration invitation result dialog', () => {
  test('clears raw codes only after the discard confirmation is accepted', () => {
    render(<ResultDialogHarness />)

    expect(screen.getByLabelText('Generated invitation codes')).toHaveValue(
      generatedInvite.code
    )

    fireEvent.click(screen.getByRole('button', { name: 'I saved the codes' }))
    expect(screen.getByRole('alertdialog')).toHaveTextContent(
      'Discard invitation codes?'
    )
    expect(screen.getByText(generatedInvite.code)).toBeInTheDocument()

    fireEvent.click(screen.getByRole('button', { name: 'Discard codes' }))

    expect(screen.getByText('Codes cleared')).toBeInTheDocument()
    expect(document.body).not.toHaveTextContent(generatedInvite.code)
  })
})
