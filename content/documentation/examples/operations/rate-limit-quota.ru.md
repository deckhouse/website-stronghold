---
title: "Ограничение частоты запросов (rate limit)"
linkTitle: "Rate limit"
description: "Защита Stronghold от перегрузки одним клиентом: квота rate limit на путь, проверка ответа 429, просмотр и удаление квоты."
weight: 130
params:
  relatedLinks:
    - title: "Квоты"
      url: ../../../admin/operations/quotas/
    - title: "API квот"
      url: ../../../reference/api/system/
---

Один неисправный или слишком активный клиент может занять все ресурсы кластера. Квота rate limit ограничивает число запросов в единицу времени, и лишние запросы получают ответ `429 Too Many Requests`.

## Цель

Ограничить частоту запросов к механизму секретов `secret/` и убедиться, что превышение лимита приводит к ответу `429`.

## Предварительные требования

- Токен Stronghold с правами на `sys/quotas/rate-limit/*`.
- Утилита `curl`.

## Шаг 1. Создайте квоту

```bash
d8 stronghold write sys/quotas/rate-limit/secret-limit \
  path="secret/" \
  rate=5 \
  interval=60
```

Параметры квоты:

| Параметр | Назначение |
| --- | --- |
| `path` | Путь, на который действует квота. Для пути монтирования квота охватывает все запросы к нему. Если параметр не указан, квота действует глобально |
| `rate` | Допустимое число запросов за `interval` |
| `interval` | Длина интервала в секундах |
| `block_interval` | Сколько секунд клиент блокируется после превышения лимита. По умолчанию блокировки нет |
| `role` | Для путей методов аутентификации: роль, на которую действует квота |

Квота считается отдельно для каждого IP-адреса клиента.

## Шаг 2. Проверьте, что лимит работает

Отправьте десять запросов подряд:

```bash
for i in $(seq 1 10); do
  curl -s -o /dev/null -w "%{http_code}\n" \
    -H "X-Vault-Token: $STRONGHOLD_TOKEN" \
    "$STRONGHOLD_ADDR/v1/secret/data/probe"
done | sort | uniq -c
```

Первые пять запросов проходят (ответ `404`, так как секрета `probe` нет, или `200`), остальные получают `429`:

```text
      5 404
      5 429
```

Запросы к другим путям, например `sys/health`, на квоту не влияют.

## Шаг 3. Просмотрите и измените квоту

```bash
d8 stronghold list sys/quotas/rate-limit
d8 stronghold read sys/quotas/rate-limit/secret-limit
d8 stronghold write sys/quotas/rate-limit/secret-limit path="secret/" rate=100 interval=1
```

После изменения квота действует сразу.

## Проверка

```bash
d8 stronghold read -field=rate sys/quotas/rate-limit/secret-limit
```

Команда возвращает `100`.

## Очистка

```bash
d8 stronghold delete sys/quotas/rate-limit/secret-limit
```
