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
import { Download, FileText } from 'lucide-react'
import { useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'

import { CopyButton } from '@/components/copy-button'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
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
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Textarea } from '@/components/ui/textarea'

import {
  downloadRegistrationInviteCodes,
  getRegistrationInviteCodesCsv,
  getRegistrationInviteCodesText,
} from '../lib'
import type { CreatedRegistrationInvite } from '../types'

type RegistrationInviteResultDialogProps = {
  invites: CreatedRegistrationInvite[]
  onDiscard: () => void
}

export function RegistrationInviteResultDialog(
  props: RegistrationInviteResultDialogProps
) {
  const { t } = useTranslation()
  const [confirmDiscardOpen, setConfirmDiscardOpen] = useState(false)
  const codesText = useMemo(
    () => getRegistrationInviteCodesText(props.invites),
    [props.invites]
  )
  const codesCsv = useMemo(
    () => getRegistrationInviteCodesCsv(props.invites),
    [props.invites]
  )

  const requestDiscard = () => setConfirmDiscardOpen(true)

  return (
    <>
      <Dialog
        open
        onOpenChange={(open) => {
          if (!open) requestDiscard()
        }}
      >
        <DialogContent className='sm:max-w-2xl'>
          <DialogHeader>
            <DialogTitle>{t('Save invitation codes')}</DialogTitle>
            <DialogDescription>
              {t('Generated {{count}} invitation code(s).', {
                count: props.invites.length,
              })}
            </DialogDescription>
          </DialogHeader>

          <Alert>
            <FileText />
            <AlertTitle>{t('These codes are shown only once')}</AlertTitle>
            <AlertDescription>
              {t(
                'Copy or download the codes now. Closing this dialog permanently clears them from this browser.'
              )}
            </AlertDescription>
          </Alert>

          <div className='space-y-2'>
            <div className='flex flex-wrap items-center justify-between gap-2'>
              <label
                htmlFor='registration-invite-codes'
                className='font-medium'
              >
                {t('Invitation codes')}
              </label>
              <CopyButton
                value={codesText}
                variant='outline'
                size='sm'
                tooltip={t('Copy all invitation codes')}
              >
                {t('Copy all')}
              </CopyButton>
            </div>
            <Textarea
              id='registration-invite-codes'
              value={codesText}
              readOnly
              aria-label={t('Generated invitation codes')}
              className='min-h-48 resize-y font-mono text-xs leading-6'
            />
          </div>

          <DialogFooter>
            <Button
              type='button'
              variant='outline'
              onClick={() => downloadRegistrationInviteCodes(codesText, 'txt')}
            >
              <FileText />
              {t('Download text')}
            </Button>
            <Button
              type='button'
              variant='outline'
              onClick={() => downloadRegistrationInviteCodes(codesCsv, 'csv')}
            >
              <Download />
              {t('Download CSV')}
            </Button>
            <Button type='button' onClick={requestDiscard}>
              {t('I saved the codes')}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <AlertDialog
        open={confirmDiscardOpen}
        onOpenChange={setConfirmDiscardOpen}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>
              {t('Discard invitation codes?')}
            </AlertDialogTitle>
            <AlertDialogDescription>
              {t(
                'The full codes cannot be recovered after this dialog is closed.'
              )}
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>{t('Keep showing codes')}</AlertDialogCancel>
            <AlertDialogAction onClick={props.onDiscard} variant='destructive'>
              {t('Discard codes')}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  )
}
