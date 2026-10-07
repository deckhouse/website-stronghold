---
title: "Развёртывание в DKP"
description: "Как модуль stronghold размещается в кластере Deckhouse Kubernetes Platform: поды на master-узлах, Raft, Ingress, доступ через Dex и автоматическое распечатывание."
weight: 20
---

В Deckhouse Platform (DP) Stronghold поставляется как модуль `stronghold`. Модуль сам разворачивает кластер Stronghold, инициализирует и распечатывает хранилище, публикует веб-интерфейс и API через выбранный инлет (по умолчанию Ingress) и в режиме `Automatic` настраивает вход через Dex. Порядок включения и настройки описан в разделе [«Настройка Stronghold»](../../../install/dkp/configuration/).

## Схема размещения

![layout_plan_dkp.png](../../../images/layout_plan_dkp.png)

Схема иллюстративная: активный узел выбирается Raft и не обязательно является `stronghold-0`. Элементы Ingress, секрет `ingress-tls`, Dex и `stronghold-keys` соответствуют конфигурации по умолчанию (инлет `Ingress` с `CertManager` и режим `Automatic`).

## Компоненты модуля

- **Неймспейс `d8-stronghold`** — в нём работают все контейнеры Stronghold и хранятся служебные секреты модуля.
- **Поды Stronghold** — каждый узел Stronghold запускается в отдельном поде. Поды управляются StatefulSet и размещаются по одному на узел (pod anti-affinity). Если `storageClass` пуст, поды запускаются на control-plane (master) узлах и на узлах с меткой `node.deckhouse.io/etcd-arbiter`. Если `storageClass` задан, размещение определяется параметром модуля `nodeSelector` (по умолчанию `node-role.kubernetes.io/control-plane=""`).
  Число реплик равно числу master-узлов плюс числу арбитров (etcd-arbiter); если сумма чётная, добавляется ещё одна реплика. Позже его можно изменить вручную.
