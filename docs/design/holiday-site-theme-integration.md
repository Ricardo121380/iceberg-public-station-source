# Seasonal site themes

Approved OpenDesign K3 design: project iceberg-holiday-site-themes-2026, final run 3acfb698-1c85-409e-8265-ec19cbd114ed. Production uses the approved CSS/color mappings and ten decoration SVGs, not the prototype HTML or mock data. Existing homepage, navigation, forms, tables and backend behavior stay in place.

ThemeProvider reuses getOpeningHoliday (same calendar as opening). It applies body[data-holiday] independently of user mode and customization cookies. The overlay is removed outside holiday windows; midnight, pageshow, focus and visibilitychange recheck the shared calendar. The opening retains its current 3.4-second mount snapshot. The 2026–2035 lunar table remains unchanged.

CSS overrides body, station-site and station-scope so portal popovers also inherit tokens. Artwork URLs use bundler-resolved references to public/iceberg/holidays-v1. Additional SVGs are unmodified from the approved design. The original 30 opening SVGs remain untouched. The preview-only status-color suggestions are not applied to production business statuses.

VChart palettes are preconverted from approved OKLCH colors to sRGB HEX (browser canvas conversion) because existing link-alpha helpers and some canvas paths require RGB/HEX. Holiday themes register per holiday and mode. Model/user/Sankey explicit color specs are recomputed when theme context changes; charts retain data, axes, labels, semantic status and vendor identity colors. No new runtime dependencies.

Validation covers direct-entry body/portal inheritance, preservation of cookies and mode choice, midnight transition, tab return fallback, listener cleanup, chart registration and HEX palettes, actual public pages under five holidays × two modes, ordinary fallback, mobile width and real notification dialog.
