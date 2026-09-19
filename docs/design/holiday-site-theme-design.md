# 冰山公益站 · 五节日主题设计说明

> 范围：与开屏 SVG 联动的全站节日主题。交付物为 `holiday-theme-tokens.json`
> （令牌全量映射）、`holiday-themes.css`（可落地增量主题层）、`assets/` 装饰
> SVG、`index.html` + `preview.css` + `preview.js`（交互预览）。令牌由
> `tools/generate-holiday-themes.py` 生成并附带 WCAG 对比度校验，改色请改
> 脚本后重跑，不要手改生成物。

## 1. 总原则

- **一个开关**：`body[data-holiday="…"]`，取值 `mid-autumn | national-day |
  new-year | spring-festival | dragon-boat`；属性缺省 = 原冰山主题。
- **与浅深模式正交**：`html.dark`（或任意 `.dark` 祖先）选择 dark 映射；
  用户的浅色/深色存储选择永不被节日逻辑触碰。
- **虾粉是品牌不是季节**：主 CTA 渐变、首页标题下划弧、连接卡顶条、步骤 2
  序号、航线 `.ice-route` 永远保持虾粉（`--station-brand: #e85887` /
  `--station-brand-deep: #bc285d`），五个节日都不动它。
- **业务语义色钉死**：`--destructive / --success / --warning / --info /
  --neutral` 不随节日变化，错误/成功/警告在任何主题下保持原语义。
- **节日感分级**：首页 = 场景换色 + 角落装饰 + 开屏同款船；控制台与密集
  数据页 = 只换色，不放装饰。

## 2. 五主题意象与配色

| 主题 | 意象 | 主色（浅） | 点缀 | 深色策略 |
|---|---|---|---|---|
| 中秋 · 月夜桂花 | 月弧 / 桂花 / 圆月 | 月夜蓝纸面 `oklch(0.972 …248)`，桂花金主按钮 `oklch(0.72 0.12 88)` 配深夜蓝字 | 链接用深夜蓝 `oklch(0.48 0.11 264)` | 深蓝夜空底 `0.19 L`，月光金提亮主色，银蓝链接 |
| 国庆 · 绛红旗映 | 旗帜折角 / 微小星芒 / 丝带 | 暖白底 + 绛红主色 `oklch(0.5 0.155 24)` 配米白字 | 香槟金进入 accent / 图表 / 星芒 | 深绛底 `0.20 L`，主色提亮为绯红，链接用香槟金 |
| 元旦 · 星轨倒数 | 倒数时钟 / 星轨 / 极少彩带 | 星空白底 + 星空蓝主色 `oklch(0.52 0.13 258)` | 银色进入边框与次系列；彩带只在 SVG 里出现两点 | 深星空蓝底 `0.18 L`，主色提亮为亮星蓝，链接用月银 |
| 春节 · 朱砂灯彩 | 灯笼 / 祥云 / 锦鲤纹 | 宣纸底 `oklch(0.965 0.016 92)` + 朱砂主色 `oklch(0.54 0.185 33)` | 暖金 accent；图表补墨青/黛粉 | 深朱底 `0.21 L`，朱砂提亮，链接用暖金 |
| 端午 · 竹叶青瓷 | 水纹 / 竹叶 / 龙鳞暗纹 | 糯米白底 + 竹叶青主色 `oklch(0.48 0.11 163)` | 青瓷进入 accent 与链接 `oklch(0.46 0.08 198)` | 深竹青底 `0.19 L`，主色提亮为嫩竹，链接用青瓷 |

浅深两套的关系：浅色 = 节日色的"纸面"（底近白、微染色）；深色 =
节日的"夜晚"（底色取该节主色相、L≈0.18–0.21、低彩度），主色与图表色
整体提亮 0.15–0.25 L 以保住对比度。

## 3. 可读性校验（WCAG，脚本自动计算）

