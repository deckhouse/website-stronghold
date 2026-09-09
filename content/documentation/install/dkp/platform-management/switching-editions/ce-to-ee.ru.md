---
title: "Переключение Stronghold на редакцию EE"
description: "Переход с базового Stronghold на Stronghold EE в DKP путём указания лицензионного ключа в ModuleConfig."
weight: 10
params:
  relatedLinks:
    - title: "Редакции"
      url: ../../../../../about/editions/
    - title: "Настройка Stronghold"
      url: ../../../configuration/
    - title: "Переключение Stronghold с EE на CSE"
      url: ../ee-to-cse/
    - title: "Создание снимка"
      url: ../../../../../admin/backups/save/
---

Модуль `stronghold` переключается с базового Stronghold на Stronghold EE указанием лицензионного ключа в параметре [`spec.settings.license`](/modules/stronghold/stable/configuration.html#parameters-license) ресурса ModuleConfig `stronghold`. Переустанавливать модуль и переносить данные не требуется: секреты, политики, механизмы секретов и настроенные методы аутентификации сохраняются.

{{< alert level="warning" >}}
Stronghold EE лицензируется отдельно и доступен для использования только в коммерческих редакциях DKP. В DKP CE переключение на Stronghold EE невозможно. Подробнее — в разделе [«Редакции»](../../../../../about/editions/).
{{< /alert >}}

{{< alert level="warning" >}}
После применения лицензионного ключа поды `stronghold-*` пересоздаются на образе Stronghold EE по очереди, поэтому доступность сервиса остаётся прежней. Тем не менее выполняйте переключение строго в окне обслуживания.
{{< /alert >}}

## Перед переключением

1. Убедитесь, что в кластере используется коммерческая редакция DKP:

   ```shell
   d8 k -n d8-system get configmap d8-deckhouse-version-info -o jsonpath='{.data.data\.json}'
   ```

   В `json` объекте должна быть указана редакция, отличная от `CE`. Пример вывода:

   ```console
   { "channel":"Stable", "version":"v1.72.5", "edition":"EE" }
   ```

   Также редакцию и версию DKP можно посмотреть в веб-интерфейсе Deckhouse на главной странице панели управления кластером (`https://console.<CLUSTER_DOMAIN>`).

   {{< alert level="info" >}}
   Если требуется сменить редакцию DKP, воспользуйтесь [инструкцией по переключению редакции DKP](/products/kubernetes-platform/documentation/v1/admin/configuration/registry/switching-editions.html).
   {{< /alert >}}

1. Убедитесь, что модуль `stronghold` включён и находится в рабочем состоянии:

   ```shell
   d8 k get module stronghold
   d8 k get moduleconfig stronghold -o yaml
   ```

   Проверьте, что:

   - для модуля `stronghold` значения `Enabled` и `Ready` равны `True`;
   - объект ModuleConfig `stronghold` существует.

1. Сохраните unseal-ключи и root-токен в защищённое хранилище:

   ```shell
   d8 k -n d8-stronghold get secret stronghold-keys -o yaml > stronghold-keys.yaml
   chmod 600 stronghold-keys.yaml
   ```

1. Создайте [снимок хранилища](../../../../admin/backups/save/):

   ```shell
   export STRONGHOLD_ADDR=https://$(d8 k -n d8-stronghold get ing stronghold -o json | jq -r '.spec.rules[0].host')
   d8 stronghold login -method=oidc -path=oidc_deckhouse
   # Либо через root-токен:
   ## d8 stronghold login -method=token
   d8 stronghold operator raft snapshot save stronghold-$(date +%F_%H-%M).snap
   ```

   Проверить снимок можно с помощью команды:

   ```shell
   ls -lh ./stronghold-*.snap
   ```

   {{< alert level="info" >}}
   Полученные файлы храните за пределами кластера DKP.
   {{< /alert >}}

## Указание лицензионного ключа

1. Получите лицензионный ключ Stronghold EE у поставщика продукта.

1. Добавьте ключ в параметр `spec.settings.license` ресурса ModuleConfig `stronghold`:

   ```shell
   d8 k patch moduleconfig stronghold --type merge --patch '{"spec":{"version":1,"settings":{"license":"<STRONGHOLD_EE_LICENSE_KEY>"}}}'
   ```

   Тот же результат можно получить, отредактировав ModuleConfig командой `d8 k edit mc stronghold`. Пример итогового манифеста:

   ```yaml
   apiVersion: deckhouse.io/v1alpha1
   kind: ModuleConfig
   metadata:
     name: stronghold
   spec:
     enabled: true
     version: 1
     settings:
       license: <STRONGHOLD_EE_LICENSE_KEY>
       management:
         mode: Automatic
         administrators:
         - type: Group
           name: admins
   ```

   {{< alert level="warning" >}}
   Лицензионный ключ хранится в ModuleConfig в открытом виде. При необходимости ограничьте доступ к ресурсу средствами [ролевой модели](../../access-control/role-model/).
   {{< /alert >}}

1. Дождитесь пересоздания подов Stronghold:

   ```shell
   d8 k -n d8-stronghold get po -w
   ```

## Проверка переключения

1. Убедитесь, что ключ сохранился в конфигурации модуля:

   ```shell
   d8 k get moduleconfig stronghold -o jsonpath='{.spec.settings.license}'
   ```

1. Проверьте поды `stronghold`:

   ```shell
   d8 k -n d8-stronghold get po
   ```

   Проверьте, что:

   - отсутствуют состояния `ImagePullBackOff`, `ErrImagePull`, `CrashLoopBackOff`;
   - поды `stronghold-*` находятся в состоянии `Running` и имеют готовность `2/2`.

1. Проверьте редакцию в стартовых строках баннера:

   ```shell
   d8 k -n d8-stronghold logs stronghold-0 | head -20
   ```

   В выводе должна быть указана версия с суффиксом `ee`. Например:

   ```console
   Version: Stronghold v1.19.0+ee
   ```

1. Убедитесь, что хранилище разблокировано:

   ```shell
   export STRONGHOLD_ADDR=https://$(d8 k -n d8-stronghold get ing stronghold -o json | jq -r '.spec.rules[0].host')
   d8 stronghold status
   ```

   В выводе значение `Sealed` должно быть равно `false`.

После переключения станут доступны возможности Stronghold EE: журнал аудита, пространства имён, репликация между кластерами, автоматические снимки, Managed Keys, управление ролями и политиками доступа через веб-интерфейс. Полный перечень отличий — в разделе [«Редакции»](../../../../about/editions/).

## Возврат к базовому Stronghold CE

{{< alert level="warning" >}}
После удаления лицензионного ключа возможности Stronghold EE перестанут работать. Заранее отключите их использование, иначе модуль может не запуститься. В частности, установите `enableAuditLog: false`, остановите репликацию между кластерами и удалите созданные пространства имён.
{{< /alert >}}

1. Удалите параметр `spec.settings.license` целиком:

   ```shell
   d8 k patch moduleconfig stronghold --type json --patch '[{"op":"remove","path":"/spec/settings/license"}]'
   ```

1. Дождитесь пересоздания подов и убедитесь, что в баннере указана версия без суффикса `ee`:

   ```shell
   d8 k -n d8-stronghold logs stronghold-0 | head -20
   ```
