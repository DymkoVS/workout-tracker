// Workout Tracker — Redesign prototype
// Mobile-first, dark, brutalist-sport, neon-lime accent.

const { useState, useMemo, useEffect } = React;

// ─────────────────────────────────────────────────────────────
// Design tokens (also surfaced in the right-hand "Design System" panel)
// ─────────────────────────────────────────────────────────────
const T = {
  bg:      '#0a0a0a',
  surface: '#141414',
  surface2:'#1b1b1b',
  hair:    '#262626',
  hair2:   '#1f1f1f',
  text:    '#fafafa',
  dim:     '#7a7a7a',
  dim2:    '#4a4a4a',
  accent:  '#D7FF1A',      // electric lime
  accentDk:'#a8c800',
  danger:  '#ff453a',
  good:    '#30d158',
  warn:    '#ff9f0a',
  // shorthand fonts
  display: '"Anton", "Bebas Neue", Impact, sans-serif',
  body:    '"Space Grotesk", system-ui, sans-serif',
  mono:    '"JetBrains Mono", ui-monospace, monospace',
};

// ─────────────────────────────────────────────────────────────
// Fake data — modeled directly on internal/model/workout.go
// ─────────────────────────────────────────────────────────────
const GYMS = [
  { id: 'g1', name: "World Class · Кутузовский" },
  { id: 'g2', name: "DDX Fitness · Авиапарк" },
];

const WELLBEING = ['', '💀', '😣', '😐', '🙂', '🔥'];
const WELLBEING_LABEL = ['', 'Разбит', 'Плохо', 'Нормально', 'Хорошо', 'Огонь'];

// each workout: id, title, date (DD.MM), gym, wellbeing, exercises[{name, sets[{w,r,rpe}]}]
const WORKOUTS = [
  {
    id: 'w1', title: 'UPPER / PUSH',  date: '12.05.2026', gym: GYMS[0].name, wellbeing: 5,
    exercises: [
      { name: 'Жим лёжа',                    sets: [{w:80,r:8,rpe:7}, {w:85,r:6,rpe:8}, {w:90,r:5,rpe:9}, {w:90,r:4,rpe:9.5}] },
      { name: 'Жим стоя',                    sets: [{w:50,r:8,rpe:7}, {w:55,r:6,rpe:8}, {w:55,r:6,rpe:8.5}] },
      { name: 'Жим гантелей на наклонной',   sets: [{w:30,r:10,rpe:7}, {w:32,r:8,rpe:8}, {w:32,r:8,rpe:8.5}] },
      { name: 'Французский жим',             sets: [{w:25,r:12,rpe:7}, {w:25,r:10,rpe:8}, {w:25,r:8,rpe:9}] },
    ],
    notes: 'Жим — лучший результат за месяц.',
  },
  {
    id: 'w2', title: 'LOWER / QUAD', date: '10.05.2026', gym: GYMS[0].name, wellbeing: 4,
    exercises: [
      { name: 'Приседания со штангой',       sets: [{w:100,r:8,rpe:7}, {w:110,r:6,rpe:8}, {w:120,r:4,rpe:9}] },
      { name: 'Жим ногами',                  sets: [{w:200,r:10,rpe:7}, {w:220,r:8,rpe:8}, {w:240,r:6,rpe:9}] },
      { name: 'Выпады с гантелями',          sets: [{w:20,r:12,rpe:7}, {w:22,r:10,rpe:8}] },
    ],
  },
  {
    id: 'w3', title: 'PULL',         date: '08.05.2026', gym: GYMS[0].name, wellbeing: 4,
    exercises: [
      { name: 'Становая тяга',               sets: [{w:120,r:5,rpe:7}, {w:140,r:3,rpe:8.5}, {w:150,r:2,rpe:9.5}] },
      { name: 'Тяга вертикального блока',    sets: [{w:60,r:10,rpe:7}, {w:65,r:8,rpe:8}, {w:70,r:6,rpe:9}] },
      { name: 'Тяга гантели в наклоне',      sets: [{w:30,r:10,rpe:7}, {w:32,r:8,rpe:8}] },
      { name: 'Сгибания на бицепс',          sets: [{w:14,r:12,rpe:7}, {w:14,r:10,rpe:8}] },
    ],
  },
  {
    id: 'w4', title: 'UPPER / PUSH', date: '05.05.2026', gym: GYMS[0].name, wellbeing: 3,
    exercises: [
      { name: 'Жим лёжа',                    sets: [{w:80,r:8,rpe:7.5}, {w:85,r:5,rpe:8.5}, {w:85,r:5,rpe:9}] },
      { name: 'Жим стоя',                    sets: [{w:50,r:8,rpe:7.5}, {w:52,r:6,rpe:8.5}] },
      { name: 'Жим гантелей на наклонной',   sets: [{w:28,r:10,rpe:7}, {w:30,r:8,rpe:8.5}, {w:30,r:7,rpe:9}] },
    ],
  },
  {
    id: 'w5', title: 'LOWER / POSTERIOR', date: '03.05.2026', gym: GYMS[1].name, wellbeing: 4,
    exercises: [
      { name: 'Румынская тяга',              sets: [{w:90,r:8,rpe:7}, {w:100,r:6,rpe:8}, {w:110,r:5,rpe:9}] },
      { name: 'Сгибания ног',                sets: [{w:50,r:12,rpe:7}, {w:55,r:10,rpe:8.5}] },
      { name: 'Подъёмы на носки',            sets: [{w:80,r:15,rpe:7}, {w:80,r:15,rpe:8.5}] },
    ],
  },
  {
    id: 'w6', title: 'PULL', date: '01.05.2026', gym: GYMS[0].name, wellbeing: 5,
    exercises: [
      { name: 'Становая тяга',               sets: [{w:120,r:5,rpe:7}, {w:135,r:3,rpe:8}, {w:145,r:2,rpe:9}] },
      { name: 'Тяга вертикального блока',    sets: [{w:60,r:10,rpe:7}, {w:65,r:8,rpe:8}] },
    ],
  },
];

// active "today" session being logged
const TODAY = {
  title: 'UPPER / PUSH',
  date: '13.05.2026',
  gym: GYMS[0].name,
  exercises: [
    {
      name: 'Жим лёжа',
      prev: '90×5 @ RPE 9',
      sets: [
        { w: 80, r: 8, rpe: 7,   done: true },
        { w: 85, r: 6, rpe: 8,   done: true },
        { w: 90, r: 6, rpe: 8.5, done: true, pr: true },
        { w: 90, r: 0, rpe: 0,   done: false, active: true },
      ],
    },
    {
      name: 'Жим стоя',
      prev: '55×6 @ RPE 8.5',
      sets: [
        { w: 50, r: 0, rpe: 0, done: false },
        { w: 55, r: 0, rpe: 0, done: false },
        { w: 55, r: 0, rpe: 0, done: false },
      ],
    },
    {
      name: 'Жим гантелей на наклонной',
      prev: '32×8 @ RPE 8.5',
      sets: [
        { w: 32, r: 0, rpe: 0, done: false },
        { w: 32, r: 0, rpe: 0, done: false },
        { w: 32, r: 0, rpe: 0, done: false },
      ],
    },
  ],
};