校验对：正文/次级文字 ≥4.5:1，主按钮/侧栏激活/accent 标签 ≥4.5:1，
焦点环与图表系列 ≥3:1。**五节日 × 浅深 = 10 套全部通过**；每套最差值：

| 主题 | 浅色最差 | 深色最差 |
|---|---|---|
| 中秋 | 3.27:1（chart-3，≥3 达标） | 4.48:1（chart-3） |
| 国庆 | 4.42:1（chart-3） | 5.65:1（chart-1） |
| 元旦 | 3.46:1（chart-3） | 5.62:1（chart-4） |
| 春节 | 4.16:1（chart-2） | 5.08:1（chart-1） |
| 端午 | 3.92:1（chart-3） | 4.31:1（chart-5） |

普通主题（线上既有值）有两处 3:1 以下的图表浅色注记（chart-3，1.96/2.84），
属锁定品牌基线，不属本次改动。所有正文/次级文字对在所有主题下均 ≥4.5:1。

已知权衡：端午竹叶青与 success 绿同色相族。缓解：主色压暗至 L0.48
（success 为 L0.60 亮绿），状态徽章一律带文字标签与圆点，不靠颜色单独表意。

### 3.1 业务语义色的验收修正（第二轮验收）

验收用 canvas 实测发现线上基线三个浅色语义色配白字不达标。**修正保持
红/绿/橙语义不变，仅压暗底色**；`*-text` 系列用于徽章淡底与独立状态
文字（如 +¥5.00 增量、状态列），同样全部过校验：

| 令牌 | 线上基线（配白） | 建议修正（配白） | 徽章/文字色（浅） | 徽章/文字色（深） |
|---|---|---|---|---|
| destructive | #d64545 → 4.38 ✗ | #cd4242 → 4.72 ✓ | #b93a3a → 5.01 ✓ | #f6a8a2 → 4.90 ✓ |
| success | #1f9d6b → 3.45 ✗ | #1a855b → 4.62 ✓ | #157347 → 5.23 ✓ | #6fe3a8 → 5.74 ✓ |
| warning | #c47f17 → 3.28 ✗ | #a16813 → 4.66 ✓ | #8a5a00 → 5.28 ✓ | #f2c264 → 5.02 ✓ |

（徽章文字色数值为 12–16% 淡色底上的实测对比度。）

**边界声明**：`holiday-themes.css` 始终不重声明业务语义色——是否采纳
上述修正取决于 theme-presets.css 的上游决策，本层不悄悄改变业务行为。
`holiday-theme-tokens.json` 的 `pinnedSemantics.shippedBaseline` 记录了
线上现值，`pinnedSemantics.light/dark` 是建议修正值；预览
（preview.css）已按建议修正值渲染。

### 3.2 图表系列分离（第二轮验收）

端午深色原本 chart-1/2/5 三条绿色系过近。已拉开明度与色相：
chart-1 竹青 `oklch(0.74 0.12 162)`、chart-2 偏青瓷蓝 `oklch(0.8 0.08 208)`、
chart-5 深松绿 `oklch(0.52 0.1 148)`；chart-4 保留品牌粉备用
`oklch(0.72 0.16 350)`。浅色侧 chart-2 同步偏蓝 `oklch(0.54 0.09 205)`、
chart-5 加深 `oklch(0.4 0.08 150)`。国庆深色 chart-3 由 55° 橙移到
48° 并压暗至 L0.66，与绛红（26°）、香槟金（85°）拉开。预览趋势图第三
条系列为虚线 + 图例文字标识，不靠颜色单独区分；中秋深色 245°/265° 两条
蓝色靠 0.18 的明度差区分。

### 3.3 375px 移动验收（第二轮）

以 375×812 视口实测六个页面 `documentElement`：
home / overview / logs / keys / wallet / showcase 全部
`clientWidth = scrollWidth = 375`，无横向溢出，越界元素扫描为 none。
修复点：各 grid 轨道改 `minmax(0, 1fr)`、`.app-content` 子项与卡片
`min-width: 0`、四张表格外包 `overflow-x: auto`（表格允许容器内横向
滚动）、预览工具条长注记允许换行、≤480px 隐藏站点导航中的"控制台"
次级入口（保留品牌 + 主 CTA，按钮不换行）。未使用 body overflow 隐藏。
`document.images` 为 0（节日船改为 div 背景图，普通主题无空 src 元素）。

