#!/usr/bin/env python3
"""
bot.py — Telegram бот для импорта тренировок в свободном формате.

Поток: текст → Claude API (парсинг в JSON) → превью → подтверждение → PostgreSQL.

На сервере: /opt/workout-bot/bot.py
Деплой:    ~/Developer/workout-tracker/bot/deploy.sh
"""

import os, re, html, json, asyncio, logging, urllib.request, urllib.error
from datetime import datetime, timezone
from pathlib import Path

from dotenv import load_dotenv

_DIR = Path(__file__).parent
(_DIR / "logs").mkdir(exist_ok=True)
load_dotenv(_DIR / ".env")

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s %(levelname)s %(message)s",
    handlers=[
        logging.FileHandler(_DIR / "logs" / "bot.log"),
        logging.StreamHandler(),
    ],
)
log = logging.getLogger(__name__)

# httpx/httpcore логируют полный URL запроса на INFO, а в нём — TELEGRAM_BOT_TOKEN
# (api.telegram.org/bot<TOKEN>/...). Глушим их до WARNING, чтобы токен не светился
# в логах. Собственные сообщения бота (INFO) остаются.
for _noisy in ("httpx", "httpcore"):
    logging.getLogger(_noisy).setLevel(logging.WARNING)

import anthropic
from telegram import Update, InlineKeyboardButton, InlineKeyboardMarkup
from telegram.ext import (
    Application, CallbackQueryHandler, CommandHandler,
    ContextTypes, MessageHandler, filters,
)

BOT_TOKEN        = os.environ["TELEGRAM_BOT_TOKEN"]
ALLOWED_USER_ID  = int(os.environ["ALLOWED_TELEGRAM_USER_ID"])
ANTHROPIC_KEY    = os.environ["ANTHROPIC_API_KEY"]
CLIENT_LOGIN     = os.getenv("CLIENT_LOGIN", "Dymko")
API_BASE_URL     = os.getenv("API_BASE_URL", "http://localhost:8080")
APP_URL          = os.getenv("APP_URL", "https://dymko.ru")
IMPORT_API_TOKEN = os.environ["IMPORT_API_TOKEN"]

anthropic_client = anthropic.Anthropic(api_key=ANTHROPIC_KEY)

# pending[telegram_user_id] = parsed workout dict ожидающий подтверждения
pending: dict[int, dict] = {}


# ── auth ──────────────────────────────────────────────────────────────────────

def _auth(update: Update) -> bool:
    return bool(update.effective_user and update.effective_user.id == ALLOWED_USER_ID)


# ── HTTP API (workout-tracker Go-приложение) ────────────────────────────────────
# Бот больше не пишет в БД напрямую: всё идёт через эндпоинты Go-приложения,
# которое сохраняет через тот же проверенный код, что и веб-интерфейс.

def _api(method: str, path: str, payload: dict | None = None) -> dict:
    data = json.dumps(payload).encode() if payload is not None else None
    req = urllib.request.Request(
        f"{API_BASE_URL}{path}", data=data, method=method,
        headers={
            "Authorization": f"Bearer {IMPORT_API_TOKEN}",
            "Content-Type": "application/json",
        },
    )
    try:
        with urllib.request.urlopen(req, timeout=20) as resp:
            return json.loads(resp.read().decode())
    except urllib.error.HTTPError as e:
        body = e.read().decode(errors="replace")
        raise RuntimeError(f"API {e.code}: {body}")


def _api_gyms() -> list[str]:
    """Список названий залов для подсказки Claude."""
    return _api("GET", "/api/gyms").get("gyms", [])


# ── Claude parsing ────────────────────────────────────────────────────────────

