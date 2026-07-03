# Обновление сервера после локальных изменений

Инструкция для ситуации: код изменён и проверен локально, нужно выложить новую версию на сервер.

Сервер: `crm.neiroot.ru`

## 1. Проверить локально

Выполни на Mac из корня проекта:

```bash
cd "/Users/nikitasmatko/Documents/Разработка/PROclients"
```

Backend:

```bash
cd backend
go build ./cmd/server
```

Frontend:

```bash
cd ../frontend
npm run build
```

Если есть ошибки — сначала исправь их локально.

## 2. Отправить изменения в GitHub

Из корня проекта на Mac:

```bash
cd "/Users/nikitasmatko/Documents/Разработка/PROclients"
git status
git add .
git commit -m "Краткое описание изменений"
git push
```

Пример:

```bash
git commit -m "Update employee access rules"
```

## 3. Зайти на сервер

```bash
ssh root@138.16.184.152
```

## 4. Забрать новый код на сервере

На сервере:

```bash
cd /opt/proclients
git pull
```

## 5. Если есть новые миграции

Если в изменениях появились новые файлы миграций, применяй только их (по порядку номера):

```bash
export PGPASSWORD='ProclientsDb2026'
psql -U proclients -d proclients -h localhost -f backend/migrations/018_allow_lead_attachments_and_activities.up.sql
psql -U proclients -d proclients -h localhost -f backend/migrations/019_backfill_lead_created_activities.up.sql
```

Если миграций нет — этот шаг пропусти.

## 6. Пересобрать backend

```bash
cd /opt/proclients/backend
go build -o /opt/proclients/proclients-api ./cmd/server
systemctl restart proclients-api
```

Проверка:

```bash
systemctl status proclients-api --no-pager
curl -s http://127.0.0.1:8080/health
```

Ожидаемый ответ:

```json
{"ok":true}
```

## 7. Пересобрать frontend

```bash
cd /opt/proclients/frontend
npm ci
npm run build
systemctl reload nginx
```

## 8. Проверить сайт

На Mac:

```bash
curl -s https://crm.neiroot.ru/health
```

Ожидаемый ответ:

```json
{"ok":true}
```

После этого открой в браузере:

```text
https://crm.neiroot.ru
```

Лучше обновить страницу через Cmd+Shift+R.

## Быстрый вариант без миграций

Если менялся только код:

```bash
cd /opt/proclients
git pull

cd backend
go build -o /opt/proclients/proclients-api ./cmd/server
systemctl restart proclients-api

cd /opt/proclients/frontend
npm ci
npm run build
systemctl reload nginx
```

## Если что-то сломалось

Backend:

```bash
journalctl -u proclients-api -n 80 --no-pager
```

Nginx:

```bash
nginx -t
systemctl status nginx --no-pager
```

Порты:

```bash
ss -tlnp | grep -E ':80|:443|:8080'
```

