# Обновление сервера после локальных изменений

Инструкция для ситуации: код изменён и проверен локально, нужно выложить новую версию на сервер.

Сервер: `crm.neiroot.ru`

## Простой деплой (рекомендуется)

На Mac:

1. Закоммить и запушить изменения.
2. Одной командой обновить сервер:

```bash
ssh root@138.16.184.152 'cd /opt/proclients && git pull && ./scripts/deploy.sh'
```

Скрипт сам:
- заберёт код (`git pull` внутри скрипта тоже есть — повторный pull безопасен);
- применит **только новые** миграции (через таблицу `schema_migrations`);
- пересоберёт и перезапустит backend;
- пересоберёт frontend и сделает `nginx reload`.

Полезные варианты:

```bash
# только frontend (логотип, UI, авторизация во вкладках)
ssh root@138.16.184.152 'cd /opt/proclients && git pull && ./scripts/deploy.sh --frontend-only'

# только backend + миграции
ssh root@138.16.184.152 'cd /opt/proclients && git pull && ./scripts/deploy.sh --backend-only'
```

Первый запуск после появления скрипта: если БД уже живая, скрипт **один раз** пометит миграции `001`–`021` как применённые (без повторного запуска) и накатит только более новые. Дальше — только новые файлы.

После деплоя в браузере лучше Cmd+Shift+R.

---

## Подробный ручной вариант (если скрипт недоступен)

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

## 5.1 Интеграция Билайн (XSI-Events → автосоздание лидов)

Если включаете интеграцию с Облачной АТС Билайн:

1) В `backend/.env` на сервере добавьте:

```env
BEELINE_API_TOKEN=...                    # токен из личного кабинета Билайн (X-MPBX-API-AUTH-TOKEN)
BEELINE_WEBHOOK_SECRET=...               # любой длинный секрет
BEELINE_CREATED_BY_USER_ID=...           # UUID пользователя в нашей БД (например, админ или отдельный “Система”)
BEELINE_WEBHOOK_DEBUG=true               # временно: писать raw body webhook в journalctl
```

Краткие строки `[beeline-webhook]` пишутся всегда. Полный raw body — только при `BEELINE_WEBHOOK_DEBUG=true`.

Просмотр логов:

```bash
journalctl -u proclients-api --since "2 hours ago" --no-pager | grep beeline-webhook
# или по странному номеру:
journalctl -u proclients-api --since "2 hours ago" --no-pager | grep -F "4721076150"
```

После разбора отключите debug: `BEELINE_WEBHOOK_DEBUG=false` и перезапустите API.

2) В личном кабинете Билайн создайте подписку на XSI-Events (пример):

```bash
cd /opt/proclients
chmod +x backend/scripts/beeline_subscribe.sh backend/scripts/beeline_subscribe_all.sh

# Один номер
BEELINE_API_TOKEN="$BEELINE_API_TOKEN" \
BEELINE_PATTERN="9613001616@rnd.so.ims.mnc099.mcc250.3gppnetwork.org" \
BEELINE_CALLBACK_URL="https://crm.neiroot.ru/api/v1/integrations/beeline/xsi-events/$BEELINE_WEBHOOK_SECRET?trafficSource=Знал%20о%20производстве" \
./backend/scripts/beeline_subscribe.sh

# Все многоканальные номера сразу (источник трафика в query trafficSource)
BEELINE_API_TOKEN="$BEELINE_API_TOKEN" \
BEELINE_WEBHOOK_SECRET="$BEELINE_WEBHOOK_SECRET" \
./backend/scripts/beeline_subscribe_all.sh
```

Маппинг номер → источник трафика:

| Номер | Источник |
|-------|----------|
| (961) 300-16-16 | Знал о производстве |
| (961) 301-50-50 | Знал о производстве |
| (961) 300-14-41 | Авито (AutoFactory) |
| (966) 206-69-59 | Визитка(авточехлы) |
| (961) 319-52-19 | Авито (Автоатрибут) |
| (906) 454-58-34 | Авито (Автоателье) |
| (906) 454-58-66 | Авито (Автоателье) |
| (903) 430-67-67 | Яндекс карты |
| (903) 436-33-36 | Instagram |
| (961) 301-14-58 | Вконтакте |
| (961) 301-14-60 | 2gis |

Важно: наш webhook проверяет секрет либо по заголовку `X-Beeline-Secret`, либо по query-параметру `?secret=...`, либо в path `/xsi-events/<secret>`. Источник трафика передаётся в `?trafficSource=...` (скрипт `beeline_subscribe_all.sh` кодирует URL автоматически).

Подписка живёт `expires` секунд (по умолчанию 3600). Чтобы она не обрывалась, настройте автопродление (раз в 30 минут):

```bash
cd /opt/proclients
chmod +x backend/scripts/beeline_renew_subscriptions.sh

sudo cp deploy/systemd/proclients-beeline-renew.service /etc/systemd/system/
sudo cp deploy/systemd/proclients-beeline-renew.timer /etc/systemd/system/

sudo systemctl daemon-reload
sudo systemctl enable --now proclients-beeline-renew.timer
```

Проверка:

```bash
systemctl status proclients-beeline-renew.timer --no-pager
systemctl list-timers --no-pager | grep beeline
tail -n 50 /var/log/proclients-beeline-renew.log
```

Ручной запуск (если нужно прямо сейчас):

```bash
sudo systemctl start proclients-beeline-renew.service
```

Альтернатива без systemd — cron (каждые 30 минут):

```cron
*/30 * * * * /opt/proclients/backend/scripts/beeline_renew_subscriptions.sh
```

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

