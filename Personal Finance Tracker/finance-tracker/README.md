# 💰 Personal Finance Tracker

Современная система для отслеживания личных финансов с REST API на Go и веб-интерфейсом.

## ✨ Функциональность

### 🔐 Аутентификация и пользователи
- ✅ Регистрация и аутентификация пользователей (JWT)
- ✅ Управление сессиями
- ✅ Профиль пользователя с настройками

### 💳 Управление финансами
- ✅ **Мультивалютность** - поддержка RUB, USD, TJS
- ✅ **Счета** - создание и управление счетами в разных валютах
- ✅ **Обмен валют** - автоматическое обновление курсов
- ✅ **Категории** - настройка категорий доходов/расходов
- ✅ **Транзакции** - добавление, просмотр, фильтрация
- ✅ **Статистика** - сводки по периодам и категориям

### 📊 Аналитика
- ✅ Фильтрация транзакций по периоду
- ✅ Статистика по категориям
- ✅ Месячные сводки
- ✅ Графики и диаграммы

## 🛠 Технологии

### Backend
- **Go 1.23** с чистой архитектурой (Clean Architecture)
- **Gin** - высокопроизводительный HTTP фреймворк
- **PostgreSQL 15** - надежная база данных
- **JWT** - безопасная аутентификация
- **Docker** - контейнеризация
- **pgx** - быстрый PostgreSQL драйвер

### Frontend
- **HTML5 + CSS3 + JavaScript** - современный веб-интерфейс
- **Responsive Design** - адаптивный дизайн
- **Chart.js** - интерактивные графики
- **Font Awesome** - иконки

### DevOps
- **Docker Compose** - оркестрация контейнеров
- **Health Checks** - мониторинг состояния
- **Multi-stage builds** - оптимизированные образы

## 🚀 Запуск проекта

### 📋 Требования
- **Docker** и **Docker Compose** (рекомендуется)
- **Go 1.23+** (для локальной разработки)
- **PostgreSQL 15+** (для локального запуска)

### 🐳 Быстрый старт (Docker) - Рекомендуется

1. **Клонируйте репозиторий**
```bash
git clone <repository-url>
cd finance-tracker
```

2. **Запустите все сервисы**
```bash
docker compose up -d --build
```

3. **Проверьте статус**
```bash
docker compose ps
```

4. **Доступ к приложению**
- 🌐 **API**: `http://localhost:8080/api/v1`
- 🏥 **Health Check**: `http://localhost:8080/api/v1/health`
- 💻 **Frontend**: Откройте `web/index.html` в браузере
- 🗄️ **Database**: `localhost:5432`

### 💻 Локальный запуск (без Docker)

1. **Установите PostgreSQL**
```bash
# Ubuntu/Debian
sudo apt-get install postgresql postgresql-contrib

# macOS
brew install postgresql

# Windows
# Скачайте с https://www.postgresql.org/download/windows/
```

2. **Создайте базу данных**
```sql
CREATE DATABASE transactions;
CREATE USER postgres WITH PASSWORD 'your_password';
GRANT ALL PRIVILEGES ON DATABASE transactions TO postgres;
```

3. **Настройте переменные окружения**

**Windows (PowerShell):**
```powershell
$env:DATABASE_URL = "host=localhost port=5432 user=postgres password=fakha dbname=transactions sslmode=disable"
$env:JWT_SECRET = "your-super-secret-jwt-key-change-in-production"
$env:GIN_MODE = "debug"
$env:PORT = "8080"
```

**Linux/macOS:**
```bash
export DATABASE_URL="host=localhost port=5432 user=postgres password=fakha dbname=transactions sslmode=disable"
export JWT_SECRET="your-super-secret-jwt-key-change-in-production"
export GIN_MODE="debug"
export PORT="8080"
```

4. **Примените миграции**
```bash
# Выполните SQL файлы в порядке:
# 1. migrations/001_init.sql
# 2. migrations/002_currencies_accounts.up.sql
```

5. **Запустите сервер**
```bash
go run ./cmd/server
```

6. **Откройте фронтенд**
Откройте `web/index.html` в браузере

### Переменные окружения
- `PORT` (по умолчанию `8080`)
- `DATABASE_URL` (пример в `docker-compose.yml`)
- `JWT_SECRET` (обязательно переопределить в проде!)
- `GIN_MODE` (`debug`/`release`)

### Миграции: up / down
В проекте используется начальная миграция `migrations/001_init.sql` и обратная `migrations/001_init.down.sql`.

