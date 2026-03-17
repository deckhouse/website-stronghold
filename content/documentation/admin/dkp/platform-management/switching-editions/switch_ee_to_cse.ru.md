---
title: "Переключение Stronghold с EE на CSE"
weight: 20
url: /documentation/admin/platform-management/switching-editions/switch-ee-to-cse/
---

Stronghold Enterprise Edition (EE) можно обновить до Stronghold Certified Security Edition (CSE) одним из следующих способов:

- в исполнении Standalone;
- в исполнении DKP.

Поддерживается обновление с Stronghold EE 1.15.x до Stronghold CSE 1.16.0.

Если используется версия Stronghold EE ниже 1.15.x, сначала [обновитесь до последней версии ветки](/products/stronghold/documentation/admin/update/update/) 1.15.x.

При обновлении до Stronghold CSE возможна временная недоступность сервиса.

## Обновление в исполнении Standalone

Перед началом обновления выполните следующие действия:

1. Проверьте текущую версию Stronghold:

   ```shell
   stronghold version
   ```

1. Сохраните unseal-ключи и root-токен в защищённое хранилище.

1. Создайте снимок (snapshot) кластера Stronghold. Пример:

   ```shell
   stronghold operator raft snapshot save stronghold-$(date +%F_%H-%M).snap
   ls -lh stronghold-*.snap
   ```

   Проверьте, что снапшот создан:

   ```shell
   ls -lh ./stronghold-*.snap
   ```

   > Полученные файлы храните вне кластера Stronghold и вне узлов, на которых он запущен.

1. Подготовьте пакет или бинарный файл Stronghold CSE 1.16.0 на каждом узле.

### Обновление кластера с одним узлом

Для обновления кластера с одним узлом выполните следующие действия:

1. Проверьте состояние кластера перед обновлением:

   ```shell
   stronghold status
   ```

   Необходимо наличие следующих значений:

   ```shell
   Initialized             true
   Sealed                  false
   Version                 1.15.0+ee
   HA Mode                 active
   ```

1. Остановите сервис `stronghold`:

   ```shell
   sudo systemctl stop stronghold
   ```

1. Замените бинарный файл Stronghold EE на бинарный файл Stronghold CSE 1.16.0.

1. Убедитесь, что права на бинарный файл остались прежними, либо выполните:

   ```shell
   sudo chmod 511 /opt/stronghold/stronghold
   sudo chown stronghold:stronghold /opt/stronghold/stronghold
   ```

1. Запустите сервис `stronghold`:

   ```shell
   sudo systemctl start stronghold
   sudo systemctl status stronghold --no-pager
   ```

1. Проверьте, что узел запустился корректно:

   ```shell
   stronghold status
   stronghold version
   ```

   Проверьте, что:

   - сервис `stronghold` находится в состоянии `Running`;
   - ошибки запуска отсутствуют;
   - версия соответствует 1.16.0 (CSE).

