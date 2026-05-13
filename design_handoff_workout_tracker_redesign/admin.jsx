// Admin & Trainer screens for Workout Tracker — Redesign
// Same dark/lime brutalist system as app.jsx

const { useState: useState_a, useMemo: useMemo_a } = React;

// Local copy of tokens to keep file independent of app.jsx symbol scope
const AT = {
  bg: '#0a0a0a', surface: '#141414', surface2: '#1b1b1b',
  hair: '#262626', hair2: '#1f1f1f',
  text: '#fafafa', dim: '#7a7a7a', dim2: '#4a4a4a',
  accent: '#D7FF1A', danger: '#ff453a', warn: '#ff9f0a', good: '#30d158',
  display: '"Anton", Impact, sans-serif',
  body: '"Space Grotesk", system-ui, sans-serif',
  mono: '"JetBrains Mono", monospace',
};

const AEyebrow = ({ children, color, style }) => (
  <div style={{
    fontSize: 10, letterSpacing: 1.8, textTransform: 'uppercase',
    color: color || AT.dim, fontWeight: 600, ...style,
  }}>{children}</div>
);

const ABack = ({ onClick, label }) => (
  <button onClick={onClick} style={{
    background: 'none', border: 'none', color: AT.text, padding: 0,
    cursor: 'pointer', display: 'flex', alignItems: 'center', gap: 8,
  }}>
    <svg width="20" height="20" viewBox="0 0 24 24" fill="none">
      <path d="M15 6l-6 6 6 6" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
    </svg>
    <span style={{ fontSize: 12, letterSpacing: 1.4, textTransform: 'uppercase' }}>{label}</span>
  </button>
);

// ─────────────────────────────────────────────────────────────
// Sample data
// ─────────────────────────────────────────────────────────────
const CLIENTS = [
  { id: 'c1', name: 'Алексей Петров',   login: 'apetrov',  workouts: 38, last: '12.05', streak: 6, status: 'on',  goal: 'Набор массы',  weekDone: 3, weekPlan: 4 },
  { id: 'c2', name: 'Мария Соколова',   login: 'msokolova', workouts: 24, last: '11.05', streak: 4, status: 'on',  goal: 'Снижение веса', weekDone: 2, weekPlan: 3 },
  { id: 'c3', name: 'Дмитрий Иванов',   login: 'divanov',  workouts: 52, last: '09.05', streak: 0, status: 'off', goal: 'Сила',         weekDone: 0, weekPlan: 4 },
  { id: 'c4', name: 'Ольга Новикова',   login: 'onovikova', workouts: 19, last: '12.05', streak: 8, status: 'on',  goal: 'Тонус',         weekDone: 3, weekPlan: 3 },
  { id: 'c5', name: 'Сергей Морозов',   login: 'smorozov', workouts: 11, last: '06.05', streak: 0, status: 'off', goal: 'Реабилитация',  weekDone: 1, weekPlan: 3 },
];

const TEMPLATES = [
  { id: 't1', title: 'UPPER / PUSH',      exercises: 5, sets: 14, type: 'Сила',     used: 18 },
  { id: 't2', title: 'UPPER / PULL',      exercises: 5, sets: 13, type: 'Сила',     used: 16 },
  { id: 't3', title: 'LOWER / QUAD',      exercises: 4, sets: 12, type: 'Сила',     used: 14 },
  { id: 't4', title: 'LOWER / POSTERIOR', exercises: 4, sets: 11, type: 'Сила',     used: 9  },
  { id: 't5', title: 'METCON · 20 МИН',   exercises: 6, sets: 6,  type: 'Кардио',   used: 4  },
  { id: 't6', title: 'CORE · 15 МИН',     exercises: 5, sets: 10, type: 'Аксессуар', used: 7 },
];

const USERS = [
  { id: 'u0', name: 'Admin',            login: 'admin',     role: 'тренер', admin: true,  active: true },
  { id: 'u1', name: 'Алексей Петров',   login: 'apetrov',   role: 'клиент', admin: false, active: true },
  { id: 'u2', name: 'Мария Соколова',   login: 'msokolova', role: 'клиент', admin: false, active: true },
  { id: 'u3', name: 'Дмитрий Иванов',   login: 'divanov',   role: 'клиент', admin: false, active: false },
  { id: 'u4', name: 'Ольга Новикова',   login: 'onovikova', role: 'клиент', admin: false, active: true },
  { id: 'u5', name: 'Сергей Морозов',   login: 'smorozov',  role: 'клиент', admin: false, active: false },
  { id: 'u6', name: 'Игорь Лебедев',    login: 'ilebedev',  role: 'тренер', admin: false, active: true },
];

