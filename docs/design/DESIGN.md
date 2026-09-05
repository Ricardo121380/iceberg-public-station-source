---
name: 冰山公益站
description: Ice sea as interface, derived from the icon with selected Liquid Glass
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
  launchpad: "#ffffffe0"
  auth-glass: "#ffffffeb"
  ice-light: "#d4eefa"
  ice-mid: "#99d9f1"
  ice-white: "#f8fdff"
  ice-blue: "#5fc0e6"
  route: "#e95383"
typography:
  display:
    fontFamily: "Public Sans Variable, Iceberg Sans CJK, PingFang SC, sans-serif"
    fontSize: "clamp(44px, 4.65vw, 67px)"
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
  action: "12px"
  nav: "18px"
  connection: "22px"
  auth-panel: "24px"
spacing:
  action-gap: "12px"
  hero-gap: "96px"
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
  launchpad:
    backgroundColor: "{colors.launchpad}"
    rounded: "{rounded.auth-panel}"
    padding: "27px 28px 18px"
  auth-panel:
    backgroundColor: "{colors.auth-glass}"
    rounded: "{rounded.auth-panel}"
    padding: "42px"
---

# Design System: 冰山公益站

## Overview

**Creative North Star: "Ice sea as interface"**

The original icon supplies faceted blue ice, navy definition and a coral wake. Those elements become a full hero landscape, clear typography, outlined route markers and practical controls. The mascot remains small in normal header/footer branding; the user rejected enlarging it as page artwork. Moderate Liquid Glass gives the navigation, API launchpad and auth panel depth over this scene.

This code-derived revision 3 supersedes the large-mascot revision. Product constraints live in [PRODUCT.md](PRODUCT.md); the approved surface direction lives in [direction.md](direction.md). Read the complete scoped stylesheet: later v3 declarations override earlier base rules.

**Key Characteristics:**

- Flat decorative ice facets and coral wake across the homepage hero and auth background.
- Strong navy sans-serif headings and a real, titled API launchpad.
- Selected glass surfaces, flatter information rows and explicit accessibility fallbacks.

## Colors

Coral/rose identifies actions and the route through the ice; station blue identifies links and small accents. Navy headings sit on ice-white and pale blue surfaces. Both hero title lines are navy; the second receives a short irregular coral underline. The homepage primary action uses `linear-gradient(170deg, #d93d61, #b72060)`; shared auth actions use the semantic primary token.

Dark mode supplies a navy background, pale foreground, rose primary and ice-blue links. Facet colors also switch to blue/navy, preserving the geometry. Full mappings and material overrides live in the sidecar. Do not turn informational sections into repeated translucent cards.

## Typography

Public Sans is used for Latin text. Chinese display headings use the self-hosted Noto Sans SC weight-750 subset exposed as `Iceberg Sans CJK`, with `font-display: swap`; body Chinese falls back to PingFang SC / Microsoft YaHei. Preserve `web/public/iceberg/OFL-sans.txt` and asset provenance. New Chinese headings may require an expanded subset.

Hero headings use the display token with two explicit lines. Section headings use `clamp(28px, 3vw, 38px)` at 1.4 line-height. Desktop auth landscape headings are 43px, falling to 37px below 1100px, 27px below 760px and 25px below 380px. Form headings are 31px/1.4 at weight 750, reducing to 28px on mobile. Supporting prose is generally 13–14px at 1.9–1.95 line-height; auth inputs remain 16px with a 46px minimum height. API text uses `ui-monospace, monospace`. The hero footer pairs ICEBERG with the translated motto, “轻舟已撞大冰山” in Chinese.

## Layout

Desktop content width is `min(1220px, calc(100% - 96px))`. The full-width hero scene contains a 1.1:1 grid, 96px gap, 186px/80px block padding and 754px minimum height. The titled launchpad is capped at 440px and sits over the landscape; the motto closes the scene. A flat model-information row leads into three outlined onboarding steps, a 1:1.15 connection workspace and a 0.8:1.2 FAQ grid. The connection anchor has a 104px scroll margin.

- At 1100px and below, content gutters become 32px; hero gap becomes 45px, top padding 155px and minimum height 720px.
- At 1000px and below, the homepage uses the existing compact navigation and mobile menu.
- At 760px and below, content gutters become 20px; hero, steps, connection and FAQ stack. Hero padding is 128px/38px with a 42px gap and no minimum height. The launchpad is upright and centered; the landscape widens and shifts behind it. Navigation becomes 60px tall. Code scrolls inside its own region.
- At 380px and below, homepage gutters become 16px and auth gutters 10px; auth panel horizontal padding becomes zero. A 320px viewport retains a full 300px for the form and Turnstile.

