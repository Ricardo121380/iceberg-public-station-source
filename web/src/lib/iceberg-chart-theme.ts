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
import type { ITheme } from '@visactor/vchart'

import type { OpeningHoliday } from '@/features/home/lib/opening-holidays'

import holidayTokens from './holiday-chart-tokens.json'

/**
 * Iceberg VChart themes. Charts are the console's visual center of gravity;
 * VChart only ships generic light/dark themes, so the station registers its
 * own pair built from the iceberg preset tokens: an ice-sea series ramp with
 * the boat pink reserved for late-series emphasis, ice-tinted axes and grid,
 * and a glass tooltip. Pure data + registration — no static vchart import, so
 * the heavy chart bundle keeps its existing lazy boundary.
 */

export const ICEBERG_CHART_THEME_LIGHT = 'iceberg-light'
export const ICEBERG_CHART_THEME_DARK = 'iceberg-dark'

type ThemeManagerLike = {
  themeExist: (name: string) => boolean
  getTheme: (name: string) => ITheme
  registerTheme: (name: string, theme: Partial<ITheme>) => void
}

const LIGHT_SERIES = [
  '#0877ac',
  '#3d9dc7',
  '#79c3e2',
  '#0e5c85',
  '#5bb2d8',
  '#bc285d',
  '#d96a92',
  '#9fd4ea',
]
const DARK_SERIES = [
  '#85d6ff',
  '#4a9bc4',
  '#2e6f99',
  '#6fb3d9',
  '#3d7ea8',
  '#ff9fbb',
  '#c4577f',
  '#b8e2f5',
]

const LIGHT_COMPONENT = {
  axis: {
    label: { style: { fill: '#506781' } },
    line: { style: { stroke: '#c9dfed' } },
    grid: { style: { line: { style: { stroke: '#e3eef7' } } } },
  },
  legend: { label: { style: { fill: '#506781' } } },
  tooltip: {
    panel: {
      backgroundColor: '#ffffffed',
      border: { color: '#c9dfed', width: 1, radius: 10 },
      shadow: { x: 0, y: 8, blur: 24, spread: -8, color: '#247ead33' },
    },
    titleLabel: { fontColor: '#172b46' },
    keyLabel: { fontColor: '#506781' },
    valueLabel: { fontColor: '#172b46' },
  },
}
const DARK_COMPONENT = {
  axis: {
    label: { style: { fill: '#a4c0d8' } },
    line: { style: { stroke: '#264563' } },
    grid: { style: { line: { style: { stroke: '#1b3350' } } } },
  },
  legend: { label: { style: { fill: '#a4c0d8' } } },
  tooltip: {
    panel: {
      backgroundColor: '#12253ced',
      border: { color: '#264563', width: 1, radius: 10 },
      shadow: { x: 0, y: 8, blur: 24, spread: -8, color: '#00000080' },
    },
    titleLabel: { fontColor: '#e8f2fc' },
    keyLabel: { fontColor: '#a4c0d8' },
    valueLabel: { fontColor: '#e8f2fc' },
  },
}

/** Register the iceberg theme pair once per loaded ThemeManager instance. */
export function registerIcebergChartThemes(
  ThemeManager: ThemeManagerLike
): void {
  if (ThemeManager.themeExist(ICEBERG_CHART_THEME_LIGHT)) return
  const baseLight = ThemeManager.getTheme('light')
  const baseDark = ThemeManager.getTheme('dark')
  ThemeManager.registerTheme(ICEBERG_CHART_THEME_LIGHT, {
    ...baseLight,
    background: 'transparent',
    colorScheme: { default: LIGHT_SERIES },
    component: {
      ...baseLight.component,
      ...LIGHT_COMPONENT,
    } as ITheme['component'],
  })
  ThemeManager.registerTheme(ICEBERG_CHART_THEME_DARK, {
    ...baseDark,
    background: 'transparent',
    colorScheme: { default: DARK_SERIES },
    component: {
      ...baseDark.component,
      ...DARK_COMPONENT,
    } as ITheme['component'],
  })
}

/**
 * Every chart owner picks its theme through here, so switching presets away
 * from iceberg restores the stock VChart themes without a reload.
 */
export function resolveChartThemeName(
  preset: string,
  resolvedTheme: 'light' | 'dark',
  holiday?: OpeningHoliday | null
): string {
  if (holiday) return `iceberg-${holiday}-${resolvedTheme}`
  if (preset === 'iceberg') {
    return resolvedTheme === 'dark'
      ? ICEBERG_CHART_THEME_DARK
      : ICEBERG_CHART_THEME_LIGHT
  }
  return resolvedTheme
}

/**
 * Series ramp for the dashboard's manual color assignment (the analytics
 * builders pick colors by index instead of the VChart theme). Returns the
 * ice-sea ramp repeated to the requested domain size.
 */
export function icebergSeriesRamp(
  resolvedTheme: 'light' | 'dark',
  size: number,
  holiday?: OpeningHoliday | null
): string[] {
  if (holiday) {
    const tokens = holidayTokens[holiday][resolvedTheme]
    const colors = [
      tokens['chart-1'],
      tokens['chart-2'],
      tokens['chart-3'],
      tokens['chart-4'],
      tokens['chart-5'],
    ]
    return Array.from({ length: size }, (_, i) => colors[i % colors.length])
  }
  const base = resolvedTheme === 'dark' ? DARK_SERIES : LIGHT_SERIES
  if (size <= base.length) return base.slice(0, size)
  return Array.from({ length: size }, (_, i) => base[i % base.length])
}

// sRGB equivalents of the approved OKLCH palette, preconverted for VChart.
export function registerHolidayChartTheme(
  manager: ThemeManagerLike,
  holiday: OpeningHoliday | null | undefined,
  mode: 'light' | 'dark'
): void {
  if (!holiday) return
  const name = resolveChartThemeName('iceberg', mode, holiday)
  if (manager.themeExist(name)) return
  const base = manager.getTheme(mode)
  const tokens = holidayTokens[holiday][mode]
  const text = tokens['muted-foreground']
  const foreground = tokens.foreground
  const border = tokens.border
  manager.registerTheme(name, {
    ...base,
    background: 'transparent',
    colorScheme: { default: icebergSeriesRamp(mode, 5, holiday) },
    component: {
      ...base.component,
      axis: {
        label: { style: { fill: text } },
        line: { style: { stroke: border } },
        grid: { style: { line: { style: { stroke: border } } } },
      },
      legend: { label: { style: { fill: text } } },
      tooltip: {
        panel: {
          backgroundColor: tokens.card,
          border: { color: border, width: 1, radius: 10 },
        },
        titleLabel: { fontColor: foreground },
        keyLabel: { fontColor: text },
        valueLabel: { fontColor: foreground },
      },
    } as ITheme['component'],
  })
}