_PARSE_SYSTEM = """\
Ты парсер тренировок. Пользователь присылает описание тренировки в произвольном формате.
Верни ТОЛЬКО валидный JSON — никаких пояснений, никакого текста вокруг.

Схема:
{
  "title": "название тренировки (строка)",
  "date": "YYYY-MM-DD",
  "gym": "название зала или null",
  "notes": "общий комментарий к тренировке (цикл, цели, самочувствие) или \"\"",
  "exercises": [
    {
      "name": "название упражнения",
      "order": 1,
      "sets": [
        {"set_num": 1, "weight": 0.0, "reps": 10, "rpe": null, "rest_seconds": null, "notes": ""}
      ]
    }
  ]
}

Правила:
- Если дата не указана → используй {today}
- set_num начинается с 1 для каждого упражнения; weight — float (0.0 без веса), reps — целое.
- Суперсет → ДВА ОТДЕЛЬНЫХ упражнения. Часто числа в суперсете идут построчно:
  строка с названиями, ниже строка чисел для ПЕРВОГО упражнения, ещё ниже — для ВТОРОГО.
  Сопоставляй строки чисел по порядку упражнений.
- Заголовок группы сам по себе — НЕ упражнение, а описание. Упражнения бери из
  перечисления: после «:» через «/», либо строками с «-». Например «Суперсет с верхнего
  блока с канатом 4х15 / - пуловер …» → ОДНО упражнение «пуловер» (а не отдельный
  «Суперсет…»). Не создавай пустое упражнение из слова «Суперсет/Трисет/Суперсерия».

ЧИСЛА ПОДХОДОВ — главное, НЕ ВЫДУМЫВАЙ:
- НИКОГДА не подставляй число «из ниоткуда». Нет значения в тексте → вес 0.0, повторы
  НЕ выдумывай. Лучше оставить 0, чем поставить случайное (никаких "1" непонятно откуда).
- ФАКТ важнее ЦЕЛИ. Если у упражнения есть и целевая схема ("15-12-10-10", "4×12-15"),
  и отдельная строка реально сделанных чисел — бери ЧИСЛА (это факт). Целевую схему/
  диапазон используй ТОЛЬКО когда фактических чисел нет.
- Куда идёт одиночная строка чисел под упражнением — РЕШАЙ ПО ТИПУ УПРАЖНЕНИЯ:
  • С ОТЯГОЩЕНИЕМ (тренажёр, блок/кроссовер, сведения, штанга, гантели, гак, платформа…)
    → это ВЕСА по подходам (даже мелкие "2 3" или "3,4,5" — это веса/уровни стопки,
    а НЕ повторы). Повторы берём из целевой схемы/диапазона (большее число), если
    повторы не записаны отдельно явно.
  • СО СВОИМ ВЕСОМ (отжимания, подтягивания, скручивания/пресс, планка, приседы/выпады
    без отягощения) → weight 0.0, а строка чисел ("8-10-8-8") = ПОВТОРЫ по подходам.
- ВЕСА ЧАСТО В ОДНОЙ СТРОКЕ со схемой повторов: «<упражнение> <схема> <веса>».
  «Тяга к поясу 15-12-10-10 10,15,20» → повторы из схемы 15-12-10-10, а хвост-числа
  «10,15,20» = ВЕСА по подходам (это НЕ повторы — НЕ теряй их!). «4×12 36» → 4 подхода
  по 12, вес 36. Числа после схемы в конце строки почти всегда веса. То же правило
  работает и когда веса на отдельной строке ниже.
- Название упражнения бери КАК В ТЕКСТЕ (хват, «по 1»/одной рукой, угол) — НЕ меняй и
  НЕ выдумывай хват (узкий/широкий/параллельный): пиши ровно то, что написано.
- "б/в", "без веса", "с собственным весом" → weight 0.0.
- Диапазон повторов без факта ("10-12", "4×12-15") → большее число (12, 15).
- "вес×повт×подходы": "3×15×4" = вес 3, по 15 повторов, 4 одинаковых подхода.
- Запятая ВНУТРИ одного числа между цифрами ("4,5") — это десятичная дробь 4.5, НЕ два
  числа; в JSON выводи через точку (4.5). Запятые МЕЖДУ разными числами — список
  ("10,15,20" = три веса). По контексту: "4,5 ×15×3" → вес 4.5.
- "70×10/11", "20×9/10×2" → ОТДЕЛЬНЫЕ подходы (10 и 11; 9 и 10).
- Дроп-сет ("дроп") → два подхода в одном упражнении.
- Если весов дано меньше, чем подходов в схеме (3 веса при схеме на 4 подхода) —
  недостающие подходы повтори с последним весом.
- Не теряй «хвосты»-вариации ("2 у окна", другой тренажёр/угол) — добавь подходом
  или в notes подхода.
- Не уверен в числе → поставь его и добавь "?" в notes подхода (я проверю в превью).
- notes подхода — КОРОТКО и НЕ ДУБЛИРУЙ одно и то же на всех подходах: ставь пометку
  только на тот подход, к которому она относится (а не копируй на каждый).
- Пометки выполнения ("отказ", "со страховкой", "с помощью") → notes ПОДХОДА, кратко.
  "Оба отказ" → обоим. Нет пометок → "".
- Общий контекст (цикл, цели, самочувствие) → notes ТРЕНИРОВКИ. Эмоции/мат — игнорировать.
- gym: если зал упомянут — ТОЧНОЕ название из списка: {gyms}
  (Fit/Фит/ФитГраунд → FitGround, Топ/TopGym → TopGym). Иначе gym: null
"""