- Применить up (в Docker — выполняется автоматически при первом старте благодаря монтированию `./migrations` в Postgres init):
  - вручную: выполнить SQL из `001_init.sql`.
- Откатить (down):
  - выполнить SQL из `001_init.down.sql`, который удаляет индексы и таблицы в обратном порядке.

Важно: откат удалит все данные. Для управляемых миграций рекомендуем использовать инструмент (например, `golang-migrate`), но текущие файлы совместимы и с ручным применением.

## 📚 API Документация

### 🔐 Аутентификация
Все защищенные эндпоинты требуют заголовок:
```
Authorization: Bearer <jwt_token>
```

### 👤 Пользователи
- `POST /api/v1/register` - Регистрация
- `POST /api/v1/login` - Вход
- `POST /api/v1/logout` - Выход
- `GET /api/v1/user/profile` - Профиль пользователя
- `GET /api/v1/user/profile-with-accounts` - Профиль с счетами
- `PUT /api/v1/user/default-currency` - Установка валюты по умолчанию

### 💰 Валюты
- `GET /api/v1/currencies` - Список валют
- `GET /api/v1/currencies/:id` - Валюта по ID

### 🏦 Счета
- `GET /api/v1/accounts` - Счета пользователя
- `POST /api/v1/accounts` - Создание счета
- `GET /api/v1/accounts/:id` - Счет по ID
- `PUT /api/v1/accounts/:id/default` - Установка счета по умолчанию

### 📂 Категории
- `GET /api/v1/categories` - Категории пользователя
- `POST /api/v1/categories` - Создание категории

### 💸 Транзакции
- `GET /api/v1/transactions` - Транзакции пользователя
- `POST /api/v1/transactions` - Создание транзакции
- `GET /api/v1/transactions/:id` - Транзакция по ID
- `GET /api/v1/transactions/period` - Транзакции по периоду
- `GET /api/v1/transactions/summary` - Сводка транзакций
- `GET /api/v1/transactions/by-category` - Статистика по категориям
- `GET /api/v1/transactions/monthly-summary` - Месячные сводки

### 🔄 Обмен валют
- `GET /api/v1/exchange/rates` - Курсы валют
- `POST /api/v1/exchange/rates/update` - Обновление курсов
- `POST /api/v1/exchange/convert` - Конвертация валют
- `GET /api/v1/exchange/balances` - Балансы пользователя
- `GET /api/v1/exchange/rate/:base/:target` - Курс между валютами

### 🏥 Система
- `GET /api/v1/health` - Проверка состояния

## 🗄️ База данных

### Структура таблиц
- **users** - Пользователи
- **currencies** - Валюты
- **accounts** - Счета пользователей
- **categories** - Категории транзакций
- **transactions** - Транзакции
- **exchange_rates** - Курсы валют
- **sessions** - Сессии пользователей

### Миграции
- `001_init.sql` - Базовая структура (пользователи, категории, транзакции)
- `002_currencies_accounts.up.sql` - Валюты и счета
- `002_currencies_accounts.down.sql` - Откат валют и счетов

## 🎨 Frontend

### Функциональность
- 🔐 **Авторизация**: логин/регистрация, хранение токена в `localStorage`
- 📊 **Дашборд**: общая статистика и графики
- 💰 **Счета**: управление мультивалютными счетами
- 📂 **Категории**: настройка категорий доходов/расходов
- 💸 **Транзакции**: добавление, просмотр, фильтрация
- 📈 **Аналитика**: статистика по периодам и категориям
- 🔄 **Обмен валют**: конвертация и обновление курсов

### Технологии
- **HTML5** - семантическая разметка
- **CSS3** - современные стили и анимации
- **JavaScript ES6+** - интерактивность
- **Chart.js** - интерактивные графики
- **Font Awesome** - иконки

## 🛠️ Разработка

### Структура проекта
```
finance-tracker/
├── cmd/server/          # Точка входа приложения
├── internal/            # Внутренний код приложения
│   ├── config/         # Конфигурация
│   ├── handler/        # HTTP обработчики
│   ├── middleware/     # Middleware
│   ├── models/         # Модели данных
│   ├── repository/     # Слой данных
│   ├── service/        # Бизнес-логика
│   └── utils/          # Утилиты
├── migrations/         # SQL миграции
├── web/               # Frontend файлы
├── docker-compose.yml # Docker конфигурация
├── Dockerfile         # Docker образ
└── README.md          # Документация
```

