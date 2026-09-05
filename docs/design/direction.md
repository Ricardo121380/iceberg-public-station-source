# Iceberg mascot and liquid glass — revision 2

User correction on 2026-09-05: the photographic/serif design does not match the existing icon. Reference the icon's style and add moderate Liquid Glass. This explicit visual reference supersedes the first photographic direction and the random catalog (seed 597bf079, assigned index 6 acknowledged).

FIRST VIEWPORT: original transparent mascot/boat/iceberg artwork at its native 512px scale, light-blue glass lens and water ripples. Strong sans-serif headings, ice-blue emphasis, coral/rose primary action reflecting the boat, navy ink reflecting the outlines. Existing information and actions retained.
GLASS: floating nav, secondary action/caption, connection workspace, authentication panel and utility controls. Use blur + saturation, specular inset edges and offset translucent shadows. Avoid full-page blur. Keep fields opaque and text contrast stable. Explicit reduced-transparency and unsupported-browser fallbacks.
MOTION: one 800ms entrance on the original artwork. No perpetual decorative movement. Reduced-motion removes entrance and button transform.
AUTH: same reference artwork and glass panel. On mobile, title and small artwork sit side-by-side above form. At 320px retain 300px inner width for Turnstile.
SCOPE: homepage and sign-in/sign-up styling only. No OAuth, registration, billing, service configuration, or monitoring changes. Original brand artwork retained byte-for-byte. Public Sans plus self-hosted Noto Sans SC title subset. Retire photographic assets and serif subset from this revision.

HOMEPAGE CAPTION: user-requested copy “轻舟已撞大冰山” beside ICEBERG. Authentication caption stays unchanged.
