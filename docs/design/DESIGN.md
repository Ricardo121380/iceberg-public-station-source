---
name: 冰山公益站
description: Original illustrated station identity with moderate Liquid Glass
colors:
  primary: "#bc285d"
  primary-foreground: "#ffffff"
  station-blue: "#0877ac"
  background: "#f5faff"
  foreground: "#172b46"
  muted: "#eaf3fa"
  muted-foreground: "#506781"
  border: "#c9dfed"
  ring: "#007bad"
  card: "#ffffff"
  accent: "#e0f2fd"
  accent-foreground: "#153b61"
  station-glass: "#ffffffb8"
  station-rim: "#ffffffed"
  station-shadow: "#247ead1c"
  station-glow: "#bbe6fc"
typography:
  display:
    fontFamily: "Public Sans Variable, Iceberg Sans CJK, PingFang SC, sans-serif"
    fontSize: "clamp(42px, 4.6vw, 66px)"
    fontWeight: 750
    lineHeight: 1.23
    letterSpacing: "-0.035em"
  body:
    fontFamily: "Public Sans Variable, PingFang SC, Microsoft YaHei, sans-serif"
    fontSize: "16px"
    lineHeight: 1.95
  action:
    fontFamily: "Public Sans Variable, PingFang SC, Microsoft YaHei, sans-serif"
    fontSize: "14px"
    fontWeight: 700
rounded:
  field: "12px"
  control: "14px"
  action: "17px"
  nav: "24px"
  connection: "26px"
  auth-panel: "30px"
spacing:
  action-gap: "12px"
  hero-gap: "48px"
  section: "64px"
components:
  button-primary:
    textColor: "{colors.primary-foreground}"
    typography: "{typography.action}"
    rounded: "{rounded.action}"
    padding: "13px 23px"
  button-secondary:
    backgroundColor: "{colors.station-glass}"
    textColor: "{colors.foreground}"
    rounded: "{rounded.action}"
    padding: "13px 23px"
  address-field:
    backgroundColor: "{colors.card}"
    textColor: "{colors.foreground}"
    rounded: "{rounded.field}"
    height: "46px"
  navigation:
    backgroundColor: "{colors.station-glass}"
    rounded: "{rounded.nav}"
    height: "64px"
    padding: "8px 20px"
  auth-panel:
    backgroundColor: "{colors.station-glass}"
    rounded: "{rounded.auth-panel}"
    padding: "42px"
---

# Design System: 冰山公益站

## Overview

**Creative North Star: "Original mascot, a little liquid glass"**

The original station icon sets the visual language: a friendly mascot, faceted blue icebergs, navy outlines and a coral/rose boat. Ice-blue surfaces and strong sans-serif headings carry that identity into the default homepage and sign-in/sign-up shell. Moderate Liquid Glass adds light-catching edges to selected controls and panels while fields remain opaque and text stays clear.

This code-derived revision 2 record supersedes the previous photographic/serif system. Product constraints live in [PRODUCT.md](PRODUCT.md); the approved surface direction lives in [direction.md](direction.md). The scoped implementation remains authoritative when documentation drifts.

**Key Characteristics:**

- Original 512px mascot artwork, blue emphasis and coral/rose actions.
- Readable sans-serif hierarchy with open spacing and selected translucent surfaces.
- Finite entrance motion, clear focus and explicit accessibility fallbacks.

## Colors

Primary coral/rose comes from the boat and identifies the main action; station blue emphasizes headings, links and step numbers. Navy foreground, ice-white background and pale blue muted surfaces provide the reading base. The homepage primary action uses a coral-to-rose gradient (`linear-gradient(170deg, #d93d61, #b72060)`), while shared auth controls use the semantic primary token.

Dark mode supplies a navy background (`#101e34`), pale foreground (`#edf5ff`), rose primary (`#ff9fbb`), ice-blue station accent (`#85d6ff`) and dark glass (`#203955cc`). Complete light/dark mappings come from the scoped CSS and are recorded in the sidecar. Glass rim, glow and shadow tokens provide material cues rather than additional content colors.

## Typography

Public Sans is used for Latin display and body text. Chinese display headings use the self-hosted Noto Sans SC weight-750 subset exposed as `Iceberg Sans CJK`, with `font-display: swap`; body Chinese falls back to PingFang SC / Microsoft YaHei. Preserve `web/public/iceberg/OFL-sans.txt` and asset provenance. New Chinese headings may require an expanded subset.

Hero headings use the display token with two explicit lines and blue emphasis on the second line. Section headings use `clamp(28px, 3vw, 38px)` at 1.4 line-height. Auth form headings are 31px/1.4 at weight 750, reducing to 28px on mobile. Supporting prose is generally 13–14px at 1.9–1.95 line-height; form inputs remain 16px with a 46px minimum height. API text uses `ui-monospace, monospace`.

## Layout

