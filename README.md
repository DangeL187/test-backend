# Как запустить

Поднять PostgreSQL (из корня проекта):
```bash
cd docker
docker compose up -d
```

Запустить сервер (из корня проекта):
```bash
go run cmd/main.go
```

[Для тестов] Поднять тестовую БД (из корня проекта):
```bash
cd docker/test
docker compose up -d
```

# Ключевые решения
- Таблица balances для атомарного изменения баланса пользователей; `PRIMARY KEY(user_id, currency)`
- Таблица ledger_entries с `UNIQUE(reference_type, reference_id)` для ссылок на любой тип операции, в т.ч. withdrawals
- Подтверждение withdrawal (`POST /v1/withdrawals/:id/confirm`)
- Создание withdrawal в транзакции
- Атормарное обновление баланса со встроенной проверкой

# Архитектура
- `DDD` + `Clean Architecture`
  - domain слой
  - handler слой (взаимодействует с usecase)
  - usecase слой (зависит от интерфейсов репозиториев, а не напрямую от БД)
  - infra слой (реализует взаимодействие с БД и внешним миром)
- `Feature-Sliced Design` - моё предпочтение для личных проектов

# Код
- Комментарии на английском, потому что привык
- Взаимодействие с БД через `gorm`
- Фреймворк для сервера - `echo`
- Логирование через библиотеку `zap`. Ошибки логируются, если это internal server error
- Tracing ошибок через мою библиотеку `erax`. В качестве альтернативы можно было использовать `fmt`