桌面端复查：首页站点导航（玻璃 nav：品牌 / 模型与额度 / 接入指南 /
常见问题 / 登录）在预览工具条下方正常可见，此前缺失是
`position: absolute` 的包含块落在了初始包含块上、被工具条盖住——已改
为锚定 `#page-home`。导航内的登录入口降为 outline 样式，避免与 hero 的
虾粉主 CTA 同屏双主按钮。

## 4. Token 接入

`holiday-theme-tokens.json` 每主题每模式含 52 个键：background /
foreground / card / card-foreground / popover / popover-foreground /
primary(-foreground) / secondary(-foreground) / muted(-foreground) /
accent(-foreground) / border / input / ring / sidebar 全家 8 个 /
chart-1..5 / overview-accent-1..3 / station-blue / station-glass /
station-rim / station-shadow / station-glow / selection(+foreground) /
skeleton-base/highlight，另附 `brand`（恒定虾粉）与 `pinnedSemantics`
（钉死的业务色）。

`holiday-themes.css` 的挂载：

```css
body[data-holiday='mid-autumn'],
body[data-holiday='mid-autumn'] .station-site,
body[data-holiday='mid-autumn'] .station-scope { /* 全套变量 */ }
.dark body[data-holiday='mid-autumn'], /* + 同上两个 scope */ { /* dark */ }
```

- **body 级**：控制台（无 scope 包裹）与所有 portal 弹层（Base UI 的
  dialog/dropdown/popover 挂在 body 下）经继承直接获得节日变量——
  这正好补上 `.station-scope` 注释里"portal 逃逸 scope"的旧缺口。
- **scope 级**：`.station-site` / `.station-scope` 在 iceberg.css /
  station-scope.css 中重声明了局部变量，必须同名再覆盖一次。
- **特异度**：light 块 (0,1,1)、dark 块 (0,2,2)、scope 块 (0,3,1)/(0,4,2)，
  稳赢 `[data-theme-preset='iceberg']` (0,1,0) 与 `.dark .station-site`
  (0,2,0)；theme-presets.css 的 overview 桥 (0,2,0) 由一条
  `body[data-holiday][data-theme-preset]` (0,2,1) 规则经
  `--holiday-overview-1..3` 中转反超。全程零 `!important`。
- **加载顺序**（必须最后）：theme.css → theme-presets.css → iceberg.css
  → iceberg-cabin.css → station-scope.css → **holiday-themes.css**。
- **不做的事**：无全局 `*` 刷色；不动布局/字号/圆角/控件尺寸；不读写
  `data-theme-preset|font|radius|scale` 及其存储；不对 SVG/Logo 加全局
  filter；`prefers-reduced-transparency` 下玻璃回退为实色 card。

## 5. 与开屏共用日期规则

不新增第二套日期表。在应用侧（如 ThemeCustomizationProvider 同级）：

```ts
import { getOpeningHoliday } from '@/lib/opening-holidays' // 现有文件

// 挂载统一节日状态；用户浅深选择完全不受影响
useEffect(() => {
  const { holiday } = getOpeningHoliday()
  applyAttribute('data-holiday', holiday) // null 时移除属性 = 普通主题
}, [])
```

规则即开屏规则：北京时间、节前 7 日至节后 2 日、中秋优先于国庆、
普通日自动恢复。预览页（index.html）的"自动"档在运行时 fetch 并解析
`reference/opening-holidays.ts` 的 lunarDates 表——单一数据源；
`file://` 直开时 fetch 会被浏览器拦截，此时自动档明示回退为普通主题，
手动选择不受影响（手动档仅用于验收，不代表已接入生产）。