// ─────────────────────────────────────────────────────────────
// Clients (trainer view)
// ─────────────────────────────────────────────────────────────
function Clients({ go }) {
  const active = CLIENTS.filter(c => c.status === 'on').length;
  return (
    <div style={{ background: AT.bg, color: AT.text, minHeight: '100%', paddingBottom: 96 }}>
      <div style={{ padding: '56px 24px 0' }}>
        <ABack onClick={() => go('profile')} label="Профиль"/>
      </div>

      <div style={{ padding: '16px 24px 0' }}>
        <AEyebrow>{CLIENTS.length} ВСЕГО · {active} АКТИВНЫХ</AEyebrow>
        <div style={{ fontFamily: AT.display, fontSize: 48, lineHeight: 1, letterSpacing: 0.5, marginTop: 6 }}>
          КЛИЕНТЫ
        </div>
        <div style={{ width: 56, height: 4, background: AT.accent, marginTop: 16 }} />
      </div>

      {/* week overview */}
      <div style={{ padding: '24px 24px 0' }}>
        <div style={{ background: AT.surface, border: `1px solid ${AT.hair}`, padding: '18px 20px' }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'baseline' }}>
            <AEyebrow>НЕДЕЛЯ · ВЫПОЛНЕНО</AEyebrow>
            <div style={{ fontFamily: AT.mono, fontSize: 10, color: AT.accent }}>9/17 трен.</div>
          </div>
          <div style={{ display: 'flex', alignItems: 'baseline', gap: 8, marginTop: 8 }}>
            <div style={{ fontFamily: AT.display, fontSize: 44, lineHeight: 0.9, letterSpacing: 0.5 }}>
              53<span style={{ fontSize: 18, color: AT.dim, marginLeft: 6 }}>%</span>
            </div>
            <div style={{ fontFamily: AT.mono, fontSize: 11, color: AT.danger }}>▼ -12%</div>
          </div>
          <div style={{ marginTop: 14, display: 'flex', gap: 2 }}>
            {CLIENTS.map(c => (
              <div key={c.id} title={c.name} style={{
                flex: c.weekPlan, height: 6,
                background: c.weekDone >= c.weekPlan ? AT.accent
                  : c.weekDone === 0 ? AT.dim2 : AT.warn,
              }} />
            ))}
          </div>
        </div>
      </div>

      {/* clients list */}
      <div style={{ padding: '24px 24px 0', display: 'flex', flexDirection: 'column', gap: 10 }}>
        {CLIENTS.map(c => (
          <button key={c.id} onClick={() => go('clientDetail', c.id)} style={{
            width: '100%', textAlign: 'left', background: AT.surface,
            border: `1px solid ${c.status === 'off' ? AT.danger + '40' : AT.hair}`,
            padding: '16px 18px', cursor: 'pointer', color: AT.text,
            display: 'flex', alignItems: 'center', gap: 14,
          }}>
            {/* avatar */}
            <div style={{
              width: 44, height: 44, background: AT.surface2,
              border: `1px solid ${AT.hair}`,
              display: 'flex', alignItems: 'center', justifyContent: 'center',
              fontFamily: AT.display, fontSize: 18, color: AT.dim, letterSpacing: 1,
            }}>
              {c.name.split(' ').map(p => p[0]).join('').slice(0,2)}
            </div>

            <div style={{ flex: 1, minWidth: 0 }}>
              <div style={{ display: 'flex', alignItems: 'baseline', gap: 8 }}>
                <div style={{ fontSize: 15, fontWeight: 500 }}>{c.name}</div>
                {c.status === 'off' && (
                  <div style={{ fontSize: 9, letterSpacing: 1.4, color: AT.danger, fontWeight: 700 }}>● ПРОПУСК</div>
                )}
              </div>
              <div style={{ fontSize: 11, color: AT.dim, marginTop: 2 }}>
                {c.goal} · @{c.login}
              </div>
              <div style={{ marginTop: 8, display: 'flex', gap: 14, fontFamily: AT.mono, fontSize: 11 }}>
                <span style={{ color: AT.dim }}>{c.workouts} <span style={{ color: AT.dim2 }}>трен.</span></span>
                <span style={{ color: AT.dim }}>послед. <span style={{ color: AT.text }}>{c.last}</span></span>
                <span style={{ color: c.streak > 0 ? AT.accent : AT.dim2 }}>
                  🔥{c.streak}
                </span>
              </div>
            </div>

            {/* week progress dots */}
            <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'flex-end', gap: 6 }}>
              <div style={{ display: 'flex', gap: 3 }}>
                {Array.from({ length: c.weekPlan }).map((_, i) => (
                  <div key={i} style={{
                    width: 6, height: 14,
                    background: i < c.weekDone ? AT.accent : AT.hair,
                  }}/>
                ))}
              </div>
              <div style={{ fontFamily: AT.mono, fontSize: 9, color: AT.dim }}>
                {c.weekDone}/{c.weekPlan} нед.
              </div>
            </div>
          </button>
        ))}
      </div>

      <div style={{ padding: '20px 24px 0' }}>
        <button onClick={() => go('templates')} style={{
          width: '100%', background: 'transparent', color: AT.text,
          border: `1px solid ${AT.hair}`, padding: '16px',
          fontSize: 12, letterSpacing: 1.6, textTransform: 'uppercase',
          fontWeight: 700, cursor: 'pointer',
        }}>Применить шаблон ко всем</button>
      </div>
    </div>
  );
}

