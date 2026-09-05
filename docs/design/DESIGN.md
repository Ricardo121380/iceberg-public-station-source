---
name: Iceberg Station
description: Ice-white reading surfaces and deep-ocean onboarding for a community AI API station
colors:
  background: "#f7fafb"
  foreground: "#173746"
  primary: "#14546b"
  primary-foreground: "#ffffff"
  muted: "#edf3f5"
  muted-foreground: "#536a75"
  border: "#d6e2e7"
  ring: "#197794"
  card: "#ffffff"
  accent: "#e5f1f5"
  station-deep: "#113746"
  dark-background: "#0f222c"
  dark-foreground: "#e8f2f5"
  dark-primary: "#9ed9e8"
  dark-primary-foreground: "#112c39"
  dark-muted: "#193541"
  dark-muted-foreground: "#aec3cb"
  dark-border: "#32505d"
  dark-accent: "#234957"
  dark-station-deep: "#0a1c25"
  ocean-text: "#eef8fa"
  ocean-muted: "#b9d0d9"
  ocean-link: "#b2e4f0"
typography:
  display:
    fontFamily: "Lora Variable, Iceberg Display CJK, Songti SC, serif"
    fontSize: "clamp(42px, 4.7vw, 68px)"
    fontWeight: 500
    lineHeight: 1.3
    letterSpacing: "-0.025em"
  headline:
    fontFamily: "Lora Variable, Iceberg Display CJK, Songti SC, serif"
    fontSize: "clamp(28px, 3vw, 40px)"
    fontWeight: 500
    lineHeight: 1.4
  body:
    fontFamily: "Public Sans Variable, PingFang SC, Microsoft YaHei, sans-serif"
    fontSize: "16px"
    lineHeight: 1.95
  button:
    fontSize: "14px"
    fontWeight: 600
  code:
    fontFamily: "ui-monospace, monospace"
    fontSize: "13px"
rounded:
  address: "6px"
  button: "8px"
  auth-art: "12px"
  hero-art: "160px 160px 12px 12px"
spacing:
  action-gap: "12px"
  hero-gap: "72px"
  container-gutter: "56px"
  mobile-gutter: "20px"
components:
  button-primary:
    backgroundColor: "{colors.primary}"
    textColor: "{colors.primary-foreground}"
    typography: "{typography.button}"
    rounded: "{rounded.button}"
    padding: "12px 22px"
  button-secondary:
    backgroundColor: "transparent"
    textColor: "{colors.foreground}"
    rounded: "{rounded.button}"
    padding: "12px 22px"
  button-secondary-hover:
    backgroundColor: "{colors.accent}"
  address-field:
    backgroundColor: "transparent"
    textColor: "#effaff"
    typography: "{typography.code}"
    height: "46px"
    padding: "0 13px"
  text-link:
    textColor: "{colors.primary}"
    typography: "{typography.button}"
    padding: "10px 0"
---

# Design System: 冰山公益站

## Overview

**Creative North Star: "Iceberg waterline"**

Ice-white reading surfaces, deep-ocean sections and restrained ice-blue details make the community API service calm and approachable. Chinese serif headings provide character; sans-serif instructions and real connection controls keep the service legible. The iceberg photograph is landscape branding, not proof of service performance.

This is a code-derived record of the approved code-led implementation, dated 2026-09-05. Its authority covers the default homepage and the opt-in sign-in/sign-up shell. Product facts live in [PRODUCT.md](PRODUCT.md); surface intent lives in [direction.md](direction.md). The implementation remains authoritative when these notes drift.

**Key Characteristics:**

- Generous reading space, serif display headings and a single photographic silhouette.
- Flat, divided information groups with a deep-ocean onboarding section.
- Useful controls with visible focus and restrained state transitions.

Source: `web/src/styles/iceberg.css`, `web/src/features/home/components/station-home.tsx`, `station-connection.tsx`, and `web/src/features/auth/station-auth-layout.tsx`. [design.json](design.json) holds extension metadata and isolated component previews. It is stored here to respect this project's documentation boundary; it is not the skill's default `.impeccable/design.json` auto-discovery path.

## Colors

The frontmatter records the actual scoped CSS values. `primary` is deep teal-blue for actions; `station-deep` supplies the large ocean section. `background` is ice white, `foreground` dark blue-green, `muted-foreground` the supporting copy color, and `border` the quiet divider. White `card` surfaces support the inherited authentication fields. `accent` is the secondary hover fill. There is no separate decorative accent family.

`.station-site` overrides existing semantic CSS variables. `secondary` equals `muted`; `card-foreground`, `accent-foreground` and `secondary-foreground` equal `foreground`. `.dark .station-site` supplies the documented dark values; dark `card` and `secondary` equal `dark-muted`, and dark `ring` equals `dark-primary`. These are scoped overrides, not a replacement for the application theme.

The ocean section uses fixed light text (`ocean-text`, `ocean-muted`, `ocean-link`) in both themes and a local ice-blue focus ring. Photo treatments also retain fixed colors. Preserve these intentional local assignments instead of assuming all text follows the global theme.

## Typography

Display: `Lora Variable, Iceberg Display CJK, Songti SC, serif`, weight 500. Body: `Public Sans Variable, PingFang SC, Microsoft YaHei, sans-serif`. API addresses use `ui-monospace, monospace`.

- Hero headings: the frontmatter display size, two explicit lines, balanced wrapping; mobile `clamp(36px, 9vw, 52px)`.
- Section display headings: `clamp(28px, 3vw, 40px)`, line-height 1.4; mobile onboarding heading 29px. FAQ and closing headings have their own 38px/30px desktop sizes.
- Hero body: 16px/1.95, at most 34em; mobile 14px. Supporting prose is generally 13–14px with 1.9–1.95 line-height.
- Authentication form heading: 34px/1.4, weight 500, left aligned; mobile 30px. Form inputs are 16px with a 46px minimum height.
- Buttons: 14px, weight 600. Utility text: 11–13px. Letter spacing is reserved for the short photographic ICEBERG label.

