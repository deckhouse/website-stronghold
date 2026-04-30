---
title: "Переключение с EE на CSE"
linkTitle: "Переключение с EE на CSE"
weight: 10
---

**Deckhouse Stronghold Enterprise Edition (EE)** можно переключить на **Deckhouse Stronghold Certified Security Edition (CSE)** одним из следующих способов:

- [в исполнении Standalone](/products/stronghold/documentation/Руководство по развертыванию/Развертывание в Standalone/Переключение редакций/Переключение с EE на CSE/);
- в исполнении **DKP**.

> Предупреждение  
> Поддерживается обновление с **Stronghold EE 1.15.x** до **Stronghold CSE 1.16.0**. Если используется версия Stronghold EE ниже `1.15.x`, сначала [обновитесь до последней версии ветки](/products/stronghold/documentation/Руководство администратора/Обновление/Обновление в DKP/).
> Предупреждение  
> При обновлении до Stronghold CSE возможна временная недоступность сервиса.

## Переключение в DKP

> Предупреждение  
> Для перехода на Stronghold CSE требуется **DKP CSE версии 1.73 и выше**.

Перед началом переключения выполните следующие действия:

1. Проверьте текущую версию Stronghold:

   ```bash
   stronghold version
   ```

1. Сохраните unseal-ключи и root-токен в защищённое хранилище:

   ```bash
   d8 k -n d8-stronghold get secret stronghold-keys -o yaml > stronghold-keys.yaml
   chmod 600 stronghold-keys.yaml
   ```

1. Создайте резервную копию или снимок кластера Stronghold:

   ```bash
   export STRONGHOLD_ADDR=https://$(d8 k -n d8-stronghold get ing stronghold -o json | jq -r '.spec.rules[0].host')
   d8 stronghold login -method=oidc -path=oidc_deckhouse
   # Либо через root-токен:
   # d8 stronghold login -method=token
   d8 stronghold operator raft snapshot save stronghold-$(date +%F_%H-%M).snap
   ```

1. Проверьте, что снимок создан:

   ```bash
   ls -lh ./stronghold-*.snap
   ```

   > Важно  
   > Храните полученные файлы за пределами кластера DKP.

1. Подготовьте пакет или бинарный файл **Stronghold CSE 1.16.0** на каждом узле.

1. Убедитесь, что `ModuleConfig` `stronghold` существует:

   ```bash
   d8 k get mc stronghold -o yaml
   ```

   Если `ModuleConfig` отсутствует, создайте его:

   ```bash
   cat <<EOF | d8 k apply -f -
   apiVersion: deckhouse.io/v1alpha1
   kind: ModuleConfig
   metadata:
     name: stronghold
   spec:
     enabled: false
   EOF
   ```

## Отключение автообновления модуля

Перед переключением рекомендуется перевести обновление модуля `stronghold` в ручной режим.

1. Определите текущий канал обновлений модуля `stronghold`:

   ```bash
   d8 k get module stronghold -o jsonpath='{.properties.releaseChannel}'
   ```

   Если команда не выводит значение, используйте канал `stable` в следующем шаге.

1. Создайте ресурс `ModuleUpdatePolicy` с ручным режимом обновления, указав в поле `releaseChannel` текущий канал обновлений, и примените его:

   ```bash
   cat <<EOF > stronghold-mup.yaml
   apiVersion: deckhouse.io/v1alpha2
   kind: ModuleUpdatePolicy
   metadata:
     name: stronghold-update-policy
   spec:
     releaseChannel: <STRONGHOLD_RELEASE_CHANNEL>
     update:
       mode: Manual
   EOF

   d8 k apply -f stronghold-mup.yaml
   d8 k get mup stronghold-update-policy
   ```

   Ожидаемый вывод:

   ```text
   NAME                       RELEASE CHANNEL   UPDATE MODE
   stronghold-update-policy   Stable            Manual
   ```

1. Укажите созданный ресурс `ModuleUpdatePolicy` в `ModuleConfig` `stronghold`:

   ```bash
   d8 k patch moduleconfig stronghold --type merge --patch '{"spec":{"updatePolicy":"stronghold-update-policy"}}'
   ```

