# LOCAL RUN

## 1) Frontend

```bash
cd frontend
cp .env.example .env
npm install
npm run dev
```

## 2) Backend

```bash
cd backend
cp .env.example .env
go mod tidy
go run ./cmd/server
```

В `.env` обязательны `DATABASE_URL` и `JWT_SECRET` (минимум 32 символа). Пример — в `backend/.env.example`.

## 3) PostgreSQL + migrations

Если PostgreSQL в Docker:

```bash
docker run --name proclients-postgres \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=proclients \
  -p 5432:5432 -d postgres:16
```

### Применить все миграции (локально, Docker)

```bash
chmod +x backend/scripts/apply_migrations.sh
./backend/scripts/apply_migrations.sh
```

Скрипт по порядку выполняет `backend/migrations/*.up.sql` (001–014). Альтернатива — вручную через `psql` или TablePlus.

### Первый администратор (вручную, один раз)

После миграций в БД ещё нет пользователей для входа. Создай админа SQL-запросом — **пароль только здесь, не в git**.

Шаблон: `backend/scripts/create_admin.sql.example` (скопируй, подставь телефон, пароль, должность и дату рождения, выполни в psql или TablePlus).

Локально (Docker):

```bash
docker exec -i proclients-postgres psql -U postgres -d proclients \
  < backend/scripts/create_admin.sql.example
```

Перед этим отредактируй файл-копию с **своими** данными или вставь SQL прямо в TablePlus.

Телефон — формат `+79001234567`. Роль `admin` — полный доступ.

### Сотрудники (через UI)

Раздел **Сотрудники** доступен только администратору:

- **Список сотрудников** — данные из PostgreSQL (`GET /api/v1/users`)
- **Добавить сотрудника** — создание в БД (`POST /api/v1/users`)
- **Редактирование / удаление** — `PATCH` и soft-delete (`DELETE`) через API

Моковые записи не используются: все операции идут через backend → PostgreSQL.

После изменений в backend или миграций — **перезапусти backend**. Frontend — **обнови страницу** (лучше Cmd+Shift+R).
