# Room Reservation System

Микросервисное приложение для бронирования номеров в отелях.

Система позволяет пользователям просматривать доступные отели и номера, выбирать подходящий вариант и оформлять бронирование на заданный период.

Пользователь может зарегистрироваться в системе в одной из двух ролей:

- **Client** — обычный пользователь, который ищет и бронирует номера.
- **Hotel Owner** — владелец отеля, который управляет отелями, номерами и ценами.

## Возможности

- регистрация и аутентификация пользователей;
- авторизация клиентов и владельцев отелей;
- просмотр доступных отелей и номеров;
- создание бронирований на выбранный период;
- управление отелями и номерами;
- имитация процесса оплаты;
- обработка результата оплаты через webhook;
- отправка событий о бронировании через Kafka;
- отправка уведомлений клиентам и владельцам отелей;
- REST API для взаимодействия с клиентскими приложениями;
- gRPC для взаимодействия между микросервисами;
- мониторинг и трассировка запросов;
- контейнеризация всех сервисов через Docker.

## Архитектура

Приложение построено на основе микросервисной архитектуры.

Взаимодействие клиента с системой происходит через **REST API**.

Взаимодействие между внутренними сервисами осуществляется преимущественно через **gRPC**.

Для асинхронного обмена событиями используется **Apache Kafka**.

Все сервисы запускаются в Docker-контейнерах и объединяются с помощью `docker-compose`.

## Основные микросервисы

### Booking Svc

Сервис бронирования является основным сервисом, управляющим процессом создания и обработки бронирований.

Основные задачи:

- создание бронирований;
- проверка информации о выбранном номере;
- получение данных об отелях и номерах из `Hotel Svc`;
- хранение информации о бронированиях в собственной PostgreSQL базе данных;
- взаимодействие с системой оплаты;
- обработка webhook от `Payment System`;
- публикация событий о бронировании в Kafka;
- передача событий в `Notification Svc` через Kafka.

---

### Hotel Svc

Сервис управления отелями и номерами.

Основные задачи:

- хранение информации об отелях;
- хранение информации о номерах;
- хранение стоимости номеров;
- получение информации о доступных номерах;
- добавление и изменение отелей;
- добавление и изменение номеров;
- предоставление данных сервису бронирования через gRPC.

Данные сервиса хранятся в отдельной PostgreSQL базе данных.

---

### Notification Svc

Сервис уведомлений.

Основные задачи:

- получение событий о бронированиях из Kafka;
- обработка событий;
- отправка уведомлений клиентам;
- отправка уведомлений владельцам отелей;
- взаимодействие с внешней системой доставки уведомлений (`Delivery System`).

---

### Auth Svc

Сервис аутентификации и авторизации пользователей.

Основные задачи:

- регистрация пользователей;
- регистрация владельцев отелей;
- аутентификация;
- выдача JWT-токенов;
- авторизация запросов;
- хранение информации о пользователях в собственной базе данных.

---

### Payment System

Сервис имитации платежной системы.

Основные задачи:

- создание заявки на оплату;
- имитация обработки платежа;
- формирование результата платежа;
- передача результата в `Booking Svc` через webhook.

## Взаимодействие сервисов

Пример процесса создания бронирования:

1. Пользователь проходит аутентификацию через `Auth Svc`.
2. Клиент запрашивает доступные отели и номера.
3. `Booking Svc` получает информацию о номере из `Hotel Svc` через gRPC.
4. `Booking Svc` создаёт бронирование.
5. Для бронирования создаётся заявка в `Payment System`.
6. `Payment System` имитирует проведение платежа.
7. Результат оплаты передаётся в `Booking Svc` через webhook.
8. После обработки бронирования `Booking Svc` публикует событие в Kafka.
9. `Notification Svc` получает событие из Kafka.
10. `Notification Svc` отправляет уведомление пользователю или владельцу отеля.

## Трассировка и метрики

OpenTelemetry SDK в `bookingsvc`, `hotelsvc` и `notificationsvc` отправляет spans пакетами по OTLP/HTTP в Jaeger v2. Jaeger принимает, сохраняет и читает трассировки из одноузлового Elasticsearch. Volume `elasticsearch-data` сохраняет трассировки при пересоздании контейнеров. На хост опубликован только UI Jaeger на loopback среди компонентов трассировки.

Серверный span создаёт ogen, а chi извлекает входящий W3C Trace Context перед вызовом ogen. Бизнес-операции и обращения к PostgreSQL получают контекст текущего запроса. gRPC stats handlers передают этот контекст между сервисами. Kafka producer добавляет `traceparent` и, если есть, `tracestate` в заголовки записи. Consumer продолжает трассировку с контекстом своего жизненного цикла. В Jaeger SMTP-попытки видны под span отправки уведомления. SQL-параметры, текст SQL, адреса и тело письма в spans не добавляются. Логи `slog`, записанные с активным span, содержат `trace_id` и `span_id`.