### Команды для разработки
```bash
# Запуск в режиме разработки
docker compose up -d

# Просмотр логов
docker compose logs -f app

# Пересборка после изменений
docker compose up -d --build

# Остановка
docker compose down

# Очистка данных
docker compose down -v
```

### Тестирование
```bash
# Проверка здоровья сервисов
curl http://localhost:8080/api/v1/health

# Тест регистрации
curl -X POST http://localhost:8080/api/v1/register \
  -H "Content-Type: application/json" \
  -d '{"username":"test","email":"test@example.com","password":"password123"}'
```

## 🔧 Конфигурация

### Переменные окружения
| Переменная | Описание | По умолчанию |
|------------|----------|--------------|
| `PORT` | Порт сервера | `8080` |
| `DATABASE_URL` | URL базы данных | `host=localhost port=5432 user=postgres password=fakha dbname=transactions sslmode=disable` |
| `JWT_SECRET` | Секретный ключ JWT | `your-super-secret-jwt-key-change-in-production` |
| `GIN_MODE` | Режим Gin | `debug` |
| `EXCHANGE_API_ENDPOINT` | API курсов валют | `https://api.exchangerate-api.com/v4/latest/USD` |

### Безопасность
- 🔐 JWT токены с истечением срока действия
- 🛡️ Хеширование паролей с bcrypt
- 🔒 CORS настройки для фронтенда
- 🚫 Защита от SQL инъекций через параметризованные запросы

## 📈 Производительность

### Оптимизации
- ⚡ Многоэтапная сборка Docker для минимального размера образа
- 🗄️ Индексы базы данных для быстрых запросов
- 🔄 Connection pooling для PostgreSQL
- 📦 Статическая компиляция Go приложения

### Мониторинг
- 🏥 Health checks для Docker контейнеров
- 📊 Логирование всех операций
- 🔍 Готовность к интеграции с системами мониторинга

## 🤝 Вклад в проект

1. Fork репозитория
2. Создайте feature branch (`git checkout -b feature/amazing-feature`)
3. Commit изменения (`git commit -m 'Add amazing feature'`)
4. Push в branch (`git push origin feature/amazing-feature`)
5. Откройте Pull Request

## 📄 Лицензия

Этот проект распространяется под лицензией MIT. См. файл `LICENSE` для подробностей.

## 🆘 Поддержка

Если у вас возникли вопросы или проблемы:

1. Проверьте [Issues](../../issues) на GitHub
2. Создайте новый Issue с подробным описанием
3. Убедитесь, что Docker и все зависимости установлены правильно

---

**Сделано с ❤️ для управления личными финансами**

## 🚀 Быстрый старт

- Склонить репозиторий и перейти в папку проекта
- Запустить через Docker Compose:

```bash
docker compose up -d --build
curl http://localhost:8080/api/v1/health
```

- Открыть фронтенд: файл `web/index.html` в браузере
  - Локально `web/script.js` автоматически ставит `API_BASE = http://localhost:8080/api/v1`

## 🧰 Локальный запуск без Docker

Требуется установленный Go и PostgreSQL.

1) Настроить БД и переменные окружения (см. `env.example`):

```bash
export DATABASE_URL="host=localhost port=5432 user=postgres password=fakha dbname=transactions sslmode=disable"
export JWT_SECRET="change-me-in-prod"
export PORT=8080
```

2) Применить миграции (вручную выполнить SQL из `migrations/`)

3) Запустить сервер:

```bash
go run ./cmd/server
```

4) Открыть `web/index.html` в браузере

## 🐳 Docker команды (кратко)

```bash
docker compose up -d            # старт
docker compose up -d --build    # пересборка и старт
docker compose logs -f app      # логи приложения
docker compose down             # стоп
docker compose down -v          # стоп и очистка volume'ов
```

## 🌐 Публикация на GitHub

1) Инициализация и первый коммит

```bash
git init
git add .
git commit -m "feat: initial release of WealthFlow Pro"
```

2) Создать пустой репозиторий на GitHub и привязать origin

```bash
git remote add origin https://github.com/<username>/<repo>.git
git branch -M main
git push -u origin main
```

### 🔒 Рекомендации по безопасности репозитория

- Не коммитьте реальные секреты. Используйте `env.example`, а `.env` держите локально
- Добавьте в `.gitignore` бинарники и локальные файлы (например, `*.exe`, `app`, `.env`)
- Поменяйте `JWT_SECRET` в проде и храните его в секретах CI/CD/хостинга