1. Проверьте, что значение сохранилось:

   ```bash
   d8 k get mc stronghold -o jsonpath='{.spec.updatePolicy}'
   ```

   Ожидаемый вывод:

   ```text
   stronghold-update-policy
   ```

1. Убедитесь, что изменения применились в модуле `stronghold`:

   ```bash
   d8 k get module stronghold -o jsonpath='{.properties.updatePolicy}'
   ```

   Ожидаемый вывод:

   ```text
   stronghold-update-policy
   ```

   > Примечание  
   > Если используется **DKP CSE 1.67 и ниже**, либо если модуль ещё ни разу не запускался, вывод команды будет пустым. Дополнительные действия в таком случае не требуются.

## Проверки перед началом переключения

1. Проверьте версию DKP. Для перехода на Stronghold CSE требуется **DKP CSE версии 1.73 и выше**:

   ```bash
   d8 k -n d8-system get configmap d8-deckhouse-version-info -o yaml
   ```

   Также проверить версию DKP можно в веб-интерфейсе Deckhouse на главной странице панели управления кластером (`https://console.<CLUSTER_DOMAIN>`).

   Если понадобится обновить или переключить редакцию DKP, воспользуйтесь инструкциями:
   - [инструкция по обновлению DKP](/products/kubernetes-platform/documentation/v1/admin/configuration/update/configuration.html);
   - [инструкция по переключению DKP EE на DKP CSE](/products/kubernetes-platform/documentation/v1/admin/configuration/registry/switching-editions.html).

1. Убедитесь, что DKP работает штатно, active-под определён, очередь пуста:

   ```bash
   d8 k -n d8-system get deploy deckhouse
   d8 k -n d8-system get pod -l leader=true
   d8 system queue list
   ```

   Ожидаемый результат:
   - `Deployment` `deckhouse` находится в состоянии `Ready`:

     ```text
     NAME        READY   UP-TO-DATE   AVAILABLE   AGE
     deckhouse   3/3     3            3           1d
     ```

   - есть рабочий под в статусе `Running`:

     ```text
     NAME                        READY   STATUS    RESTARTS   AGE
     deckhouse-596d944f95-d82m8  2/2     Running   0          1d
     ```

   - очередь задач пуста:

     ```text
     Summary:
     - 'main' queue: empty.
     - 121 other queues (0 active, 121 empty): 0 tasks.
     - no tasks to handle.
     ```

1. Убедитесь, что модуль `stronghold` включён и находится в рабочем состоянии:

   ```bash
   d8 k get module stronghold
   d8 k get moduleconfig stronghold -o yaml
   ```

   Проверьте, что:
   - для `stronghold` значения `Enabled` и `Ready` равны `True`;
   - объект `ModuleConfig` `stronghold` существует;
   - в `spec.settings.license` указан лицензионный ключ.

1. Убедитесь, что используется Stronghold EE версии `1.15.x`. Это можно сделать одним из следующих способов:

   - через веб-интерфейс Deckhouse — откройте главную страницу панели управления кластером (`https://console.<CLUSTER_DOMAIN>`) и проверьте, что в нижней части страницы указана версия `1.15.x` с суффиксом `ee`;

   - по манифесту или образу пода:

     ```bash
     d8 k -n d8-stronghold get pod -o yaml | grep version

     ```

     В лейблах подов должна быть указана версия `1.15.x` с суффиксом `ee`;

   - по журналам:

     ```bash
     d8 k -n d8-stronghold logs stronghold-0 | head -20
     ```

     В стартовых строках баннера должна быть указана версия с редакцией:

     ```text
     Version: Stronghold v1.15.0+ee
     ```

## Подготовка модуля к установке

### Подготовьте Stronghold CSE в хранилище образов контейнеров

Этот шаг выполняется на машине, имеющей доступ к загрузке образов в хранилище образов контейнеров.