// ─────────────────────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────────────────────
function tonnage(workout) {
  let t = 0;
  for (const e of workout.exercises) for (const s of (e.sets || [])) if (s.w && s.r) t += s.w * s.r;
  return t;
}
function totalSets(workout) {
  return workout.exercises.reduce((n, e) => n + (e.sets ? e.sets.length : 0), 0);
}

// ─────────────────────────────────────────────────────────────
// Tiny SVG icons (stroke, 24×24, hairline)
// ─────────────────────────────────────────────────────────────
const Icon = {
  home:   (p) => <svg viewBox="0 0 24 24" {...p}><path d="M3 11l9-8 9 8v10a1 1 0 01-1 1h-5v-7H10v7H5a1 1 0 01-1-1z" fill="none" stroke="currentColor" strokeWidth="1.8"/></svg>,
  list:   (p) => <svg viewBox="0 0 24 24" {...p}><path d="M3 6h18M3 12h18M3 18h18" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round"/></svg>,
  chart:  (p) => <svg viewBox="0 0 24 24" {...p}><path d="M4 20V8M10 20V4M16 20v-8M22 20H2" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round"/></svg>,
  user:   (p) => <svg viewBox="0 0 24 24" {...p}><circle cx="12" cy="8" r="4" fill="none" stroke="currentColor" strokeWidth="1.8"/><path d="M4 21c0-4.4 3.6-8 8-8s8 3.6 8 8" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round"/></svg>,
  plus:   (p) => <svg viewBox="0 0 24 24" {...p}><path d="M12 5v14M5 12h14" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round"/></svg>,
  back:   (p) => <svg viewBox="0 0 24 24" {...p}><path d="M15 6l-6 6 6 6" stroke="currentColor" strokeWidth="2" fill="none" strokeLinecap="round" strokeLinejoin="round"/></svg>,
  fire:   (p) => <svg viewBox="0 0 24 24" {...p}><path d="M12 3s4 4 4 9a4 4 0 11-8 0c0-2 1-3 1-3s-2-1-2-4c2 1 3-1 5-2zM12 21a5 5 0 005-5c0-3-2-4-2-4s-1 2-3 2-3-2-3-2-2 1-2 4a5 5 0 005 5z" fill="currentColor"/></svg>,
  check:  (p) => <svg viewBox="0 0 24 24" {...p}><path d="M5 12.5l4.5 4.5L19 7" stroke="currentColor" strokeWidth="2.5" fill="none" strokeLinecap="round" strokeLinejoin="round"/></svg>,
  bolt:   (p) => <svg viewBox="0 0 24 24" {...p}><path d="M13 2L4 14h7l-1 8 9-12h-7l1-8z" fill="currentColor"/></svg>,
  dot:    (p) => <svg viewBox="0 0 24 24" {...p}><circle cx="12" cy="12" r="3" fill="currentColor"/></svg>,
  timer:  (p) => <svg viewBox="0 0 24 24" {...p}><circle cx="12" cy="13" r="8" fill="none" stroke="currentColor" strokeWidth="1.8"/><path d="M12 9v4l2.5 2.5M9 2h6" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round"/></svg>,
  edit:   (p) => <svg viewBox="0 0 24 24" {...p}><path d="M4 20h4l10-10-4-4L4 16v4z" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinejoin="round"/></svg>,
  trash:  (p) => <svg viewBox="0 0 24 24" {...p}><path d="M5 7h14M10 11v6M14 11v6M6 7l1 13a2 2 0 002 2h6a2 2 0 002-2l1-13M9 7V4h6v3" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinejoin="round" strokeLinecap="round"/></svg>,
  arrow:  (p) => <svg viewBox="0 0 24 24" {...p}><path d="M5 12h14M13 6l6 6-6 6" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/></svg>,
  trend:  (p) => <svg viewBox="0 0 24 24" {...p}><path d="M3 17l6-6 4 4 8-9M14 6h7v7" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/></svg>,
};

// ─────────────────────────────────────────────────────────────
// Tiny atoms
// ─────────────────────────────────────────────────────────────
function Stat({ value, unit, label, color }) {
  return (
    <div>
      <div style={{ fontFamily: T.display, fontSize: 44, lineHeight: 0.9, color: color || T.text, letterSpacing: 0.5 }}>
        {value}<span style={{ fontSize: 18, color: T.dim, marginLeft: 4, letterSpacing: 0 }}>{unit}</span>
      </div>
      <div style={{ fontSize: 10, color: T.dim, letterSpacing: 1.5, textTransform: 'uppercase', marginTop: 6 }}>{label}</div>
    </div>
  );
}

function Eyebrow({ children, color = T.dim, style }) {
  return (
    <div style={{
      fontSize: 10, letterSpacing: 1.8, textTransform: 'uppercase',
      color, fontWeight: 600, ...style,
    }}>{children}</div>
  );
}

function HairRule({ color = T.hair }) {
  return <div style={{ height: 1, background: color }} />;
}

// ─────────────────────────────────────────────────────────────
// Screen 1 — Login
// ─────────────────────────────────────────────────────────────
function Login({ onEnter }) {
  return (
    <div style={{ height: '100%', background: T.bg, color: T.text, padding: '120px 28px 32px', display: 'flex', flexDirection: 'column' }}>
      <div style={{ flex: 1 }}>
        <Eyebrow color={T.accent}>v 1.0 · MOBILE</Eyebrow>
        <div style={{ fontFamily: T.display, fontSize: 96, lineHeight: 0.88, letterSpacing: 1, marginTop: 16 }}>
          WORK<br/>OUT.
        </div>
        <div style={{ width: 56, height: 6, background: T.accent, marginTop: 24 }} />
        <div style={{ marginTop: 24, color: T.dim, fontSize: 14, lineHeight: 1.45, maxWidth: 280 }}>
          Личный журнал тренировок.<br/>
          Подходы, RPE, тоннаж, прогресс.
        </div>
      </div>

      <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
        <Field label="ЛОГИН" value="admin" />
        <Field label="ПАРОЛЬ" value="••••••••" type="password" />
        <button onClick={onEnter} style={{
          marginTop: 8, height: 60, background: T.accent, color: '#000', border: 'none',
          fontFamily: T.display, fontSize: 22, letterSpacing: 2, cursor: 'pointer',
        }}>
          ВОЙТИ →
        </button>
        <div style={{ color: T.dim2, fontSize: 11, letterSpacing: 1.4, textTransform: 'uppercase', textAlign: 'center', marginTop: 6 }}>
          admin / admin123
        </div>
      </div>
    </div>
  );
}

function Field({ label, value, type='text' }) {
  return (
    <label style={{ display: 'block' }}>
      <div style={{ fontSize: 10, letterSpacing: 1.6, color: T.dim, marginBottom: 6 }}>{label}</div>
      <div style={{ position: 'relative' }}>
        <input
          readOnly type={type} defaultValue={value}
          style={{
            width: '100%', background: 'transparent', color: T.text, border: 'none',
            borderBottom: `1px solid ${T.hair}`, padding: '10px 0 12px', fontSize: 20,
            fontFamily: T.mono, outline: 'none', letterSpacing: 1,
          }}
        />
      </div>
    </label>
  );
}

