```markdown
# Raft Key-Value Store

[![Go Report Card](https://goreportcard.com/badge/github.com/yourusername/raft-kv-store)](https://goreportcard.com/report/github.com/yourusername/raft-kv-store)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Распределенное хранилище ключ-значение с консенсусом Raft. Учебный проект для демонстрации работы распределенных систем.

## Особенности

- 🚀 Реализация алгоритма Raft (Leader Election, Log Replication)
- 🔄 Автоматическое восстановление после сбоев
- 📦 Поддержка снапшотов состояния
- 📊 Мониторинг через Prometheus/Grafana
- 🐳 Готовые Docker образы и docker-compose конфигурации

## Технологии

![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)
![gRPC](https://img.shields.io/badge/gRPC-1.50+-000?logo=grpc)
![Docker](https://img.shields.io/badge/Docker-20.10+-2496ED?logo=docker)
![Prometheus](https://img.shields.io/badge/Prometheus-2.30+-E6522C?logo=prometheus)

## Быстрый старт

### Требования
- Go 1.21+
- Docker 20.10+
- protoc 3.20+

```bash
# Клонирование репозитория
git clone https://github.com/AlexAntinomy/raft
cd raft-kv-store

# Сборка и запуск кластера
docker-compose up --build
```

## Использование

Пример работы через HTTP API:
```bash
# Запись значения
curl -X PUT "http://localhost:8080/key/foo?value=bar"

# Чтение значения
curl "http://localhost:8080/key/foo"

# Статус кластера
curl "http://localhost:8080/cluster/status"
```

## Конфигурация

Основные параметры в `configs/cluster.yaml`:
```yaml
cluster:
  nodes:
    - id: 1
      address: "node1:9090"
      http_port: 8080

replication:
  heartbeat_interval: "500ms"
  election_timeout_min: "1500ms"
```

## Мониторинг

Grafana Dashboard:
![Raft Dashboard](docs/images/grafana-dashboard.png)

Запуск мониторинга:
```bash
docker-compose -f docker-compose-monitoring.yml up
```

## Тестирование

Интеграционные тесты:
```bash
go test -v ./tests/integration
```

Нагрузочное тестирование:
```bash
go test -v ./tests -run TestLoad
```

## Развертывание в AWS

Используйте скрипт для создания кластера:
```bash
./scripts/deploy-cluster.sh
```

## Лицензия

MIT License. Подробнее в файле [LICENSE](LICENSE).

---

**Совместимость**: Go 1.21+, Linux/macOS  
**Поддержка**: [Сообщить о проблеме](https://github.com/yourusername/raft-kv-store/issues)
```