1. При необходимости [распечатайте Stronghold](/products/stronghold/documentation/admin/standalone/raft_lost_quorum_recovery/#распечатывание-stronghold).

### Обновление кластера с несколькими узлами

Обновление кластера с несколькими узлами выполняется поочерёдно, чтобы кластер оставался доступным.

1. Определите текущий leader-узел:

   ```shell
   stronghold status
   ```

   В выводе будет указан адрес текущего leader-узла в поле `HA Cluster`:

   ```console
   ...
   HA Cluster              https://10.241.32.36:8201
   ...
   ```

   Зафиксируйте leader-узел.

1. Поочерёдно обновите все follower-узлы. На каждом follower-узле:

   - Остановите сервис:

     ```shell
     sudo systemctl stop stronghold
     ```

   - Замените бинарный файл Stronghold EE на Stronghold CSE 1.16.0.
     При необходимости восстановите права:

     ```shell
     sudo chmod 511 /opt/stronghold/stronghold
     sudo chown stronghold:stronghold /opt/stronghold/stronghold
     ```

   - Запустите сервис:

     ```shell
     sudo systemctl start stronghold
     sudo systemctl status stronghold --no-pager
     ```

   - При необходимости [распечатайте узел Stronghold](/products/stronghold/documentation/admin/standalone/raft_lost_quorum_recovery/#распечатывание-stronghold).

   - Проверьте состояние узла:

     ```shell
     stronghold status
     stronghold version
     ```

   - Проверьте, что:

     - сервис `stronghold` находится в состоянии `Running`;
     - ошибки запуска отсутствуют;
     - версия соответствует 1.16.0 (CSE).

   - Переходите к следующему follower-узлу только после успешной проверки текущего.

1. После обновления всех follower-узлов выполните смену leader-узла:

   ```shell
   stronghold operator step-down
   # Подождите некоторое время.
   stronghold status
   ```

   Ожидаемый результат: leader-узел сменился, в поле `HA Cluster` указано другое значение.

1. Обновите прежний leader-узел:

   ```shell
   sudo systemctl stop stronghold
   # Замените бинарный файл Stronghold EE на Stronghold CSE 1.16.0.
   # При необходимости восстановите права:
   # sudo chmod 511 /opt/stronghold/stronghold
   # sudo chown stronghold:stronghold /opt/stronghold/stronghold
   sudo systemctl start stronghold
   sudo systemctl status stronghold --no-pager
   stronghold status
   stronghold version
   ```

   Проверьте, что:

   - сервис `stronghold` находится в состоянии `Running`;
   - ошибки запуска отсутствуют;
   - версия соответствует 1.16.0 (CSE).

1. При необходимости [распечатайте последний узел Stronghold](/products/stronghold/documentation/admin/standalone/raft_lost_quorum_recovery/#распечатывание-stronghold).

1. Выполните финальную проверку кластера:

   ```shell
   stronghold status
   stronghold version
   ```

   Проверьте, что:

   - все узлы кластера доступны;
   - отсутствуют ошибки репликации и проблемы с кворумом;
   - на узлах установлена версия 1.16.0 (CSE).

## Обновление в исполнении DKP

> Поддерживается обновление с Stronghold EE 1.15.x до Stronghold CSE 1.16.0. Для перехода на Stronghold CSE требуется DKP CSE версии 1.73 и выше.

Перед началом обновления выполните следующие действия:

1. Проверьте текущую версию Stronghold:

   ```shell
   stronghold version
   ```

1. Сохраните unseal-ключи и root-токен в защищённое хранилище. Пример:

   ```shell
   d8 k -n d8-stronghold get secret stronghold-keys -o yaml > stronghold-keys.yaml
   chmod 600 stronghold-keys.yaml
   ```

1. Создайте резервную копию или снимок (snapshot) кластера Stronghold. Пример:

   ```sh
   export STRONGHOLD_ADDR=https://$(d8 k -n d8-stronghold get ing stronghold -o json | jq -r '.spec.rules[0].host')
   d8 stronghold login -method=oidc -path=oidc_deckhouse
   # Либо через root-токен:
   ## d8 stronghold login -method=token
   d8 stronghold operator raft snapshot save stronghold-$(date +%F_%H-%M).snap
   ```

   Проверить снимок можно с помощью команды:

   ```sh
   ls -lh ./stronghold-*.snap
   ```

   > Полученные файлы храните за пределами кластера DKP.

1. Подготовьте пакет или бинарный файл Stronghold CSE 1.16.0 на каждом узле.

1. Убедитесь, что ModuleConfig `stronghold` существует:

   ```shell
   d8 k get mc stronghold -o yaml
   ```

   Если ModuleConfig отсутствует, создайте его:

   
   cat | d8 k apply -f - <<EOF
   apiVersion: deckhouse.io/v1alpha1
   kind: ModuleConfig
   metadata:
     name: stronghold
   spec:
     enabled: false
   EOF
   ```

### Отключение автообновления модуля

Для отключения автообновления модуля выполните следующие действия:

1. Определите текущий канал обновлений модуля `stronghold`:

   ```shell
   d8 k get module stronghold -o jsonpath='{.properties.releaseChannel}'
   ```

   Если команда не выводит значение, используйте канал `stable` в следующем шаге.

1. Создайте ресурс ModuleUpdatePolicy c ручным режимом обновления, указав в поле `releaseChannel` текущий канал обновлений, и примените его:

   
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

   
   NAME                       RELEASE CHANNEL   UPDATE MODE
   stronghold-update-policy   Stable            Manual
   ```

1. Укажите созданный ресурс ModuleUpdatePolicy в ModuleConfig `stronghold`:

   ```shell
   d8 k patch moduleconfig stronghold --type merge --patch '{"spec":{"updatePolicy":"stronghold-update-policy"}}'
   ```

1. Проверьте, что значение сохранилось:

   ```shell
   d8 k get mc stronghold -o jsonpath='{.spec.updatePolicy}'
   ```

   Ожидаемый вывод:

   ```text
   stronghold-update-policy
   ```

1. Убедитесь, что изменения применились в модуле `stronghold`:

   ```shell
   d8 k get module stronghold -o jsonpath='{.properties.updatePolicy}'
   ```

   Ожидаемый вывод:

   ```text
   stronghold-update-policy
   ```

   > Если используется DKP CSE 1.67 и ниже, либо если модуль ещё ни разу не запускался, вывод команды будет пустым. Дополнительные действия в таком случае не требуются.

### Проверки перед началом обновления

1. Проверьте версию DKP. Для перехода на Stronghold CSE требуется DKP CSE версии 1.73 и выше. Проверить версию DKP можно с помощью команды:

   ```shell
   d8 k -n d8-system get configmap d8-deckhouse-version-info -o yaml
   ```

   Также проверить версию DKP можно в веб-интерфейсе Deckhouse на главной странице панели управления кластером (`https://console.<CLUSTER_DOMAIN>`).

   > Если понадобится обновить или переключить редакцию DKP, воспользуйтесь инструкциями:
   >
   > - [инструкция по обновлению DKP](/products/kubernetes-platform/documentation/v1/admin/configuration/update/configuration.html);
   > - [инструкция по переключению DKP EE на DKP CSE](/products/kubernetes-platform/documentation/v1/admin/configuration/registry/switching-editions.html).

1. Убедитесь, что DKP работает штатно, leader-узел определён, очередь пуста:

   ```shell
   d8 k -n d8-system get deploy deckhouse
   d8 k -n d8-system get pod -l leader=true
   d8 system queue list
   ```

   Проверьте, что:

   - Deployment `deckhouse` находится в состоянии `Ready`;
   - есть под в статусе `leader`;
   - очередь задач пуста.

1. Убедитесь, что модуль `stronghold` включён и находится в рабочем состоянии:

   ```shell
   d8 k get module stronghold
   d8 k get moduleconfig stronghold -o yaml
   ```

   Проверьте, что:

   - для `stronghold` значения `Enabled` и `Ready` равны `True`;
   - объект ModuleConfig `stronghold` существует;
   - в `spec.settings.license` указан лицензионный ключ.

1. Убедитесь, что используется Stronghold EE версии 1.15.x. Это можно сделать одним из следующих способов:

   - Через веб-интерфейс Deckhouse — откройте главную страницу панели управления кластером (`https://console.<CLUSTER_DOMAIN>`) и проверьте, что в нижней части страницы указана версия 1.15.x с суффиксом `ee`.

   - По манифесту или образу пода. Для этого выполните команду:

     ```shell
     d8 k -n d8-stronghold get pod -o yaml | grep version
     ```

     В лейблах подов должна быть указана версия 1.15.x с суффиксом `ee`.

   - По логам. Для этого выполните команду:

     ```shell
     d8 k -n d8-stronghold logs stronghold-0 | head -20
     ```

     В стартовых строках баннера должна быть указана версия с редакцией:

     ```console
     Version: Stronghold v1.15.0+ee
     ```

### Подготовка модуля к установке

1. Подготовьте Stronghold CSE в вашем хранилище образов (выполняется на машине с доступом к загрузке образов в хранилище). Проверьте контрольную сумму архива и загрузите бандл модуля Stronghold CSE:

   ```shell
   # Проверьте контрольную сумму архива:
   gost12sum module-stronghold.tar

   # Переместите архив в директорию modules:
   mkdir modules
   mv module-stronghold.tar modules/.

   # Адрес хоста хранилища образов. Например, 10.129.0.18:5000 или my-registry.com
   export REGISTRY_HOST="<REGISTRY_HOST:PORT>"

   # Адрес хранилища образов DKP. Например, 10.129.0.18:5000/dkp-cse/stable
   export MODULES_MODULE_REPO="${REGISTRY_HOST}/<PATH_TO_DKP_REPO>"

   # Используйте учётную запись с правами на запись в хранилище образов.
   d8 mirror push modules $MODULES_MODULE_REPO -u <USERNAME> -p <PASSWORD>
   ```

1. Создайте ModuleSource для Stronghold CSE. Убедитесь средствами вашей инфраструктуры, что хранилище образов доступно с каждого master-узла. Подготовьте и примените ModuleSource:

   ```shell
   # Путь к модулям в хранилище образов.
   MODULES_MODULE_SOURCE="$MODULES_MODULE_REPO/modules"

   # Имя пользователя, имеющего доступ на чтение образов.
   REGISTRY_USER="<REGISTRY_USER>"

   # Пароль пользователя.
   REGISTRY_PASSWORD="<REGISTRY_PASSWORD>"

   # Сертификат УЦ домена, используемого для хранилища образов.
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

1. Проверьте статус ModuleSource:

   ```shell
   d8 k get modulesource stronghold-cse -o yaml
   ```

   Проверьте, что в статусе отсутствуют ошибки, а поле `status.message` пустое:

   ```shell
   ...
   status:
     message: ""
   ...
   ```

### Переключение модуля Stronghold с редакции EE на CSE

Обновите ModuleConfig `stronghold` для использования нового ModuleSource:

1. Проверьте наличие ModuleConfig `stronghold`:

   ```shell
   d8 k get mc stronghold -o yaml
   ```

1. Добавьте в ModuleConfig поле `spec.source` со значением `stronghold-cse`:

   ```shell
   d8 k patch moduleconfig stronghold --type merge --patch '{"spec":{"source":"stronghold-cse"}}'
   d8 k get moduleconfig stronghold -o yaml
   ```

   Проверьте, что изменения применились:

   - в ModuleConfig `stronghold` присутствует значение `spec.source: stronghold-cse`;
   - остальные поля в `spec.settings` не изменились.

1. Если после переключения источника модуль сообщает о проблеме с лицензией, обновите только поле лицензии (`CSE_LICENSE`):

   ```shell
   d8 k patch moduleconfig stronghold --type merge --patch '{"spec":{"settings":{"license":"<CSE_LICENSE>"},"version":1}}'
   ```

1. Если модуль ранее был выключен, включите его:

   ```shell
   d8 system module enable stronghold
   ```

1. Дождитесь стабилизации DKP и пустой очереди:

   ```shell
   d8 k -n d8-system get pods -l app=deckhouse
   d8 system queue list
   ```

   Проверьте, что:

   - под `deckhouse` находится в состоянии `Running`/`Ready`;
   - очередь пуста.

Переключите ModuleUpdatePolicy в режим `AutoPatch`:

1. Измените режим обновления:

   ```shell
   d8 k patch mup stronghold-update-policy --type merge --patch '{"spec":{"update":{"mode":"AutoPatch"}}}'
   ```

1. Проверьте, что режим обновления изменился:

   ```shell
   d8 k get mup stronghold-update-policy
   ```

   Ожидаемый вывод:

   ```text
   NAME                       RELEASE CHANNEL   UPDATE MODE
   stronghold-update-policy   Stable            AutoPatch
   ```

Проверьте завершение перехода на редакцию CSE:

1. Проверьте состояние модуля:

   ```shell
   d8 k get modules stronghold
   ```

   Ожидаемый вывод:

   ```text
   NAME            STAGE                  SOURCE           PHASE   ENABLED   READY
   stronghold      General Availability   stronghold-cse   Ready   True      True
   ```

1. Проверьте источник модуля:

   ```shell
   d8 k get modulesource stronghold-cse -o yaml
   d8 k get moduleconfig stronghold -o yaml
   d8 k get module stronghold -o yaml
   ```

   Проверьте, что:

   - отсутствуют ошибки `auth`, `tls`, `x509`, `timeout`;
   - отсутствуют предупреждающие события;
   - в ModuleConfig `stronghold` указано `spec.source: stronghold-cse`;
   - в Module `stronghold` указано `properties.source: stronghold-cse`.

1. Проверьте поды `stronghold`:

   ```shell
   d8 k -n d8-stronghold get po
   ```

   Проверьте, что:

   - отсутствуют состояния `ImagePullBackOff`, `ErrImagePull`, `CrashLoopBackOff`;
   - поды `stronghold-*` находятся в состоянии `Running` и имеют готовность `2/2`.

### Альтернативный способ через установку нового кластера требуемой версии

Если не удалось пройти [проверки перед началом обновления](#проверки-перед-началом-обновления), то есть привести текущий кластер DKP и модуль `stronghold` к требуемым версиям DKP CSE 1.73 и выше и Stronghold EE 1.15.x, можно перенести кластер Stronghold на новый кластер DKP CSE 1.73 путём восстановления резервной копии.

Для этого выполните следующие действия:

1. Выполните [резервное копирование данных кластера](#обновление-в-исполнении-dkp).
1. Разверните кластер DKP CSE 1.73.
1. Начиная с раздела [Отключение автообновления модуля](#отключение-автообновления-модуля),  последовательно выполните все шаги на новом кластере.
1. Восстановите резервную копию, unseal-ключи и root-токен. Команды выполняются на хосте, где находятся файлы резервной копии `stronghold-*.snap` и ключи `stronghold-keys.yaml`. Пример восстановления резервной копии кластера Stronghold:

   - Восстановите снимок (snapshot):

     ```shell
     export STRONGHOLD_ADDR=https://$(d8 k -n d8-stronghold get ing stronghold -o json | jq -r '.spec.rules[0].host')
     d8 stronghold login -method=oidc -path=oidc_deckhouse
     # Либо через root-токен:
     ## d8 stronghold login -method=token
     d8 stronghold operator raft snapshot restore -force stronghold-<SNAPSHOT_DATE>.snap
     ```

   - Восстановите unseal-ключи и root-токен:

     ```shell
     d8 k -n d8-stronghold delete secret stronghold-keys
     d8 k -n d8-stronghold create -f stronghold-keys.yaml
     ```

   - Проверьте поды `stronghold`:

     ```shell
     d8 k -n d8-stronghold get po
     ```

     Проверьте, что:

     - отсутствуют состояния `ImagePullBackOff`, `ErrImagePull`, `CrashLoopBackOff`;
     - поды `stronghold-*` находятся в состоянии `Running` и имеют готовность `2/2`.
