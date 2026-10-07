---
title: "Квоты"
description: "Ограничение частоты запросов и количества аренд в Stronghold с помощью квот sys/quotas: настройка, примеры и проверка."
weight: 70
---

Квоты защищают Stronghold от перегрузки отдельными клиентами: ошибочно настроенное приложение, которое аутентифицируется на каждый запрос или бесконтрольно создаёт аренды, может снизить доступность всего кластера.

Stronghold поддерживает два типа квот:

- **квоты ограничения частоты запросов** (rate limit) — ограничивают число запросов в единицу времени;
- **квоты количества аренд** (lease count) — ограничивают число одновременно существующих аренд.

Для управления квотами требуется токен с правами на пути `sys/quotas/*`.

## Квоты ограничения частоты запросов

Квота применяется глобально, к пространству имён, к точке монтирования или к роли метода аутентификации. Параметры квоты описаны в справочнике API [`/sys/quotas/rate-limit/{name}`](../../../reference/api/system/#post-sysquotasrate-limitname):

| Параметр | Описание |
| --- | --- |
| `rate` | Максимальное число запросов за интервал. Положительное число |
| `interval` | Интервал, за который считается `rate`. По умолчанию `1s` |
| `block_interval` | Если задан, клиент, превысивший лимит, блокируется на указанное время |
| `path` | Путь точки монтирования или пространства имён. Пустое значение — глобальная квота |
| `role` | Роль метода аутентификации, к которой применяется квота |

При превышении квоты Stronghold возвращает HTTP-код `429`.

### Примеры

Глобальная квота на весь кластер:

```shell
d8 stronghold write sys/quotas/rate-limit/global rate=1000
```

Квота на метод аутентификации AppRole — защита от шторма повторных входов:

```shell
d8 stronghold write sys/quotas/rate-limit/approle-login \
  path=auth/approle \
  rate=50 \
  interval=1s
```

Квота на отдельную роль с блокировкой нарушителя на минуту:

```shell
d8 stronghold write sys/quotas/rate-limit/ci-role \
  path=auth/approle \
  role=ci \
  rate=10 \
  interval=1s \
  block_interval=60s
```

Квота на точку монтирования механизма секретов:

```shell
d8 stronghold write sys/quotas/rate-limit/kv-apps \
  path=secret/ \
  rate=200
```

Квота на пространство имён (Stronghold EE, см. [«Пространства имён»](../../namespaces/overview/)):

```shell
d8 stronghold write sys/quotas/rate-limit/team-a \
  path=team-a/ \
  rate=300
```

Если для запроса подходят несколько квот, применяется наиболее специфичная: роль, затем точка монтирования, затем пространство имён, затем глобальная квота.

### Просмотр и удаление

```shell
d8 stronghold list sys/quotas/rate-limit
d8 stronghold read sys/quotas/rate-limit/approle-login
d8 stronghold delete sys/quotas/rate-limit/approle-login
```

## Общие настройки квот

Эндпоинт [`/sys/quotas/config`](../../../reference/api/system/#post-sysquotasconfig) задаёт общие параметры:

| Параметр | Описание |
| --- | --- |
| `rate_limit_exempt_paths` | Пути, на которые не распространяются квоты ограничения частоты |
| `enable_rate_limit_audit_logging` | Записывать в журнал аудита запросы, отклонённые из-за квоты |
| `enable_rate_limit_response_headers` | Добавлять в ответы HTTP-заголовки с информацией о лимите |

Пример — исключить проверку состояния из квот и включить заголовки ответов:

```shell
d8 stronghold write sys/quotas/config \
  rate_limit_exempt_paths="sys/health" \
  enable_rate_limit_response_headers=true \
  enable_rate_limit_audit_logging=true
```

{{< alert level="info" >}}
Параметр `rate_limit_exempt_paths` заменяет весь список. Перед изменением прочитайте текущие значения командой `d8 stronghold read sys/quotas/config`.
{{< /alert >}}

## Квоты количества аренд

Квоты количества аренд ограничивают число одновременно существующих аренд на точке монтирования, в пространстве имён или для роли. При достижении лимита новые запросы, создающие аренды, отклоняются до истечения или отзыва существующих аренд.

{{< alert level="warning" >}}
Квоты количества аренд входят в Stronghold EE и включаются функцией лицензии. Без неё эндпоинты [`/sys/quotas/lease-count/{name}`](../../../reference/api/system/#post-sysquotaslease-countname) недоступны (в справочнике API они отмечены как `enterprise-stub`). Параметры квоты: `path`, `role` и `max_leases`.
{{< /alert >}}

Пример по аналогии с upstream Vault:

```shell
d8 stronghold write sys/quotas/lease-count/db-creds \
  path=database/ \
  max_leases=5000
```

Для точного подсчёта аренд по ролям не включайте параметр сервера `imprecise_lease_role_tracking` (см. [«Настройка»](../../../install/standalone/configuration/)).

## Подбор значений

1. Соберите фактическую нагрузку по метрикам `stronghold_core_handle_request` и `stronghold_expire_num_leases` (см. [«Мониторинг»](../monitoring/)) за несколько недель.
1. Установите лимиты с запасом относительно пиковых значений.
1. Включите `enable_rate_limit_audit_logging` и проанализируйте отклонённые запросы.
1. Постепенно снижайте лимиты для клиентов с аномальной нагрузкой.
