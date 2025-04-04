# gRPC File Server на Go

Проект представляет собой файловый сервер, реализованный с использованием gRPC. Сервис позволяет:

- Загружать файлы (Upload)
- Скачивать файлы (Download)
- Получать список всех файлов с датами (List)

Поддерживается ограничение на количество одновременных подключений, логирование через `slog`, а также интерфейс командной строки для клиента.

## Пример использования клиента

### Загрузка файла
```bash
$ go run ./client --upload ./assets/test.txt
```

### Скачивание файла
```bash
$ go run ./client --download test.txt --out ./downloads/test_copy.txt
```

### Получить список файлов
```bash
$ go run ./client --list
```

## Стек технологий

- Язык: Go 1.24.1
- gRPC: Google Protocol Buffers + `grpc-go`
- Логирование: `slog`
- Ограничения: Semaphore-based лимитер
- Файлы: Локальное хранилище (filesystem)

## Возможности для доработки

- Добавить TLS (шифрование трафика)
- Аутентификация пользователей
- Docker-образ и Compose
- Unit-тесты и CI
