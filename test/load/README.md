# Load Testing

Нагрузочное тестирование сервиса с использованием k6.

## Требования

- [k6](https://k6.io/docs/getting-started/installation/) установлен
- Сервис запущен и доступен на `http://localhost:8080`

## Установка k6

### Linux
```bash
sudo gpg -k
sudo gpg --no-default-keyring --keyring /usr/share/keyrings/k6-archive-keyring.gpg --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys C5AD17C747E3415A3642D57D77C6C491D6AC1D69
echo "deb [signed-by=/usr/share/keyrings/k6-archive-keyring.gpg] https://dl.k6.io/deb stable main" | sudo tee /etc/apt/sources.list.d/k6.list
sudo apt-get update
sudo apt-get install k6
```

### macOS
```bash
brew install k6
```

### Базовый тест

**Из корня проекта:**
```bash
cd /home/emir/GolandProjects/avito-tech-internship
k6 run test/load/k6_test.js
```