#!/bin/bash

# --- Настройки ---
BINARY_NAME="./rsshub"
TEST_FEED_NAME="test-feed-$(date +%s)"
TEST_FEED_URL="https://news.ycombinator.com/rss"

# Цвета для вывода
GREEN="\033[0;32m"
RED="\033[0;31m"
YELLOW="\033[0;33m"
NC="\033[0m" # No Color

# --- Хелперы ---
info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

fail() {
    echo -e "${RED}[FAIL]${NC} $1"
    exit 1
}

check_binary() {
    if [ ! -f "$BINARY_NAME" ]; then
        fail "Бинарный файл '$BINARY_NAME' не найден. Сначала скомпилируйте проект (go build -o rsshub ./cmd/cli/main.go)"
    fi
}

cleanup() {
    warn "--- Зачистка ---"
    if [ ! -z "$FETCH_PID" ]; then
        info "Останавливаем фоновый процесс (PID: $FETCH_PID)..."
        kill $FETCH_PID
        # Даем ему время завершиться
        sleep 2
    fi
    info "Удаляем тестовый фид '$TEST_FEED_NAME'..."
    $BINARY_NAME delete --name "$TEST_FEED_NAME"
    info "Зачистка завершена."
}

# Перехватываем выход (Ctrl+C или exit) для зачистки
trap cleanup EXIT

# --- Тесты ---
check_binary
info "Начало E2E тестирования..."

# 1. Справка
info "Тест 1: Проверка --help..."
$BINARY_NAME --help > /dev/null || fail "Команда --help завершилась с ошибкой"

# 2. Добавление фида
info "Тест 2: Добавление фида '$TEST_FEED_NAME'..."
$BINARY_NAME add --name "$TEST_FEED_NAME" --url "$TEST_FEED_URL" || fail "Не удалось добавить фид"

# 3. Листинг фида
info "Тест 3: Проверка листинга фида..."
$BINARY_NAME list | grep "$TEST_FEED_NAME" > /dev/null || fail "Добавленный фид не найден в 'list'"
info "Фид '$TEST_FEED_NAME' успешно добавлен и найден."

# 4. Запуск fetch в фоне
info "Тест 4: Запуск 'fetch' в фоновом режиме..."
# Убедимся, что миграции запущены (на всякий случай)
$BINARY_NAME migrate-up > /dev/null
# Запускаем в фоне и сохраняем PID
$BINARY_NAME fetch &
FETCH_PID=$!
info "Процесс 'fetch' запущен с PID: $FETCH_PID"
# Даем ему время на запуск и захват блокировки
sleep 3

# 5. Проверка блокировки (второй запуск fetch)
info "Тест 5: Проверка блокировки (второй 'fetch' должен завершиться)..."
# Этот тест будет "подвисать" до фикса бага из пункта 1
OUTPUT=$($BINARY_NAME fetch)
if ! echo "$OUTPUT" | grep -q "Background process is already running"; then
    fail "Второй 'fetch' не сообщил, что процесс уже запущен."
fi
info "Блокировка работает: второй 'fetch' корректно сообщил и завершился."

# 6. Тестирование IPC (set-interval)
info "Тест 6: Проверка 'set-interval'..."
$BINARY_NAME set-interval --duration 10m || fail "Команда 'set-interval' не удалась"
info "Команда 'set-interval' отправлена."
# В реальном тесте нужно было бы проверить лог 'fetch' (PID: $FETCH_PID)
# Но мы просто проверяем, что команда выполнилась

# 7. Тестирование IPC (set-workers)
info "Тест 7: Проверка 'set-workers'..."
$BINARY_NAME set-workers --count 10 || fail "Команда 'set-workers' не удалась"
info "Команда 'set-workers' отправлена."

info "----------------"
info "${GREEN}✓ ВСЕ ТЕСТЫ ПРОЙДЕНЫ УСПЕШНО${NC}"
info "----------------"

# cleanup() будет вызван автоматически при выходе