`display-cjk.woff2` is a self-hosted Noto Serif SC weight-500 **text subset**, with `font-display: swap`. New Chinese display copy may need an expanded subset; otherwise the serif fallback renders missing glyphs. Preserve the OFL file and asset provenance in `web/public/iceberg/ASSETS.md`.

## Layout

The desktop homepage container is `min(1240px, calc(100% - 112px))`. The hero uses 1.16:1 columns with a 72px gap and 142px/68px block padding, increasing top padding to 160px at widths of 1600px and above. Its photo is 496px tall. The content after the hero remains in normal document flow: model link row, three onboarding steps, connection controls, FAQ, closing link and inherited footer.

The onboarding steps use three columns. Connection controls use a 1:1.15 grid; FAQ uses a 0.8:1.2 grid. Dividers separate groups without enclosing them in repeated cards. The connection anchor has a 70px scroll margin.

- At 1100px and below: container side gutters become 32px, hero gap 40px and image height 440px; authentication gutters and panel padding reduce.
- At 760px and below: container side gutters become 20px; hero, onboarding, connection and FAQ stack. Copy precedes a 290px photo. CTA buttons wrap. Section spacing tightens. Code scrolls inside its own `pre` region.
- At 360px and below: homepage gutters become 16px and auth gutters 10px; CTA padding and font size reduce.

Authentication has a 92px header, a two-column main area with a 1500px maximum width, a photo with 640px minimum height, and a form capped at 390px. The page uses `min-height`, allowing registration to scroll. At 760px and below, the header becomes 80px, the photo becomes a 174px banner, the form stacks below it, the secondary photo copy and footer disappear, and the explicit back-home text is hidden. The brand remains a home link, with language and theme controls available.

## Elevation & Depth

The station's custom surfaces are flat. Dividers, tonal sections and photography establish depth; there is no new card-shadow system. Photo captions use restrained dark gradients for contrast, and the homepage caption has `text-shadow: 0 1px 4px #123444`. The selected client tab uses an inset 2px ice-blue underline. Existing UI primitives retain their own focus rings and state styles.

## Shapes

Homepage actions have 8px corners; the API address frame has 6px corners. The auth photo has 12px corners. The hero photo is the distinctive shape: `160px 160px 12px 12px`, becoming `120px 120px 12px 12px` at 1100px and `90px 90px 8px 8px` at 760px. A one-pixel waterline crosses it at 67% height. Do not turn that photograph-specific silhouette into a generic card shape.

## Components

**Homepage actions.** `.station-button` uses a 48px minimum height, 12px 22px padding, a 1px border and 14px icon gap. Primary hover mixes the primary color with 15% foreground; secondary hover fills with `accent`. Both move up 2px on hover. Authentication buttons remain the existing shared `Button`, with only a scoped 46px minimum-height override; they are not the homepage button class.

**Connection controls.** `StationConnection` shares the existing accessible Tabs primitives. A read-only, selectable Base URL field pairs with a 46px copy button; copy success changes the icon and updates a reserved `role=status` region. Failure offers manual selection. Tabs are flat, with a selected underline and a 235px minimum-height panel (250px on mobile). Preserve the three client guides and horizontally scrollable, keyboard-focusable code example.

**FAQ.** Native `details`/`summary` rows use bottom dividers, generous 22px vertical summary padding and a plus/minus indicator. Expanded prose remains in document flow; no additional animation or JavaScript accordion is required.

**Authentication.** `StationAuthLayout` wraps sign-in and sign-up only. It reuses configured logo/name, `LanguageSwitcher`, `ThemeSwitch`, existing forms, `TermsFooter` and the real invitation/OAuth/Turnstile/consent behavior. Input error, disabled and validation styles continue to come from existing UI components and theme variables. Requirement labels are a wrapping plain-text list, not chips. Keep New API / QuantumNous attribution visible.

**Navigation and scope.** The homepage retains `PublicLayout` and its existing configuration-driven global header/navigation. Only the default homepage enters `.station-site`; configured URL, HTML and Markdown custom homepages retain their existing branches. Loading, other public pages, dashboard and other authentication routes keep their incumbent layout. Portalled controls may inherit the application theme outside `.station-site`; do not describe these as fully rebranded surfaces.

**Focus and motion.** Scoped links, buttons, inputs, summaries and code blocks receive a 2px `ring` outline with 5px offset on `:focus-visible`. Skip links become visible on focus. The hero photo alone enters through `station-water-reveal`: 1.1s `cubic-bezier(0.16, 1, 0.3, 1)`, clip inset 6% → 0, opacity 0.7 → 1. Text is visible by default. Reduced-motion removes this animation and disables scoped transitions and smooth scrolling. The static hover translation remains; the current rule removes its transition, not the transform itself.

## Do's and Don'ts

### Do:

- **Do** extend station styles through the opt-in scope and existing semantic variables.
- **Do** keep real model, quota, registration and account state authoritative in the interface.
- **Do** preserve accessible control semantics, manual-copy fallback and normal page scrolling.
- **Do** keep source/license records with the photograph and display font.

### Don't:

- **Don't** invent availability figures, quota promises or service proof from the landscape photograph.
- **Don't** spread the hero silhouette, photographic gradients or nested panels into routine controls.
- **Don't** replace shared authentication behavior or configuration-driven navigation to achieve a visual match.
- **Don't** treat this scoped first phase as a completed application-wide redesign.
