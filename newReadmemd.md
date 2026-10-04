# Gophermart — накопительная система лояльности

HTTP API системы лояльности интернет-магазина «Гофермарт»: регистрация и аутентификация пользователей, приём и обработка номеров заказов, начисление и списание баллов лояльности.

Сервис асинхронно взаимодействует с внешней системой расчёта вознаграждений (**Accrual**), которая определяет, сколько баллов положено за каждый заказ.

## Содержание

- [Архитектура](#архитектура)
- [Стек](#стек)
- [Потоки запросов](#потоки-запросов)
  - [Регистрация](#регистрация)
  - [Аутентификация](#аутентификация)
  - [Загрузка номера заказа](#загрузка-номера-заказа)
  - [Получение списка заказов](#получение-списка-заказов)
  - [Получение баланса](#получение-баланса)
  - [Списание баллов](#списание-баллов)
  - [История списаний](#история-списаний)
  - [Фоновая обработка заказов](#фоновая-обработка-заказов)
- [Архитектурные решения](#архитектурные-решения)
  - [Деньги как decimal.Decimal, а не float](#деньги-как-decimaldecimal-а-не-float)
  - [Фоновый воркер: push + polling](#фоновый-воркер-push--polling)
  - [Rate limiting на внешнюю систему расчёта](#rate-limiting-на-внешнюю-систему-расчёта)
- [API](#api)
- [Запуск](#запуск)
- [Конфигурация](#конфигурация)

---

- [Accrual — система расчёта баллов](#accrual--система-расчёта-баллов)
  - [Архитектура](#архитектура-1)
  - [Таблицы](#таблицы)
  - [Потоки запросов](#потоки-запросов-1)
    - [Регистрация механики вознаграждения](#регистрация-механики-вознаграждения)
    - [Регистрация заказа](#регистрация-заказа)
    - [Получение статуса расчёта](#получение-статуса-расчёта)
    - [Фоновый расчёт вознаграждения](#фоновый-расчёт-вознаграждения)
  - [Архитектурные решения](#архитектурные-решения-1)
    - [Кэш правил вознаграждения](#кэш-правил-вознаграждения)
    - [Чем воркер Accrual проще воркера Gophermart](#чем-воркер-accrual-проще-воркера-gophermart)
  - [API (Accrual)](#api-accrual)
  - [Запуск](#запуск-1)
  - [Конфигурация](#конфигурация-1)

## Архитектура

```mermaid
flowchart LR
    Client(["Клиент"]) -->|HTTP| API["Gophermart API"]
    API --> DB[("PostgreSQL")]
    API -.->|уведомление о новом заказе| Worker["Background Worker"]
    Worker -->|читает NEW / PROCESSING| DB
    Worker -->|GET /api/orders/number| Accrual["Accrual API"]
    Worker -->|обновляет статус и баланс| DB
```

Сервис состоит из HTTP API и фонового воркера, работающих в одном процессе. API отвечает на запросы пользователя синхронно и быстро; всё, что требует ожидания внешней системы (расчёт баллов), уходит в фоновую обработку.

## Стек

| Компонент | Технология |
|---|---|
| Язык | Go |
| HTTP-роутер | chi |
| База данных | PostgreSQL (pgx/v5, пул соединений) |
| Миграции | goose |
| Аутентификация | JWT |
| Хеширование паролей | Argon2 |
| Деньги | shopspring/decimal |
| HTTP-клиент к Accrual | resty |
| Конфигурация | флаги + переменные окружения + YAML |

## Потоки запросов

### Регистрация

`POST /api/user/register`

```mermaid
sequenceDiagram
    actor U as Пользователь
    participant H as Handler
    participant S as Service
    participant R as Repository
    participant DB as PostgreSQL

    U->>H: login, password
    H->>S: RegisterUser(login, password)
    S->>S: хеширование пароля (Argon2)
    S->>R: CreateUser(user)
    R->>DB: BEGIN
    R->>DB: INSERT INTO users
    R->>DB: INSERT INTO balance (0, 0)
    R->>DB: COMMIT
    R-->>S: ok
    S->>S: выпуск JWT
    S-->>H: token
    H-->>U: 200 OK + токен (автологин)
```

Пользователь и его нулевой баланс создаются одной транзакцией: учётная запись без баланса — невалидное состояние, поэтому их существование гарантировано атомарно.

### Аутентификация

`POST /api/user/login`

```mermaid
sequenceDiagram
    actor U as Пользователь
    participant H as Handler
    participant S as Service
    participant R as Repository

    U->>H: login, password
    H->>S: LoginUser(login, password)
    S->>R: GetUserByLogin(login)
    R-->>S: user (password hash)
    S->>S: сверка пароля
    S->>S: выпуск JWT
    S-->>H: token
    H-->>U: 200 OK + токен
```

### Загрузка номера заказа

`POST /api/user/orders`

```mermaid
sequenceDiagram
    actor U as Пользователь
    participant H as Handler
    participant S as Service
    participant R as Repository
    participant W as Background Worker

    U->>H: номер заказа
    H->>H: валидация алгоритмом Луна
    H->>S: UploadOrder(number, userID)
    S->>R: CreateOrder(status=NEW)
    R-->>S: ok
    S-->>W: Notify(order) (неблокирующе)
    S-->>H: ok
    H-->>U: 202 Accepted
```

Ответ пользователю не ждёт обработки заказа — она происходит асинхронно в воркере. `Notify` — лёгкая постановка в очередь, не блокирующая HTTP-ответ.

### Получение списка заказов

`GET /api/user/orders`

```mermaid
sequenceDiagram
    actor U as Пользователь
    participant H as Handler
    participant S as Service
    participant R as Repository

    U->>H: запрос списка
    H->>S: GetOrders(userID)
    S->>R: GetOrdersByUser(userID)
    R-->>S: заказы (сортировка по uploaded_at)
    S-->>H: список
    H-->>U: 200 OK / 204 No Content
```

### Получение баланса

`GET /api/user/balance`

```mermaid
sequenceDiagram
    actor U as Пользователь
    participant H as Handler
    participant S as Service
    participant R as Repository

    U->>H: запрос баланса
    H->>S: GetBalance(userID)
    S->>R: GetBalance(userID)
    R-->>S: current, withdrawn
    S-->>H: баланс
    H-->>U: 200 OK
```

### Списание баллов

`POST /api/user/balance/withdraw`

```mermaid
sequenceDiagram
    actor U as Пользователь
    participant H as Handler
    participant S as Service
    participant R as Repository
    participant DB as PostgreSQL

    U->>H: order, sum
    H->>S: Withdraw(userID, order, sum)
    S->>R: Withdraw(userID, order, sum)
    R->>DB: BEGIN
    R->>DB: SELECT balance FOR UPDATE
    alt средств недостаточно
        R-->>S: ErrInsufficientFunds
        S-->>H: ошибка
        H-->>U: 402 Payment Required
    else средств достаточно
        R->>DB: UPDATE balance (current -, withdrawn +)
        R->>DB: INSERT INTO transactions
        R->>DB: COMMIT
        R-->>S: ok
        S-->>H: ok
        H-->>U: 200 OK
    end
```

`SELECT ... FOR UPDATE` блокирует строку баланса на время транзакции — это исключает гонку, при которой два одновременных списания проходят проверку платёжеспособности раньше, чем кто-то из них спишет средства.

### История списаний

`GET /api/user/withdrawals`

```mermaid
sequenceDiagram
    actor U as Пользователь
    participant H as Handler
    participant S as Service
    participant R as Repository

    U->>H: запрос истории
    H->>S: GetWithdrawals(userID)
    S->>R: GetWithdrawals(userID)
    R-->>S: список списаний (сортировка по processed_at)
    S-->>H: список
    H-->>U: 200 OK / 204 No Content
```

### Фоновая обработка заказов

Воркер запускается двумя путями: по уведомлению сразу после загрузки заказа и по таймеру (polling), который подхватывает всё, что уведомление могло пропустить — таким образом надёжность обработки не зависит от доставки уведомления.

```mermaid
sequenceDiagram
    participant W as Worker
    participant DB as PostgreSQL
    participant A as Accrual API

    alt по уведомлению
        Note over W: заказ только что загружен
    else по таймеру
        W->>DB: SELECT заказы NEW / PROCESSING
    end

    W->>A: GET /api/orders/{number}
    alt ещё не зарегистрирован в Accrual
        A-->>W: 204
        Note over W: ничего не меняем, подхватит следующий проход
    else считается
        A-->>W: REGISTERED / PROCESSING
        W->>DB: обновить статус на PROCESSING
    else расчёт завершён
        A-->>W: PROCESSED / INVALID, accrual
        W->>DB: BEGIN
        W->>DB: UPDATE orders (статус, accrual)
        opt статус PROCESSED
            W->>DB: UPDATE balance (+ accrual)
        end
        W->>DB: COMMIT
    end
```

Обновление статуса заказа и начисление баллов происходят в одной транзакции — пользователь никогда не увидит заказ со статусом `PROCESSED`, но без зачисленных баллов на балансе.

## Архитектурные решения

### Деньги как decimal.Decimal, а не float

Баллы лояльности представлены типом `decimal.Decimal` (`github.com/shopspring/decimal`) во всех денежных операциях — балансе, начислениях, списаниях.

`float64` не подходит для денежных расчётов, потому что хранит числа в двоичной системе с плавающей запятой, в которой большинство десятичных дробей (например, `0.1`) не представимы точно. После нескольких операций сложения и вычитания накопленная погрешность может привести к расхождению баланса на доли копейки — недопустимо для системы, где баланс должен сходиться с точностью до цента при любом количестве операций.

`decimal.Decimal` хранит число как целочисленную мантиссу с явным показателем степени (десятичную дробь, а не двоичную), поэтому арифметика над ним точна в той же мере, в какой точны десятичные дроби, с которыми работает человек.

Дополнительно: `Accrual` в заказе — это `*decimal.Decimal` (указатель), а не значение. Это позволяет различить два разных состояния в JSON-ответе — «начисление ещё не рассчитано» (поле отсутствует, `nil`) и «начисление рассчитано и равно нулю» (поле присутствует со значением `0`). При маршалинге отключены кавычки вокруг числа (`decimal.MarshalJSONWithoutQuotes = true`), чтобы в ответе API баллы передавались как JSON-число (`500.5`), а не как строка (`"500.5"`).

### Фоновый воркер: push + polling

Обработка заказа запускается двумя независимыми механизмами:

- **Push.** Сразу после успешной загрузки заказа хендлер неблокирующе уведомляет воркер — это минимизирует задержку между загрузкой заказа и первым обращением к Accrual.
- **Polling.** Воркер периодически сам читает из БД все заказы в статусах `NEW`/`PROCESSING`.

Polling — не запасной вариант «на всякий случай», а источник истины: корректность обработки заказов не зависит от того, дошло ли уведомление. Если уведомление потеряно (например, при перезапуске процесса между загрузкой заказа и уведомлением), заказ всё равно будет найден и обработан на следующем проходе polling-цикла. Push существует только как оптимизация задержки поверх уже надёжного механизма.

Число одновременно обрабатываемых заказов ограничено пулом воркеров фиксированного размера — это защищает Accrual от резкого всплеска запросов при массовой загрузке заказов.

### Rate limiting на внешнюю систему расчёта

Accrual ограничивает число запросов в единицу времени и сообщает об этом ответом `429 Too Many Requests` с заголовком `Retry-After`. При получении такого ответа воркер приостанавливает **все** свои обращения к Accrual на указанное время — не только по заказу, вызвавшему ограничение, а глобально для всего сервиса, — и возобновляет их по истечении паузы. Это предотвращает ситуацию, при которой несколько параллельных воркеров продолжают получать `429` от одного и того же ограничения, вместо того чтобы скоординированно подождать.

## API

| Метод | Путь | Auth | Описание |
|---|---|---|---|
| POST | `/api/user/register` | нет | Регистрация пользователя |
| POST | `/api/user/login` | нет | Аутентификация |
| POST | `/api/user/orders` | да | Загрузка номера заказа |
| GET | `/api/user/orders` | да | Список загруженных заказов |
| GET | `/api/user/balance` | да | Текущий баланс баллов |
| POST | `/api/user/balance/withdraw` | да | Списание баллов |
| GET | `/api/user/withdrawals` | да | История списаний |

Аутентификация — JWT, передаётся в заголовке `Authorization`.

## Запуск

```bash
go run ./cmd/gophermart \
  -a :8080 \
  -d "postgres://user:pass@localhost:5432/gophermart" \
  -r "http://localhost:8081" \
  -config config/gophermart.yaml
```

## Конфигурация

| Переменная | Флаг | Описание |
|---|---|---|
| `RUN_ADDRESS` | `-a` | Адрес и порт запуска |
| `DATABASE_URI` | `-d` | Строка подключения к PostgreSQL |
| `ACCRUAL_SYSTEM_ADDRESS` | `-r` | Адрес сервиса Accrual |
| `GOPHERMART_CONFIG_PATH` | `-config` | Путь к YAML-конфигу (логирование, rate limit, воркер) |

Приоритет: переданные флаги/переменные окружения переопределяют значения из YAML-конфига, который отвечает за настройки, не входящие в обязательный по спецификации набор (логирование, лимиты запросов, параметры фонового воркера).

---

# Accrual — система расчёта баллов

Внутренний сервис в доверенном контуре, который рассчитывает вознаграждение за заказ по составу купленных товаров. У Accrual собственная база данных и собственный HTTP API; единственная точка соприкосновения с Gophermart — эндпоинт `GET /api/orders/{number}`, который периодически опрашивает фоновый воркер Gophermart.

По спецификации регистрацию совершённого заказа (`POST /api/orders`) выполняет интернет-магазин в момент покупки — Accrual выступает пассивным получателем этих данных. Gophermart никогда не вызывает `POST /api/orders` сам, он только читает результат расчёта через `GET /api/orders/{number}`.

## Архитектура

```mermaid
flowchart LR
    Shop(["Интернет-магазин"]) -->|POST /api/orders| API["Accrual API"]
    Manager(["Менеджер"]) -->|POST /api/goods| API
    Gophermart(["Gophermart Worker"]) -->|GET /api/orders/number| API
    API --> DB[("PostgreSQL")]
    API -.->|уведомление о новом заказе| Worker["Background Worker"]
    Worker -->|читает REGISTERED / PROCESSING| DB
    Worker -->|поиск совпадений| Cache["Кэш правил вознаграждения"]
    Worker -->|обновляет статус и accrual| DB
```

## Таблицы

```sql
CREATE TYPE o_status AS ENUM ('REGISTERED', 'PROCESSING', 'INVALID', 'PROCESSED');

CREATE TABLE orders (
    order_num    VARCHAR(255) PRIMARY KEY,
    order_status o_status DEFAULT 'REGISTERED' NOT NULL,
    accrual      NUMERIC(15, 2),
    uploaded_at  TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE goods (
    good_id     UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    description VARCHAR(255) NOT NULL,
    price       NUMERIC(15, 2) NOT NULL,
    order_num   VARCHAR(255) REFERENCES orders(order_num)
);

CREATE TABLE rewards (
    uuid         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    match        VARCHAR(255) NOT NULL UNIQUE,
    reward_value NUMERIC(15, 2) NOT NULL,
    reward_type  VARCHAR(10) NOT NULL  -- '%' — процент от цены, 'pt' — фиксированные баллы
);
```

`rewards.match` уникален — одна механика вознаграждения регистрируется на один ключ поиска. `goods.order_num` связывает товары с заказом (один заказ — много товаров).

## Потоки запросов

### Регистрация механики вознаграждения

`POST /api/goods`

```mermaid
sequenceDiagram
    actor M as Менеджер
    participant H as Handler
    participant S as Service
    participant R as Repository
    participant C as Reward Rules Cache

    M->>H: match, reward_value, reward_type
    H->>S: RegisterRule(rule)
    S->>R: Create(rule)
    alt match уже зарегистрирован
        R-->>S: ErrRuleAlreadyExists
        S-->>H: ошибка
        H-->>M: 409 Conflict
    else успешно
        R-->>S: ok
        S->>C: Add(rule)
        S-->>H: ok
        H-->>M: 200 OK
    end
```

Правило добавляется в кэш сразу после успешной записи в БД — отдельного похода в БД за свежими правилами не требуется, объект уже есть на руках в момент вставки.

### Регистрация заказа

`POST /api/orders`

```mermaid
sequenceDiagram
    actor Shop as Интернет-магазин
    participant H as Handler
    participant S as Service
    participant R as Repository
    participant DB as PostgreSQL
    participant W as Background Worker

    Shop->>H: order, goods[]
    H->>S: RegisterOrder(order, goods)
    S->>R: CreateOrder(order, goods)
    R->>DB: BEGIN
    R->>DB: INSERT INTO orders (status=REGISTERED)
    alt заказ уже зарегистрирован
        R-->>S: ErrOrderAlreadyProcessing
        S-->>H: ошибка
        H-->>Shop: 409 Conflict
    else новый заказ
        R->>DB: COPY INTO goods
        R->>DB: COMMIT
        R-->>S: ok
        S-->>W: Notify(order) (неблокирующе)
        S-->>H: ok
        H-->>Shop: 202 Accepted
    end
```

Товары заказа вставляются пакетно через `COPY`, а не построчными `INSERT` — это заметно быстрее при большом составе заказа и не требует отдельных round-trip'ов к БД на каждый товар.

### Получение статуса расчёта

`GET /api/orders/{number}`

```mermaid
sequenceDiagram
    participant Gm as Gophermart Worker
    participant H as Handler
    participant S as Service
    participant R as Repository

    Gm->>H: номер заказа
    H->>S: GetOrder(number)
    S->>R: GetOrder(number)
    alt заказ не зарегистрирован
        R-->>S: not found
        S-->>H: not found
        H-->>Gm: 204 No Content
    else заказ найден
        R-->>S: status, accrual
        S-->>H: результат
        H-->>Gm: 200 OK (status, accrual)
    end
```

### Фоновый расчёт вознаграждения

Так же, как и в Gophermart, обработка запускается и уведомлением сразу после регистрации заказа, и периодическим опросом БД (`REGISTERED`/`PROCESSING`) — polling остаётся источником истины на случай, если уведомление потеряно или заказ завис между сменой статуса и расчётом.

```mermaid
sequenceDiagram
    participant W as Worker
    participant DB as PostgreSQL
    participant C as Reward Rules Cache

    alt по уведомлению
        Note over W: заказ только что зарегистрирован
    else по таймеру
        W->>DB: SELECT заказы REGISTERED / PROCESSING
    end

    alt состав заказа пуст
        W->>DB: UPDATE orders (status=INVALID)
    else есть товары
        W->>DB: UPDATE orders (status=PROCESSING)
        W->>C: Get() — все правила вознаграждения
        loop по каждому товару заказа
            loop по каждому правилу
                W->>W: совпадение match в description?
                opt совпало
                    W->>W: accrual += вознаграждение
                end
            end
        end
        W->>DB: UPDATE orders (status=PROCESSED, accrual)
    end
```

Обновление статуса и `accrual` — один атомарный `UPDATE` с условием `status NOT IN (PROCESSED, INVALID)`: если заказ уже финализирован (например, параллельно сработали уведомление и polling), повторная запись просто не применится, это предотвращает повторный расчёт одного и того же заказа.

## Архитектурные решения

### Кэш правил вознаграждения

Правила вознаграждения (`rewards`) держатся в памяти процесса — потокобезопасной структуре со слайсом правил под `sync.RWMutex` — а не вычитываются из БД на каждый расчёт.

Правила меняются только через один канал — `POST /api/goods` — и ни одним другим. Это значит, что кэш не обязан гадать, протух ли он: он точно знает единственный момент, когда нужно обновиться — непосредственно после успешной записи нового правила в БД. Поэтому кэш заполняется один раз при старте приложения (`ListRules` из БД) и далее обновляется инкрементально (`Add`) при каждой регистрации нового правила, без переинициализации всего набора.

Альтернатива — запрашивать правила из БД при каждом расчёте заказа — была отклонена: это самый частый путь (а не исключение), и платить round-trip'ом к БД за операцию, которая по сути является статичным поиском по небольшому набору строк, не оправдано на объёмах учебного проекта. То же касается и более сложных структур данных для поиска подстроки (суффиксные деревья, Aho-Corasick) — при текущем количестве правил линейный перебор `strings.Contains` занимает микросекунды, усложнение здесь было бы преждевременной оптимизацией.

### Чем воркер Accrual проще воркера Gophermart

Воркер Accrual устроен по тому же паттерну push + polling + bounded worker pool, что и в Gophermart, но лишён двух самых сложных частей:

- **Нет кросс-табличной транзакции.** Gophermart должен был атомарно обновить и статус заказа, и баланс пользователя — две разные таблицы. Accrual обновляет только одну таблицу (`orders`), поэтому достаточно одного атомарного `UPDATE` с условием, без оборачивающей транзакции.
- **Нет исходящего rate limiting.** Воркер Gophermart обращается к внешнему сервису (Accrual) и обязан уважать его `429 Too Many Requests`. Воркер Accrual никуда не стучится — вся его работа это чтение из своей же БД и сопоставление с кэшем в памяти, — поэтому `gate`-механизм с паузой по `Retry-After` ему не нужен.

## API (Accrual)

| Метод | Путь | Описание |
|---|---|---|
| GET | `/api/orders/{number}` | Статус расчёта начислений по заказу |
| POST | `/api/orders` | Регистрация совершённого заказа |
| POST | `/api/goods` | Регистрация механики вознаграждения за товар |

Аутентификация не требуется — сервис работает в доверенном контуре.

## Запуск

```bash
go run ./cmd/accrual \
  -a :8081 \
  -d "postgres://user:pass@localhost:5432/accrual"
```

## Конфигурация

| Переменная | Флаг | Описание |
|---|---|---|
| `RUN_ADDRESS` | `-a` | Адрес и порт запуска |
| `DATABASE_URI` | `-d` | Строка подключения к PostgreSQL |

В отличие от Gophermart, Accrual не взаимодействует с внешними сервисами, поэтому в его обязательном наборе параметров нет аналога `ACCRUAL_SYSTEM_ADDRESS`.
