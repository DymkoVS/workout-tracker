# Handoff: Workout Tracker — Mobile-first Dark Redesign

Hand-off for the existing repo **`DymkoVS/workout-tracker`** (Go + `html/template` + Tailwind + htmx + Chart.js + Flatpickr, deployed at `dymko.ru`). The current UI is stock Tailwind on light grey with blue buttons and emoji icons. This handoff redesigns it as a **mobile-first, dark, brutalist-sport** product with a single neon-lime accent.

---

## About the design files

The files in this bundle are **design references created as a browser-side React/JSX prototype**. They are **not production code to copy directly**. Your task is to recreate these designs in the existing codebase — i.e. update the Go `html/template` files under `web/templates/` to match this visual system, keeping the existing routes, htmx flows, and data model intact.

You may keep Tailwind (the cleanest path) and just swap the palette and reach for custom utilities, or you can move to a small custom stylesheet — your call based on how much reuse there is. The current Tailwind CDN script (`<script src="https://cdn.tailwindcss.com">`) supports inline config which is enough to declare the design tokens below.

The prototype lives in:

- `Workout Tracker - Redesign.html` — page chrome, audit panels, design-system panel, screen-jump pills
- `app.jsx` — main app screens (Login, Home, History, Detail, Log, Analytics, Profile) + bottom tab nav
- `admin.jsx` — trainer/admin screens (Clients, ClientDetail, Templates, TemplateApply, Users, Assign)
- `ios-frame.jsx` — iPhone device bezel for presentation (NOT to be reproduced — it's just the showcase frame)

Open `Workout Tracker - Redesign.html` in a browser to explore the prototype interactively. The pills above the phone let you jump between screens.

## Fidelity

**Hi-fi.** Colors, type, spacing, sizes, and component states are all final. Recreate pixel-perfectly using the codebase's existing patterns (Tailwind + htmx). The data model is unchanged; the routes are unchanged; **only the templates and styling change**.

---

## Design tokens

### Colors

```
--bg:        #0a0a0a   /* page background */
--surface:   #141414   /* cards, list rows */
--surface-2: #1b1b1b   /* nested chips, num pads */
--hair:      #262626   /* primary borders / dividers */
--hair-2:    #1f1f1f   /* in-card dividers (softer) */
--text:      #fafafa   /* primary text */
--dim:       #7a7a7a   /* secondary text */
--dim-2:     #4a4a4a   /* tertiary text, disabled */
--accent:    #D7FF1A   /* THE single accent — actions, PR, tonnage, active states */
--accent-dk: #a8c800   /* hover/pressed accent */
--danger:    #ff453a   /* destructive, missed plan */
--good:      #30d158   /* rare — success confirmations */
--warn:      #ff9f0a   /* warning bars in compliance grid */
```

**Rule:** the accent is **never** used decoratively. It is only on: primary CTAs, the active set in a logged workout, PR flags, tonnage values, completed-set markers in compliance grids, and "current" indicators (badges, dots).

### Typography

Three families, loaded from Google Fonts:

```
"Anton"               — display: weights, reps, screen titles, big numbers
"Space Grotesk"       — body/UI: 400/500/600/700
"JetBrains Mono"      — tabular: timer, RPE, dates, ratios (e.g. 9/17), deltas
```

Google Fonts URL used in prototype:

```
https://fonts.googleapis.com/css2?family=Anton&family=Space+Grotesk:wght@400;500;600;700&family=JetBrains+Mono:wght@400;500;600&display=swap
```

Type scale:

| Token       | Family        | Size | Line-height | Letter-spacing | Use                                                |
|-------------|---------------|------|-------------|----------------|----------------------------------------------------|
| display-xl  | Anton         | 96px | 0.88        | 1px            | Login wordmark only                                |
| display-l   | Anton         | 56px | 0.95        | 0.5px          | Workout detail title                               |
| display-m   | Anton         | 48px | 1           | 0.5px          | Screen titles (History, Progress, Profile, …)      |
| display-s   | Anton         | 32px | 1           | 0.5px          | Section CTA titles, large stats                    |
| display-xs  | Anton         | 22px | 1.05        | 0.5px          | Exercise names, template names                     |
| num-l       | Anton         | 64px | 0.9         | 0.5px          | Hero metric in analytics                           |
| num-m       | Anton         | 44px | 0.9         | 0.5px          | Stat blocks on Home, frequency hero                |
| num-s       | Anton         | 30px | 0.9         | 0.5px          | Small stat blocks (4-up rows)                      |
| num-xs      | Anton         | 22px | 1           | 0.5px          | Set values (weight, reps)                          |
| body        | Space Grotesk | 14px | 1.45        | 0              | Default body text                                  |
| body-s      | Space Grotesk | 12px | 1.45        | 0              | Secondary copy                                     |
| eyebrow     | Space Grotesk | 10px | 1           | 1.8px          | UPPERCASE section labels, weight 600              |
| eyebrow-xs  | Space Grotesk | 9px  | 1           | 1.4–1.8px      | Tiny labels under stats and chips, weight 600     |
| mono-m      | JetBrains Mono| 22px | 1           | 1px            | Active workout timer                               |
| mono-s      | JetBrains Mono| 13px | 1           | 0              | RPE values, exercise volume                        |
| mono-xs     | JetBrains Mono| 11px | 1           | 0              | Dates, ratios, deltas                              |
| mono-xxs    | JetBrains Mono| 10–9 | 1           | 1.4px          | Chart labels                                       |

### Spacing

The prototype uses straight pixel values, not a strict scale. Approximate rhythm:

- Screen horizontal padding: **24px**
- Section vertical padding (between blocks): **20–28px**
- Card inner padding: **18–20px**
- Row inner padding: **14–18px**
- Top padding on first screen content (below status bar): **56–64px**
- Bottom padding to clear tab bar / sticky CTA: **96–120px**

### Borders & shapes

- **No rounded corners.** Cards, buttons, inputs are sharp (`border-radius: 0`).
- Hairline borders: `1px solid var(--hair)` for cards; `1px solid var(--hair-2)` for in-card dividers.
- **No shadows.** Depth is communicated by colour and stroke only.
- Active state on a "current set": `1.5px solid var(--accent)` with `background: #000` (forces a deeper black against the card surface).
- A 4px-wide vertical accent strip on a card means "this is the most important card on the screen" (used on Home — last workout, and Detail — title).

### Icons

Outline only, 1.8px stroke. The prototype's `Icon` object in `app.jsx` is the source of truth. No emoji as functional icons. (Emoji are used **only** for the 5-point wellbeing scale on workouts — see "Iconography" below.)

---

## Screens

Each screen maps to one or more existing Go templates. I list the file you'll be editing, the route, and the rebuild spec.

### Iconography (used across screens)

- **Wellbeing scale** (1–5, stored as `*int` in `Workout.Wellbeing`): keep emoji `💀 😣 😐 🙂 🔥` (your `wellbeingEmoji` template helper already does this). On dark surfaces they read well at 18–24px.
- **PR flag**: 9px Anton "PR" in lime on a transparent background, never an emoji or a star.
- **Streak**: `🔥{n}` rendered as Space Grotesk text, lime when n>0, dim-2 when n=0.
- **Status dot**: a 6×6 square (not circle) in the relevant color, with `●` glyph used only inside text labels (e.g. "● ПРОПУСК", "● ИДЁТ ТРЕНИРОВКА").

---

### 1. Login — `web/templates/login.html`

Route: `GET /login`. Form posts to `/login`. Error message rendered via existing `.Error` field (`invalid` / `inactive`).

**Layout:**

- Full viewport, `background: var(--bg)`. No nav (user not authenticated).
- Padding `120px 28px 32px` (more top padding, anchored bottom).
- Content area (top): eyebrow `v 1.0 · MOBILE` in lime; below it the wordmark `WORK` / `OUT.` on two lines, Anton 96px, line-height 0.88; a 56×6 lime stripe; a small dim caption "Личный журнал тренировок. / Подходы, RPE, тоннаж, прогресс."
- Form (bottom): two `Field` inputs (Логин, Пароль), then a full-width lime submit button.

**Field component:**

- Tiny uppercase label (`eyebrow`, 10px, color `--dim`).
- Input has **no box**: transparent background, no border, only `border-bottom: 1px solid var(--hair)`, padding `10px 0 12px`, font 20px JetBrains Mono, letter-spacing 1px, white text.

**Submit button:**

- Height **60px**, full width.
- Background `--accent`, text `#000`.
- Anton 22px, letter-spacing 2px, text `"ВОЙТИ →"`.

**Below the button:** a dim hint `admin / admin123`, eyebrow style, centered. (Keep this for now since it's a personal tool.)

**Error state:** when `.Error == "invalid"` or `"inactive"`, render a slim red strip above the form: `padding: 12px 14px; background: rgba(255,69,58,0.12); color: var(--danger); font-size: 12px; letter-spacing: 1.2px; text-transform: uppercase`. No rounded corners.

---

### 2. Home (Dashboard) — `web/templates/dashboard.html`

Route: `GET /`. Reuses `.CurrentUser` from `internal/handler/dashboard.go`. This screen needs **new data** the current handler doesn't compute — see "Data additions" at the bottom.

**Top bar (no global nav element on this screen — see "Navigation"):**

- Padding `64px 24px 0`.
- Left: eyebrow with today's date `"13 МАЯ · СРЕДА"`; below it Anton 36px `"ПОРА В ЗАЛ."` (or `"СНОВА В ЗАЛЕ."` if returning the same day).
- Right: a 44×44 square avatar — `background: var(--surface); border: 1px solid var(--hair)`. Inside, a user-outline icon. Tapping this opens Profile.

**Week stats (3-up grid):**

- Padding `24px 24px 0`.
- Eyebrow `"ЭТА НЕДЕЛЯ"`.
- Grid: `4` (трен.) / `12.5т` (тоннаж, in lime) / `6дн` (стрик).
- Each stat: Anton 44px number, 18px dim unit, 10px eyebrow label below.
- Below the row a 1px `var(--hair)` rule.

**Primary CTA — "Next Workout" card:**

- Padding `20px 24px 0`.
- Full-width button. Background `--accent`, color `#000`. Padding `22px 24px`.
- Layout: flex row, space-between.
- Left column: tiny eyebrow `"СЛЕДУЮЩАЯ"` (700, opacity 0.7); below it Anton 32px workout title `"UPPER / PUSH"`; below it 12px `"3 упражнения · ~ 50 мин"`.
- Right column: 52×52 black-bordered (`2px solid #000`) square holding an arrow-right outline icon.
- This is **the single most important affordance on the home screen.** Hover/active: darken accent to `--accent-dk`.

**Last workout:**

- Eyebrow `"ПОСЛЕДНЯЯ"` with right-aligned `Все →` link to `/workouts`.
- Renders a `WorkoutCard` (see Components below) of the most recent workout. Tap opens that workout's detail page.

**Fresh PRs:**

- Eyebrow `"СВЕЖИЕ РЕКОРДЫ"`.
- A single surface card containing 3 rows, separated by `--hair-2` hairlines.
- Each row: 16px lime bolt icon, exercise name, reps subtitle on the left; on the right an Anton 28px weight value (e.g. `90кг`) and a tiny mono `▲ +2.5 кг` in lime.

---

### 3. History — `web/templates/workouts/list.html`

Route: `GET /workouts`.

**Top bar:**

- Eyebrow with total count `"23 ТРЕНИРОВОК"`; below it Anton 48px `"ЖУРНАЛ"`.
- Right: a 44×44 square filter button with a "filter / sort" outline icon (3 lines, decreasing length).

**Trend strip:**

- A surface card containing a sparkline. Eyebrow `"ТОННАЖ · 6 ПОСЛЕДНИХ"` on the left, `+14% к мес.` in mono lime on the right.
- The sparkline is an SVG path drawn in lime, stroke-width 2, with circle dots at each point. Container height 60px, padding 4px inside.

**Workouts list:**

- Grouped by month. Each month header is an eyebrow `"МАЙ 2026"`.
- 10px gap between cards inside a group.
- Each item is a `WorkoutCard` (see Components).

**Empty state** (no workouts): keep the current copy but render in the dark style — Anton 24px title `"ТРЕНИРОВОК ПОКА НЕТ"`, dim caption, and a single lime CTA `"+ ДОБАВИТЬ ПЕРВУЮ"`.

---

### 4. Workout Detail — `web/templates/workouts/show.html`

Route: `GET /workouts/{id}`.

**Top bar:**

- A left-aligned `← Журнал` back button (lined arrow + 12px uppercase label).
- Right-aligned a row of two outline icons: `edit` and `trash`. Trash triggers existing delete confirm.

**Title block:**

- Eyebrow with date and gym: `"12.05.2026 · World Class · Кутузовский"`.
- Below it Anton 56px workout title — uppercase rendering (`{{ .Title | toUpper }}`).
- A 56×4 lime stripe under the title.

**Summary stats (4-up):**

- `Stat` blocks: упражнения, подходов, тоннаж (lime), состояние (the emoji from wellbeing).

**Exercises (one card per exercise):**

Each exercise card has 3 parts:

1. **Header** (`18px 20px 12px`):
   - Left: eyebrow `"УПР. 01"` (zero-padded number), Anton 22px exercise name uppercased.
   - Right: eyebrow `"ОБЪЁМ"`, below it a mono 13px lime value computed as `sum(set.weight * set.reps)`.

2. **Sets table** (`0 8px 12px`):
   - 4-column grid: `32px 1fr 1fr 1fr` — #, ВЕС, ПОВТ., RPE (right-aligned header).
   - Each row 10px vertical padding.
   - **The top set** (the one with max `weight * reps`) is highlighted: `background: rgba(215,255,26,0.06)` and `border-left: 2px solid var(--accent)` (instead of `2px solid transparent` for non-top rows).
   - Values: number in mono dim; weight + reps in Anton 24px with mini "кг" suffix; RPE in mono — **lime if this row is the top set, dim otherwise**.

3. **Per-exercise note** (if present): italic dim 14px below the table.

**Workout note (if present):**

- Below all exercises.
- Background `--surface-2`, padding 16/18, **3px lime left border**.
- Eyebrow `"ЗАМЕТКА"` in lime; body text below in white.

---

### 5. Log Workout (active session) — **NEW**

This is the most important screen and currently does not exist. The existing `web/templates/workouts/form.html` is a giant single form. Split it into:

- A **start screen** (still `form.html`, simplified) that captures title / date / gym / wellbeing / optionally pulled from a template.
- A new **active session** template — call it `workouts/active.html` — that renders the in-progress workout as a structured set-by-set tracker.

Route suggestion: `GET /workouts/{id}/active`. The active session reads `Workout` rows where `EndedAt IS NULL`. See "Data additions" below.

**Sticky header:**

- Position sticky at top, `border-bottom: 1px solid var(--hair)`, background `var(--bg)`. Padding `56px 20px 14px`.
- Row 1: back-arrow (left), center lime eyebrow `"● ИДЁТ ТРЕНИРОВКА"`, right danger button `"ЗАВЕРШИТЬ"` (background `--danger`, white, 10px letter-spaced 1.6, padding `6px 12px`).
- Row 2 (3-up): ВРЕМЯ (mono 22px white, format `MM:SS`) · ПОДХОДЫ (mono 22px `N/M` with `/M` dimmed) · ТОННАЖ (mono 22px lime).
- Row 3 (rest timer — visible only when `restSeconds > 0`): a lime strip with timer icon, label `"ОТДЫХ"`, mono 22px `"0:56"`, and a `"ПРОПУСТИТЬ"` button on the right.

**Exercise blocks (one per exercise):**

Each block has 3 states:

- **All sets done** — block opacity 0.55, border `1px solid var(--dim-2)`. Header shows a lime check icon next to the `"УПР. 01"` eyebrow. The sets table is collapsed (or shown but greyed).
- **In progress** — full opacity, regular `--hair` border. Sets table is expanded.
- **Locked / future** — same as in progress but no active set highlighted; all rows shown at 0.45 opacity.

Header (any state):

- Left: eyebrow with `"УПР. 01"`, Anton 22px exercise name uppercased, below it a mono 11px "Прошлый раз: 90×5 @ RPE 9" pulled from the user's last workout containing this exercise name (server-side lookup).
- Right: Anton 22px dim `"3/4"` (sets done / total).

Sets table — for **each set**, render one of three row variants:

1. **Done set:**
   - 5-column grid `36px 1fr 1fr 1fr 48px`.
   - Number in mono dim.
   - Anton 22px weight + reps.
   - Mono 13px dim RPE.
   - Right cell: either a dim 16px check icon **or**, if this is a PR, the lime label `"PR"` and the row gets `background: rgba(215,255,26,0.07)` + `border-left: 2px solid var(--accent)`.

2. **Active set (the "money" row):**
   - `margin: 6px 4px; padding: 14px; background: #000; border: 1.5px solid var(--accent); position: relative`.
   - A tiny lime tab `"ТЕКУЩИЙ"` floating on the top-left edge (`position: absolute; top: -8px; left: 12px; padding: 1px 8px; background: var(--accent); color: #000; font-size: 9px; letter-spacing: 1.4px; font-weight: 700`).
   - Inside: a 3-column grid of `NumPad` panels (ВЕС, ПОВТ, RPE). Each pad: `background: var(--surface-2); padding: 10px 4px; text-align: center`; eyebrow label dim-2; Anton 30px lime number with mini "кг" suffix.
   - Below the pads, a full-width lime button: Anton 18px, letter-spacing 2px, text `"✓ ПОДХОД ВЫПОЛНЕН"`. On click: mark the set done, advance the active state to the next set (next exercise if last set), and start a **90s rest timer**.

3. **Queued (future) set:**
   - Same 5-col grid, all values dimmed (opacity 0.45). The weight/reps are the **suggested** numbers (from template or previous workout).

Add-set button per exercise:

- Full-width dashed-border button under the sets: `border: 1px dashed var(--hair); color: var(--dim); padding: 10px; font-size: 11px; letter-spacing: 1.5px; text-transform: uppercase; text: "+ Подход"`. Add a new queued set on click (htmx POST or local state, your call).

Add-exercise button at the very bottom of the list (same dashed style, slightly larger padding, text `"+ Добавить упражнение"`).

**Interactions to implement** (the existing system has none of these):

- Pressing the active-set CTA: mark set done, advance active to next set, start a 90s rest timer.
- "Skip" on rest timer: zero it.
- Auto-prefill weight/reps from the last set of the previous workout for this exercise (or from a template if applied).
- "PR" detection: a set is a PR if `weight * reps` strictly exceeds any previous set of the same exercise for this user.

---

### 6. Analytics — `web/templates/analytics/index.html`

Route: `GET /analytics`. The existing template uses Chart.js with blue/green/amber. Replace with custom SVG (matching the rest of the design) — or restyle Chart.js with lime + transparent fills if you prefer to keep the library.

**Top bar:**

- Eyebrow `"ПОСЛЕДНИЕ 90 ДНЕЙ"`.
- Anton 48px `"ПРОГРЕСС"`.
- For trainers, the existing client picker stays — but restyled to look like the rest of the dark UI (a surface chip with a chevron, not a stock `<select>`).

**Tonnage hero card:**

- Surface card, padding 20px. Eyebrow `"ТОННАЖ · 90 ДНЕЙ"`.
- Big Anton 64px value (e.g. `"184.2т"`), with mono lime delta `"▲ +18%"` next to it.
- Below: an SVG area chart, stroke lime 2px, fill `url(#fadeAcc)` — a vertical gradient from `rgba(215,255,26,0.25)` to transparent. X labels in mono 9px dim-2 (dates DD.MM, every 2nd point).

**Exercise progress card:**

- Same structure as tonnage. Adds a styled `<select>` for exercise picker.
- Hero value is the all-time max weight for that exercise; chart is the per-workout max for the last 90 days.

**Frequency card:**

- Eyebrow `"ЧАСТОТА · 8 НЕДЕЛЬ"`.
- Anton 44px `"3.8 /нед"`.
- Vertical bar chart, 8 bars, gap 6px. Past bars in `--dim-2`, the **last bar (current week)** in `--accent`. Mono 9px value labels above each bar; mono 8px `W{i+1}` labels below.

**Split breakdown card:**

- Eyebrow `"СПЛИТ ПО МЫШЕЧНЫМ ГРУППАМ"`.
- 4 rows (Грудь/Трицепс, Спина/Бицепс, Ноги, Плечи/Кор). Each row: label + mono percent on the right; below them a 6px-tall horizontal bar (filled to `pct%`). The first row uses lime, the rest use white, dim, dim-2 (visual hierarchy by intensity, not chroma).

---

### 7. Profile — **NEW** (replaces simple right-side dropdown of `base.html`)

There is currently no profile screen — admin/trainer affordances are scattered in the navbar. Create `web/templates/profile.html` at route `GET /profile`. The bottom tab bar (see Navigation) routes here.

**Top bar:**

- Eyebrow `"АДМИНИСТРАТОР · ТРЕНЕР"` (depending on roles).
- Anton 48px login uppercased (`"ADMIN"`).
- A 56×4 lime stripe.

**Account stats (2×2 grid):**

- Всего тренировок · Тоннаж (lime) · Макс. стрик · Залов.

**Sections:**

1. **ТРЕНЕРСКОЕ** (visible if `IsTrainer`) — surface card with rows:
   - `Клиенты` → `/trainer/clients` (badge: count).
   - `Шаблоны тренировок` → `/templates` (badge: count).
   - `Залы` → `/gyms`.

2. **АДМИНИСТРАЦИЯ** (visible if `IsAdmin`) — surface card with rows:
   - `Пользователи` → `/admin/users` (badge: count).
   - `Назначения` → `/admin/assign` (badge: count of trainer↔client links).

Each row: 16px label, mono 11px badge dim, → arrow dim-2 on the right.

3. **Logout button** at the bottom — transparent with `1px solid var(--hair)`, uppercase 12px letter-spaced 1.6, danger color text. Posts to `/logout`.

---

### 8. Clients (trainer) — `web/templates/trainer/clients.html`

Route: `GET /trainer/clients`. Adds new computed fields per client — see "Data additions".

**Top bar:**

- Back button to `/profile`.
- Eyebrow `"5 ВСЕГО · 4 АКТИВНЫХ"`.
- Anton 48px `"КЛИЕНТЫ"` + 56×4 lime stripe.

**Week pulse card:**

- Eyebrow `"НЕДЕЛЯ · ВЫПОЛНЕНО"` + right-aligned mono lime `"9/17 трен."`.
- Anton 44px `"53%"` with mono danger `"▼ -12%"`.
- Below: a horizontal bar split into segments per client, where each segment's width is proportional to that client's `weekPlan` and whose color is:
  - lime if `weekDone >= weekPlan`
  - warn (`--warn`) if 0 < weekDone < weekPlan
  - dim-2 if weekDone == 0

**Clients list:**

- Each client is a surface card row, 16/18 padding.
- If `status == 'off'` (no workouts in the last 5 days), border becomes `1px solid rgba(255,69,58,0.25)` instead of `--hair`, and a tiny right-aligned lime/danger eyebrow `"● ПРОПУСК"` appears next to the name.
- Layout: 44×44 surface-2 initials avatar (Anton 18px, dim, 2 initials) → name + meta → week progress on the right.
- Meta line (below name): `{goal} · @{login}` in 11px dim.
- Stats line (below meta): mono 11px row — `38 трен.` · `послед. 12.05` · `🔥6` (lime if streak > 0, else dim-2).
- Right side: a vertical column of mini "pills" — N rectangles 6×14 each, where N is `weekPlan`. Filled (lime) for the first `weekDone`, hair for the rest. Below them mono 9px `"3/4 нед."`.

**Bottom CTA:** ghost-style full-width button `"ПРИМЕНИТЬ ШАБЛОН КО ВСЕМ"` linking to `/templates` (or directly to the template-apply flow).

---

### 9. Client detail (trainer) — **NEW**

There is currently no per-client detail page — the existing `trainer/client_workouts.html` is just the workouts list. Create `web/templates/trainer/client.html` at `GET /trainer/clients/{id}`.

**Title block:**

- Back to `/trainer/clients`.
- Eyebrow `"{goal} · @{login}"`.
- Anton 44px client name uppercased.
- 56×4 lime stripe.

**Stats (4-up):** Trainings · Week ratio (lime if met) · Streak (lime if > 0) · Average tonnage.

**Compliance grid** (this is a small custom element — implement as inline HTML or SVG):

- Eyebrow `"СОБЛЮДЕНИЕ ПЛАНА · 4 НЕД."` with right-aligned mono lime percent.
- A horizontal flex row of 16 cells, each 28px tall, lime if the planned workout happened, hair if missed. 4px gap. Mini mono labels under the grid: `нед -3`, `нед -2`, `нед -1`, `сейчас`.

**Recent workouts** — a 4-row card list. Each row: date in mono dim, Anton 18px title, mono "тоннаж · 4.8т" (lime value), wellbeing emoji on the right.

**Actions** (sticky-feeling block at the bottom):

- Primary lime CTA `"+ НАЗНАЧИТЬ ТРЕНИРОВКУ"` → `/templates` (or directly to the apply flow scoped to this client).
- Below it a 2-column row of ghost buttons: `"НАПИСАТЬ"`, `"ПЛАН НЕДЕЛИ"` (these can link to placeholders for now).

Existing route `GET /trainer/clients/{id}/workouts` can either redirect to this new page or stay as a sub-page linked from "Все →".

---

### 10. Templates — `web/templates/templates/list.html`

Route: `GET /templates`.

**Top bar:**

- Back to `/profile`.
- Eyebrow with count, Anton 48px `"ШАБЛОНЫ"`.
- 44×44 lime square `+` button on the right linking to `/templates/new` (Anton 28px `+`, lime bg, black glyph).

**Filter tabs:** `Все · Сила · Кардио · Аксессуар` — a row of small chips. The active chip is `var(--text)` background with `#000` text. Inactive chips have `var(--hair)` border, dim text.

(NB: the current model doesn't have a "type" field. Add `type` to the `WorkoutTemplate` struct — see "Data additions". Default existing templates to "Сила".)

**Template cards:**

- Each template is a surface card with padding 18/20.
- Eyebrow top-left: the type uppercased; top-right: mono 10px `"применён ×N"` (where N is a count of `Workout` rows derived from this template — needs a new query or a `template_id` column on workouts, see "Data additions").
- Below: Anton 28px template title.
- A row of 3 minis: УПР., ПОДХ., ВРЕМЯ (rough estimate).
- Two-button row: a primary lime `"ПРИМЕНИТЬ"` (50% width) → `/templates/{id}/apply`, and a ghost `"ОТКРЫТЬ"` → `/templates/{id}`.

---

### 11. Template apply — `web/templates/templates/apply.html`

Route: `GET /templates/{id}/apply`, posts to same path. The current template uses flatpickr for date and a flat list of checkboxes. Replace with:

**Title block:** back to `/templates`, lime eyebrow `"ПРИМЕНИТЬ ШАБЛОН"`, Anton 36px template title, 12px dim sub `"5 упражнений · 14 подходов · ~45 мин"`.

**Date + gym card:**

- Eyebrow `"ДАТА"`.
- **A 7-day horizontal date picker** (replaces flatpickr). Render the next 7 days as 7 equal-width cells in a flex row. Each cell:
  - Anton 22px day number (e.g. `14`).
  - Eyebrow 9px day-of-week (`ЧТ`).
  - Border: `1px solid var(--hair)`. Selected cell: `background: var(--accent); color: #000; border: 1px solid var(--accent)`.
- Below: eyebrow `"ЗАЛ"` and a row of gym chips (one per gym). Selected chip has lime left dot and lime border, others ghost.

**Clients select:**

- Eyebrow with ratio `"КЛИЕНТЫ · 2/5"` and a right-aligned lime link `"Выбрать всех"` (toggles to `"Снять"` when all selected).
- Surface card with one row per assigned client. Each row:
  - A custom 22×22 checkbox square — `background: var(--accent)` when checked with a 14px black check icon, `border: 1.5px solid var(--dim-2)` when not.
  - Name in 14px white, goal in 10px dim below.
  - Mono 10px `"3/4 нед."` on the right.

**Sticky bottom CTA:**

- `position: absolute; left: 0; right: 0; bottom: 0; padding: 14px 20px 28px; background: var(--bg); border-top: 1px solid var(--hair)`.
- Lime button Anton 20px `"СОЗДАТЬ N ТРЕНИРОВКУ/ТРЕНИРОВКИ/ТРЕНИРОВОК →"` (Russian pluralisation: 1 → тренировку, 2–4 → тренировки, 5+ → тренировок). Disabled (opacity 0.3) when zero selected.

---

### 12. Admin users — `web/templates/admin/users.html`

Route: `GET /admin/users`. The current template renders an HTML table. Replace with a mobile card list.

**Top bar:**

- Back to `/profile`.
- Eyebrow count.
- Anton 48px `"ПОЛЬЗОВАТЕЛИ"` + a 44×44 lime `+` button → `/admin/users/new`.

**Counters (3-up):** ТРЕНЕРЫ · КЛИЕНТЫ (lime accent) · НЕАКТ. (danger color). Each as a small surface card with Anton 28px number and 9px eyebrow.

**Filter tabs:** `Все · Тренеры · Клиенты · Неактивные` — same chip style as Templates. Horizontally scrollable on overflow.

**User rows:**

- Surface card list, one row per user.
- 36×36 surface-2 initials avatar on the left.
- Name + small ADMIN badge (8px Anton in `#000`, lime background, padding `2px 6px`, letter-spacing 1.5px).
- Below: mono 10px `"@{login} · {role}"`, with role in lime if `тренер`. If inactive, append `"● деактивирован"` in danger.
- Whole row opacity 0.5 if `!IsActive`.
- 3-dot menu icon on the right → links to edit + activate/deactivate.

**Footer CTA:** ghost button `"Назначения тренер ↔ клиент →"` → `/admin/assign`.

---

### 13. Admin assign — `web/templates/admin/assign.html`

Route: `GET /admin/assign`, POST `/admin/assign` for new, POST `/admin/assign/{trainerID}/{clientID}/remove` for removal.

**Title block:** back to `/admin/users`, eyebrow with totals, Anton 44px `"НАЗНАЧЕНИЯ"`, 56×4 lime stripe.

**Quick assign card:**

- Eyebrow lime `"+ НОВАЯ СВЯЗКА"`.
- Two custom-styled `<select>`-equivalents: a label, then a surface-2 box with the current selection + a `▾` glyph on the right. (Restyle the native `<select>` with `appearance: none` + custom chevron, or use a dropdown component if available in the codebase.)
- Lime button `"НАЗНАЧИТЬ →"`.

**Current pairings:**

- Group by trainer. Each trainer is a surface card with:
  - Header row (with `--hair-2` bottom border): eyebrow `"ТРЕНЕР"`, Anton 22px trainer name uppercased; on the right an Anton 28px lime count and 9px `"КЛИЕНТОВ"` eyebrow.
  - One row per client: a 6×6 lime square, the client name in 14px, and a danger `"Убрать"` link on the right (10px letter-spaced uppercase). Each row uses an htmx POST to the existing remove endpoint with the same confirmation.

---

## Components

### `WorkoutCard`

Used on Home (last workout) and History (all workouts). A button-shaped surface card, full width.

- Layout: 4px lime vertical stripe on the left (`align-self: stretch`), then the content with 20px padding.
- Header row: eyebrow with date (`12.05.2026`) on the left, wellbeing emoji 18px on the right.
- Anton 24–32px title (`big` variant when standalone, 24px in lists).
- Gym name in dim 11px.
- 3-mini row: упр., подх., тоннаж (`X.Yт` lime if you want to draw the eye to volume).

### Bottom tab bar (`TabBar`)

- Fixed `position: absolute; bottom: 0; left: 0; right: 0; padding-bottom: 28px; background: var(--bg); border-top: 1px solid var(--hair); display: grid; grid-template-columns: repeat(5, 1fr)`.
- 5 items: Главная / Журнал / **+ Новая** / Прогресс / Профиль.
- Center FAB: a 50×50 lime square overlapping the bar top edge by 22px, with a 6px `--bg` ring (`box-shadow: 0 0 0 6px var(--bg)`). 26px black `+` glyph.
- All tab labels: 8px Anton-ish (Space Grotesk 600) uppercase letter-spaced 1.4px. Active tab text is white, inactive is `--dim-2`.

The tab bar is **not** shown on:

- Login.
- Active log session (it has its own header chrome and a `ЗАВЕРШИТЬ` button instead).
- Workout detail (just a back arrow is enough).
- All trainer/admin screens (Clients, ClientDetail, Templates, TemplateApply, Users, Assign — they have their own back-arrow nav to `/profile`).

This is currently the topmost `<nav>` in `base.html`. **Move it to the bottom**, restyle it, and only render it on the screens listed above.

### Eyebrow

```css
font-family: "Space Grotesk", system-ui;
font-size: 10px;
letter-spacing: 1.8px;
text-transform: uppercase;
color: var(--dim);
font-weight: 600;
```

Use `--accent` color when the eyebrow signals an active state (e.g. `"● ИДЁТ ТРЕНИРОВКА"`, `"+ НОВАЯ СВЯЗКА"`, `"ЗАМЕТКА"`). Use `--dim-2` when it's a tertiary label (e.g. `"УПР. 01"` inside a card).

### Stat

A small reusable display unit:

```
[44px Anton number][18px dim unit]
[10px eyebrow letter-spaced 1.5 UPPERCASE label]
```

Sizes scale: 64/22 (hero), 44/18 (large), 30/12 (small). Tonnage stats always lime.

---

## Interactions & behavior

### Navigation

The current `base.html` has a top nav with role-dependent links. **Replace it with the bottom tab bar** (see Components). Move trainer/admin entry points into the Profile screen. Keep all existing routes — only the chrome changes.

### Active workout flow

This is the largest behavioural change. Today's flow is "fill in a big form once". The new flow:

1. User taps `+ НОВАЯ` (center tab) on any screen.
2. A start screen captures title / date / gym / wellbeing / optional template — this is essentially today's `form.html`, simplified (no exercises section). On submit, a `Workout` row is created with `started_at = now()` and `ended_at = null`.
3. User lands on `/workouts/{id}/active`. This is the new active-session template — the screen with the rest timer and active-set highlighting.
4. As the user marks sets done (htmx PATCH per set), the server stores them with timestamps. Rest timer is purely client-side.
5. Tapping `ЗАВЕРШИТЬ` sets `ended_at = now()` and redirects to the detail page.

The existing single-shot create form should still work for editing past workouts and for trainers logging on behalf of clients.

### Animations and transitions

Sparse. The design relies on contrast, not motion. Allowed:

- `transition: opacity 150ms` on tab changes.
- `transition: background 100ms` on button hover/active.
- The accent button on hover: `background: var(--accent-dk)`.

Do **not** add card-rise/shadow hovers, slide-in panels, or skeleton shimmer. The aesthetic is "control panel", not "consumer".

### Form validation

The current backend validation rules apply unchanged. Surface errors as the danger strip described in Login.

---

## Data additions

The following fields/queries do not exist yet and are needed for the redesign. None of them require destructive migrations.

1. **`Workout.StartedAt` and `Workout.EndedAt` (`*time.Time`)** — for the active-session flow. Migration adds two nullable timestamp columns.
2. **`Workout.TemplateID` (`*uuid.UUID`)** — link a logged workout back to the template it was created from. Powers the "применён ×N" counter on Templates list.
3. **`WorkoutTemplate.Type` (`string`)** — one of `сила | кардио | аксессуар`. Defaults to `сила` for existing rows. Drives the Templates filter tabs.
4. **A "last set for exercise" lookup** — server helper that, given a user + exercise name, returns the most recent `Set` from any of that user's workouts. Powers the "Прошлый раз: 90×5 @ RPE 9" hint and the auto-prefill of suggested weight/reps in the active session.
5. **A "PR detection" check** — given a candidate set, return true iff `weight * reps` strictly exceeds any previous set of the same exercise for the same user. Used to flag PRs in the set table.
6. **Per-client week computations** for the Clients screen:
   - `WeekPlan` (int): number of planned workouts this week. For now, infer as 4 (or derive from `WorkoutTemplate` applications scheduled this week) — your call.
   - `WeekDone` (int): count of workouts in the current ISO week.
   - `Status` (`on` | `off`): `off` if no workouts in the last 5 days.
   - `Streak` (int): consecutive days/weeks with workouts (define which).
7. **Recent PRs query** — for the Home screen's "СВЕЖИЕ РЕКОРДЫ" section, return the user's last 3 PRs in the last 30 days with the prior best value (to compute the `+2.5 кг` delta).
8. **Tonnage delta query** — for analytics, return total tonnage for the last 90 days **and** the 90 days prior, to compute the `▲ +18%` delta.

---

## Tailwind config (suggested)

If you're keeping Tailwind, drop this inline config block in `base.html` right after the CDN script:

```html
<script>
  tailwind.config = {
    theme: {
      extend: {
        colors: {
          bg:       '#0a0a0a',
          surface:  '#141414',
          surface2: '#1b1b1b',
          hair:     '#262626',
          hair2:    '#1f1f1f',
          dim:      '#7a7a7a',
          dim2:     '#4a4a4a',
          accent:   '#D7FF1A',
          accentDk: '#a8c800',
          danger:   '#ff453a',
          warn:     '#ff9f0a',
        },
        fontFamily: {
          display: ['"Anton"', 'Impact', 'sans-serif'],
          sans:    ['"Space Grotesk"', 'system-ui', 'sans-serif'],
          mono:    ['"JetBrains Mono"', 'ui-monospace', 'monospace'],
        },
      },
    },
  };
</script>
```

Then in `base.html` `<head>`, swap the Tailwind Preflight body class to `bg-bg text-white font-sans`.

---

## Assets

There are **no external image assets** in this design. Everything is rendered with type, color, SVG icons, and SVG charts. The 5-point wellbeing scale (💀 😣 😐 🙂 🔥) uses native emoji — keep your existing `wellbeingEmoji` template helper.

---

## Files in this bundle

| File                                | Purpose                                                            |
|-------------------------------------|--------------------------------------------------------------------|
| `Workout Tracker - Redesign.html`   | Open this in a browser to explore the interactive prototype.       |
| `app.jsx`                           | React components for screens 1–7 (main app + bottom tab bar).      |
| `admin.jsx`                         | React components for screens 8–13 (trainer & admin).               |
| `ios-frame.jsx`                     | iPhone bezel used by the prototype showcase — do NOT reproduce.    |
| `README.md`                         | This file.                                                          |

The JSX files are the most precise reference for sizes, paddings, and component logic. Open them side-by-side with the corresponding Go template when implementing.

---

## Out of scope

These were noticed during the audit but **not** redesigned. Mention them to the user before implementing:

- The gyms form (`web/templates/gyms/form.html`) and gyms list (`web/templates/gyms/list.html`) — should follow the same dark/lime card style. Apply by analogy from the User and Template screens.
- The admin user form (`web/templates/admin/user_form.html`) — same.
- The template form (`web/templates/templates/form.html`) and show (`web/templates/templates/show.html`) — same. The show page can borrow the structure from Workout Detail.
- Push notifications / rest-timer notifications — not in scope. The current rest timer is purely visual in the prototype.
