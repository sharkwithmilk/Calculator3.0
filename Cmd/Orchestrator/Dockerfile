# Используем официальный образ Go
FROM golang:1.23

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем go.mod и go.sum для установки зависимостей
COPY go.mod ./
COPY go.sum ./

# Устанавливаем зависимости
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем приложение
RUN go build -o orchestrator ./Cmd/Orchestrator

# Открываем порт для HTTP-запросов
EXPOSE 8080

# Запускаем оркестратор
CMD ["./orchestrator"]