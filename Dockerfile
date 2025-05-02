FROM golang:1.21-alpine

WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o main ./cmd/api

EXPOSE 8080
CMD ["./main"]



# # Используем официальный образ Go
# FROM golang:1.21 as builder

# # Устанавливаем рабочую директорию внутри контейнера
# WORKDIR /app

# # Копируем go.mod и go.sum для установки зависимостей
# COPY go.mod go.sum ./
# RUN go mod download

# # Копируем оставшиеся файлы проекта
# COPY . .

# # Сборка приложения
# RUN go build -o server ./cmd/main.go

# # --- Фаза выполнения ---
# FROM debian:bullseye-slim

# # Создаем директорию приложения
# WORKDIR /app

# # Копируем бинарник из builder-слоя
# COPY --from=builder /app/server .

# # Копируем .env, если он нужен в контейнере (например, для локальных тестов)
# COPY .env .env

# # Устанавливаем переменные окружения (можно переопределять через Docker Compose)
# ENV PORT=8080

# # Открываем порт
# EXPOSE 8080

# # Запускаем сервер
# CMD ["./server"]