1. Проверьте контрольную сумму архива и загрузите bundle модуля Stronghold CSE:

   ```bash
   # Проверьте контрольную сумму архива:
   gost12sum module-stronghold.tar

   # Переместите архив в директорию modules:
   mkdir modules
   mv module-stronghold.tar modules/.

   # Адрес хоста хранилища образов контейнеров.
   export REGISTRY_HOST="<REGISTRY_HOST:PORT>"

   # Адрес хранилища образов DKP.
   export MODULES_MODULE_REPO="${REGISTRY_HOST}/<PATH_TO_DKP_REPO>"

   # Используйте учётную запись с правами на запись в хранилище образов контейнеров.
   d8 mirror push modules $MODULES_MODULE_REPO -u <USERNAME> -p <PASSWORD>
   ```

### Создайте ModuleSource для Stronghold CSE

1. Убедитесь средствами вашей инфраструктуры, что хранилище образов контейнеров доступно с каждого master-узла.

1. Подготовьте и примените `ModuleSource`:

   ```bash
   # Путь к модулям в хранилище образов контейнеров.
   MODULES_MODULE_SOURCE="$MODULES_MODULE_REPO/modules"

   # Имя пользователя, имеющего доступ на чтение образов.
   REGISTRY_USER="<REGISTRY_USER>"

   # Пароль пользователя.
   REGISTRY_PASSWORD="<REGISTRY_PASSWORD>"

   # Сертификат УЦ домена, используемого для хранилища образов контейнеров.
   REGISTRY_CA_CERT="<CA_CERT_FOR_REGISTRY>"

   AUTH_STRING=$(echo -n "${REGISTRY_USER}:${REGISTRY_PASSWORD}" | base64)
   DOCKER_CFG=$(echo -n '{"auths":{"'${REGISTRY_HOST}'":{"username":"'${REGISTRY_USER}'","password":"'${REGISTRY_PASSWORD}'","auth":"'${AUTH_STRING}'"}}}' | base64 -w0)

   cat <<EOF > stronghold-cse-ms.yaml
   apiVersion: deckhouse.io/v1alpha1
   kind: ModuleSource
   metadata:
     name: stronghold-cse
   spec:
     registry:
       ca: "$REGISTRY_CA_CERT"
       dockerCfg: $DOCKER_CFG
       repo: $MODULES_MODULE_SOURCE
       scheme: HTTPS
   EOF

   # Проверьте конфигурацию.
   cat stronghold-cse-ms.yaml

   # Примените ресурс.
   d8 k apply -f stronghold-cse-ms.yaml
   ```

1. Проверьте статус `ModuleSource`:

   ```bash
   d8 k get modulesource stronghold-cse -o yaml
   ```

   Убедитесь, что в статусе отсутствуют ошибки, а поле `status.message` пустое:

   ```yaml
   status:
     message: ""
   ```

## Переключение модуля Stronghold с EE на CSE

Для переключения обновите `ModuleConfig` `stronghold` и укажите новый `ModuleSource`.

1. Проверьте наличие `ModuleConfig` `stronghold`:

   ```bash
   d8 k get mc stronghold -o yaml
   ```

1. Добавьте в `ModuleConfig` поле `spec.source` со значением `stronghold-cse`:

   ```bash
   d8 k patch moduleconfig stronghold --type merge --patch '{"spec":{"source":"stronghold-cse"}}'
   d8 k get moduleconfig stronghold -o yaml
   ```

   Проверьте, что изменения применились:
   - в `ModuleConfig` `stronghold` присутствует значение `spec.source: stronghold-cse`;
   - остальные поля в `spec.settings` не изменились.

1. Если после переключения источника модуль сообщает о проблеме с лицензией, обновите только поле лицензии (`CSE_LICENSE`):

   ```bash
   d8 k patch moduleconfig stronghold --type merge --patch '{"spec":{"settings":{"license":"<CSE_LICENSE>"},"version":1}}'
   ```

1. Если модуль ранее был выключен, включите его:

   ```bash
   d8 system module enable stronghold
   ```

1. Дождитесь стабилизации DKP и пустой очереди:

   ```bash
   d8 k -n d8-system get pods -l app=deckhouse
   d8 system queue list
   ```

   Проверьте, что:
   - под `deckhouse` находится в состоянии `Running`/`Ready`;
   - очередь пуста.

