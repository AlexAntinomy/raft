# Этап сборки
FROM golang:1.21-alpine AS builder

# Установка зависимостей
RUN apk add --no-cache git make

# Копирование исходного кода
WORKDIR /app
COPY . .

# Сборка приложения
RUN make build

# Этап запуска
FROM alpine:latest

# Установка зависимостей времени выполнения
RUN apk --no-cache add ca-certificates tzdata

# Копирование бинарника и конфигов
COPY --from=builder /app/bin/raft-kv-store /usr/local/bin/raft-kv-store
COPY --from=builder /app/configs /etc/raft-kv-store/configs

# Настройка рабочей директории
WORKDIR /var/lib/raft-kv-store

# Порт для HTTP API и gRPC
EXPOSE 8080 9090

# Точка входа
ENTRYPOINT ["raft-kv-store"]
CMD ["--config", "/etc/raft-kv-store/configs/defaults.toml"]