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
import type { RankingsSnapshot } from '../types'

/**
 * Internal safety/abuse canary models (e.g. `iceberg-abuse-canary-*`) exist so
 * the backend can probe upstreams; they carry a token or two of probe traffic
 * and must not show up on the public leaderboard. The durable server-side fix
 * is tracked in docs/runbooks/70-deferred-work.md; until then the public page
 * drops them at the boundary.
 */
const INTERNAL_MODEL_PATTERN = /canary/i

export function isInternalModelName(modelName: string): boolean {
  return INTERNAL_MODEL_PATTERN.test(modelName)
}

/**
 * Return a snapshot without internal probe models. Model rows are re-ranked
 * densely (1..n) so removing a row never leaves a visible gap; history series
 * drop the model from both points and the legend list.
 */
export function dropInternalModels(
  snapshot: RankingsSnapshot
): RankingsSnapshot {
  const models = snapshot.models
    .filter((row) => !isInternalModelName(row.model_name))
    .map((row, index) => ({ ...row, rank: index + 1 }))

  const movers = snapshot.top_movers.filter(
    (row) => !isInternalModelName(row.model_name)
  )
  const droppers = snapshot.top_droppers.filter(
    (row) => !isInternalModelName(row.model_name)
  )

  const history = snapshot.models_history
  const historyPoints = history.points.filter(
    (point) => !isInternalModelName(point.model)
  )
  const historyModels = history.models.filter(
    (entry) => !isInternalModelName(entry.name)
  )

  return {
    ...snapshot,
    models,
    top_movers: movers,
    top_droppers: droppers,
    models_history: {
      ...history,
      points: historyPoints,
      models: historyModels,
    },
  }
}
