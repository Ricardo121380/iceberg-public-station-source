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
import { Plus } from 'lucide-react'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'

import { SectionPageLayout } from '@/components/layout'
import { Button } from '@/components/ui/button'

import { RegistrationInviteCreateDialog } from './components/registration-invite-create-dialog'
import { RegistrationInviteResultDialog } from './components/registration-invite-result-dialog'
import { RegistrationInviteRevokeDialog } from './components/registration-invite-revoke-dialog'
import { RegistrationInvitesTable } from './components/registration-invites-table'
import type { CreatedRegistrationInvite, RegistrationInvite } from './types'

export function RegistrationInvites() {
  const { t } = useTranslation()
  const [createDialogOpen, setCreateDialogOpen] = useState(false)
  const [createdInvites, setCreatedInvites] = useState<
    CreatedRegistrationInvite[] | null
  >(null)
  const [inviteToRevoke, setInviteToRevoke] =
    useState<RegistrationInvite | null>(null)
  const [refreshKey, setRefreshKey] = useState(0)

  const refresh = () => setRefreshKey((current) => current + 1)

  const handleCreated = (invites: CreatedRegistrationInvite[]) => {
    setCreatedInvites(invites)
    refresh()
  }

  return (
    <>
      <SectionPageLayout fixedContent>
        <SectionPageLayout.Title>
          {t('Registration Invites')}
        </SectionPageLayout.Title>
        <SectionPageLayout.Actions>
          <Button
            type='button'
            size='sm'
            onClick={() => setCreateDialogOpen(true)}
          >
            <Plus />
            {t('Create invitation codes')}
          </Button>
        </SectionPageLayout.Actions>
        <SectionPageLayout.Content>
          <RegistrationInvitesTable
            refreshKey={refreshKey}
            onCreate={() => setCreateDialogOpen(true)}
            onRevoke={setInviteToRevoke}
          />
        </SectionPageLayout.Content>
      </SectionPageLayout>

      <RegistrationInviteCreateDialog
        open={createDialogOpen}
        onOpenChange={setCreateDialogOpen}
        onCreated={handleCreated}
      />
      {createdInvites ? (
        <RegistrationInviteResultDialog
          invites={createdInvites}
          onDiscard={() => setCreatedInvites(null)}
        />
      ) : null}
      <RegistrationInviteRevokeDialog
        invite={inviteToRevoke}
        onOpenChange={(open) => !open && setInviteToRevoke(null)}
        onRefresh={refresh}
      />
    </>
  )
}