## 6. iceberg.css 硬编码颜色 → 变量替换清单（有限集）

以下是目前唯一硬编码、需要节日感知的位置；holiday-themes.css 已用
`color-mix` 从令牌派生覆盖，无需改原文件。若未来要正本清源，可将这些
位置改为引用变量：

| 位置 | 现状 | 节日层处理 |
|---|---|---|
| `.station-site ::selection` `#ffd0df/#42122d` | 粉 | `var(--selection)` |
| `.station-site` scrollbar `#81bdd8` | 蓝 | station-blue 45% 混入 |
| header nav 底线 `#4a99c029` | 蓝 | station-blue 22% 混入 |
| `.station-button-secondary` 底线 `#5295b540` | 蓝 | station-blue 30% 混入 |
| `.station-hero-scene` / `.station-error` 渐变 | 冰蓝 | station-blue + station-glow 派生 |
| `.station-auth` 两团径向光 | 蓝+粉 | station-blue + `--station-brand`（粉保留） |
| `.station-launchpad` / `.station-auth-panel` 玻璃底与阴影 | 白/深蓝 | `var(--card)` + rim/shadow 派生 |
| `.launchpad-plug` `#f8dce8/#a8295c`（dark `#653047/#ffc1d9`） | 粉 | `var(--accent)` 对（随节日） |
| `.station-auth-route` 轨道 `#84b6cf`、当前步 `#df8fac` | 蓝/粉 | border / primary 派生 |
| `.ice-*` 场景填充 14 类 × 浅深 | 冰蓝系 | station-blue/background/foreground 混比派生 |
| iceberg-cabin 卡片阴影 `rgb(8 119 172 / …)` | 蓝 | station-blue 派生 |
| `.station-error-boat` 投影 `#247ead2e` | 蓝 | station-blue 派生 |

**明确不改**（品牌虾粉）：`.station-button` 渐变、hero 下划弧
`#e85887`、`.station-connection::before` 顶条、步骤 2 描边、
`.ice-route` 航线。

## 7. 装饰 SVG

`assets/holiday-*.svg` 共 10 个：每节日 1 主 1 辅（圆月+桂花枝、
旗帜+星芒、倒数钟+星轨、灯笼+祥云/锦鲤纹、竹叶+水纹龙鳞）。全部纯矢量、
无位图、无外链、无 filter，沿用五船的多边形切面 + 描边语言，配色直接
取自各节日已批准 chip/boat 的调色板，浅深底上均可读。挂载方式为
`.station-hero-scene::before/::after` 绝对定位伪元素（上右主、下右辅），
不挡数字、不占布局；≤760px 收缩为单个小装饰；`prefers-reduced-motion`
下漂移动画关闭。节日期间首页还会在同层引用
`reference/holidays/<id>/boat.svg`（只读，未改动）。

## 8. 图表

chart-1..5 每主题色相间隔 ≥40°（如春节：朱砂 33 / 暖金 82 / 墨青 205 /
黛粉 350 / 竹青 165），浅深两套均 ≥3:1 于卡片底。概览三卡经
overview-accent-1..3 独立映射。预览页趋势图为三条系列（面+线、线、
虚线），全部读 `var(--chart-*)`，切主题即时联动；图例常驻，不依赖
颜色单独表意。

## 9. 评审结论

- 信息架构：五个控制台页 + 首页结构与原站一致，节日只换"皮肤层"。
- 对比度：10 套状态全部通过脚本校验（见 §3）。
- 键盘：所有可交互元素有 `:focus-visible` 环（`var(--ring)` 随节日），
  Esc 关闭弹层，触控目标 ≥44px。
- 移动端：375px 无横向滚动（表格在窄屏隐藏次要列，侧栏折为横排）。
- 限制：预览的"自动"档在 `file://` 下回退普通主题（日期表需经 HTTP
  读取）；真实接入只需应用侧加 §5 的四行挂载 + 引入
  `holiday-themes.css`。