// ─────────────────────────────────────────────────────────────
// Screen 2 — Home
// ─────────────────────────────────────────────────────────────
function Home({ go }) {
  const weekTon = useMemo(() => Math.round(WORKOUTS.slice(0,4).reduce((s,w)=>s+tonnage(w),0)/100)/10, []);
  const last = WORKOUTS[0];

  return (
    <div style={{ background: T.bg, color: T.text, minHeight: '100%', paddingBottom: 96 }}>
      {/* top bar */}
      <div style={{ padding: '64px 24px 0', display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
        <div>
          <Eyebrow>13 МАЯ · СРЕДА</Eyebrow>
          <div style={{ fontFamily: T.display, fontSize: 36, lineHeight: 1, letterSpacing: 0.5, marginTop: 6 }}>
            ПОРА В ЗАЛ.
          </div>
        </div>
        <div style={{ width: 44, height: 44, borderRadius: 22, background: T.surface, border: `1px solid ${T.hair}`, display: 'flex', alignItems: 'center', justifyContent: 'center', color: T.dim }}>
          <Icon.user width="20" height="20"/>
        </div>
      </div>

      {/* week stats */}
      <div style={{ padding: '24px 24px 0' }}>
        <Eyebrow style={{ marginBottom: 14 }}>ЭТА НЕДЕЛЯ</Eyebrow>
        <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr 1fr', gap: 18 }}>
          <Stat value="4"      unit=""    label="ТРЕНИРОВКИ" />
          <Stat value={weekTon} unit="т"  label="ТОННАЖ" color={T.accent} />
          <Stat value="6"      unit="дн"  label="СТРИК" />
        </div>
        <HairRule />
      </div>

      {/* primary CTA */}
      <div style={{ padding: '20px 24px 0' }}>
        <button onClick={() => go('log')} style={{
          width: '100%', background: T.accent, color: '#000', border: 'none',
          padding: '22px 24px', display: 'flex', alignItems: 'center', justifyContent: 'space-between',
          cursor: 'pointer', textAlign: 'left',
        }}>
          <div>
            <div style={{ fontSize: 10, letterSpacing: 2, fontWeight: 700, opacity: 0.7 }}>СЛЕДУЮЩАЯ</div>
            <div style={{ fontFamily: T.display, fontSize: 32, lineHeight: 1, letterSpacing: 0.5, marginTop: 6 }}>
              UPPER / PUSH
            </div>
            <div style={{ fontSize: 12, opacity: 0.8, marginTop: 8 }}>3 упражнения · ~ 50 мин</div>
          </div>
          <div style={{ width: 52, height: 52, border: '2px solid #000', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
            <Icon.arrow width="22" height="22"/>
          </div>
        </button>
      </div>

      {/* last workout */}
      <div style={{ padding: '24px 24px 0' }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'baseline', marginBottom: 14 }}>
          <Eyebrow>ПОСЛЕДНЯЯ</Eyebrow>
          <button onClick={() => go('history')} style={{ background: 'none', border: 'none', color: T.dim, fontSize: 11, letterSpacing: 1.4, cursor: 'pointer', textTransform: 'uppercase' }}>
            Все →
          </button>
        </div>
        <WorkoutCard w={last} onClick={() => go('detail', last.id)} />
      </div>

      {/* records */}
      <div style={{ padding: '28px 24px 0' }}>
        <Eyebrow style={{ marginBottom: 14 }}>СВЕЖИЕ РЕКОРДЫ</Eyebrow>
        <div style={{ background: T.surface, border: `1px solid ${T.hair}` }}>
          <PrRow exercise="Жим лёжа" value="90" unit="кг" reps="×5" delta="+2.5" />
          <HairRule color={T.hair2}/>
          <PrRow exercise="Становая тяга" value="150" unit="кг" reps="×2" delta="+5" />
          <HairRule color={T.hair2}/>
          <PrRow exercise="Жим ногами" value="240" unit="кг" reps="×6" delta="+20" />
        </div>
      </div>
    </div>
  );
}

function PrRow({ exercise, value, unit, reps, delta }) {
  return (
    <div style={{ padding: '18px 18px', display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
      <div style={{ display: 'flex', alignItems: 'center', gap: 14 }}>
        <Icon.bolt width="16" height="16" style={{ color: T.accent }} />
        <div>
          <div style={{ fontSize: 14, color: T.text }}>{exercise}</div>
          <div style={{ fontSize: 11, color: T.dim, marginTop: 2 }}>{reps} повторений</div>
        </div>
      </div>
      <div style={{ textAlign: 'right' }}>
        <div style={{ fontFamily: T.display, fontSize: 28, lineHeight: 1, letterSpacing: 0.5 }}>
          {value}<span style={{ fontSize: 12, color: T.dim, marginLeft: 4 }}>{unit}</span>
        </div>
        <div style={{ fontSize: 10, color: T.accent, marginTop: 4, fontFamily: T.mono }}>▲ {delta} кг</div>
      </div>
    </div>
  );
}

// Reusable workout card
function WorkoutCard({ w, onClick, big }) {
  const ton = Math.round(tonnage(w));
  return (
    <button onClick={onClick} style={{
      width: '100%', textAlign: 'left', background: T.surface,
      border: `1px solid ${T.hair}`, padding: '20px',
      display: 'flex', alignItems: 'stretch', gap: 18, cursor: 'pointer', color: T.text,
    }}>
      <div style={{ width: 4, background: T.accent, alignSelf: 'stretch' }} />
      <div style={{ flex: 1 }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'baseline' }}>
          <Eyebrow color={T.dim}>{w.date}</Eyebrow>
          <div style={{ fontSize: 18 }}>{WELLBEING[w.wellbeing]}</div>
        </div>
        <div style={{ fontFamily: T.display, fontSize: big ? 32 : 24, lineHeight: 1, letterSpacing: 0.5, marginTop: 8 }}>
          {w.title}
        </div>
        <div style={{ fontSize: 11, color: T.dim, marginTop: 8 }}>{w.gym}</div>
        <div style={{ display: 'flex', gap: 22, marginTop: 16 }}>
          <Mini value={w.exercises.length} label="УПР." />
          <Mini value={totalSets(w)} label="ПОДХ." />
          <Mini value={(ton/1000).toFixed(1)+'т'} label="ТОННАЖ" />
        </div>
      </div>
    </button>
  );
}

function Mini({ value, label }) {
  return (
    <div>
      <div style={{ fontFamily: T.display, fontSize: 22, lineHeight: 1, letterSpacing: 0.5 }}>{value}</div>
      <div style={{ fontSize: 9, color: T.dim, letterSpacing: 1.4, marginTop: 3 }}>{label}</div>
    </div>
  );
}

// ─────────────────────────────────────────────────────────────
// Screen 3 — History list
// ─────────────────────────────────────────────────────────────
function History({ go }) {
  const grouped = useMemo(() => {
    const g = {};
    for (const w of WORKOUTS) {
      const k = w.date.slice(3); // MM.YYYY
      (g[k] ??= []).push(w);
    }
    return Object.entries(g);
  }, []);
  const months = { '05.2026': 'МАЙ 2026', '04.2026': 'АПРЕЛЬ 2026' };

  return (
    <div style={{ background: T.bg, color: T.text, minHeight: '100%', paddingBottom: 96 }}>
      <div style={{ padding: '64px 24px 0', display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
        <div>
          <Eyebrow>{WORKOUTS.length} ТРЕНИРОВОК</Eyebrow>
          <div style={{ fontFamily: T.display, fontSize: 48, lineHeight: 1, letterSpacing: 0.5, marginTop: 6 }}>
            ЖУРНАЛ
          </div>
        </div>
        <div style={{ width: 44, height: 44, border: `1px solid ${T.hair}`, display: 'flex', alignItems: 'center', justifyContent: 'center', color: T.text }}>
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none"><path d="M3 6h18M6 12h12M9 18h6" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/></svg>
        </div>
      </div>

      {/* mini sparkline of recent tonnage */}
      <div style={{ padding: '24px 24px 0' }}>
        <div style={{ background: T.surface, border: `1px solid ${T.hair}`, padding: '18px 20px' }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'baseline' }}>
            <Eyebrow>ТОННАЖ · 6 ПОСЛЕДНИХ</Eyebrow>
            <div style={{ fontFamily: T.mono, fontSize: 10, color: T.accent }}>+14% к мес.</div>
          </div>
          <Spark data={WORKOUTS.slice().reverse().map(tonnage)} />
        </div>
      </div>

      {grouped.map(([k, list]) => (
        <div key={k} style={{ padding: '28px 24px 0' }}>
          <Eyebrow style={{ marginBottom: 12 }}>{months[k] || k}</Eyebrow>
          <div style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
            {list.map(w => <WorkoutCard key={w.id} w={w} onClick={() => go('detail', w.id)} />)}
          </div>
        </div>
      ))}
    </div>
  );
}

function Spark({ data }) {
  const max = Math.max(...data);
  const min = Math.min(...data);
  const w = 320, h = 60, pad = 4;
  const pts = data.map((v, i) => {
    const x = pad + (w - pad*2) * (i / (data.length - 1));
    const y = pad + (h - pad*2) * (1 - (v - min) / (max - min || 1));
    return [x, y];
  });
  const d = pts.map((p,i) => (i===0?'M':'L') + p[0].toFixed(1) + ' ' + p[1].toFixed(1)).join(' ');
  return (
    <svg viewBox={`0 0 ${w} ${h}`} style={{ width: '100%', height: 60, marginTop: 12, display: 'block' }} preserveAspectRatio="none">
      <path d={d} stroke={T.accent} strokeWidth="2" fill="none"/>
      {pts.map((p,i) => <circle key={i} cx={p[0]} cy={p[1]} r="2.5" fill={T.accent}/>)}
    </svg>
  );
}

// ─────────────────────────────────────────────────────────────
// Screen 4 — Workout Detail
// ─────────────────────────────────────────────────────────────
function Detail({ id, go }) {
  const w = WORKOUTS.find(x => x.id === id) || WORKOUTS[0];
  const ton = tonnage(w);
  const sets = totalSets(w);

  return (
    <div style={{ background: T.bg, color: T.text, minHeight: '100%', paddingBottom: 40 }}>
      <div style={{ padding: '56px 24px 0', display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
        <button onClick={() => go('history')} style={{ background: 'none', border: 'none', color: T.text, padding: 0, cursor: 'pointer', display: 'flex', alignItems: 'center', gap: 8 }}>
          <Icon.back width="20" height="20"/>
          <span style={{ fontSize: 12, letterSpacing: 1.4, textTransform: 'uppercase' }}>Журнал</span>
        </button>
        <div style={{ display: 'flex', gap: 12, color: T.dim }}>
          <Icon.edit width="20" height="20"/>
          <Icon.trash width="20" height="20"/>
        </div>
      </div>

      <div style={{ padding: '20px 24px 0' }}>
        <Eyebrow>{w.date} · {w.gym}</Eyebrow>
        <div style={{ fontFamily: T.display, fontSize: 56, lineHeight: 0.95, letterSpacing: 0.5, marginTop: 8 }}>
          {w.title}
        </div>
        <div style={{ width: 56, height: 4, background: T.accent, marginTop: 16 }} />
      </div>

      {/* summary stats */}
      <div style={{ padding: '24px 24px 0' }}>
        <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr 1fr 1fr', gap: 10 }}>
          <Stat value={w.exercises.length} unit="" label="УПР." />
          <Stat value={sets} unit="" label="ПОДХОДОВ" />
          <Stat value={(ton/1000).toFixed(1)} unit="т" label="ТОННАЖ" color={T.accent}/>
          <Stat value={WELLBEING[w.wellbeing]} unit="" label="СОСТ." />
        </div>
      </div>

      {/* exercises */}
      <div style={{ padding: '28px 24px 0', display: 'flex', flexDirection: 'column', gap: 14 }}>
        {w.exercises.map((e, i) => (
          <div key={i} style={{ background: T.surface, border: `1px solid ${T.hair}` }}>
            <div style={{ padding: '18px 20px 12px', display: 'flex', justifyContent: 'space-between', alignItems: 'baseline' }}>
              <div>
                <Eyebrow color={T.dim2}>УПР. {String(i+1).padStart(2,'0')}</Eyebrow>
                <div style={{ fontFamily: T.display, fontSize: 22, lineHeight: 1.05, letterSpacing: 0.5, marginTop: 4 }}>
                  {e.name.toUpperCase()}
                </div>
              </div>
              <div style={{ textAlign: 'right' }}>
                <div style={{ fontSize: 10, color: T.dim }}>ОБЪЁМ</div>
                <div style={{ fontFamily: T.mono, fontSize: 13, color: T.accent, marginTop: 2 }}>
                  {e.sets.reduce((s,x)=>s+x.w*x.r,0)} кг
                </div>
              </div>
            </div>

            <div style={{ padding: '0 8px 12px' }}>
              <div style={{ display: 'grid', gridTemplateColumns: '32px 1fr 1fr 1fr', padding: '6px 12px', fontSize: 9, color: T.dim2, letterSpacing: 1.4 }}>
                <span>#</span><span>ВЕС</span><span>ПОВТ.</span><span style={{ textAlign: 'right' }}>RPE</span>
              </div>
              {e.sets.map((s, j) => {
                const top = Math.max(...e.sets.map(x => x.w*x.r));
                const isTop = s.w*s.r === top;
                return (
                  <div key={j} style={{
                    display: 'grid', gridTemplateColumns: '32px 1fr 1fr 1fr',
                    alignItems: 'center', padding: '10px 12px',
                    background: isTop ? `${T.accent}10` : 'transparent',
                    borderLeft: isTop ? `2px solid ${T.accent}` : '2px solid transparent',
                  }}>
                    <span style={{ fontFamily: T.mono, color: T.dim, fontSize: 12 }}>{j+1}</span>
                    <span style={{ fontFamily: T.display, fontSize: 24, letterSpacing: 0.5 }}>
                      {s.w}<span style={{ fontSize: 11, color: T.dim, marginLeft: 3 }}>кг</span>
                    </span>
                    <span style={{ fontFamily: T.display, fontSize: 24, letterSpacing: 0.5 }}>
                      ×{s.r}
                    </span>
                    <span style={{ fontFamily: T.mono, color: isTop ? T.accent : T.dim, fontSize: 13, textAlign: 'right' }}>
                      {s.rpe}
                    </span>
                  </div>
                );
              })}
            </div>
          </div>
        ))}
      </div>

      {w.notes && (
        <div style={{ padding: '20px 24px 0' }}>
          <div style={{ background: T.surface2, padding: '16px 18px', borderLeft: `3px solid ${T.accent}` }}>
            <Eyebrow color={T.accent}>ЗАМЕТКА</Eyebrow>
            <div style={{ marginTop: 6, fontSize: 14, color: T.text, lineHeight: 1.45 }}>{w.notes}</div>
          </div>
        </div>
      )}
    </div>
  );
}

// ─────────────────────────────────────────────────────────────
// Screen 5 — Log Workout (active session — THE money screen)
// ─────────────────────────────────────────────────────────────
function Log({ go }) {
  const [data, setData] = useState(TODAY);
  const [elapsedS, setElapsedS] = useState(28 * 60 + 14);
  const [restS, setRestS] = useState(56);
  useEffect(() => {
    const t = setInterval(() => { setElapsedS(s => s + 1); setRestS(s => Math.max(0, s - 1)); }, 1000);
    return () => clearInterval(t);
  }, []);
  const fmt = (s) => `${String(Math.floor(s/60)).padStart(2,'0')}:${String(s%60).padStart(2,'0')}`;

  const totalSets = data.exercises.reduce((n,e)=>n+e.sets.length, 0);
  const doneSets  = data.exercises.reduce((n,e)=>n+e.sets.filter(s=>s.done).length, 0);
  const ton = data.exercises.reduce((s,e) => s + e.sets.filter(x=>x.done).reduce((a,x)=>a+x.w*x.r,0), 0);

  const completeSet = (ei, si) => {
    setData(d => {
      const next = JSON.parse(JSON.stringify(d));
      next.exercises[ei].sets[si].done = true;
      next.exercises[ei].sets[si].active = false;
      const nextSet = next.exercises[ei].sets[si+1];
      if (nextSet) nextSet.active = true;
      else {
        const ne = next.exercises[ei+1];
        if (ne && ne.sets[0]) ne.sets[0].active = true;
      }
      return next;
    });
    setRestS(90);
  };

  return (
    <div style={{ background: T.bg, color: T.text, minHeight: '100%', paddingBottom: 110 }}>
      {/* sticky head */}
      <div style={{
        position: 'sticky', top: 0, zIndex: 5,
        background: T.bg, padding: '56px 20px 14px',
        borderBottom: `1px solid ${T.hair}`,
      }}>
        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
          <button onClick={() => go('home')} style={{ background: 'none', border: 'none', color: T.dim, padding: 0, cursor: 'pointer' }}>
            <Icon.back width="22" height="22"/>
          </button>
          <Eyebrow color={T.accent}>● ИДЁТ ТРЕНИРОВКА</Eyebrow>
          <button style={{ background: T.danger, color: '#fff', border: 'none', padding: '6px 12px', fontSize: 10, letterSpacing: 1.6, fontWeight: 700, cursor: 'pointer' }}>
            ЗАВЕРШИТЬ
          </button>
        </div>

        <div style={{ marginTop: 14, display: 'grid', gridTemplateColumns: '1fr 1fr 1fr', gap: 6 }}>
          <div>
            <Eyebrow color={T.dim2}>ВРЕМЯ</Eyebrow>
            <div style={{ fontFamily: T.mono, fontSize: 22, color: T.text, letterSpacing: 1, marginTop: 4 }}>{fmt(elapsedS)}</div>
          </div>
          <div>
            <Eyebrow color={T.dim2}>ПОДХОДЫ</Eyebrow>
            <div style={{ fontFamily: T.mono, fontSize: 22, color: T.text, letterSpacing: 1, marginTop: 4 }}>{doneSets}<span style={{ color: T.dim }}>/{totalSets}</span></div>
          </div>
          <div>
            <Eyebrow color={T.dim2}>ТОННАЖ</Eyebrow>
            <div style={{ fontFamily: T.mono, fontSize: 22, color: T.accent, letterSpacing: 1, marginTop: 4 }}>{ton}<span style={{ color: T.dim2, fontSize: 13 }}>кг</span></div>
          </div>
        </div>

        {/* rest timer */}
        {restS > 0 && (
          <div style={{ marginTop: 14, background: T.accent, color: '#000', padding: '10px 14px', display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
              <Icon.timer width="18" height="18"/>
              <span style={{ fontSize: 11, letterSpacing: 1.6, fontWeight: 700 }}>ОТДЫХ</span>
            </div>
            <div style={{ fontFamily: T.mono, fontSize: 22, fontWeight: 600 }}>0:{String(restS).padStart(2,'0')}</div>
            <button style={{ background: '#000', color: T.accent, border: 'none', padding: '4px 10px', fontSize: 10, letterSpacing: 1.5, fontWeight: 700, cursor: 'pointer' }}>
              ПРОПУСТИТЬ
            </button>
          </div>
        )}
      </div>

      {/* exercises */}
      <div style={{ padding: '16px 20px 0', display: 'flex', flexDirection: 'column', gap: 14 }}>
        {data.exercises.map((e, ei) => {
          const expanded = e.sets.some(s => s.active) || e.sets.every(s => !s.done);
          const allDone = e.sets.every(s => s.done);
          return (
            <div key={ei} style={{ background: T.surface, border: `1px solid ${allDone ? T.dim2 : T.hair}`, opacity: allDone ? 0.55 : 1 }}>
              <div style={{ padding: '14px 18px', display: 'flex', justifyContent: 'space-between', alignItems: 'center', borderBottom: `1px solid ${T.hair2}` }}>
                <div>
                  <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                    {allDone && <Icon.check width="14" height="14" style={{ color: T.accent }}/>}
                    <Eyebrow color={T.dim2}>УПР. {String(ei+1).padStart(2,'0')}</Eyebrow>
                  </div>
                  <div style={{ fontFamily: T.display, fontSize: 22, lineHeight: 1, letterSpacing: 0.5, marginTop: 4 }}>
                    {e.name.toUpperCase()}
                  </div>
                  <div style={{ fontSize: 11, color: T.dim, marginTop: 6, fontFamily: T.mono }}>
                    Прошлый раз: {e.prev}
                  </div>
                </div>
                <div style={{ fontFamily: T.display, fontSize: 22, color: T.dim }}>
                  {e.sets.filter(s=>s.done).length}/{e.sets.length}
                </div>
              </div>

              {expanded && (
                <div style={{ padding: '6px 8px 12px' }}>
                  <div style={{ display: 'grid', gridTemplateColumns: '36px 1fr 1fr 1fr 48px', padding: '6px 12px', fontSize: 9, color: T.dim2, letterSpacing: 1.4 }}>
                    <span>#</span><span>ВЕС</span><span>ПОВТ.</span><span>RPE</span><span></span>
                  </div>
                  {e.sets.map((s, si) => <LogSetRow key={si} s={s} si={si} onDone={() => completeSet(ei, si)} />)}
                  <button style={{
                    width: 'calc(100% - 16px)', margin: '8px 8px 0',
                    background: 'transparent', border: `1px dashed ${T.hair}`,
                    color: T.dim, padding: '10px', fontSize: 11, letterSpacing: 1.5, cursor: 'pointer', textTransform: 'uppercase',
                  }}>+ Подход</button>
                </div>
              )}
            </div>
          );
        })}

        <button style={{
          background: 'transparent', border: `1px dashed ${T.hair}`,
          color: T.dim, padding: '18px', fontSize: 12, letterSpacing: 1.6, cursor: 'pointer', textTransform: 'uppercase',
        }}>+ Добавить упражнение</button>
      </div>
    </div>
  );
}

function LogSetRow({ s, si, onDone }) {
  if (s.done) {
    return (
      <div style={{
        display: 'grid', gridTemplateColumns: '36px 1fr 1fr 1fr 48px',
        alignItems: 'center', padding: '10px 12px',
        background: s.pr ? `${T.accent}12` : 'transparent',
        borderLeft: s.pr ? `2px solid ${T.accent}` : '2px solid transparent',
      }}>
        <span style={{ fontFamily: T.mono, color: T.dim, fontSize: 12 }}>{si+1}</span>
        <span style={{ fontFamily: T.display, fontSize: 22, letterSpacing: 0.5 }}>{s.w}<span style={{ fontSize: 11, color: T.dim, marginLeft: 3 }}>кг</span></span>
        <span style={{ fontFamily: T.display, fontSize: 22, letterSpacing: 0.5 }}>×{s.r}</span>
        <span style={{ fontFamily: T.mono, fontSize: 13, color: T.dim }}>{s.rpe}</span>
        <span style={{ textAlign: 'right' }}>
          {s.pr ? <span style={{ fontSize: 9, color: T.accent, letterSpacing: 1.4, fontWeight: 700 }}>PR</span>
                : <Icon.check width="16" height="16" style={{ color: T.dim, opacity: 0.7 }}/>}
        </span>
      </div>
    );
  }
  if (s.active) {
    return (
      <div style={{
        margin: '6px 4px', padding: '14px 14px',
        background: '#000', border: `1.5px solid ${T.accent}`, position: 'relative',
      }}>
        <div style={{ position: 'absolute', top: -8, left: 12, background: T.accent, color: '#000', padding: '1px 8px', fontSize: 9, fontWeight: 700, letterSpacing: 1.4 }}>
          ТЕКУЩИЙ
        </div>
        <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr 1fr', gap: 12, marginTop: 4 }}>
          <NumPad label="ВЕС"  value={s.w} unit="кг" />
          <NumPad label="ПОВТ" value={6}   unit="" />
          <NumPad label="RPE"  value={8}   unit="" />
        </div>
        <button onClick={onDone} style={{
          marginTop: 14, width: '100%', background: T.accent, color: '#000',
          border: 'none', padding: '14px', fontFamily: T.display, fontSize: 18, letterSpacing: 2, cursor: 'pointer',
        }}>
          ✓ ПОДХОД ВЫПОЛНЕН
        </button>
      </div>
    );
  }
  // queued
  return (
    <div style={{
      display: 'grid', gridTemplateColumns: '36px 1fr 1fr 1fr 48px',
      alignItems: 'center', padding: '10px 12px', opacity: 0.45,
    }}>
      <span style={{ fontFamily: T.mono, color: T.dim, fontSize: 12 }}>{si+1}</span>
      <span style={{ fontFamily: T.display, fontSize: 22, color: T.dim, letterSpacing: 0.5 }}>{s.w}<span style={{ fontSize: 11, marginLeft: 3 }}>кг</span></span>
      <span style={{ fontFamily: T.display, fontSize: 22, color: T.dim, letterSpacing: 0.5 }}>×{s.r || 6}</span>
      <span style={{ fontFamily: T.mono, fontSize: 13, color: T.dim2 }}>—</span>
      <span/>
    </div>
  );
}

function NumPad({ label, value, unit }) {
  return (
    <div style={{ textAlign: 'center', background: T.surface2, padding: '10px 4px' }}>
      <Eyebrow color={T.dim2}>{label}</Eyebrow>
      <div style={{ fontFamily: T.display, fontSize: 30, color: T.accent, letterSpacing: 0.5, marginTop: 6, lineHeight: 1 }}>
        {value}{unit && <span style={{ fontSize: 11, color: T.dim, marginLeft: 3 }}>{unit}</span>}
      </div>
    </div>
  );
}

// ─────────────────────────────────────────────────────────────
// Screen 6 — Analytics
// ─────────────────────────────────────────────────────────────
function Analytics() {
  const [exercise, setExercise] = useState('Жим лёжа');
  const benchData = [
    { d: '01.04', w: 75 },{ d: '08.04', w: 77.5 },{ d: '15.04', w: 80 },
    { d: '22.04', w: 80 },{ d: '29.04', w: 82.5 },{ d: '05.05', w: 85 },
    { d: '12.05', w: 90 },
  ];
  const freq = [3, 4, 3, 4, 4, 3, 5, 4]; // last 8 weeks

  return (
    <div style={{ background: T.bg, color: T.text, minHeight: '100%', paddingBottom: 96 }}>
      <div style={{ padding: '64px 24px 0' }}>
        <Eyebrow>ПОСЛЕДНИЕ 90 ДНЕЙ</Eyebrow>
        <div style={{ fontFamily: T.display, fontSize: 48, lineHeight: 1, letterSpacing: 0.5, marginTop: 6 }}>
          ПРОГРЕСС
        </div>
      </div>

      {/* Big tonnage card */}
      <div style={{ padding: '24px 24px 0' }}>
        <div style={{ background: T.surface, border: `1px solid ${T.hair}`, padding: '20px' }}>
          <Eyebrow>ТОННАЖ · 90 ДНЕЙ</Eyebrow>
          <div style={{ display: 'flex', alignItems: 'baseline', gap: 10, marginTop: 8 }}>
            <div style={{ fontFamily: T.display, fontSize: 64, lineHeight: 0.9, letterSpacing: 0.5 }}>
              184.2<span style={{ fontSize: 22, color: T.dim, marginLeft: 6 }}>т</span>
            </div>
            <div style={{ fontFamily: T.mono, fontSize: 13, color: T.accent }}>▲ +18%</div>
          </div>
          <BigChart data={benchData.map(x => ({...x, w: x.w * 250 / 90}))} unit="т" />
        </div>
      </div>

      {/* Exercise progress */}
      <div style={{ padding: '20px 24px 0' }}>
        <div style={{ background: T.surface, border: `1px solid ${T.hair}`, padding: '20px' }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'baseline' }}>
            <Eyebrow>МАКС. ВЕС</Eyebrow>
            <select value={exercise} onChange={e => setExercise(e.target.value)} style={{
              background: T.surface2, color: T.text, border: `1px solid ${T.hair}`,
              padding: '4px 8px', fontSize: 11, letterSpacing: 1, textTransform: 'uppercase',
              fontFamily: T.body, cursor: 'pointer',
            }}>
              <option>Жим лёжа</option>
              <option>Приседания</option>
              <option>Становая тяга</option>
            </select>
          </div>
          <div style={{ display: 'flex', alignItems: 'baseline', gap: 10, marginTop: 8 }}>
            <div style={{ fontFamily: T.display, fontSize: 64, lineHeight: 0.9, letterSpacing: 0.5, color: T.accent }}>
              90<span style={{ fontSize: 22, color: T.dim, marginLeft: 6 }}>кг</span>
            </div>
            <div style={{ fontFamily: T.mono, fontSize: 13, color: T.accent }}>▲ +15 кг</div>
          </div>
          <BigChart data={benchData} unit="кг" />
        </div>
      </div>

      {/* Frequency */}
      <div style={{ padding: '20px 24px 0' }}>
        <div style={{ background: T.surface, border: `1px solid ${T.hair}`, padding: '20px' }}>
          <Eyebrow>ЧАСТОТА · 8 НЕДЕЛЬ</Eyebrow>
          <div style={{ display: 'flex', alignItems: 'baseline', gap: 10, marginTop: 8 }}>
            <div style={{ fontFamily: T.display, fontSize: 44, lineHeight: 0.9, letterSpacing: 0.5 }}>
              3.8<span style={{ fontSize: 18, color: T.dim, marginLeft: 6 }}>/нед</span>
            </div>
          </div>
          <BarChart data={freq} />
        </div>
      </div>

      {/* split breakdown */}
      <div style={{ padding: '20px 24px 0' }}>
        <div style={{ background: T.surface, border: `1px solid ${T.hair}`, padding: '20px' }}>
          <Eyebrow>СПЛИТ ПО МЫШЕЧНЫМ ГРУППАМ</Eyebrow>
          <div style={{ marginTop: 14, display: 'flex', flexDirection: 'column', gap: 12 }}>
            <SplitBar label="Грудь / Трицепс" pct={32} color={T.accent} />
            <SplitBar label="Спина / Бицепс"  pct={28} color="#fff" />
            <SplitBar label="Ноги"            pct={26} color={T.dim} />
            <SplitBar label="Плечи / Кор"     pct={14} color={T.dim2} />
          </div>
        </div>
      </div>
    </div>
  );
}

function BigChart({ data, unit }) {
  const w = 320, h = 120, pad = 8;
  const max = Math.max(...data.map(d => d.w));
  const min = Math.min(...data.map(d => d.w));
  const pts = data.map((d, i) => {
    const x = pad + (w - pad*2) * (i / (data.length - 1));
    const y = pad + (h - pad*2) * (1 - (d.w - min) / (max - min || 1));
    return [x, y, d];
  });
  const path = pts.map((p,i) => (i===0?'M':'L') + p[0].toFixed(1) + ' ' + p[1].toFixed(1)).join(' ');
  const fill = `${path} L ${pts[pts.length-1][0]} ${h-pad} L ${pts[0][0]} ${h-pad} Z`;
  return (
    <svg viewBox={`0 0 ${w} ${h+24}`} style={{ width: '100%', height: 140, marginTop: 16, display: 'block' }} preserveAspectRatio="none">
      <defs>
        <linearGradient id="fadeAcc" x1="0" y1="0" x2="0" y2="1">
          <stop offset="0%"  stopColor={T.accent} stopOpacity="0.25"/>
          <stop offset="100%" stopColor={T.accent} stopOpacity="0"/>
        </linearGradient>
      </defs>
      <path d={fill} fill="url(#fadeAcc)" />
      <path d={path} stroke={T.accent} strokeWidth="2" fill="none"/>
      {pts.map((p,i) => (
        <g key={i}>
          <circle cx={p[0]} cy={p[1]} r="3" fill={T.bg} stroke={T.accent} strokeWidth="1.5"/>
        </g>
      ))}
      {/* x labels */}
      {pts.filter((_,i) => i % 2 === 0 || i === pts.length - 1).map((p,i) => (
        <text key={i} x={p[0]} y={h + 16} fill={T.dim2} fontSize="9" fontFamily="JetBrains Mono,monospace" textAnchor="middle" letterSpacing="1">
          {p[2].d}
        </text>
      ))}
    </svg>
  );
}

function BarChart({ data }) {
  const max = Math.max(...data);
  return (
    <div style={{ display: 'flex', alignItems: 'flex-end', gap: 6, height: 100, marginTop: 16 }}>
      {data.map((v, i) => {
        const h = (v / max) * 100;
        const isLast = i === data.length - 1;
        return (
          <div key={i} style={{ flex: 1, display: 'flex', flexDirection: 'column', alignItems: 'center', gap: 4 }}>
            <div style={{ fontFamily: T.mono, fontSize: 9, color: isLast ? T.accent : T.dim }}>{v}</div>
            <div style={{ width: '100%', background: isLast ? T.accent : T.dim2, height: `${h}%`, minHeight: 2 }} />
            <div style={{ fontFamily: T.mono, fontSize: 8, color: T.dim2 }}>W{i+1}</div>
          </div>
        );
      })}
    </div>
  );
}

function SplitBar({ label, pct, color }) {
  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'baseline', marginBottom: 5 }}>
        <div style={{ fontSize: 12, color: T.text }}>{label}</div>
        <div style={{ fontFamily: T.mono, fontSize: 11, color }}>{pct}%</div>
      </div>
      <div style={{ height: 6, background: T.surface2 }}>
        <div style={{ height: '100%', width: `${pct}%`, background: color }} />
      </div>
    </div>
  );
}

// ─────────────────────────────────────────────────────────────
// Profile (lightweight, for nav target)
// ─────────────────────────────────────────────────────────────
function Profile({ go }) {
  return (
    <div style={{ background: T.bg, color: T.text, minHeight: '100%', paddingBottom: 96 }}>
      <div style={{ padding: '64px 24px 0' }}>
        <Eyebrow>АДМИНИСТРАТОР · ТРЕНЕР</Eyebrow>
        <div style={{ fontFamily: T.display, fontSize: 48, lineHeight: 1, letterSpacing: 0.5, marginTop: 6 }}>
          ADMIN
        </div>
        <div style={{ width: 56, height: 4, background: T.accent, marginTop: 16 }} />
      </div>

      <div style={{ padding: '24px 24px 0' }}>
        <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 10 }}>
          <Stat value={WORKOUTS.length} unit="" label="ВСЕГО ТРЕН." />
          <Stat value="184" unit="т" label="ТОННАЖ" color={T.accent}/>
          <Stat value="42" unit="дн" label="МАКС. СТРИК" />
          <Stat value="2" unit="" label="ЗАЛА" />
        </div>
      </div>

      <div style={{ padding: '28px 24px 0' }}>
        <Eyebrow style={{ marginBottom: 12 }}>ТРЕНЕРСКОЕ</Eyebrow>
        <div style={{ background: T.surface, border: `1px solid ${T.hair}` }}>
          {[
            { l: 'Клиенты',            k: 'clients',  badge: '5' },
            { l: 'Шаблоны тренировок', k: 'templates', badge: '6' },
            { l: 'Залы',               k: null,       badge: '2' },
          ].map((x, i, a) => (
            <button key={x.l} onClick={() => x.k && go(x.k)} style={{
              width: '100%', background: 'transparent', border: 'none', cursor: x.k ? 'pointer' : 'default',
              padding: '16px 18px', display: 'flex', justifyContent: 'space-between', alignItems: 'center',
              borderBottom: i < a.length - 1 ? `1px solid ${T.hair2}` : 'none', color: T.text,
              opacity: x.k ? 1 : 0.5,
            }}>
              <span style={{ fontSize: 14 }}>{x.l}</span>
              <span style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
                <span style={{ fontFamily: T.mono, fontSize: 11, color: T.dim }}>{x.badge}</span>
                <span style={{ color: T.dim2 }}>→</span>
              </span>
            </button>
          ))}
        </div>
      </div>

      <div style={{ padding: '20px 24px 0' }}>
        <Eyebrow style={{ marginBottom: 12 }}>АДМИНИСТРАЦИЯ</Eyebrow>
        <div style={{ background: T.surface, border: `1px solid ${T.hair}` }}>
          {[
            { l: 'Пользователи',  k: 'users',  badge: '7' },
            { l: 'Назначения',    k: 'assign', badge: '5' },
          ].map((x, i, a) => (
            <button key={x.l} onClick={() => x.k && go(x.k)} style={{
              width: '100%', background: 'transparent', border: 'none', cursor: 'pointer',
              padding: '16px 18px', display: 'flex', justifyContent: 'space-between', alignItems: 'center',
              borderBottom: i < a.length - 1 ? `1px solid ${T.hair2}` : 'none', color: T.text,
            }}>
              <span style={{ fontSize: 14 }}>{x.l}</span>
              <span style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
                <span style={{ fontFamily: T.mono, fontSize: 11, color: T.dim }}>{x.badge}</span>
                <span style={{ color: T.dim2 }}>→</span>
              </span>
            </button>
          ))}
        </div>
      </div>

      <div style={{ padding: '20px 24px 0' }}>
        <button style={{
          width: '100%', background: 'transparent', color: T.danger,
          border: `1px solid ${T.hair}`, padding: '16px', fontSize: 12, letterSpacing: 1.6,
          textTransform: 'uppercase', cursor: 'pointer', fontWeight: 700,
        }}>Выйти</button>
      </div>
    </div>
  );
}

// ─────────────────────────────────────────────────────────────
// Bottom tab nav
// ─────────────────────────────────────────────────────────────
function TabBar({ tab, go }) {
  const items = [
    { k: 'home',      icon: Icon.home,  label: 'ГЛАВНАЯ' },
    { k: 'history',   icon: Icon.list,  label: 'ЖУРНАЛ' },
    { k: 'log',       icon: Icon.plus,  label: 'НОВАЯ', center: true },
    { k: 'analytics', icon: Icon.chart, label: 'ПРОГРЕСС' },
    { k: 'profile',   icon: Icon.user,  label: 'ПРОФИЛЬ' },
  ];
  return (
    <div style={{
      position: 'absolute', left: 0, right: 0, bottom: 0,
      paddingBottom: 28, background: T.bg, borderTop: `1px solid ${T.hair}`,
      display: 'grid', gridTemplateColumns: 'repeat(5, 1fr)',
      zIndex: 10,
    }}>
      {items.map(it => {
        const active = tab === it.k || (tab === 'detail' && it.k === 'history');
        if (it.center) {
          return (
            <button key={it.k} onClick={() => go('log')} style={{
              background: 'none', border: 'none', cursor: 'pointer',
              display: 'flex', flexDirection: 'column', alignItems: 'center',
              padding: '10px 0 6px',
            }}>
              <div style={{
                width: 50, height: 50, background: T.accent, color: '#000',
                display: 'flex', alignItems: 'center', justifyContent: 'center',
                marginTop: -22, boxShadow: `0 0 0 6px ${T.bg}`,
              }}>
                <it.icon width="26" height="26"/>
              </div>
              <div style={{ fontSize: 8, letterSpacing: 1.4, color: T.accent, marginTop: 6, fontWeight: 700 }}>{it.label}</div>
            </button>
          );
        }
        return (
          <button key={it.k} onClick={() => go(it.k)} style={{
            background: 'none', border: 'none', cursor: 'pointer',
            display: 'flex', flexDirection: 'column', alignItems: 'center',
            padding: '12px 0 6px', color: active ? T.text : T.dim2,
          }}>
            <it.icon width="22" height="22"/>
            <div style={{ fontSize: 8, letterSpacing: 1.4, marginTop: 6, fontWeight: 600 }}>{it.label}</div>
          </button>
        );
      })}
    </div>
  );
}

// ─────────────────────────────────────────────────────────────
// Phone shell — routes between screens
// ─────────────────────────────────────────────────────────────
function Phone({ initial = 'login' }) {
  const [tab, setTab] = useState(initial);
  const [detailId, setDetailId] = useState('w1');

  const go = (k, id) => { setTab(k); if (id) setDetailId(id); };

  const showTab = !['login', 'log', 'detail'].includes(tab);
  const showTabAlways = !['login'].includes(tab); // also show in log/detail? we'll hide in login only

  let screen;
  if (tab === 'login')     screen = <Login onEnter={() => setTab('home')} />;
  else if (tab === 'home') screen = <Home go={go} />;
  else if (tab === 'history') screen = <History go={go} />;
  else if (tab === 'detail')  screen = <Detail id={detailId} go={go} />;
  else if (tab === 'log')     screen = <Log go={go} />;
  else if (tab === 'analytics') screen = <Analytics />;
  else if (tab === 'profile')   screen = <Profile go={go} />;
  else if (tab === 'clients')      screen = <Clients go={go} />;
  else if (tab === 'clientDetail') screen = <ClientDetail id={detailId} go={go} />;
  else if (tab === 'templates')    screen = <Templates go={go} />;
  else if (tab === 'templateApply') screen = <TemplateApply id={detailId} go={go} />;
  else if (tab === 'users')        screen = <Users go={go} />;
  else if (tab === 'assign')       screen = <Assign go={go} />;

  return (
    <IOSDevice dark={true}>
      <div style={{ position: 'relative', height: '100%', background: T.bg, overflow: 'hidden', fontFamily: T.body }}>
        <div style={{ position: 'absolute', inset: 0, overflow: 'auto' }}>
          {screen}
        </div>
        {tab !== 'login' && !['clients','clientDetail','templates','templateApply','users','assign'].includes(tab) && <TabBar tab={tab} go={go} />}
      </div>
    </IOSDevice>
  );
}

window.WorkoutPhone = Phone;
