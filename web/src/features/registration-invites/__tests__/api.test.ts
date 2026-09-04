/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.
*/
import { beforeEach, describe, expect, test, vi } from 'vitest'

import { api } from '@/lib/api'

import { getRegistrationInviteStats } from '../api'

vi.mock('@/lib/api', () => ({
  api: {
    get: vi.fn(),
  },
}))

describe('registration invitation statistics', () => {
  beforeEach(() => {
    vi.mocked(api.get).mockReset()
  })

  test('requests every status without emitting one global toast per request', async () => {
    vi.mocked(api.get).mockResolvedValue({
      data: { success: true, data: { items: [], total: 0 } },
    })

    await expect(getRegistrationInviteStats()).resolves.toEqual({
      success: true,
      data: { active: 0, expired: 0, used: 0, revoked: 0 },
    })

    expect(api.get).toHaveBeenCalledTimes(4)
    for (const [, config] of vi.mocked(api.get).mock.calls) {
      expect(config).toMatchObject({ skipBusinessError: true })
    }
  })
})