## Переключите ModuleUpdatePolicy в режим AutoPatch

1. Измените режим обновления:

   ```bash
   d8 k patch mup stronghold-update-policy --type merge --patch '{"spec":{"update":{"mode":"AutoPatch"}}}'
   ```

1. Проверьте, что режим обновления изменился:

   ```bash
   d8 k get mup stronghold-update-policy
   ```

   Ожидаемый вывод:

   ```text
   NAME                       RELEASE CHANNEL   UPDATE MODE
   stronghold-update-policy   Stable            AutoPatch
   ```

## Проверьте завершение перехода на CSE

1. Проверьте состояние модуля:

   ```bash
   d8 k get modules stronghold
   ```

   Ожидаемый вывод:

   ```text
   NAME         STAGE                  SOURCE           PHASE   ENABLED   READY
   stronghold   General Availability   stronghold-cse   Ready   True      True
   ```

1. Проверьте источник модуля:

   ```bash
   d8 k get modulesource stronghold-cse -o yaml
   d8 k get moduleconfig stronghold -o yaml
   d8 k get module stronghold -o yaml
   ```

   Проверьте, что:
   - отсутствуют ошибки `auth`, `tls`, `x509`, `timeout`;
   - отсутствуют предупреждающие события;
   - в `ModuleConfig` `stronghold` указано `spec.source: stronghold-cse`;
   - в `Module` `stronghold` указано `properties.source: stronghold-cse`.

1. Проверьте поды `stronghold`:

   ```bash
   d8 k -n d8-stronghold get po
   ```

   Проверьте, что:
   - отсутствуют состояния `ImagePullBackOff`, `ErrImagePull`, `CrashLoopBackOff`;
   - поды `stronghold-*` находятся в состоянии `Running` и имеют готовность `2/2`.

## Альтернативный способ: перенос на новый кластер DKP CSE

Если не удалось пройти [проверки перед началом переключения](#проверки-перед-началом-переключения), то есть привести текущий кластер к требуемым версиям **DKP CSE 1.73 и выше** и **Stronghold EE 1.15.x**, можно перенести кластер Stronghold на новый кластер DKP CSE `1.73` путём восстановления резервной копии.

Для этого выполните следующие действия:

1. Выполните резервное копирование данных кластера Stronghold, как описано выше.

1. Разверните кластер **DKP CSE 1.73**.

1. Начиная с раздела [Отключение автообновления модуля](#отключение-автообновления-модуля), последовательно выполните все шаги на новом кластере.

1. Восстановите резервную копию, unseal-ключи и root-токен. Команды выполняются на хосте, где находятся файлы резервной копии `stronghold-*.snap` и ключи `stronghold-keys.yaml`.

### Восстановление снимка

```bash
export STRONGHOLD_ADDR=https://$(d8 k -n d8-stronghold get ing stronghold -o json | jq -r '.spec.rules[0].host')
d8 stronghold login -method=oidc -path=oidc_deckhouse
# Либо через root-токен:
# d8 stronghold login -method=token
d8 stronghold operator raft snapshot restore -force stronghold-<SNAPSHOT_DATE>.snap
```

### Восстановление unseal-ключей и root-токена

```bash
d8 k -n d8-stronghold delete secret stronghold-keys
d8 k -n d8-stronghold create -f stronghold-keys.yaml
```

### Проверка подов Stronghold

```bash
d8 k -n d8-stronghold get po
```

Проверьте, что:
- отсутствуют состояния `ImagePullBackOff`, `ErrImagePull`, `CrashLoopBackOff`;
- поды `stronghold-*` находятся в состоянии `Running` и имеют готовность `2/2`.

## Результат

После успешного выполнения шагов модуль Stronghold в составе DKP будет переключён с **Enterprise Edition** на **Certified Security Edition**.

Для дополнительной проверки состояния после переключения используйте раздел [Проверка работоспособности](/products/stronghold/documentation/deployment/functionality-check/).
