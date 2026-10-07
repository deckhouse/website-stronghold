---
title: "Мониторинг"
description: "Включение телеметрии Stronghold, сбор метрик в Prometheus, ключевые метрики, примеры правил алертинга и проверка состояния через sys/health."
weight: 10
---

Мониторинг Stronghold строится на трёх источниках данных:

- метрики телеметрии, доступные через эндпоинт `/v1/sys/metrics`;
- эндпоинт проверки состояния `/v1/sys/health`;
- [серверные журналы](../logs/) и [журналы аудита](../../audit/overview/).

## Включение телеметрии

{{< alert level="info" >}}
Этот раздел и следующие до [«Мониторинг в DP»](#мониторинг-в-dp) описывают автономную (standalone) установку. В DP конфигурацией управляет модуль (ConfigMap `stronghold-config`), см. [«Мониторинг в DP»](#мониторинг-в-dp).
{{< /alert >}}

Параметры телеметрии задаются в блоке `telemetry` конфигурационного файла сервера (см. [«Настройка»](../../../install/standalone/configuration/)). Чтобы метрики были доступны в формате Prometheus, задайте время хранения метрик в памяти:

```hcl
telemetry {
  prometheus_retention_time = "24h"
  disable_hostname          = true
}
```

- `prometheus_retention_time` — время, в течение которого метрики хранятся для выдачи в формате Prometheus. Значение `0` отключает Prometheus-формат. По умолчанию `24h`.
- `disable_hostname` — не добавлять имя узла в префикс метрик. Рекомендуется включить, чтобы имена метрик на всех узлах совпадали.
- `metrics_prefix` — префикс имён метрик. По умолчанию метрики имеют префикс `stronghold`; модуль DP явно задаёт то же значение.

После изменения конфигурации перезапустите сервис (standalone):

```shell
systemctl restart stronghold
```

## Доступ к эндпоинту sys/metrics

Эндпоинт [`GET /sys/metrics`](../../../reference/api/system/#get-sysmetrics) возвращает агрегированные метрики. Для формата Prometheus передайте параметр `format=prometheus`:

```shell
curl \
  --header "X-Vault-Token: ${STRONGHOLD_TOKEN}" \
  "${STRONGHOLD_ADDR}/v1/sys/metrics?format=prometheus"
```

### Доступ по токену

Создайте отдельную политику с минимальными правами для сборщика метрик:

```hcl
path "sys/metrics" {
  capabilities = ["read"]
}
```

```shell
d8 stronghold policy write prometheus-metrics prometheus-metrics.hcl
d8 stronghold token create -policy=prometheus-metrics -orphan -period=24h
```

Используйте периодический токен и настройте его продление, либо выпускайте токен через метод аутентификации (например, [Kubernetes](../../../user/auth/kubernetes/) или [AppRole](../../../user/auth/approle/)).

### Доступ без аутентификации

Если сеть, из которой собираются метрики, доверенная, разрешите неаутентифицированный доступ на уровне слушателя:

```hcl
listener "tcp" {
  address       = "0.0.0.0:8200"
  tls_cert_file = "/opt/stronghold/tls/node-1-cert.pem"
  tls_key_file  = "/opt/stronghold/tls/node-1-key.pem"

  telemetry {
    unauthenticated_metrics_access = true
  }
}
```

{{< alert level="warning" >}}
Метрики раскрывают сведения о структуре хранилища (пути монтирования, количество аренд и токенов). Разрешайте доступ без аутентификации только при ограничении сетевого доступа к слушателю.
{{< /alert >}}

### Пример задания для Prometheus

```yaml
scrape_configs:
  - job_name: stronghold
    metrics_path: /v1/sys/metrics
    params:
      format: ["prometheus"]
    scheme: https
    tls_config:
      ca_file: /etc/prometheus/stronghold-ca.pem
    authorization:
      credentials_file: /etc/prometheus/stronghold-token
    static_configs:
      - targets:
          - raft-node-1.demo.tld:8200
          - raft-node-2.demo.tld:8200
          - raft-node-3.demo.tld:8200
```

Собирайте метрики со всех узлов кластера, а не только с адреса балансировщика: часть метрик (Raft, runtime) относится к конкретному узлу.

Каждый узел, включая standby, отдаёт собственные метрики: запрос `sys/metrics` не перенаправляется на активный узел.

## Мониторинг в DP

В DP Stronghold работает в пространстве имён `d8-stronghold`, а его конфигурацией (включая телеметрию) управляет модуль через ConfigMap `stronghold-config`. Метрики собираются следующим образом:

- в настройках телеметрии задан `metrics_prefix = "stronghold"`, поэтому все метрики имеют префикс `stronghold_` (например, `stronghold_core_unsealed`);
- метрики отдаёт loopback-listener `127.0.0.1:8400`, недоступный снаружи Pod'а;
- sidecar `kube-rbac-proxy` публикует их на порту `9889` (`https-metrics`); для доступа требуется право `get` на `statefulsets/prometheus-metrics` (ресурс `stronghold`) в пространстве имён `d8-stronghold`. Модуль выдаёт его ServiceAccount `prometheus` в `d8-monitoring`.

Объекты мониторинга модуль создаёт сам, настраивать ничего не нужно:

- PodMonitor `stronghold` в пространстве имён `d8-monitoring`;
- PrometheusRule с алертами модуля (см. ниже);
- три дашборда Grafana: `stronghold.json`, `replication.json` и `snapshot_auto.json`.

Проверить наличие объекта сбора метрик:

```shell
d8 k -n d8-monitoring get podmonitor stronghold
```

### Алерты модуля в DP

Алерты заданы в PrometheusRule модуля и используют метку `severity_level`, как и остальные алерты DP:

| Алерт | Условие | `for` | `severity_level` |
| --- | --- | --- | --- |
| `D8StrongholdNoReadyPod` | Ни один Pod StatefulSet `stronghold` не находится в состоянии Ready | 3m | 4 |
| `D8StrongholdNoActiveNodes` | `stronghold_core_active` равна `0` на всех узлах (или метрики отсутствуют) | 1m | 3 |
| `D8StrongholdSealedNodesPresent` | Количество узлов, отдающих метрики, больше количества распечатанных узлов | 5m | 7 |
| `D8StrongholdClusterNotHealthy` | Сумма `stronghold_autopilot_healthy` равна `0` | 5m | 7 |
| `D8StrongholdQuorumInCriticalState` | `stronghold_autopilot_failure_tolerance` равна `0`; создаётся только если в кластере больше 2 master-узлов (`clusterMasterCount > 2`) | 3m | 4 |
| `D8StrongholdAbsentMetrics` | Количество готовых контейнеров `kube-rbac-proxy` больше количества узлов, отдающих метрики | 1m | 3 |
| `D8StrongholdAutoSnapshotFailed` | `stronghold_core_snapshot_auto_failed` равна `1` для конфигурации автоматических снимков | 5m | 6 |
| `D8StrongholdAutoSnapshotRotationFailed` | `stronghold_autosnapshots_rotate_failed` равна `1`: ротация старых снимков не удалась после успешного резервного копирования | 15m | 8 |

Собственные алерты в DP задаются ресурсом [CustomPrometheusRules](/modules/prometheus/cr.html#customprometheusrules): перенесите в него содержимое `spec.groups` из примеров ниже.

<!-- TODO(verify): актуальные ссылки на документацию DKP по сбору метрик и CustomPrometheusRules. -->

## Ключевые метрики

Названия метрик приведены с префиксом `stronghold_`: по умолчанию Stronghold использует его и в standalone-установке, и в DP. Префикс можно изменить параметром `metrics_prefix` блока `telemetry`.

| Метрика | Что показывает | На что обратить внимание |
| --- | --- | --- |
| `stronghold_core_unsealed` | Узел распечатан (`1`) или запечатан (`0`) | Любое значение `0` |
| `stronghold_core_active` | Узел является активным (`1`) | Сумма по кластеру должна быть равна `1` |
| `stronghold_core_handle_request` | Время обработки запросов, мс | Рост 99-го перцентиля |
| `stronghold_core_handle_login_request` | Время обработки запросов аутентификации, мс | Рост 99-го перцентиля |
| `stronghold_expire_num_leases` | Количество активных аренд | Постоянный рост без снижения |
| `stronghold_token_count` | Количество токенов | Постоянный рост без снижения |
| `stronghold_audit_log_request_failure` | Ошибки записи запросов в аудит-устройства | Любое ненулевое приращение |
| `stronghold_audit_log_response_failure` | Ошибки записи ответов в аудит-устройства | Любое ненулевое приращение |
| `stronghold_raft_leader_lastContact` | Время с последнего контакта standby-узла с лидером Raft, мс | Значения выше сотен миллисекунд |
| `stronghold_raft_state_candidate` | Переходы узла в состояние кандидата Raft | Частые приращения — признак нестабильных выборов лидера |
| `stronghold_raft_commitTime` | Время фиксации записи в Raft, мс | Рост — признак медленного диска или сети |
| `stronghold_autopilot_healthy` | Состояние кластера по данным Autopilot | Значение `0` |
| `stronghold_runtime_alloc_bytes` | Используемая процессом память | Приближение к лимиту памяти |

## Примеры правил алертинга

Пример ресурса PrometheusRule с базовым набором алертов. Это шаблон для standalone-установки: он использует метку `severity` и пути standalone. Для DP см. [«Алерты модуля в DP»](#алерты-модуля-в-dp). Пороговые значения подберите под свою нагрузку.

```yaml
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: stronghold
spec:
  groups:
    - name: stronghold.availability
      rules:
        - alert: StrongholdSealed
          expr: stronghold_core_unsealed == 0
          for: 1m
          labels:
            severity: critical
          annotations:
            summary: "Узел Stronghold {{ $labels.instance }} запечатан"
            description: "Узел не обслуживает запросы. Выполните распечатывание или проверьте auto unseal."
        - alert: StrongholdNoActiveNode
          expr: sum(stronghold_core_active) < 1
          for: 2m
          labels:
            severity: critical
          annotations:
            summary: "В кластере Stronghold нет активного узла"
            description: "Проверьте состояние Raft и кворума командой d8 stronghold operator raft list-peers."
        - alert: StrongholdLeaderFlapping
          expr: increase(stronghold_raft_state_candidate[15m]) > 3
          labels:
            severity: warning
          annotations:
            summary: "Частые выборы лидера Raft на {{ $labels.instance }}"
            description: "Проверьте сетевую связность между узлами и задержки диска."
        - alert: StrongholdAutopilotUnhealthy
          expr: stronghold_autopilot_healthy == 0
          for: 5m
          labels:
            severity: warning
          annotations:
            summary: "Autopilot сообщает о неработоспособном кластере Stronghold"
    - name: stronghold.performance
      rules:
        - alert: StrongholdHighRequestLatency
          expr: stronghold_core_handle_request{quantile="0.99"} > 500
          for: 10m
          labels:
            severity: warning
          annotations:
            summary: "99-й перцентиль времени обработки запросов выше 500 мс"
        - alert: StrongholdLeaseCountGrowth
          expr: delta(stronghold_expire_num_leases[1h]) > 10000
          labels:
            severity: warning
          annotations:
            summary: "Количество аренд выросло более чем на 10000 за час"
            description: "Проверьте TTL аренд и клиентов, которые не переиспользуют токены."
        - alert: StrongholdLeaseCountHigh
          expr: stronghold_expire_num_leases > 250000
          for: 15m
          labels:
            severity: warning
          annotations:
            summary: "Количество активных аренд превышает 250000"
    - name: stronghold.audit
      rules:
        - alert: StrongholdAuditFailures
          expr: increase(stronghold_audit_log_request_failure[5m]) > 0 or increase(stronghold_audit_log_response_failure[5m]) > 0
          labels:
            severity: critical
          annotations:
            summary: "Ошибки записи в аудит-устройства Stronghold"
            description: "При отказе всех аудит-устройств Stronghold перестаёт обслуживать запросы."
    - name: stronghold.certificates
      rules:
        - alert: StrongholdTLSCertificateExpiringSoon
          expr: (probe_ssl_earliest_cert_expiry{job="stronghold-tls"} - time()) / 86400 < 21
          for: 1h
          labels:
            severity: warning
          annotations:
            summary: "TLS-сертификат Stronghold {{ $labels.instance }} истекает менее чем через 21 день"
    - name: stronghold.storage
      rules:
        - alert: StrongholdRaftDiskUsageHigh
          expr: |
            (1 - node_filesystem_avail_bytes{mountpoint="/opt/stronghold/data"}
              / node_filesystem_size_bytes{mountpoint="/opt/stronghold/data"}) > 0.8
          for: 15m
          labels:
            severity: warning
          annotations:
            summary: "Раздел с данными Raft на {{ $labels.instance }} заполнен более чем на 80%"
```

Комментарии к примеру:

- Алерт на срок действия TLS-сертификата использует метрику `probe_ssl_earliest_cert_expiry` из blackbox_exporter, который опрашивает HTTPS-адрес Stronghold. Аналогично контролируйте срок действия корневых и промежуточных сертификатов [PKI](../../../user/secrets-engines/pki/), выпущенных в Stronghold, например с помощью экспортера сертификатов, проверяющего файлы или адреса сервисов.
- Алерт на заполнение диска использует метрики node-exporter. Укажите точку монтирования каталога, заданного в параметре `path` блока `storage "raft"`. В DP данные хранятся в каталоге `/var/lib/deckhouse/stronghold` на master-узлах (внутри Pod'а он смонтирован в `/stronghold/data`). Если в конфигурации модуля задан параметр `storageClass`, данные хранятся в PVC `data-stronghold-N`, и точка монтирования из примера неприменима.
- Для DP перенесите `spec.groups` в ресурс CustomPrometheusRules.

Stronghold не экспортирует метрики срока действия сертификатов PKI (есть только метрики очистки `secrets_pki_tidy_*`). Контролируйте сроки сертификатов внешними средствами, например с помощью blackbox-exporter.

## Проверка состояния через sys/health

Эндпоинт [`GET /sys/health`](../../../reference/api/system/#get-syshealth) доступна без аутентификации и подходит для проверок балансировщика и внешнего мониторинга:

```shell
curl -s -o /dev/null -w "%{http_code}\n" "${STRONGHOLD_ADDR}/v1/sys/health"
```

Коды ответа:

| Код | Состояние |
| --- | --- |
| `200` | Инициализирован, распечатан, активный узел |
| `429` | Распечатан, standby-узел |
| `472` | Узел — вторичный кластер DR-репликации |
| `473` | Performance standby-узел |
| `423` | Активный узел ещё не готов (только при `isleaderreadyok=true`) |
| `425` | Raft Autopilot ещё не готов (только при `raftautopilotok=true`) |
| `501` | Не инициализирован |
| `503` | Запечатан |

Коды ответа можно переопределить параметрами запроса (поведение upstream Vault):

- `standbyok=true` — возвращать `200` для standby-узлов. Используйте для балансировщика, который распределяет запросы на все распечатанные узлы;
- `activecode`, `standbycode`, `sealedcode`, `uninitcode` — задать собственный код для соответствующего состояния.

Пример проверки для балансировщика, который должен направлять трафик только на активный узел:

```shell
curl -s -o /dev/null -w "%{http_code}\n" "${STRONGHOLD_ADDR}/v1/sys/health?standbycode=503"
```

Для проверки состояния Raft используйте команды:

```shell
d8 stronghold operator raft list-peers
d8 stronghold operator raft autopilot state
```