| Операция | Дочерние операции |
| --- | --- |
| `CreateBooking` | BookingService, gRPC `Rooms.Get`, PostgreSQL обоих сервисов, Kafka publish/process, уведомление `booking.created`, SMTP |
| `CancelBooking` | BookingService, PostgreSQL, Kafka publish/process, уведомление `booking.cancelled`, SMTP |
| `GetAvailableRooms` | BookingService, gRPC `Rooms.GetAll`, PostgreSQL обоих сервисов |
| `GetBookings`, `GetBookingByID` | BookingService, PostgreSQL bookingsvc |

В Compose заданы `OTEL_SERVICE_NAME` и `OTEL_EXPORTER_OTLP_ENDPOINT=http://jaeger:4318` для каждого сервиса. Локальный SDK записывает 100% spans и отправляет буфер при штатной остановке. Запросы `/metrics` не создают spans. Метрики Kafka Exporter и приложений доступны в [Prometheus](http://localhost:9090/targets), дашборды — в [Grafana](http://localhost:3000).

### Запуск и безопасная проверка

Для обычного запуска используйте `docker compose up -d --build` с настроенным `notificationsvc/.env`. Чтобы проверить создание и отмену бронирования без отправки реальных писем, запустите стек с Mailpit. Override направляет SMTP на локальный приёмник с STARTTLS и тестовыми адресами.

```bash
export SMTP_TEST_CERT_DIR="$(mktemp -d)"
openssl req -x509 -newkey rsa:2048 -nodes -days 7 \
  -subj '/CN=smtp-test' -addext 'subjectAltName=DNS:smtp-test' \
  -keyout "$SMTP_TEST_CERT_DIR/key.pem" -out "$SMTP_TEST_CERT_DIR/cert.pem"
docker compose -f docker-compose.yml -f observability/compose.smtp-test.yml config --quiet
docker compose -f docker-compose.yml -f observability/compose.smtp-test.yml up -d --build --wait
```

Для проверки нужен отель со свободным номером. Создайте тестовый отель и номер, если подходящих данных ещё нет. REST API hotelsvc отвечает на создание статусом 201 без идентификатора. Запрос к PostgreSQL ниже только читает ID созданного отеля.

```bash
hotel_name="Tracing test $(date +%s)"
curl -fsS -X POST http://localhost:8080/hotels -H 'Content-Type: application/json' \
  -d "{\"name\":\"$hotel_name\",\"address\":\"Test address\",\"rating\":3,\"rooms_num\":1,\"rooms_occupied\":0}"
hotel_id="$(docker compose exec -T hotels-db psql -At -U booking -d hotels \
  -c "SELECT id FROM hotels WHERE name = '$hotel_name' ORDER BY id DESC LIMIT 1")"
curl -fsS -X POST "http://localhost:8080/hotels/$hotel_id/rooms" \
  -H 'Content-Type: application/json' \
  -d '{"room_num":101,"type":"standard","cost":1000,"is_available":1}'

python3 observability/verify_traces.py "$hotel_id"
```

Проверка выполняет пять HTTP-запросов, сверяет статусы, TraceID, родительские связи и доставку двух писем в Mailpit. Подождите несколько секунд, пока SDK и Jaeger запишут пакеты. В [Jaeger UI](http://localhost:16686) выберите сервис `bookingsvc` и имя операции либо откройте `http://localhost:16686/trace/<trace_id>`. Ту же трассировку можно получить через Query API по адресу `http://localhost:16686/api/traces/<trace_id>`. Письма видны в [Mailpit](http://localhost:8025).

## Infrastructure

Все компоненты приложения запускаются в Docker-контейнерах.

Для локального запуска всей инфраструктуры используется:

```bash
docker compose up -d --build
```

В контейнерах запускаются:

- микросервисы приложения;
- PostgreSQL;
- Apache Kafka;
- Jaeger и Elasticsearch;
- Prometheus и Grafana;
- связанные инфраструктурные компоненты.

## Tech Stack

### Backend

- Go
- REST API
- gRPC
- JWT

### Databases

- PostgreSQL

### Messaging

- Apache Kafka

### Observability

- OpenTelemetry, Jaeger v2 и Elasticsearch
- Prometheus, Grafana и Kafka Exporter
- slog

### Infrastructure

- Docker
- Docker Compose

### Development

- Git
