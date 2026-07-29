---
title: "Disaster recovery"
linkTitle: "Disaster recovery"
weight: 30
params:
  edition: ee
description: "Настройка репликации Disaster Recovery, повышение DR-secondary при аварийном переключении и церемония выпуска DR operation token."
---

Репликация Disaster Recovery (DR) поддерживает горячий резервный кластер, который зеркалирует **весь** non-ignored keyspace primary, включая локальные данные, такие как токены и лизы. DR-secondary не обслуживает клиентские запросы (кроме unseal и небольшого внутреннего набора) — он ожидает promote и берёт нагрузку на себя при отказе primary.

## Перед началом

- Убедитесь, что оба кластера работают на Stronghold EE с integrated Raft storage и что репликация включена (смотрите [Обзор](../overview/)).
- Убедитесь, что кластерный порт primary доступен с secondary и что у вас есть CA-сертификат primary для TLS.
- Подготовьте токен с правами на `sys/replication/*` на primary.
- Держите наготове держателей долей unseal- или recovery-ключей — они нужны для церемонии promote.

В примерах ниже `${PRIMARY_ADDR}` и `${SECONDARY_ADDR}` — API-адреса кластеров, а `${VAULT_TOKEN}` — токен с правами на `sys/replication/*`.

## Шаг 1. Включите DR primary

```shell
d8 stronghold write -force sys/replication/dr/primary/enable
```

## Шаг 2. Создайте activation-токен для DR secondary

```shell
d8 stronghold write sys/replication/dr/primary/secondary-token id=dr-1
```

Команда возвращает wrapping-токен в поле `wrap_info.token` — передайте на secondary именно его.

## Шаг 3. Включите DR secondary

```shell
curl \
  --header "X-Vault-Token: ${VAULT_TOKEN}" \
  --request POST \
  --data @dr-secondary-enable.json \
  "${SECONDARY_ADDR}/v1/sys/replication/dr/secondary/enable"
```

Пример `dr-secondary-enable.json`:

```json
{
  "token": "<wrapping_token_from_step_2>",
  "primary_api_addr": "<primary_api_address>",
  "ca_cert": "<primary_ca_certificate_in_pem>"
}
```

Для окружений с самоподписанными сертификатами параметр `ca_cert` (CA primary в формате PEM) обязателен. DR-secondary реплицирует весь keyspace, включая локальные данные, и не обслуживает клиентские запросы.

## Шаг 4. Проверьте статус

```shell
d8 stronghold read -address="${SECONDARY_ADDR}" sys/replication/dr/status
```

## Повышение DR-secondary

Повышение DR-secondary требует **DR operation token**, который выпускается через многошаговую церемонию, аналогичную generate-root: она объединяет доли unseal- или recovery-ключей с одноразовым паролем (OTP).

1. Запустите церемонию. В ответе возвращаются `nonce` и `otp`:

   ```shell
   curl \
     --request PUT \
     --data '{}' \
     "${SECONDARY_ADDR}/v1/sys/replication/dr/secondary/generate-operation-token/attempt"
   ```

1. Внесите долю ключа. Повторите для каждой требуемой доли:

   ```shell
   curl \
     --request PUT \
     --data '{"key":"<unseal_or_recovery_key_share>","nonce":"<nonce>"}' \
     "${SECONDARY_ADDR}/v1/sys/replication/dr/secondary/generate-operation-token/update"
   ```

   Когда `complete` становится `true`, в ответе появляется `encoded_token`. Расшифруйте его с помощью `otp`, чтобы получить DR operation token.

1. Повысьте secondary с помощью DR operation token:

   ```shell
   curl \
     --header "X-Vault-Token: ${VAULT_TOKEN}" \
     --request POST \
     --data '{"dr_operation_token":"<dr_operation_token>"}' \
     "${SECONDARY_ADDR}/v1/sys/replication/dr/secondary/promote"
   ```

После успешного promote бывший secondary становится активным DR primary и обслуживает клиентские запросы.

## Операции управления

| Действие | Эндпоинт |
| --- | --- |
| Статус | `d8 stronghold read sys/replication/dr/status` |
| Отозвать secondary | `d8 stronghold write sys/replication/dr/primary/revoke-secondary id=dr-1` |
| Отключить на primary | `d8 stronghold write -force sys/replication/dr/primary/disable` |
| Понизить primary (demote) | `d8 stronghold write -force sys/replication/dr/primary/demote` |
| Перенаправить secondary | `d8 stronghold write sys/replication/dr/secondary/update-primary token=<activation_token>` |

Примечания:

- `demote` понижает DR primary до отключённого DR secondary, **сохраняя** cluster ID и локальные данные, поэтому он готов к переподключению без обнуления.
- Для контролируемого переключения выполните `demote` на старом primary и `promote` на резерве; используйте `update-primary`, чтобы переподключить старый primary к только что повышенному.
