---
title: "Устранение неполадок"
description: "Диагностика и устранение типичных проблем Stronghold: ошибки доступа и TLS, запечатанные узлы, потеря кворума Raft, медленные ответы, аудит, аутентификация и поды в DKP."
weight: 30
---

На странице собраны типичные проблемы эксплуатации Stronghold, способы их диагностики и устранения. Перед началом диагностики проверьте общее состояние узла и кластера:

```shell
d8 stronghold status
d8 stronghold operator raft list-peers
```

## Permission denied

Симптом — ошибка `permission denied` (HTTP-код `403`) при обращении к пути.

Диагностика:

1. Проверьте, какие политики привязаны к токену:

   ```shell
   d8 stronghold token lookup
   ```

1. Проверьте возможности (capabilities) токена на конкретном пути:

   ```shell
   d8 stronghold token capabilities secret/data/app/config
   ```

   Для чужого токена передайте его явно или используйте accessor через эндпоинт [`/sys/capabilities-accessor`](../../../reference/api/system/#post-syscapabilities-accessor):

   ```shell
   d8 stronghold token capabilities <TOKEN> secret/data/app/config
   ```

1. Просмотрите содержимое политик:

   ```shell
   d8 stronghold policy read <POLICY_NAME>
   ```

Типичные причины:

- для KV v2 политика указана без сегмента `data/` или `metadata/` (например, `secret/app/*` вместо `secret/data/app/*`), см. [KV v2](../../../user/secrets-engines/kv/kv-v2/);
- путь закрыт явным правилом `deny`, которое имеет приоритет;
- запрос выполняется в другом [пространстве имён](../../namespaces/overview/);
- у токена истёк срок действия — `d8 stronghold token lookup` вернёт ошибку.

Подробнее о синтаксисе политик — в разделе [«Политики»](../../../concepts/policy/).

## Ошибки x509

Симптомы — ошибки `x509: certificate signed by unknown authority`, `x509: certificate is valid for ..., not ...`, `x509: certificate has expired`.

| Ошибка | Причина | Решение |
| --- | --- | --- |
| `certificate signed by unknown authority` | Клиент не доверяет CA, выпустившему сертификат сервера | Укажите CA через `STRONGHOLD_CACERT` или флаг `-ca-cert` |
| `certificate is valid for X, not Y` | Адрес в `STRONGHOLD_ADDR` отсутствует в SAN сертификата | Используйте имя из SAN или перевыпустите сертификат с нужными SAN |
| `certificate has expired` | Истёк срок действия сертификата слушателя | Замените сертификат по инструкции [«TLS-сертификаты»](../tls-certificates/) |

Проверить сертификат, который отдаёт сервер:

```shell
openssl s_client -connect stronghold.example.com:8200 -showcerts </dev/null \
  | openssl x509 -noout -subject -issuer -dates -ext subjectAltName
```

Для межузлового взаимодействия Raft (`retry_join`) те же ошибки появляются в журналах standby-узлов. Проверьте параметры `leader_ca_cert_file`, `leader_client_cert_file` и `leader_client_key_file` (см. [«Установка»](../../../install/standalone/installation/)).

## Узел запечатан после перезапуска

При использовании seal на основе Shamir каждый узел после перезапуска запускается запечатанным — это штатное поведение (см. [«Seal»](../../../concepts/seal/)).

1. Проверьте состояние:

   ```shell
   d8 stronghold status
   ```

1. Распечатайте узел, введя пороговое количество ключей:

   ```shell
   d8 stronghold operator unseal
   ```

Если настроено автоматическое распечатывание ([KMS или HSM](../../kms-hsm/hsm/)), а узел остался запечатанным, проверьте в [серверных журналах](../logs/) ошибки обращения к KMS/HSM: доступность сервиса, учётные данные, PIN или слот HSM.

В DP модуль в режиме `Automatic` инициализирует кластер с одной долей ключа и порогом 1. Ключ распечатывания и root-токен хранятся в секрете `stronghold-keys` (`unsealKey`, `rootToken`) в пространстве имён `d8-stronghold`. Узлы распечатывает Deployment `stronghold-automatic` (период проверки по умолчанию — 60 секунд). Если узел остаётся запечатанным, проверьте наличие секрета и журналы распечатывателя:

```shell
d8 k -n d8-stronghold get secret stronghold-keys
d8 k -n d8-stronghold logs deploy/stronghold-automatic
```

В режиме `Manual` нет ни секрета, ни распечатывателя: распечатывайте узлы вручную.

## Потеря кворума Raft

Симптом — ошибка `local node not active but active cluster node not found`, кластер не выбирает лидера.

Кворум требует доступности большинства голосующих узлов: для 3 узлов — 2, для 5 узлов — 3. Если большинство узлов потеряно безвозвратно, выполните восстановление по инструкции:

- [Восстановление после потери кворума в Linux](../../../install/standalone/raft-lost-quorum-recovery/);
- [Восстановление после потери кворума в DP](../../../install/dkp/raft-lost-quorum-recovery/).

Перед восстановлением убедитесь, что недоступные узлы действительно не вернутся: запуск старых узлов после восстановления может привести к расхождению данных.

## Нестабильные выборы лидера

Симптомы — в журналах часто появляются сообщения о переходе в состояние кандидата, клиенты получают кратковременные ошибки, активный узел меняется.

Диагностика:

- проверьте метрики `stronghold_raft_leader_lastContact` и `stronghold_raft_state_candidate` (см. [«Мониторинг»](../monitoring/));
- проверьте сетевую задержку и потери пакетов между узлами на порту кластера (`cluster_addr`, по умолчанию `8201`);
- проверьте задержку диска с данными Raft (`iostat -x 1`), см. [«Планирование ресурсов»](../sizing/);
- проверьте синхронизацию времени (NTP) на всех узлах;
- проверьте состояние Autopilot:

  ```shell
  d8 stronghold operator raft autopilot state
  ```

Типичные причины — перегруженный или сетевой диск с высокой задержкой, нехватка CPU при всплесках нагрузки, размещение узлов в разных площадках с большой задержкой, паузы процесса из-за нехватки памяти.

## Медленные ответы и рост числа аренд

Симптомы — растёт время ответа, увеличивается потребление памяти и размер хранилища, медленно проходит запуск и смена лидера.

Частая причина — неконтролируемый рост количества аренд и токенов: клиенты заново аутентифицируются на каждый запрос вместо переиспользования токена, либо TTL аренд слишком большой.

Диагностика:

1. Проверьте общее количество аренд через эндпоинт [`/sys/leases/count`](../../../reference/api/system/#get-sysleasescount) и метрику `stronghold_expire_num_leases`.
1. Найдите префиксы с наибольшим количеством аренд:

   ```shell
   d8 stronghold list sys/leases/lookup/auth/approle/login
   ```

1. Проверьте TTL в настройках методов аутентификации и механизмов секретов:

   ```shell
   d8 stronghold read sys/auth/approle/tune
   ```

Устранение:

- уменьшите `default_lease_ttl` и `max_lease_ttl` для проблемных точек монтирования;
- настройте клиентов на переиспользование и продление токенов (например, через [Stronghold Agent](../../../user/agent/overview/));
- отзовите лишние аренды по префиксу: `d8 stronghold lease revoke -prefix <PATH>`;
- ограничьте нагрузку [квотами](../quotas/).

## Аудит-устройство блокирует запросы

Если ни одно аудит-устройство не может записать событие, Stronghold отклоняет запрос или запросы зависают до устранения проблемы (см. [«Аудит в Stronghold»](../../audit/overview/)).

Диагностика:

1. Проверьте список аудит-устройств:

   ```shell
   d8 stronghold audit list -detailed
   ```

1. Для устройства `file` проверьте свободное место и права на запись в каталог журнала.
1. Для устройств `syslog` и `socket` проверьте доступность приёмника.
1. Проверьте метрики `stronghold_audit_log_request_failure` и `stronghold_audit_log_response_failure`.

Устранение — освободите место или восстановите приёмник. Держите не менее двух аудит-устройств, чтобы отказ одного не останавливал обслуживание запросов.

## Проблемы входа

### OIDC

- **Ошибка `redirect_uri` или `invalid redirect`**: URI перенаправления должен совпадать в параметре `allowed_redirect_uris` роли и в настройках провайдера. Для CLI используется `http://localhost:8250/oidc/callback`, для веб-интерфейса — адрес вида `https://<STRONGHOLD_ADDR>/ui/stronghold/auth/oidc/oidc/callback`. Подробнее — в разделе [«OIDC»](../../../user/auth/oidc/).
- **Ошибки проверки токена провайдера**: проверьте `oidc_discovery_url`, `oidc_client_id`, а при собственном CA — `oidc_discovery_ca_pem`.
- **Нет прав после входа**: проверьте `bound_claims`, `groups_claim` и сопоставление групп с политиками.

### LDAP

- **Ошибка привязки (bind)**: проверьте `binddn` и `bindpass`, а также доступность сервера из всех узлов Stronghold.
- **Ошибки TLS**: при использовании `ldaps://` или `starttls` укажите CA сервера в параметре `certificate`; не используйте `insecure_tls` в production.
- **Пользователь входит, но не получает политики**: проверьте `groupdn`, `groupfilter` и `groupattr`.

Подробнее — в разделе [«LDAP»](../../../user/auth/ldap/).

## Проблемы подов в DP

Проверьте состояние подов и события:

```shell
d8 k -n d8-stronghold get pods -o wide
d8 k -n d8-stronghold describe pod stronghold-0
d8 k -n d8-stronghold get events --sort-by=.lastTimestamp
```

| Симптом | Что проверить |
| --- | --- |
| Под в состоянии `ContainerCreating`, нет секрета `ingress-tls` (`ingress-tls-customcertificate` в режиме `CustomCertificate`) | В режиме `CertManager` — выпуск сертификата cert-manager (Certificate `stronghold` существует только при инлете `Ingress`); в режиме `CustomCertificate` — секрет в `d8-system`. См. [«Настройка Stronghold»](../../../install/dkp/configuration/#устранение-неполадок) |
| Под в состоянии `CrashLoopBackOff` | Журналы текущего и предыдущего запуска: `d8 k -n d8-stronghold logs stronghold-0 --previous` |
| Под в состоянии `Pending` | Доступность master-узлов и ресурсов на них |
| Под работает, но узел запечатан | Наличие секрета `stronghold-keys` и журналы `stronghold-automatic` (только режим `Automatic`) |

Состояние модуля:

```shell
d8 k get module stronghold
```

<!-- TODO(verify): команда проверки статуса модуля (d8 k get module stronghold) и поля статуса. -->

## Сбор диагностических данных

### Журналы уровня debug

Временно повысьте уровень логирования без перезапуска через эндпоинт `/sys/loggers` (см. [«Серверные журналы»](../logs/#изменение-уровня-логирования)) и верните его после сбора данных.

### Пакет debug

Команда `d8 stronghold debug` собирает в архив состояние узла, метрики, профили pprof и журналы за заданный период:

```shell
d8 stronghold debug -duration=5m -interval=30s -output=stronghold-debug.tar.gz
```

Токен должен иметь права на чтение `sys/metrics`, `sys/pprof/*`, `sys/host-info`, `sys/in-flight-req` и `sys/monitor`. Передавайте полученный архив в поддержку по защищённому каналу: он содержит сведения о конфигурации и структуре кластера.
