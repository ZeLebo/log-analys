# log-analys

CLI-инструмент для интерактивного анализа JSON-логов из `stdin` с фильтрами в реальном времени.

Основные возможности:
- парсинг JSON-логов с настраиваемой схемой (`config/schema.yaml`);
- выделение базовых полей (`timestamp`, `level`, `service`, `message`);
- добавление любых пользовательских полей через `custom_fields` в схеме (без изменений кода);
- поддержка поля `extra` и фильтрации по вложенным ключам (`extra.some.path`);
- корректная обработка stack trace: не-JSON строки могут приклеиваться в `raw` предыдущего события;
- вывод в обычном формате (`show`) и pretty JSON (`showjson`).

## Требования

- Go `1.18+`
- Unix-подобная среда с `/dev/tty` (команды читаются из tty, а логи — из `stdin`)

## Быстрый старт

Сборка:

```bash
go build -o log-analys ./app
```

Запуск на файле логов:

```bash
cat config/output.log | ./log-analys
```

После запуска вводи команды в терминал (промпт `>`).

## Структура проекта

- `app/main.go`, `app/handler.go` — CLI-цикл и обработка команд.
- `domain/` — бизнес-логика (парсер, фильтрация, ring-buffer, загрузка схемы).
- `models/` — модели данных (`Event`, `Filter`, `Schema`).
- `utils/` — утилиты общего назначения.
- `tests/` — тесты.
- `config/schema.yaml`, `config/output.log`, `config/mock_output.log` — конфиг и тестовые данные.

## Команды

- `show [n]` — показать последние `n` подходящих событий (по умолчанию `50`)
- `showjson [n]` — показать последние `n` событий в формате pretty JSON (по умолчанию `20`)
- `eq <field> <value>` — точное совпадение
- `like <field> <value>` — подстрока
- `clear` — очистить все активные фильтры
- `exit` / `quit` — выйти

Примеры:

```text
eq company_name some-company
like msg Saving
like raw Traceback
like extra.nested.k v
eq ticket T-42
like field.user alex
show 30
showjson 5
```

## Поля для фильтрации

Поддерживаются:
- `level`
- `service`
- `msg`
- `raw`
- `extra` (весь объект как строка)
- `extra.<path>` (вложенные поля в `extra`)
- `attr.<path>` (вложенные поля исходного JSON-события)
- `<custom_field_name>` (любое поле из `custom_fields`)
- `field.<custom_field_name>` (явный префикс для custom-поля)

`<path>` задается через точку, например: `extra.user.id`.

## Настройка формата логов: `config/schema.yaml`

Файл `config/schema.yaml` управляет тем, какие ключи из входного JSON считаются целевыми полями.

Текущий шаблон:

```yaml
buffer_size: 50000

fields:
  timestamp:
    - ts
    - time
    - timestamp
  level:
    - level
  service:
    - logger
    - service
    - filename
  message:
    - msg
    - message

custom_fields:
  call_id:
    - call_id
  company_name:
    - company_name

non_json:
  append_to_previous_raw: true
  create_event_if_no_last: true
```

### Пояснения

- `buffer_size` — размер кольцевого буфера событий.
- `fields.*` — список ключей-кандидатов в порядке приоритета.
- `custom_fields` — произвольные сервисные поля (алиас -> список ключей-кандидатов).
- `custom_fields` позволяет добавить/заменить любые бизнес-поля только через конфиг.
- `non_json.append_to_previous_raw`:
  - `true`: не-JSON строка добавляется в `raw` последнего события;
  - `false`: не добавляется к предыдущему событию.
- `non_json.create_event_if_no_last`:
  - `true`: если не-JSON строку некуда приклеить, создается отдельное событие;
  - `false`: такая строка игнорируется.

Если `config/schema.yaml` отсутствует или пустой, используются встроенные значения по умолчанию.

## Как обрабатываются stack trace и ошибки

- JSON-строки создают обычные события.
- Не-JSON строки (например, Python traceback) по настройке приклеиваются в `raw` предыдущего события.
- В `showjson` поле `raw` специально не выводится, чтобы не раздувать вывод.
- При этом `raw` хранится внутри и доступен для фильтров (`like raw Traceback`).

## Тесты

```bash
go test ./...
```
