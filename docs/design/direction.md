# Ice sea as interface — revision 3

User rejected the enlarged central logo: derive the frontend design from the icon instead of placing an oversized icon in the page. Original mascot remains only in normal header/footer branding; StationArt is removed. Prior user caption “轻舟已撞大冰山” stays beside ICEBERG.

FIRST VIEWPORT: a full-width pale ice sea, explicit faceted background geometry, coral wake and navy typography. The hero text and real interactive API launchpad form one composed scene. Geometry is authored as flat SVG paths, never a traced or enlarged logo; it is decorative and inaccessible to assistive technology. No stock photo or replacement raster. The translucent panel contains a real selectable/copyable URL and model/key links. Signed-out key link preserves /keys as login redirect; signed-in goes directly to /keys.

TRANSLATION FROM ICON: triangular ice planes become the landscape and angular step markers; navy outline becomes thin component definition; the coral boat becomes the CTA and curved route; white wake becomes long scroll composition; blue sea becomes surfaces. Glass exists primarily on the practical launchpad, floating nav and auth panel.

OTHER SURFACES: onboarding uses an outlined route sequence and flatter information rows. Login/signup get a cross-page ice landscape, route steps and existing form in glass. At mobile widths the form follows a compact headline; no large icon or mascot is inserted anywhere.

BOUNDARIES: all previous factual copy preserved. No backend/auth/payment/monitoring changes. Code-led concrete user direction; seed 926b7822 acknowledged, user correction overrides catalog. Retain reduced-motion, reduced-transparency, dark theme and 320px verification width. New launchpad functionality requires clipboard failure and signed-in/out destination tests.
