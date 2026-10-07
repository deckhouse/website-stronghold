---
title: "Сценарии установки"
description: "Типовые сценарии установки Stronghold в Deckhouse Kubernetes Platform: тестовый кластер с одним master-узлом, production-кластер с тремя master-узлами, выделенные узлы и закрытый контур."
hidden: true
weight: 40
---

<!-- TODO(verify): страница скрыта до проверки параметров размещения модуля stronghold на выделенных узлах; после проверки удалите hidden: true. -->

Модуль `stronghold` запускает узлы Stronghold на master-узлах кластера Deckhouse Platform (DP) и хранит данные в каталоге `/var/lib/deckhouse/stronghold` на этих узлах (см. [«Настройка Stronghold»](../../configuration/)). Выбор сценария определяется назначением кластера и требованиями к доступности.

| Сценарий | Назначение | Отказоустойчивость Stronghold |
| --- | --- | --- |
| [Один master-узел](#тестовый-кластер-с-одним-master-узлом) | Тестирование, демонстрационные стенды | Нет |
| [Три master-узла](#production-кластер-с-тремя-master-узлами) | Production | Отказ одного узла |
| [Выделенные узлы](#выделенные-узлы-для-stronghold) | Изоляция Stronghold от других нагрузок | Зависит от числа узлов |
| [Закрытый контур](#установка-в-закрытом-контуре) | Инфраструктура без доступа в интернет | Зависит от выбранной топологии |

Общий порядок установки платформы описан в разделах [«Подготовка окружения»](../steps/prepare/), [«Установка платформы»](../steps/install/) и [«Первичная настройка доступа»](../steps/access/).

## Тестовый кластер с одним master-узлом

Подходит для знакомства с продуктом и функционального тестирования.

1. Установите DP с одним master-узлом по инструкции [«Установка платформы»](../steps/install/).
1. Включите модуль `stronghold` (см. [«Настройка Stronghold»](../../configuration/#включение-модуля)):

   ```shell
   d8 system module enable stronghold
   ```

1. Проверьте состояние:

   ```shell
   d8 k get modules stronghold
   d8 k -n d8-stronghold get pods
   ```

{{< alert level="warning" >}}
С одним master-узлом Stronghold работает без отказоустойчивости: при недоступности узла недоступно и хранилище секретов. Не используйте этот сценарий в production.
{{< /alert >}}

## Production-кластер с тремя master-узлами

Типовая production-конфигурация. Используйте нечётное количество master-узлов, чтобы сохранялся кворум (см. [«Мастер узлы»](../../platform-management/control-plane/masters/)).

1. Установите DP с тремя master-узлами или [добавьте master-узлы](../../platform-management/control-plane/masters/#добавление-master-узла) в существующий кластер. Перед добавлением следующего узла дождитесь статуса `Ready` для всех master-узлов:

   ```shell
   d8 k get no -l node-role.kubernetes.io/control-plane=
   ```

1. Включите модуль `stronghold` и настройте доступ администраторов (см. [«Настройка Stronghold»](../../configuration/)).
1. Проверьте, что все узлы Stronghold вошли в кластер Raft:

   ```shell
   d8 stronghold operator raft list-peers
   ```

По умолчанию число реплик Stronghold равно числу master-узлов плюс числу арбитров (etcd-arbiter); если сумма чётная, добавляется ещё одна реплика. Кворум Raft — `floor(n/2)+1`, где `n` — число реплик.

Настройте регулярное резервное копирование: [ручные снимки](../../../../admin/backups/save/) или [автоматические снимки](../../../../admin/backups/automated-snapshots/) (Stronghold EE).

## Выделенные узлы для Stronghold

Начиная с версии `1.19` модуль `stronghold` поддерживает настройку `nodeSelector` и `storageClass`, а также миграцию хранилища для HA-установок (см. [историю изменений](../../../../release-notes/)). Это позволяет запускать Stronghold на узлах, выделенных только под него. Если `nodeSelector` не задан, поды размещаются на узлах с меткой `node-role.kubernetes.io/control-plane=""`. Учтите: при пустом `storageClass` поды запускаются на control-plane и etcd-arbiter узлах независимо от `nodeSelector`, поэтому задайте оба параметра.

1. Создайте NodeGroup для узлов Stronghold с меткой и taint, чтобы на этих узлах не запускались другие нагрузки (см. [«Группы узлов»](../../platform-management/node-management/node-group/)):

   ```yaml
   apiVersion: deckhouse.io/v1
   kind: NodeGroup
   metadata:
     name: stronghold
   spec:
     nodeType: Static
     nodeTemplate:
       labels:
         node-role.deckhouse.io/stronghold: ""
       taints:
         - effect: NoExecute
           key: dedicated.deckhouse.io
           value: stronghold
   ```

1. Добавьте в группу нечётное количество узлов, например три (см. [«Добавление узла»](../../platform-management/node-management/adding-node/)).
1. Укажите в ModuleConfig `stronghold` селектор узлов и StorageClass:

   ```yaml
   apiVersion: deckhouse.io/v1alpha1
   kind: ModuleConfig
   metadata:
     name: stronghold
   spec:
     enabled: true
     version: 1
     settings:
       nodeSelector:
         node-role.deckhouse.io/stronghold: ""
       storageClass: <STORAGE_CLASS_NAME>
   ```

<!-- TODO(verify): нужно ли задавать tolerations для taint dedicated.deckhouse.io, порядок миграции хранилища с master-узлов на выделенные узлы и ссылка на описание параметров модуля (/modules/stronghold/stable/configuration.html). -->

## Установка в закрытом контуре

В закрытом контуре узлы кластера не имеют доступа к `registry.deckhouse.ru`. Образы платформы и модуля `stronghold` загружаются в собственное хранилище образов.

1. На машине с доступом в интернет загрузите образы DP и модулей с помощью `d8 mirror`, перенесите их в закрытый контур и загрузите в хранилище образов:

   ```shell
   d8 mirror push modules <REGISTRY_HOST:PORT>/<PATH_TO_DKP_REPO> -u <USERNAME> -p <PASSWORD>
   ```

   <!-- TODO(verify): команды d8 mirror pull/push для полной поставки DKP и модуля stronghold и ссылка на инструкцию DKP по установке в закрытом окружении. -->

1. Установите DP, указав в InitConfiguration параметры доступа к собственному хранилищу образов (см. [«Установка платформы»](../steps/install/)).
1. Убедитесь, что хранилище образов доступно с каждого master-узла. Если модуль поставляется отдельно, создайте ресурс ModuleSource с адресом модулей в хранилище образов (пример — в разделе [«Переключение Stronghold с EE на CSE»](../../platform-management/switching-editions/ee-to-cse/)).
1. Включите модуль `stronghold` и проверьте, что поды не находятся в состояниях `ImagePullBackOff` или `ErrImagePull`:

   ```shell
   d8 k -n d8-stronghold get pods
   ```

Для доступа к веб-интерфейсу в закрытом контуре без публичного центра сертификации используйте ClusterIssuer с самоподписанным CA или собственный сертификат (см. [«Способы организации доступа через инлет Ingress»](../../configuration/#способы-организации-доступа-через-инлет-ingress)).