def _parse(text: str) -> dict:
    today = datetime.now(timezone.utc).strftime("%Y-%m-%d")
    gym_names = ", ".join(_api_gyms()) or "нет данных"
    system = _PARSE_SYSTEM.replace("{today}", today).replace("{gyms}", gym_names)
    # Модель изредка возвращает невалидный JSON (случайная кавычка и т.п.) —
    # повторяем разок, прежде чем сдаться. Парсинг детерминирован по правилам,
    # так что повтор почти всегда отдаёт корректный JSON.
    last_err: json.JSONDecodeError | None = None
    for attempt in range(2):
        msg = anthropic_client.messages.create(
            model="claude-sonnet-4-6",
            max_tokens=8192,  # длинные тренировки (много подходов) не влезали в 2048 → обрыв JSON
            system=system,
            messages=[{"role": "user", "content": text}],
        )
        raw = msg.content[0].text.strip()
        raw = re.sub(r"^```(?:json)?\s*", "", raw)
        raw = re.sub(r"\s*```$", "", raw)
        try:
            return json.loads(raw)
        except json.JSONDecodeError as e:
            last_err = e
            log.warning("parse: невалидный JSON (попытка %d): %s", attempt + 1, e)
    raise last_err


# ── preview ───────────────────────────────────────────────────────────────────

_WEEKDAYS_RU = ["Пн", "Вт", "Ср", "Чт", "Пт", "Сб", "Вс"]


def _fmt_date(iso: str) -> str:
    """'2026-06-08' → '08.06.2026 (Вс)' — день недели помогает поймать ошибку даты."""
    try:
        d = datetime.strptime(iso, "%Y-%m-%d")
        return f"{d.strftime('%d.%m.%Y')} ({_WEEKDAYS_RU[d.weekday()]})"
    except (ValueError, TypeError):
        return str(iso)


def _fmt_set(s: dict) -> str:
    """Один подход: '75×4 @9 (отказ)' / 'б/в×12'."""
    base = f"{s['weight']:g}×{s['reps']}" if s.get("weight") else f"б/в×{s['reps']}"
    if s.get("rpe") is not None:
        base += f" @{s['rpe']:g}"
    if s.get("notes"):
        base += f" ({s['notes']})"
    return base


def _preview(w: dict) -> str:
    """HTML-превью (экранируем всё пользовательское — Markdown ломался на * и _)."""
    lines = [
        f"📋 <b>{html.escape(w['title'])}</b>",
        f"📅 {_fmt_date(w['date'])}",
        f"🏋 {html.escape(w.get('gym') or 'зал не указан')}",
    ]
    if w.get("notes"):
        lines.append(f"📝 {html.escape(w['notes'])}")
    lines.append("")
    total_sets = 0
    for ex in w["exercises"]:
        sets_str = ", ".join(html.escape(_fmt_set(s)) for s in ex["sets"])
        lines.append(f"• <b>{html.escape(ex['name'])}</b>: {sets_str}")
        total_sets += len(ex["sets"])
    lines += ["", f"<i>{len(w['exercises'])} упр., {total_sets} подх.</i>"]
    return "\n".join(lines)


# ── Import через API ────────────────────────────────────────────────────────────

def _api_import(w: dict) -> str:
    """Отправляет разобранную тренировку в Go-приложение, возвращает её id.
    Числа отдаём как есть — Go-сторона валидирует и пишет через общий Create
    (параметризованный SQL), поэтому ручная коэрция/экранирование больше не нужны.
    """
    payload = {
        "login":  CLIENT_LOGIN,
        "title":  w.get("title", ""),
        "date":   w.get("date", ""),
        "gym":    w.get("gym"),
        "notes":  w.get("notes", ""),
        "exercises": [
            {
                "name": ex["name"],
                "sets": [
                    {
                        "weight":       s.get("weight"),
                        "reps":         s.get("reps"),
                        "rpe":          s.get("rpe"),
                        "rest_seconds": s.get("rest_seconds"),
                        "notes":        s.get("notes") or "",
                    }
                    for s in ex.get("sets", [])
                ],
            }
            for ex in w.get("exercises", [])
        ],
    }
    return _api("POST", "/api/import", payload).get("id", "")


# ── handlers ──────────────────────────────────────────────────────────────────

async def cmd_start(update: Update, _: ContextTypes.DEFAULT_TYPE):
    if not _auth(update):
        return
    await update.message.reply_text(
        "Пришли текст тренировки в любом формате — разберу и добавлю в журнал.\n"
        "/cancel — отменить текущее добавление"
    )


async def cmd_cancel(update: Update, _: ContextTypes.DEFAULT_TYPE):
    if not _auth(update):
        return
    pending.pop(update.effective_user.id, None)
    await update.message.reply_text("Отменено.")


