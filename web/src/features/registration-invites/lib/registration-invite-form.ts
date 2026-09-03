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
import type { TFunction } from 'i18next'
import { z } from 'zod'

import {
  REGISTRATION_INVITE_DEFAULT_VALUES,
  REGISTRATION_INVITE_VALIDATION,
} from '../constants'
import type { RegistrationInviteCreateParams } from '../types'

export type RegistrationInviteFormValues = {
  count: number
  valid_days: number
  note: string
}

export function getRegistrationInviteFormSchema(t: TFunction) {
  return z.object({
    count: z
      .number()
      .int(t('Invitation count must be a whole number'))
      .min(
        REGISTRATION_INVITE_VALIDATION.COUNT_MIN,
        t('Invitation count must be between {{min}} and {{max}}', {
          min: REGISTRATION_INVITE_VALIDATION.COUNT_MIN,
          max: REGISTRATION_INVITE_VALIDATION.COUNT_MAX,
        })
      )
      .max(
        REGISTRATION_INVITE_VALIDATION.COUNT_MAX,
        t('Invitation count must be between {{min}} and {{max}}', {
          min: REGISTRATION_INVITE_VALIDATION.COUNT_MIN,
          max: REGISTRATION_INVITE_VALIDATION.COUNT_MAX,
        })
      ),
    valid_days: z
      .number()
      .int(t('Validity period must be a whole number'))
      .min(
        REGISTRATION_INVITE_VALIDATION.VALID_DAYS_MIN,
        t('Validity period must be between {{min}} and {{max}} days', {
          min: REGISTRATION_INVITE_VALIDATION.VALID_DAYS_MIN,
          max: REGISTRATION_INVITE_VALIDATION.VALID_DAYS_MAX,
        })
      )
      .max(
        REGISTRATION_INVITE_VALIDATION.VALID_DAYS_MAX,
        t('Validity period must be between {{min}} and {{max}} days', {
          min: REGISTRATION_INVITE_VALIDATION.VALID_DAYS_MIN,
          max: REGISTRATION_INVITE_VALIDATION.VALID_DAYS_MAX,
        })
      ),
    note: z.string().refine(
      (value) =>
        new TextEncoder().encode(value).length <=
        REGISTRATION_INVITE_VALIDATION.NOTE_MAX_BYTES,
      t('Note must not exceed {{count}} bytes', {
        count: REGISTRATION_INVITE_VALIDATION.NOTE_MAX_BYTES,
      })
    ),
  })
}

export const REGISTRATION_INVITE_FORM_DEFAULT_VALUES: RegistrationInviteFormValues =
  REGISTRATION_INVITE_DEFAULT_VALUES

export function toRegistrationInviteCreateParams(
  values: RegistrationInviteFormValues
): RegistrationInviteCreateParams {
  return {
    count: values.count,
    valid_days: values.valid_days,
    note: values.note.trim(),
  }
}
