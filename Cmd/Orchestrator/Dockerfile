# Используем официальный образ Go
FROM golang:1.23

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем go.mod для установки зависимостей
COPY go.mod ./

# Копируем исходный код
COPY . .

# Собираем приложение
RUN go build -o orchestrator ./cmd/orchestrator

# Устанавливаем переменные окружения по умолчанию
ENV TIME_ADDITION_MS=100
ENV TIME_SUBTRACTION_MS=100
ENV TIME_MULTIPLICATIONS_MS=200
ENV TIME_DIVISIONS_MS=200

# Открываем порт для HTTP-запросов
EXPOSE 8080

# Запускаем оркестратор
CMD ["./orchestrator"]