async def handle_text(update: Update, _: ContextTypes.DEFAULT_TYPE):
    if not _auth(update):
        await update.message.reply_text("⛔ Доступ запрещён.")
        return

    uid  = update.effective_user.id
    text = update.message.text.strip()
    msg  = await update.message.reply_text("⏳ Разбираю тренировку...")

    try:
        # _parse делает блокирующий I/O (Claude API + psql) — в отдельном
        # потоке, чтобы не вешать event loop бота на время инференса.
        workout = await asyncio.to_thread(_parse, text)
    except json.JSONDecodeError as e:
        await msg.edit_text(f"❌ Claude вернул невалидный JSON: {e}\n\nПопробуй переформулировать.")
        return
    except Exception as e:
        log.exception("parse error")
        await msg.edit_text(f"❌ Ошибка при разборе: {e}")
        return

    pending[uid] = workout

    # Если зал не распознан — сперва спросим его кнопками (иначе тренировка уходит
    # без зала, что и ломало отчёт + теряло данные). Если Claude зал нашёл — сразу превью.
    if not workout.get("gym"):
        try:
            gyms = await asyncio.to_thread(_api_gyms)
        except Exception:
            gyms = []
        rows = [[InlineKeyboardButton(g, callback_data="gym:" + g)] for g in gyms]
        rows.append([InlineKeyboardButton("— без зала", callback_data="gym:")])
        await msg.edit_text(_preview(workout) + "\n\n🏋 <b>В каком зале была тренировка?</b>",
                            parse_mode="HTML", reply_markup=InlineKeyboardMarkup(rows))
        return

    await msg.edit_text(_preview(workout), parse_mode="HTML", reply_markup=_confirm_kb())


def _confirm_kb() -> InlineKeyboardMarkup:
    return InlineKeyboardMarkup([[
        InlineKeyboardButton("✅ Добавить", callback_data="confirm"),
        InlineKeyboardButton("❌ Отмена",  callback_data="cancel"),
    ]])


async def handle_callback(update: Update, _: ContextTypes.DEFAULT_TYPE):
    if not _auth(update):
        return

    q      = update.callback_query
    await q.answer()
    uid    = update.effective_user.id
    action = q.data

    # выбор зала (когда Claude его не распознал) → проставляем и показываем превью
    if action.startswith("gym:"):
        workout = pending.get(uid)
        if not workout:
            await q.edit_message_text("Нет данных. Пришли тренировку заново.")
            return
        gym = action[len("gym:"):]
        workout["gym"] = gym or None  # пусто = без зала
        await q.edit_message_text(_preview(workout), parse_mode="HTML", reply_markup=_confirm_kb())
        return

    if action == "cancel":
        pending.pop(uid, None)
        await q.edit_message_text("❌ Отменено.")
        return

    if action == "confirm":
        workout = pending.get(uid)
        if not workout:
            await q.edit_message_text("Нет данных. Пришли тренировку заново.")
            return

        await q.edit_message_text("⏳ Сохраняю в базу...")

        # 1. Собственно сохранение. Только сбой ЗДЕСЬ — это «ошибка сохранения».
        try:
            wid = await asyncio.to_thread(_api_import, workout)
        except Exception as e:
            log.exception("insert error")
            await q.edit_message_text(f"❌ Ошибка сохранения: {e}")
            return  # pending не трогаем — пусть можно повторить

        pending.pop(uid, None)
        log.info(f"inserted workout {wid}: {workout['title']} {workout['date']}")

        # 2. Обновление сообщения в Telegram — best-effort. Тренировка уже в базе,
        # поэтому таймаут/сбой Telegram API здесь НЕ выдаём за ошибку сохранения
        # (иначе пользователь жмёт «Добавить» повторно → дубликаты).
        open_kb = InlineKeyboardMarkup([[
            InlineKeyboardButton("↗ Открыть в приложении", url=f"{APP_URL}/workouts/{wid}"),
        ]])
        try:
            await q.edit_message_text(
                f"✅ <b>{html.escape(workout['title'])}</b> — сохранена, {_fmt_date(workout['date'])}.",
                parse_mode="HTML",
                reply_markup=open_kb,
            )
        except Exception:
            log.exception("post-save message edit failed (workout %s already saved)", wid)


# ── main ──────────────────────────────────────────────────────────────────────

def main():
    app = (
        Application.builder()
        .token(BOT_TOKEN)
        .connect_timeout(15)
        .read_timeout(20)
        .write_timeout(20)
        .pool_timeout(20)
        .build()
    )
    app.add_handler(CommandHandler("start",  cmd_start))
    app.add_handler(CommandHandler("cancel", cmd_cancel))
    app.add_handler(MessageHandler(filters.TEXT & ~filters.COMMAND, handle_text))
    app.add_handler(CallbackQueryHandler(handle_callback))
    log.info("workout-bot started")
    app.run_polling(drop_pending_updates=True)


if __name__ == "__main__":
    main()