- **Хранилище** — интегрированное хранилище Raft в режиме HA включено по умолчанию. Если `storageClass` пуст, данные каждого узла хранятся на локальном диске узла в каталоге `/var/lib/deckhouse/stronghold` (может использоваться подкаталог с именем из `localPathUUID` модуля). Если `storageClass` задан, данные хранятся в PersistentVolumeClaim `data-stronghold-N` StatefulSet.
- **Внутренний сервис** — узлы обращаются друг к другу по именам вида `stronghold-0.stronghold-internal`. Внутренний API слушает порт `8300` (адреса Raft-пиров — `8301`); `retry_join` между узлами и распечатывание inner-cluster идут на порт `8300`. Внешний API слушает `8200` (кластерный адрес `8201`). См. раздел [«Порты»](../ports/).
- **Инлет** — параметр `inlet` задаёт способ публикации сервиса: `Ingress` (по умолчанию), `GatewayAPI`, `LoadBalancer`, `NodePort` или `None`. Для `LoadBalancer`, `NodePort` и `None` обязательно указать `https.mode: CustomCertificate`. При инлете `Ingress` веб-интерфейс и API публикуются по адресу, который получается из шаблона [`publicDomainTemplate`](/products/kubernetes-platform/documentation/v1/reference/api/global.html#parameters-modules-publicdomaintemplate) заменой `%s` на `stronghold`, например `stronghold.mycompany.tld`.
- **TLS-сертификат** — секрет `ingress-tls` в неймспейсе `d8-stronghold`. При инлете `Ingress` и `https.mode: CertManager` его выпускает cert-manager по ClusterIssuer, заданному в настройках DP; для остальных инлетов, а также при `https.mode: CustomCertificate`, используется сертификат из `customCertificate`. Без этого секрета поды остаются в состоянии `ContainerCreating`.

## Режим работы и распечатывание

Режим задаётся параметром `management.mode`. По умолчанию используется `Automatic`; также поддерживается `Manual`. В режиме `Manual` автоматическая инициализация отключена, интеграции с Dex и Kubernetes не настраиваются, секрета `stronghold-keys` нет, а параметр `management.administrators` недоступен. В режиме `Automatic`:

1. При первом запуске модуль автоматически инициализирует хранилище.
1. Ключ распечатывания и root-токен сохраняются в секрет `stronghold-keys` неймспейса `d8-stronghold`.
1. Модуль автоматически распечатывает узлы Stronghold после инициализации и при каждом перезапуске подов.

<!-- TODO(verify): использует ли модуль в DKP механизм seal "inner-cluster" (release notes v1.18: «автоматическое распечатывание с хранением ключей в памяти узлов Stronghold») или распечатывает узлы по секрету stronghold-keys. -->

{{< alert level="warning" >}}
Секрет `stronghold-keys` содержит root-токен и ключ распечатывания. Любой, кто может прочитать секреты в неймспейсе `d8-stronghold`, получает полный доступ к Stronghold. Ограничьте доступ к этому неймспейсу средствами RBAC DP и сохраните копию секрета в защищённом месте: при выключении модуля секрет удаляется, а без него восстановить доступ к данным нельзя.
{{< /alert >}}

HSM (`seal "pkcs11"`) и Yandex Cloud KMS (`seal "yandexcloudkms"`) в DP не поддерживаются — они доступны только в [standalone-развёртывании](../deployment-standalone/).

## Доступ пользователей

- После инициализации модуль создаёт в Stronghold роль `deckhouse_administrators` и включает вход в веб-интерфейс через OIDC-аутентификацию [Dex](/modules/user-authn/) (модуль `user-authn`).
- Администраторы Stronghold задаются в ModuleConfig модуля в параметре `management.administrators` (только при `management.mode: Automatic`) — группами (`Group`) или отдельными пользователями (`User`). Пользователь должен входить хотя бы в одну группу, иначе вход через OIDC не сработает.
- Остальных пользователей и их права настраивайте встроенными средствами Stronghold: методами аутентификации и политиками.

## Интеграция с кластером

Модуль автоматически подключает текущий кластер DP к Stronghold для работы модуля [`secrets-store-integration`](/modules/secrets-store-integration/stable/), который доставляет секреты в поды приложений.

<!-- TODO(verify): какой метод аутентификации (kubernetes auth) и какой путь mount модуль настраивает для secrets-store-integration. -->

## Кворум и отказоустойчивость

Кворум Raft рассчитывается по формуле `floor(n/2)+1`, где `n` — число узлов Stronghold (PodDisruptionBudget модуля использует `minAvailable = replicas/2 + 1`). По умолчанию `n` равно числу master-узлов плюс числу арбитров, округлённому до нечётного. В кластере из трёх master-узлов для чтения и записи нужны как минимум два работающих пода. Кластер из одного master-узла не отказоустойчив. Порядок действий при потере кворума — на странице [«Восстановление после потери кворума»](../../../install/dkp/raft-lost-quorum-recovery/).

## Выключение модуля и данные

При выключении модуля удаляются все контейнеры Stronghold из неймспейса `d8-stronghold` и секрет `stronghold-keys`. Если `storageClass` пуст, данные на узлах в каталоге `/var/lib/deckhouse/stronghold` сохраняются. Если `storageClass` задан, PVC удаляются вместе с StatefulSet (`persistentVolumeClaimRetentionPolicy.whenDeleted: Delete`), поэтому данные не сохраняются. Чтобы вернуть доступ, включите модуль и поместите в неймспейс сохранённую копию секрета `stronghold-keys`.

## Требования

- Для Stronghold `1.19` нужна DP версии `1.76` или новее.
- Stronghold EE включается лицензионным ключом в ModuleConfig и доступен только в коммерческих редакциях DP. Подробнее — в разделе [«Редакции»](../../../about/editions/).
