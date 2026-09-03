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
import { zodResolver } from '@hookform/resolvers/zod'
import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'

import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { handleServerError } from '@/lib/handle-server-error'

import { createRegistrationInvites } from '../api'
import {
  getRegistrationInviteFormSchema,
  REGISTRATION_INVITE_FORM_DEFAULT_VALUES,
  toRegistrationInviteCreateParams,
  type RegistrationInviteFormValues,
} from '../lib'
import type { CreatedRegistrationInvite } from '../types'

type RegistrationInviteCreateDialogProps = {
  open: boolean
  onOpenChange: (open: boolean) => void
  onCreated: (invites: CreatedRegistrationInvite[]) => void
}

export function RegistrationInviteCreateDialog(
  props: RegistrationInviteCreateDialogProps
) {
  const { t } = useTranslation()
  const [isSubmitting, setIsSubmitting] = useState(false)
  const form = useForm<RegistrationInviteFormValues>({
    resolver: zodResolver(getRegistrationInviteFormSchema(t)),
    defaultValues: REGISTRATION_INVITE_FORM_DEFAULT_VALUES,
  })

  const handleOpenChange = (open: boolean) => {
    if (!open) {
      form.reset(REGISTRATION_INVITE_FORM_DEFAULT_VALUES)
    }
    props.onOpenChange(open)
  }

  const onSubmit = async (values: RegistrationInviteFormValues) => {
    setIsSubmitting(true)
    try {
      const response = await createRegistrationInvites(
        toRegistrationInviteCreateParams(values)
      )
      const invites = response.data?.invites

      if (!response.success || !invites || invites.length === 0) {
        toast.error(response.message || t('Failed to create invitation codes'))
        return
      }

      form.reset(REGISTRATION_INVITE_FORM_DEFAULT_VALUES)
      props.onOpenChange(false)
      props.onCreated(invites)
    } catch (error: unknown) {
      handleServerError(error)
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <Dialog open={props.open} onOpenChange={handleOpenChange}>
      <DialogContent className='sm:max-w-md'>
        <DialogHeader>
          <DialogTitle>{t('Create invitation codes')}</DialogTitle>
          <DialogDescription>
            {t(
              'New invitation codes are shown only once after they are generated.'
            )}
          </DialogDescription>
        </DialogHeader>

        <Form {...form}>
          <form
            id='registration-invite-create-form'
            onSubmit={form.handleSubmit(onSubmit)}
            className='space-y-4'
            aria-busy={isSubmitting}
          >
            <FormField
              control={form.control}
              name='count'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>{t('Quantity')}</FormLabel>
                  <FormControl>
                    <Input
                      {...field}
                      type='number'
                      min={1}
                      max={100}
                      inputMode='numeric'
                      onChange={(event) =>
                        field.onChange(event.currentTarget.valueAsNumber)
                      }
                    />
                  </FormControl>
                  <FormDescription>
                    {t('Choose between 1 and 100 codes.')}
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name='valid_days'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>{t('Valid for (days)')}</FormLabel>
                  <FormControl>
                    <Input
                      {...field}
                      type='number'
                      min={1}
                      max={365}
                      inputMode='numeric'
                      onChange={(event) =>
                        field.onChange(event.currentTarget.valueAsNumber)
                      }
                    />
                  </FormControl>
                  <FormDescription>
                    {t('Invitation codes are valid for 7 days by default.')}
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name='note'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>{t('Note')}</FormLabel>
                  <FormControl>
                    <Textarea
                      {...field}
                      maxLength={255}
                      placeholder={t('Optional note for this batch')}
                    />
                  </FormControl>
                  <FormDescription>
                    {t('This note is visible to Root administrators only.')}
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />
          </form>
        </Form>

        <DialogFooter>
          <Button
            type='button'
            variant='outline'
            disabled={isSubmitting}
            onClick={() => handleOpenChange(false)}
          >
            {t('Cancel')}
          </Button>
          <Button
            type='submit'
            form='registration-invite-create-form'
            disabled={isSubmitting}
          >
            {isSubmitting ? t('Creating...') : t('Create invitation codes')}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
