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
import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'

import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog'
import { handleServerError } from '@/lib/handle-server-error'

import { revokeRegistrationInvite } from '../api'
import type { RegistrationInvite } from '../types'

type RegistrationInviteRevokeDialogProps = {
  invite: RegistrationInvite | null
  onOpenChange: (open: boolean) => void
  onRefresh: () => void
}

export function RegistrationInviteRevokeDialog(
  props: RegistrationInviteRevokeDialogProps
) {
  const { t } = useTranslation()
  const [isRevoking, setIsRevoking] = useState(false)

  const handleRevoke = async () => {
    if (!props.invite) return

    setIsRevoking(true)
    try {
      const response = await revokeRegistrationInvite(props.invite.id)
      if (!response.success) {
        toast.error(response.message || t('Failed to revoke invitation code'))
        return
      }

      toast.success(
        response.data?.revoked
          ? t('Invitation code revoked')
          : t('Invitation code was already revoked')
      )
      props.onOpenChange(false)
    } catch (error: unknown) {
      handleServerError(error)
    } finally {
      setIsRevoking(false)
      props.onRefresh()
    }
  }

  return (
    <AlertDialog
      open={props.invite !== null}
      onOpenChange={(open) => !open && props.onOpenChange(false)}
    >
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>{t('Revoke invitation code?')}</AlertDialogTitle>
          <AlertDialogDescription>
            {t(
              'The invitation code with prefix {{prefix}} can no longer be used after revocation.',
              { prefix: props.invite?.code_prefix || '' }
            )}
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel disabled={isRevoking}>
            {t('Cancel')}
          </AlertDialogCancel>
          <AlertDialogAction
            variant='destructive'
            disabled={isRevoking}
            onClick={handleRevoke}
          >
            {isRevoking ? t('Revoking...') : t('Revoke')}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  )
}