// ─────────────────────────────────────────────────────────────
// Client Detail
// ─────────────────────────────────────────────────────────────
function ClientDetail({ id, go }) {
  const c = CLIENTS.find(x => x.id === id) || CLIENTS[0];
  const recent = [
    { date: '12.05', title: 'UPPER / PUSH',     ton: 4.8, well: '🔥' },
    { date: '10.05', title: 'LOWER / QUAD',     ton: 6.2, well: '🙂' },
    { date: '08.05', title: 'PULL',             ton: 5.1, well: '🙂' },
    { date: '05.05', title: 'UPPER / PUSH',     ton: 4.4, well: '😐' },
  ];

  return (
    <div style={{ background: AT.bg, color: AT.text, minHeight: '100%', paddingBottom: 40 }}>
      <div style={{ padding: '56px 24px 0', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <ABack onClick={() => go('clients')} label="Клиенты" />
        <div style={{ fontSize: 10, letterSpacing: 1.4, color: AT.dim, textTransform: 'uppercase' }}>Редактировать</div>
      </div>

      <div style={{ padding: '20px 24px 0' }}>
        <AEyebrow>{c.goal} · @{c.login}</AEyebrow>
        <div style={{ fontFamily: AT.display, fontSize: 44, lineHeight: 0.95, letterSpacing: 0.5, marginTop: 6 }}>
          {c.name.toUpperCase()}
        </div>
        <div style={{ width: 56, height: 4, background: AT.accent, marginTop: 16 }} />
      </div>

      <div style={{ padding: '24px 24px 0' }}>
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(4, 1fr)', gap: 10 }}>
          <SmallStat value={c.workouts} unit="" label="ТРЕН." />
          <SmallStat value={`${c.weekDone}/${c.weekPlan}`} unit="" label="НЕДЕЛЯ" color={c.weekDone>=c.weekPlan ? AT.accent : AT.text}/>
          <SmallStat value={c.streak} unit="дн" label="СТРИК" color={c.streak>0?AT.accent:AT.dim2}/>
          <SmallStat value="4.6" unit="т" label="СРЕДН." />
        </div>
      </div>

      {/* compliance */}
      <div style={{ padding: '24px 24px 0' }}>
        <div style={{ background: AT.surface, border: `1px solid ${AT.hair}`, padding: '18px 20px' }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'baseline' }}>
            <AEyebrow>СОБЛЮДЕНИЕ ПЛАНА · 4 НЕД.</AEyebrow>
            <div style={{ fontFamily: AT.mono, fontSize: 10, color: AT.accent }}>87%</div>
          </div>
          <div style={{ display: 'flex', gap: 4, marginTop: 14 }}>
            {[true,true,true,false, true,true,true,true, true,false,true,true, true,true,true,true].map((d,i) => (
              <div key={i} style={{ flex: 1, height: 28, background: d ? AT.accent : AT.hair }} />
            ))}
          </div>
          <div style={{ display: 'flex', justifyContent: 'space-between', marginTop: 6, fontFamily: AT.mono, fontSize: 9, color: AT.dim2 }}>
            <span>нед -3</span><span>нед -2</span><span>нед -1</span><span>сейчас</span>
          </div>
        </div>
      </div>

      {/* recent workouts */}
      <div style={{ padding: '28px 24px 0' }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'baseline', marginBottom: 12 }}>
          <AEyebrow>ПОСЛЕДНИЕ ТРЕНИРОВКИ</AEyebrow>
          <button style={{ background: 'none', border: 'none', color: AT.dim, fontSize: 11, letterSpacing: 1.4, cursor: 'pointer', textTransform: 'uppercase' }}>Все →</button>
        </div>
        <div style={{ background: AT.surface, border: `1px solid ${AT.hair}` }}>
          {recent.map((r, i) => (
            <div key={i} style={{
              padding: '14px 18px',
              borderBottom: i < recent.length - 1 ? `1px solid ${AT.hair2}` : 'none',
              display: 'flex', alignItems: 'center', justifyContent: 'space-between',
            }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: 14 }}>
                <div style={{ fontFamily: AT.mono, fontSize: 11, color: AT.dim }}>{r.date}</div>
                <div>
                  <div style={{ fontFamily: AT.display, fontSize: 18, lineHeight: 1, letterSpacing: 0.5 }}>{r.title}</div>
                  <div style={{ fontSize: 11, color: AT.dim, marginTop: 4 }}>тоннаж · <span style={{ fontFamily: AT.mono, color: AT.accent }}>{r.ton}т</span></div>
                </div>
              </div>
              <div style={{ fontSize: 18 }}>{r.well}</div>
            </div>
          ))}
        </div>
      </div>

      {/* actions */}
      <div style={{ padding: '24px 24px 0', display: 'flex', flexDirection: 'column', gap: 10 }}>
        <button onClick={() => go('templateApply', 't1')} style={{
          width: '100%', background: AT.accent, color: '#000', border: 'none',
          padding: '18px', fontFamily: AT.display, fontSize: 18, letterSpacing: 2, cursor: 'pointer',
        }}>+ НАЗНАЧИТЬ ТРЕНИРОВКУ</button>
        <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 10 }}>
          <button style={{
            background: 'transparent', color: AT.text,
            border: `1px solid ${AT.hair}`, padding: '14px',
            fontSize: 11, letterSpacing: 1.6, textTransform: 'uppercase', cursor: 'pointer', fontWeight: 700,
          }}>Написать</button>
          <button style={{
            background: 'transparent', color: AT.text,
            border: `1px solid ${AT.hair}`, padding: '14px',
            fontSize: 11, letterSpacing: 1.6, textTransform: 'uppercase', cursor: 'pointer', fontWeight: 700,
          }}>План недели</button>
        </div>
      </div>
    </div>
  );
}

