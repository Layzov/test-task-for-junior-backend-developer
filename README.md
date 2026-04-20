# Task Service

Сервис для управления задачами с HTTP API на Go.

## Требования

- Go `1.23+`
- Docker и Docker Compose

## Быстрый запуск через Docker Compose

```bash
docker compose up --build
```

После запуска сервис будет доступен по адресу `http://localhost:8080`.

Если `postgres` уже запускался ранее со старой схемой, пересоздай volume:

```bash
docker compose down -v
docker compose up --build
```

Причина в том, что SQL-файлы из `migrations/` монтируются в `docker-entrypoint-initdb.d` и применяются только при инициализации пустого data volume.

## Swagger

Swagger UI:

```text
http://localhost:8080/swagger/
```

OpenAPI JSON:

```text
http://localhost:8080/swagger/openapi.json
```

## API

Базовый префикс API:

```text
/api/v1
```

Основные маршруты:

- `POST /api/v1/tasks`
- `GET /api/v1/tasks`
- `POST /api/v1/tasks/generate?date=YYYY-MM-DD`
- `GET /api/v1/tasks/{id}`
- `PUT /api/v1/tasks/{id}`
- `DELETE /api/v1/tasks/{id}`
- `POST /api/v1/task-templates`
- `GET /api/v1/task-templates`
- `GET /api/v1/task-templates/{id}`
- `PUT /api/v1/task-templates/{id}`
- `DELETE /api/v1/task-templates/{id}`

## Периодические задачи

Добавлена поддержка шаблонов периодических задач и генерации задач-экземпляров на конкретную дату.

### Типы периодичности

- `daily_interval` — каждый `interval_days` день от `starts_on`
- `monthly_days` — по указанным числам месяца (`1..30`)
- `specific_dates` — только на конкретные даты
- `odd_even` — только четные или нечетные дни месяца

### Принцип работы фичи

Решение построено на разделении двух сущностей:

- `task_templates` — шаблоны периодических задач (правило повторения).
- `tasks` — реальные рабочие задачи на конкретную дату.

Поток работы:

1. Пользователь создает шаблон (`POST /api/v1/task-templates`).
2. Вызывается генерация задач на нужную дату (`POST /api/v1/tasks/generate?date=YYYY-MM-DD`).
3. Сервис проверяет шаблоны и создает задачи в `tasks`, если дата подходит под правило:
   - если `template_id` не передан — проверяются все активные шаблоны;
   - если `template_id` передан — проверяется только указанный шаблон.
4. Персонал работает уже с обычными задачами (`GET/PUT/DELETE /api/v1/tasks...`).

Связь между сущностями:

- `tasks.template_id` указывает, из какого шаблона создана задача.
- `tasks.scheduled_for` хранит дату, на которую задача создана.
- Уникальный индекс `(template_id, scheduled_for)` защищает от дублей при повторной генерации.

### Быстрые примеры API

Создание шаблона: каждые 3 дня.

```bash
curl -X POST 'http://localhost:8080/api/v1/task-templates' \
  -H 'Content-Type: application/json' \
  -d '{
    "title": "Обзвон пациентов",
    "description": "Проверить состояние после выписки",
    "recurrence": {
      "type": "daily_interval",
      "interval_days": 3
    },
    "starts_on": "2026-04-01",
    "is_active": true
  }'
```

Создание шаблона: на 15 и 30 число каждого месяца.

```bash
curl -X POST 'http://localhost:8080/api/v1/task-templates' \
  -H 'Content-Type: application/json' \
  -d '{
    "title": "Подготовка отчетности",
    "recurrence": {
      "type": "monthly_days",
      "month_days": [15, 30]
    },
    "starts_on": "2026-04-01"
  }'
```

Создание шаблона: только на конкретные даты.

```bash
curl -X POST 'http://localhost:8080/api/v1/task-templates' \
  -H 'Content-Type: application/json' \
  -d '{
    "title": "Инвентаризация",
    "recurrence": {
      "type": "specific_dates",
      "specific_dates": ["2026-04-22", "2026-05-06"]
    },
    "starts_on": "2026-04-01"
  }'
```

Создание шаблона: только по четным дням месяца.

```bash
curl -X POST 'http://localhost:8080/api/v1/task-templates' \
  -H 'Content-Type: application/json' \
  -d '{
    "title": "Проверка оборудования",
    "recurrence": {
      "type": "odd_even",
      "odd_even_type": "even"
    },
    "starts_on": "2026-04-01"
  }'
```

Генерация задач на дату (по всем активным шаблонам).

```bash
curl -X POST 'http://localhost:8080/api/v1/tasks/generate?date=2026-04-20'
```

Точечная генерация на дату только по одному шаблону.

```bash
curl -X POST 'http://localhost:8080/api/v1/tasks/generate?date=2026-04-20&template_id=1'
```

Список задач на дату.

```bash
curl 'http://localhost:8080/api/v1/tasks?scheduled_for=2026-04-20'
```

### Принятые допущения

- Все вычисления дат выполняются в `UTC`.
- Для `monthly_days` допустимы только значения `1..30` (по условию задания).
- Повторная генерация на ту же дату не создает дублей благодаря уникальному индексу `(template_id, scheduled_for)` и `upsert`.
- Изменение шаблона не переписывает уже созданные задачи автоматически.
- `POST /api/v1/tasks/generate` поддерживает опциональный `template_id` для точечной генерации; без него генерация выполняется по всем активным шаблонам.

## Автор

- GitHub: [Layzov](https://github.com/Layzov)
