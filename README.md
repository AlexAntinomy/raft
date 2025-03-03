# Raft Key-Value Store

[![Go Report Card](https://goreportcard.com/badge/github.com/yourusername/raft-kv-store)](https://goreportcard.com/report/github.com/yourusername/raft-kv-store)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Распределенное хранилище ключ-значение с консенсусом Raft. Учебный проект для демонстрации работы распределенных систем.

## Особенности

- 🚀 **Реализация алгоритма Raft** (Leader Election, Log Replication)
- 🔄 **Автоматическое восстановление** после сбоев
- 📦 **Поддержка снапшотов состояния**
- 📊 **Мониторинг** через Prometheus/Grafana
- 🐳 **Готовые Docker образы** и docker-compose конфигурации

## Технологии

![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)
![gRPC](https://img.shields.io/badge/gRPC-1.50+-000?logo=grpc)
![Docker](https://img.shields.io/badge/Docker-20.10+-2496ED?logo=docker)
![Prometheus](https://img.shields.io/badge/Prometheus-2.30+-E6522C?logo=prometheus)

## Быстрый старт

### Требования

- **Go** 1.21+
- **Docker** 20.10+
- **protoc** 3.20+

### Клонирование репозитория и запуск кластера

```bash
# Клонирование репозитория
git clone https://github.com/AlexAntinomy/raft
cd raft-kv-store

# Сборка и запуск кластера
docker-compose up --build
