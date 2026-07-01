# PROclients

CRM для автомобильного ателье.

## Стек

- **Frontend:** Vue 3, TypeScript, Vite, Naive UI
- **Backend:** Go (HTTP API)
- **БД:** PostgreSQL

Подробнее: [ARCHITECTURE.md](./ARCHITECTURE.md), [DESIGN.md](./DESIGN.md)

## Запуск frontend

```bash
cd frontend
cp .env.example .env
npm install
npm run dev
```

Приложение откроется на http://localhost:5173

Подробная пошаговая шпаргалка запуска: `LOCAL_RUN.md`

## Структура layout

- Верхняя панель — статичное горизонтальное меню на всех страницах
- Боковое меню — выезжает слева по кнопке «гамбургер» (три полоски)
- Пункты бокового меню: Лиды, Сделки, Задачи
