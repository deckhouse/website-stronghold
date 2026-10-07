---
title: "Service mesh"
description: "Использование PKI Stronghold как внешнего центра сертификации для Istio через cert-manager и istio-csr."
weight: 50
---

{{< alert level="warning" >}}
Сценарий описывает интеграцию сторонних компонентов и не проверялся в составе Deckhouse Platform. <!-- TODO(verify): поддержка внешнего CA в модуле istio DKP и совместимость istio-csr со Stronghold -->
{{< /alert >}}

По умолчанию Istio выпускает сертификаты рабочих нагрузок для mTLS собственным центром сертификации в istiod. Чтобы сертификаты выпускались от промежуточного CA из Stronghold, используйте цепочку:

1. [Механизм секретов PKI](../../../user/secrets-engines/pki/) Stronghold хранит промежуточный CA для mesh и роль для выпуска сертификатов SPIFFE.
1. cert-manager обращается к Stronghold через Issuer типа `vault` (см. [«cert-manager»](../cert-manager/)).
1. Компонент istio-csr проекта cert-manager принимает запросы на сертификаты от прокси Istio и выпускает их через cert-manager.

## Настройка Stronghold

Создайте роль PKI, разрешающую URI SAN в формате SPIFFE, которые использует Istio:

```bash
d8 stronghold write pki/roles/istio-mesh \
  allowed_uri_sans="spiffe://cluster.local/*" \
  allow_any_name=true \
  require_cn=false \
  max_ttl=24h
```

Разрешите Issuer cert-manager подписывать запросы по этой роли (`pki/sign/istio-mesh`), как описано в разделе [«cert-manager»](../cert-manager/). Если используется ClusterIssuer, аудитория токена в роли Kubernetes auth — `vault://<имя ClusterIssuer>`.

## Настройка cert-manager и Istio

1. Создайте ClusterIssuer `stronghold-mesh` типа `vault` с `path: pki/sign/istio-mesh`.
1. Установите istio-csr и укажите в его параметрах ClusterIssuer `stronghold-mesh`.
1. Настройте Istio на использование istio-csr как CA: укажите адрес сервиса istio-csr в `global.caAddress` и отключите встроенный CA istiod (переменная `ENABLE_CA_SERVER=false`).

Пример параметров Helm-чарта istio-csr:

```yaml
app:
  certmanager:
    issuer:
      name: stronghold-mesh
      kind: ClusterIssuer
      group: cert-manager.io
```

Корневой сертификат цепочки, которым должны доверять прокси, можно получить из Stronghold:

```bash
d8 stronghold read -field=certificate pki/cert/ca
```

Задавайте короткий `max_ttl` для сертификатов рабочих нагрузок: они перевыпускаются автоматически.