Desktop content width is `min(1220px, calc(100% - 96px))`. The hero uses 1.08:1 columns, a 48px gap and 162px/50px block padding. Models sit in a single horizontal panel; three onboarding steps precede a 1:1.15 connection grid. FAQ uses 0.8:1.2 columns. The connection anchor has a 104px scroll margin.

- At 1100px and below, content gutters become 32px, hero gap 24px and connection/auth spacing tightens.
- At 1000px and below, the homepage switches to the existing compact navigation and mobile menu.
- At 760px and below, content gutters become 20px and hero, steps, connection and FAQ stack. Hero headings use `clamp(37px, 9vw, 52px)`; artwork is capped at 360px. The nav becomes 60px tall. Code scrolls within its own region.
- At 380px and below, homepage gutters become 16px and auth gutters 10px; auth panel horizontal padding becomes zero. The 320px verification width therefore retains 300px for Turnstile and the form.

Authentication has a 96px header, a two-column main area capped at 1200px, and a form capped at 390px. It uses minimum height so registration can scroll. On mobile the header becomes 84px; title and a 144px mascot sit side-by-side above the form (120px mascot column below 380px). Secondary illustration copy/footer and back-home text hide; the home-linked brand, language and theme controls remain.

## Elevation & Depth

**The Limited Glass Rule.** Reserve blur for the floating navigation, secondary action, artwork caption, connection workspace, auth utility controls and auth panel. Model and step surfaces share translucent fill and inset rims without blur; text fields use opaque card backgrounds.

Blur ranges from 14px on the secondary action to 24px on the auth panel. White inset rims, subtle lower edges and blue-tinted offset shadows create the material. The mascot sits over a circular lens and ripples with two small colored droplets. These accents remain subordinate to the unchanged artwork.

Reduced transparency sets glass to opaque light (`#f7fbff`) or dark (`#203955`) and removes the listed backdrop filters. Browsers supporting neither standard nor WebKit backdrop filtering receive the same opaque fills. Decorative lens gradients remain; this is a glass-surface fallback, not an all-transparency removal.

## Shapes

Rounded controls and broad panels echo the illustration's friendly character: fields 12px, auth/header controls 14px, homepage actions 17px, floating navigation/model panel 24px, connection workspace 26px and auth panel 30px. Mobile radii tighten. The illustration lens, ripples and droplets are circular or elliptical; retain these as artwork accents.

## Components

**Original artwork.** `StationArt` reuses `/iceberg-station-mark-v2.png` with intrinsic dimensions of 512 × 512 and descriptive translated alt text. Lens, ripples and droplets are decorative and hidden from accessibility APIs. The image scales within available space; do not redraw or replace the source asset.

**Actions and navigation.** Homepage actions have a 50px minimum height, weight 700, inset highlights and a 2px hover lift; press resets the lift and scales to 0.98. Secondary actions use glass. Auth buttons remain the shared components with scoped 46px minimum height and inset highlight. The floating header retains configuration-driven navigation and the existing menu behavior.

**Connection and disclosure.** The selectable read-only address has an opaque field and a 44px copy control. Shared accessible Tabs use a muted track with an opaque selected tab; panel minimum height is 230px (265px on mobile). Preserve copy status/manual fallback and keyboard-focusable code. Native FAQ details/summary rows use bottom dividers and plus/minus indicators.

**Authentication and scope.** The shell wraps sign-in/sign-up while retaining configured logo/name, existing forms, language/theme controls, invitation/OAuth/Turnstile/consent logic and New API / QuantumNous attribution. Original auth and route logic is unchanged. Custom homepages, dashboard and other routes retain their existing branches and styling; portalled controls may inherit the broader application theme.

**Focus and motion.** Scoped links, buttons, inputs, summaries and code receive a 2px ring with 5px offset on visible focus. Skip links appear on focus. The mascot has one 800ms entrance (`cubic-bezier(0.16, 1, 0.3, 1)`), from 12px lower / −2° / 0.8 opacity to its resting position. No perpetual animation is introduced. Reduced motion removes the entrance, transitions, smooth scrolling and button hover/press transforms.

The v2 visual review compared the original icon with six captures and returned ship. The completed review included 320px as the narrow verification width; this document records the implementation and does not imply new browser or authentication testing.

## Do's and Don'ts

### Do:

- **Do** reuse the original illustration and preserve its colors, proportions and transparent silhouette.
- **Do** extend the opt-in station scope with existing semantic tokens and maintain light/dark parity.
- **Do** preserve authentic registration, model, quota and account state, accessible controls and normal page scrolling.
- **Do** keep font licensing and expand the Chinese title subset when adding display copy.

### Don't:

- **Don't** restore photographic hero art or serif typography for these surfaces.
- **Don't** apply blur to the whole page or nest glass panels throughout routine content.
- **Don't** replace authentication or configuration-driven navigation behavior to achieve a visual match.
- **Don't** treat this scoped homepage/auth revision as an application-wide redesign.