function SmallStat({ value, unit, label, color }) {
  return (
    <div>
      <div style={{ fontFamily: AT.display, fontSize: 30, lineHeight: 0.9, color: color || AT.text, letterSpacing: 0.5 }}>
        {value}{unit && <span style={{ fontSize: 12, color: AT.dim, marginLeft: 3 }}>{unit}</span>}
      </div>
      <div style={{ fontSize: 9, color: AT.dim, letterSpacing: 1.4, marginTop: 5 }}>{label}</div>
    </div>
  );
}

// ─────────────────────────────────────────────────────────────
// Templates list
// ─────────────────────────────────────────────────────────────
function Templates({ go }) {
  const [filter, setFilter] = useState_a('Все');
  const tabs = ['Все', 'Сила', 'Кардио', 'Аксессуар'];
  const list = filter === 'Все' ? TEMPLATES : TEMPLATES.filter(t => t.type === filter);
  return (
    <div style={{ background: AT.bg, color: AT.text, minHeight: '100%', paddingBottom: 96 }}>
      <div style={{ padding: '56px 24px 0' }}>
        <ABack onClick={() => go('profile')} label="Профиль"/>
      </div>

      <div style={{ padding: '16px 24px 0', display: 'flex', justifyContent: 'space-between', alignItems: 'flex-end' }}>
        <div>
          <AEyebrow>{TEMPLATES.length} ШАБЛОНОВ</AEyebrow>
          <div style={{ fontFamily: AT.display, fontSize: 48, lineHeight: 1, letterSpacing: 0.5, marginTop: 6 }}>
            ШАБЛОНЫ
          </div>
        </div>
        <button style={{
          width: 44, height: 44, background: AT.accent, color: '#000', border: 'none',
          fontFamily: AT.display, fontSize: 28, cursor: 'pointer', lineHeight: 1,
        }}>+</button>
      </div>

      <div style={{ padding: '20px 24px 0', display: 'flex', gap: 6 }}>
        {tabs.map(t => (
          <button key={t} onClick={() => setFilter(t)} style={{
            background: filter === t ? AT.text : 'transparent',
            color: filter === t ? '#000' : AT.dim,
            border: `1px solid ${filter === t ? AT.text : AT.hair}`,
            padding: '6px 12px', fontSize: 10, letterSpacing: 1.4, fontWeight: 700,
            textTransform: 'uppercase', cursor: 'pointer',
          }}>{t}</button>
        ))}
      </div>

      <div style={{ padding: '20px 24px 0', display: 'flex', flexDirection: 'column', gap: 10 }}>
        {list.map(t => (
          <div key={t.id} style={{
            background: AT.surface, border: `1px solid ${AT.hair}`,
            padding: '18px 20px',
          }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'baseline' }}>
              <AEyebrow color={AT.dim2}>{t.type.toUpperCase()}</AEyebrow>
              <div style={{ fontFamily: AT.mono, fontSize: 10, color: AT.dim }}>применён ×{t.used}</div>
            </div>
            <div style={{ fontFamily: AT.display, fontSize: 28, lineHeight: 1, letterSpacing: 0.5, marginTop: 8 }}>
              {t.title}
            </div>
            <div style={{ display: 'flex', gap: 18, marginTop: 12 }}>
              <Mini2 value={t.exercises} label="УПР." />
              <Mini2 value={t.sets} label="ПОДХ." />
              <Mini2 value="~45м" label="ВРЕМЯ" />
            </div>
            <div style={{ display: 'flex', gap: 8, marginTop: 16 }}>
              <button onClick={() => go('templateApply', t.id)} style={{
                flex: 1, background: AT.accent, color: '#000', border: 'none',
                padding: '12px', fontSize: 11, letterSpacing: 1.6, fontWeight: 700,
                textTransform: 'uppercase', cursor: 'pointer',
              }}>Применить</button>
              <button style={{
                background: 'transparent', color: AT.text,
                border: `1px solid ${AT.hair}`, padding: '12px 16px',
                fontSize: 11, letterSpacing: 1.6, textTransform: 'uppercase',
                fontWeight: 700, cursor: 'pointer',
              }}>Открыть</button>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

function Mini2({ value, label }) {
  return (
    <div>
      <div style={{ fontFamily: AT.display, fontSize: 22, lineHeight: 1, letterSpacing: 0.5 }}>{value}</div>
      <div style={{ fontSize: 9, color: AT.dim, letterSpacing: 1.4, marginTop: 3 }}>{label}</div>
    </div>
  );
}

// ─────────────────────────────────────────────────────────────
// Template Apply
// ─────────────────────────────────────────────────────────────
function TemplateApply({ id, go }) {
  const t = TEMPLATES.find(x => x.id === id) || TEMPLATES[0];
  const [picked, setPicked] = useState_a(new Set(['c1', 'c2']));
  const eligibleClients = CLIENTS.filter(c => c.status === 'on' || c.id === 'c5');

  const toggle = (id) => setPicked(p => {
    const n = new Set(p);
    if (n.has(id)) n.delete(id); else n.add(id);
    return n;
  });

  return (
    <div style={{ background: AT.bg, color: AT.text, minHeight: '100%', paddingBottom: 120 }}>
      <div style={{ padding: '56px 24px 0' }}>
        <ABack onClick={() => go('templates')} label={t.title}/>
      </div>

      <div style={{ padding: '16px 24px 0' }}>
        <AEyebrow color={AT.accent}>ПРИМЕНИТЬ ШАБЛОН</AEyebrow>
        <div style={{ fontFamily: AT.display, fontSize: 36, lineHeight: 0.95, letterSpacing: 0.5, marginTop: 6 }}>
          {t.title}
        </div>
        <div style={{ fontSize: 12, color: AT.dim, marginTop: 6 }}>
          {t.exercises} упражнений · {t.sets} подходов · ~45 мин
        </div>
      </div>

      {/* date + gym */}
      <div style={{ padding: '24px 24px 0' }}>
        <div style={{ background: AT.surface, border: `1px solid ${AT.hair}`, padding: '18px 20px' }}>
          <div>
            <AEyebrow color={AT.dim2}>ДАТА</AEyebrow>
            <div style={{ display: 'flex', gap: 6, marginTop: 10 }}>
              {[
                { d: '13', m: 'СР', sel: false },
                { d: '14', m: 'ЧТ', sel: true },
                { d: '15', m: 'ПТ', sel: false },
                { d: '16', m: 'СБ', sel: false },
                { d: '17', m: 'ВС', sel: false },
                { d: '18', m: 'ПН', sel: false },
                { d: '19', m: 'ВТ', sel: false },
              ].map((day, i) => (
                <div key={i} style={{
                  flex: 1, padding: '10px 4px', textAlign: 'center',
                  background: day.sel ? AT.accent : 'transparent',
                  color: day.sel ? '#000' : AT.text,
                  border: `1px solid ${day.sel ? AT.accent : AT.hair}`,
                  cursor: 'pointer',
                }}>
                  <div style={{ fontFamily: AT.display, fontSize: 22, lineHeight: 1, letterSpacing: 0.5 }}>{day.d}</div>
                  <div style={{ fontSize: 9, letterSpacing: 1.4, marginTop: 4, opacity: 0.7 }}>{day.m}</div>
                </div>
              ))}
            </div>
          </div>

          <div style={{ marginTop: 18 }}>
            <AEyebrow color={AT.dim2}>ЗАЛ</AEyebrow>
            <div style={{ display: 'flex', gap: 8, marginTop: 10 }}>
              {['World Class · Кутузовский', 'DDX · Авиапарк'].map((g, i) => (
                <button key={i} style={{
                  flex: 1, background: i === 0 ? AT.surface2 : 'transparent',
                  color: AT.text,
                  border: `1px solid ${i === 0 ? AT.accent : AT.hair}`,
                  padding: '12px', fontSize: 11, cursor: 'pointer',
                  fontWeight: 500, letterSpacing: 0.4, textAlign: 'left',
                }}>
                  {i === 0 && <span style={{ color: AT.accent, marginRight: 6 }}>●</span>}
                  {g}
                </button>
              ))}
            </div>
          </div>
        </div>
      </div>

      {/* clients select */}
      <div style={{ padding: '24px 24px 0' }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'baseline', marginBottom: 12 }}>
          <AEyebrow>КЛИЕНТЫ · {picked.size}/{eligibleClients.length}</AEyebrow>
          <button onClick={() => {
            const all = new Set(eligibleClients.map(c => c.id));
            setPicked(picked.size === all.size ? new Set() : all);
          }} style={{
            background: 'none', border: 'none', color: AT.accent,
            fontSize: 10, letterSpacing: 1.4, fontWeight: 700,
            textTransform: 'uppercase', cursor: 'pointer',
          }}>{picked.size === eligibleClients.length ? 'Снять' : 'Выбрать всех'}</button>
        </div>

        <div style={{ background: AT.surface, border: `1px solid ${AT.hair}` }}>
          {eligibleClients.map((c, i) => {
            const on = picked.has(c.id);
            return (
              <label key={c.id} style={{
                display: 'flex', alignItems: 'center', gap: 14,
                padding: '14px 18px',
                borderBottom: i < eligibleClients.length - 1 ? `1px solid ${AT.hair2}` : 'none',
                cursor: 'pointer',
              }}>
                <div onClick={() => toggle(c.id)} style={{
                  width: 22, height: 22,
                  background: on ? AT.accent : 'transparent',
                  border: `1.5px solid ${on ? AT.accent : AT.dim2}`,
                  display: 'flex', alignItems: 'center', justifyContent: 'center',
                }}>
                  {on && <svg width="14" height="14" viewBox="0 0 24 24"><path d="M5 12.5l4.5 4.5L19 7" stroke="#000" strokeWidth="3" fill="none" strokeLinecap="round" strokeLinejoin="round"/></svg>}
                </div>
                <div style={{ flex: 1 }} onClick={() => toggle(c.id)}>
                  <div style={{ fontSize: 14, fontWeight: 500 }}>{c.name}</div>
                  <div style={{ fontSize: 10, color: AT.dim, marginTop: 2, letterSpacing: 0.4 }}>{c.goal}</div>
                </div>
                <div style={{ fontFamily: AT.mono, fontSize: 10, color: AT.dim }}>
                  {c.weekDone}/{c.weekPlan} нед.
                </div>
              </label>
            );
          })}
        </div>
      </div>

      {/* sticky bottom CTA */}
      <div style={{
        position: 'absolute', left: 0, right: 0, bottom: 0,
        padding: '14px 20px 28px', background: AT.bg,
        borderTop: `1px solid ${AT.hair}`, zIndex: 8,
      }}>
        <button style={{
          width: '100%', background: AT.accent, color: '#000', border: 'none',
          padding: '18px', fontFamily: AT.display, fontSize: 20, letterSpacing: 2, cursor: 'pointer',
          opacity: picked.size === 0 ? 0.3 : 1,
        }}>
          СОЗДАТЬ {picked.size} {picked.size === 1 ? 'ТРЕНИРОВКУ' : picked.size < 5 ? 'ТРЕНИРОВКИ' : 'ТРЕНИРОВОК'} →
        </button>
      </div>
    </div>
  );
}

// ─────────────────────────────────────────────────────────────
// Admin users
// ─────────────────────────────────────────────────────────────
function Users({ go }) {
  const [filter, setFilter] = useState_a('Все');
  const tabs = ['Все', 'Тренеры', 'Клиенты', 'Неактивные'];
  const filtered = USERS.filter(u => {
    if (filter === 'Тренеры') return u.role === 'тренер';
    if (filter === 'Клиенты') return u.role === 'клиент';
    if (filter === 'Неактивные') return !u.active;
    return true;
  });

  return (
    <div style={{ background: AT.bg, color: AT.text, minHeight: '100%', paddingBottom: 96 }}>
      <div style={{ padding: '56px 24px 0' }}>
        <ABack onClick={() => go('profile')} label="Профиль"/>
      </div>

      <div style={{ padding: '16px 24px 0', display: 'flex', justifyContent: 'space-between', alignItems: 'flex-end' }}>
        <div>
          <AEyebrow>{USERS.length} АККАУНТОВ</AEyebrow>
          <div style={{ fontFamily: AT.display, fontSize: 48, lineHeight: 1, letterSpacing: 0.5, marginTop: 6 }}>
            ПОЛЬЗОВАТЕЛИ
          </div>
        </div>
        <button style={{
          width: 44, height: 44, background: AT.accent, color: '#000', border: 'none',
          fontFamily: AT.display, fontSize: 28, cursor: 'pointer', lineHeight: 1,
        }}>+</button>
      </div>

      {/* counters */}
      <div style={{ padding: '20px 24px 0' }}>
        <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr 1fr', gap: 8 }}>
          <Counter label="ТРЕНЕРЫ" value={USERS.filter(u=>u.role==='тренер').length}/>
          <Counter label="КЛИЕНТЫ" value={USERS.filter(u=>u.role==='клиент').length} accent/>
          <Counter label="НЕАКТ." value={USERS.filter(u=>!u.active).length} color={AT.danger}/>
        </div>
      </div>

      <div style={{ padding: '20px 24px 0', display: 'flex', gap: 6, overflow: 'auto' }}>
        {tabs.map(t => (
          <button key={t} onClick={() => setFilter(t)} style={{
            background: filter === t ? AT.text : 'transparent',
            color: filter === t ? '#000' : AT.dim,
            border: `1px solid ${filter === t ? AT.text : AT.hair}`,
            padding: '6px 12px', fontSize: 10, letterSpacing: 1.4, fontWeight: 700,
            textTransform: 'uppercase', cursor: 'pointer', whiteSpace: 'nowrap',
          }}>{t}</button>
        ))}
      </div>

      <div style={{ padding: '16px 24px 0' }}>
        <div style={{ background: AT.surface, border: `1px solid ${AT.hair}` }}>
          {filtered.map((u, i) => (
            <div key={u.id} style={{
              padding: '14px 18px',
              borderBottom: i < filtered.length - 1 ? `1px solid ${AT.hair2}` : 'none',
              display: 'flex', alignItems: 'center', gap: 12,
              opacity: u.active ? 1 : 0.5,
            }}>
              <div style={{
                width: 36, height: 36, background: AT.surface2,
                border: `1px solid ${AT.hair}`,
                display: 'flex', alignItems: 'center', justifyContent: 'center',
                fontFamily: AT.display, fontSize: 14, color: AT.dim, letterSpacing: 1,
              }}>
                {u.name.split(' ').map(p => p[0]).join('').slice(0,2)}
              </div>
              <div style={{ flex: 1, minWidth: 0 }}>
                <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                  <span style={{ fontSize: 14, fontWeight: 500 }}>{u.name}</span>
                  {u.admin && (
                    <span style={{
                      fontSize: 8, padding: '2px 6px',
                      background: AT.accent, color: '#000',
                      letterSpacing: 1.5, fontWeight: 700,
                    }}>ADMIN</span>
                  )}
                </div>
                <div style={{ fontFamily: AT.mono, fontSize: 10, color: AT.dim, marginTop: 3 }}>
                  @{u.login} · <span style={{ color: u.role === 'тренер' ? AT.accent : AT.dim }}>{u.role}</span>
                  {!u.active && <span style={{ color: AT.danger, marginLeft: 6 }}>● деактивирован</span>}
                </div>
              </div>
              <button style={{ background: 'none', border: 'none', color: AT.dim, cursor: 'pointer', padding: 4 }}>
                <svg width="18" height="18" viewBox="0 0 24 24" fill="none"><circle cx="5" cy="12" r="2" fill="currentColor"/><circle cx="12" cy="12" r="2" fill="currentColor"/><circle cx="19" cy="12" r="2" fill="currentColor"/></svg>
              </button>
            </div>
          ))}
        </div>
      </div>

      <div style={{ padding: '20px 24px 0' }}>
        <button onClick={() => go('assign')} style={{
          width: '100%', background: 'transparent', color: AT.text,
          border: `1px solid ${AT.hair}`, padding: '16px',
          fontSize: 12, letterSpacing: 1.6, textTransform: 'uppercase',
          fontWeight: 700, cursor: 'pointer',
        }}>Назначения тренер ↔ клиент →</button>
      </div>
    </div>
  );
}

function Counter({ label, value, accent, color }) {
  return (
    <div style={{ background: AT.surface, border: `1px solid ${AT.hair}`, padding: '14px 12px' }}>
      <div style={{ fontFamily: AT.display, fontSize: 28, lineHeight: 1, letterSpacing: 0.5, color: color || (accent ? AT.accent : AT.text) }}>
        {value}
      </div>
      <div style={{ fontSize: 9, color: AT.dim, letterSpacing: 1.4, marginTop: 4 }}>{label}</div>
    </div>
  );
}

// ─────────────────────────────────────────────────────────────
// Assign trainer ↔ client
// ─────────────────────────────────────────────────────────────
function Assign({ go }) {
  const trainers = USERS.filter(u => u.role === 'тренер');
  const assigns = [
    { trainer: 'Admin',         clients: ['Алексей Петров', 'Мария Соколова', 'Ольга Новикова'] },
    { trainer: 'Игорь Лебедев', clients: ['Дмитрий Иванов', 'Сергей Морозов'] },
  ];

  return (
    <div style={{ background: AT.bg, color: AT.text, minHeight: '100%', paddingBottom: 96 }}>
      <div style={{ padding: '56px 24px 0' }}>
        <ABack onClick={() => go('users')} label="Пользователи"/>
      </div>

      <div style={{ padding: '16px 24px 0' }}>
        <AEyebrow>{trainers.length} ТРЕНЕРОВ · {assigns.reduce((s,a) => s + a.clients.length, 0)} КЛИЕНТОВ</AEyebrow>
        <div style={{ fontFamily: AT.display, fontSize: 44, lineHeight: 1, letterSpacing: 0.5, marginTop: 6 }}>
          НАЗНАЧЕНИЯ
        </div>
        <div style={{ width: 56, height: 4, background: AT.accent, marginTop: 16 }} />
      </div>

      {/* quick assign */}
      <div style={{ padding: '24px 24px 0' }}>
        <div style={{ background: AT.surface, border: `1px solid ${AT.hair}`, padding: '18px 20px' }}>
          <AEyebrow color={AT.accent}>+ НОВАЯ СВЯЗКА</AEyebrow>
          <div style={{ marginTop: 14 }}>
            <div style={{ fontSize: 10, letterSpacing: 1.6, color: AT.dim, marginBottom: 6 }}>ТРЕНЕР</div>
            <div style={{
              background: AT.surface2, border: `1px solid ${AT.hair}`,
              padding: '12px 14px', fontFamily: AT.mono, fontSize: 14,
              display: 'flex', justifyContent: 'space-between', alignItems: 'center',
            }}>
              <span>Игорь Лебедев</span>
              <span style={{ color: AT.dim }}>▾</span>
            </div>
          </div>
          <div style={{ marginTop: 12 }}>
            <div style={{ fontSize: 10, letterSpacing: 1.6, color: AT.dim, marginBottom: 6 }}>КЛИЕНТ</div>
            <div style={{
              background: AT.surface2, border: `1px solid ${AT.hair}`,
              padding: '12px 14px', fontFamily: AT.mono, fontSize: 14,
              display: 'flex', justifyContent: 'space-between', alignItems: 'center',
              color: AT.dim,
            }}>
              <span>— выберите клиента —</span>
              <span>▾</span>
            </div>
          </div>
          <button style={{
            marginTop: 14, width: '100%', background: AT.accent, color: '#000', border: 'none',
            padding: '14px', fontSize: 12, letterSpacing: 1.8, cursor: 'pointer',
            fontWeight: 700, textTransform: 'uppercase',
          }}>Назначить →</button>
        </div>
      </div>

      {/* current assignments */}
      <div style={{ padding: '28px 24px 0' }}>
        <AEyebrow style={{ marginBottom: 12 }}>ТЕКУЩИЕ СВЯЗКИ</AEyebrow>
        <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
          {assigns.map(a => (
            <div key={a.trainer} style={{ background: AT.surface, border: `1px solid ${AT.hair}` }}>
              <div style={{ padding: '14px 18px', borderBottom: `1px solid ${AT.hair2}`, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                <div>
                  <AEyebrow color={AT.dim2}>ТРЕНЕР</AEyebrow>
                  <div style={{ fontFamily: AT.display, fontSize: 22, lineHeight: 1, letterSpacing: 0.5, marginTop: 4 }}>
                    {a.trainer.toUpperCase()}
                  </div>
                </div>
                <div style={{ textAlign: 'right' }}>
                  <div style={{ fontFamily: AT.display, fontSize: 28, lineHeight: 1, color: AT.accent, letterSpacing: 0.5 }}>{a.clients.length}</div>
                  <div style={{ fontSize: 9, color: AT.dim, letterSpacing: 1.4, marginTop: 4 }}>КЛИЕНТОВ</div>
                </div>
              </div>
              {a.clients.map((c, i) => (
                <div key={i} style={{
                  padding: '12px 18px',
                  borderBottom: i < a.clients.length - 1 ? `1px solid ${AT.hair2}` : 'none',
                  display: 'flex', alignItems: 'center', justifyContent: 'space-between',
                }}>
                  <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
                    <div style={{
                      width: 6, height: 6, background: AT.accent,
                    }} />
                    <span style={{ fontSize: 14 }}>{c}</span>
                  </div>
                  <button style={{ background: 'none', border: 'none', color: AT.danger, fontSize: 10, letterSpacing: 1.4, fontWeight: 700, cursor: 'pointer', textTransform: 'uppercase' }}>
                    Убрать
                  </button>
                </div>
              ))}
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}

Object.assign(window, { Clients, ClientDetail, Templates, TemplateApply, Users, Assign });
