# Holiday opening themes

Approved themes: Mid-Autumn (moon, mooncake, osmanthus), National Day (red/gold, red flag, fireworks), New Year (event year, stars, ribbons), Spring Festival (lanterns, red envelopes, auspicious clouds), Dragon Boat (zongzi, dragon-boat trim, water patterns). Preserve shrimp/boat/iceberg, 3400 ms inside-to-outside collision, every-visit playback and immediate input dismissal.

## Calendar contract

Use Beijing UTC+08:00 civil dates, independent of browser locale/timezone. Start at00:00 seven days before the festival; end at00:00 three days after it (festival+2 included). Date refers to the festival day, not the statutory leave period. Mid-Autumn wins any overlapping National Day window. New Year decorations show the year of January1, including the preceding December window.

Lunar dates are pinned from Hong Kong Observatory Gregorian–Lunar Conversion Tables (2026–2035). Runtime does not fetch calendars or rely on potentially different ICU lunar implementations. Sources: https://www.hko.gov.hk/tc/gts/time/conversion.htm and https://www.hko.gov.hk/tc/gts/time/calendar/text/files/T2026c.txt through T2035c.txt. Extend this table with verified HKO dates before2036; unsupported lunar years fall back to ordinary artwork. Fixed January1/October1 events continue automatically. Window selection uses visitor system time converted to Beijing time. Existing open animations finish their3.4s sequence; each new homepage mount reevaluates the theme.

2026 schedule: Mid-Autumn September18–27; National Day September28–October3 after applying overlap priority. New Year2027 starts December25,2026 and ends January3,2027.

## Generation provenance

Commissioned through the running OpenDesign application's API with its existing deepseek-harness runtime, explicitly model kimi-coding/k3. Project iceberg-holiday-opening-2026, run a0304fae-e0b0-4b4d-8ceb-18cba888297e, request2fa102b1-8104-44d0-a56b-1e3b0d575475. Original five SVG sources and animation CSS provided as reference. No alternate model or external credential handling. Final artwork must be validated as self-contained SVG before shipping.

Holiday assets use versioned paths. Load only the current theme; failed holiday loads try the original artwork within the existing1.5s load budget. If both fail or load too slowly, render the homepage without the optional animation. No backend/auth changes or scheduled server job are needed.

## Approved artwork (v3)

Approved 2026-09-19 after two K3 revisions: five distinct vehicle silhouettes (dragon boat, crescent moon boat, red/gold parade boat, star countdown boat, koi lantern boat). The shrimp lower body is behind the foreground hull; the koi has a recessed cockpit. Final OpenDesign run `d668c405-ad08-4d9f-b855-f86d26687ed8` succeeded using `kimi-coding/k3`, conversation `eddbe652-bea8-445c-9d6b-b78282febbc1`. All 30 SVGs are copied unchanged from the approved v3 output into `web/public/iceberg/intro/holidays-v1/`. Preview gallery and its scripts are not shipped. Existing production motion and click-through behavior remain authoritative.

The New Year badge matches the SVG rectangle at (252,20), size 116x44 within 620x496. The displayed year comes from the selected holiday occurrence, including January 2–3.