Authentication has a 96px header and two-column main area capped at 1180px with a 100px gap, reducing to 48px below 1100px. The form is capped at 390px. Minimum height allows registration to scroll. Below 760px, the header becomes 84px; a compact heading precedes the form. Route steps, supporting landscape copy/footer and back-home text hide; home-linked branding, language and theme controls remain. There is no mobile mascot column.

## Elevation & Depth

**The Limited Glass Rule.** Reserve blur for floating navigation, the secondary action, the launchpad, auth utility controls and the auth panel. Models are a flat divided row; the connection workspace uses a muted opaque fill and border. Auth and connection fields use opaque card surfaces; the launchpad address is transparent over its high-opacity panel.

Navigation uses the shared glass token with 22px blur; the secondary action uses 14px and auth tools 18px. The launchpad uses 20px blur and an approximately 88% white fill (92% dark fill). The auth panel uses 24px blur and approximately 92% white fill (93% dark fill). Inset rims and restrained blue shadows distinguish these layers. The launchpad rests at a 1° angle on desktop and upright on mobile.

Reduced transparency uses opaque light (`#f7fbff`) or dark (`#203955`) fills and removes those backdrop filters. Browsers supporting neither standard nor WebKit filtering receive opaque fills. Decorative landscape opacity remains; this fallback concerns glass surfaces.

## Shapes

Faceted SVG planes and asymmetrically rounded outlined step markers translate the icon into interface geometry. Fields and homepage actions use 12px corners, auth controls 14px, navigation 18px, connection workspace 22px, and launchpad/auth panels 24px. The hero title underline and coral wake are curved accents. Keep the original logo at ordinary branding scale.

## Components

**Landscape.** `StationLandscape` is a flat decorative SVG with a 1440 × 860 viewBox, authored facets, shore, wake and coral route. It spans the hero and auth background, with responsive positioning. It is nonfocusable, hidden from accessibility APIs and ignores pointer events. `StationArt` and its enlarged logo, lens, ripples and droplets have been removed.

**Launchpad.** `StationLaunchpad` has a real heading, labelled selectable read-only API address and clipboard action. Success and manual-copy failure instructions use a status region. Model links go to `/pricing`; signed-in key links go to `/keys`, while signed-out links go to `/sign-in` with `/keys` as the redirect. The guide jumps to `#connect`. All visible strings continue through the existing translations.

**Actions and navigation.** Homepage actions have a 50px minimum height, weight 700, inset highlights and a 2px hover lift; press resets the lift and scales to 0.98. Signed-in primary actions go to the dashboard; signed-out actions go to sign-in. Registration appears only when permitted by current configuration. Shared auth buttons retain a scoped 46px minimum height. Navigation remains configuration-driven with existing menu behavior.

**Connection and disclosure.** The connection address has an opaque field and 44px copy control. Shared accessible Tabs have a muted track and opaque selected tab; panels have a 230px minimum height (265px on mobile). Preserve copy status/manual fallback and keyboard-focusable code. Native FAQ details/summary rows retain bottom dividers and plus/minus indicators.

**Authentication and scope.** The sign-in/sign-up shell retains configured logo/name, existing forms, language/theme controls, invitation/OAuth/Turnstile/consent logic and New API / QuantumNous attribution. Existing authentication and route logic is retained. Custom homepages, dashboard and other routes retain their branches and styling; portalled controls may inherit the broader application theme. Models, quota and eligibility must reflect real configuration.

**Focus and motion.** Scoped links, buttons, inputs, summaries and code receive a 2px ring with 5px offset on visible focus; skip links appear on focus. The launchpad has one 700ms entrance using `cubic-bezier(0.16, 1, 0.3, 1)`, from 12px lower and 0.8 opacity to its resting position. No perpetual animation is introduced. Reduced motion removes this entrance, scoped transitions, smooth scrolling and primary action hover/press transforms. The launchpad link arrow still shifts on hover without a transition.

The v3 reviewer inspected all eight desktop, mobile, auth, narrow, dark and English/tablet captures plus source and returned ship. The parent task reports 24 related tests, typecheck, lint and build passing, with browser clipboard and signed-out key redirect verified. This is source/screenshot coverage, not live OAuth authentication. Evidence is recorded in the sidecar.

## Do's and Don'ts

### Do:

- **Do** translate the icon into ice facets, navy definition and coral wake while keeping branding small.
- **Do** extend the opt-in station scope with existing semantic tokens and maintain light/dark parity.
- **Do** preserve authentic registration, model, quota and account state, accessible controls and normal page scrolling.
- **Do** keep font licensing and expand the Chinese title subset when adding display copy.

### Don't:

- **Don't** enlarge the mascot as hero or authentication artwork.
- **Don't** restore photographic hero art or serif typography for these surfaces.
- **Don't** apply blur to the whole page or nest glass panels throughout routine content.
- **Don't** replace authentication or configuration-driven navigation behavior to achieve a visual match.
- **Don't** treat this scoped homepage/auth revision as an application-wide redesign.
