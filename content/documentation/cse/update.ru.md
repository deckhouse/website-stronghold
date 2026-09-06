---
title: "Обновление Deckhouse Stronghold Certified Security Edition"
description: "Порядок обновления Stronghold CSE в Deckhouse Kubernetes Platform с версии 1.16.0 до 1.16.25."
hidden: true
---

В данном руководстве описан процесс обновления ПО «Deckhouse Stronghold Certified Security Edition» (далее — Stronghold CSE), развёрнутого в Deckhouse Kubernetes Platform, с версии v1.16.0 до v1.16.25.

## Минимальные требования

Для обновления требуется Deckhouse Kubernetes Platform Certified Security Edition (DKP CSE) версии v1.73.4 или новее.

## Процесс обновления

Чтобы обновить Stronghold CSE, выполните следующие шаги:

1. Перед обновлением убедитесь, что в кластере отсутствуют алерты и ошибки в очередях DKP CSE.

   Чтобы проверить состояние очередей, выполните следующую команду:

   ```shell
   d8 system queue list
   ```

   Пример вывода при отсутствии ошибок и активных задач в очереди:

   ```text
   Summary:
   - 'main' queue: empty.
   - 103 other queues (0 active, 103 empty): 0 tasks.
   - no tasks to handle.
   ```

2. Убедитесь, что у вас есть доступ к ключам распечатки Stronghold CSE, которые при установке были сохранены в Secret `d8-stronghold/stronghold-keys`.

3. Загрузите предоставленное вам обновление в ваше хранилище образов контейнеров.

   Скопируйте файлы поставки с USB-флеш накопителя или DVD-диска на компьютер, с которого есть доступ до хранилища образа контейнеров.

   ```shell
   d8 mirror pull \
     ${PACKAGE_DIR_PATH} \
     --no-packages \
     --no-installer \
     --no-security-db \
     --no-platform \
     --include-module stronghold@=v1.16.25 \
     --gost-digest \
     --source="registry-cse.deckhouse.ru/stronghold/cse" \
     --license="${LICENSE_KEY}"
   ```

   Здесь:

   * `${PACKAGE_DIR_PATH}` — директория для сохранения образов;
   * `${LICENSE_KEY}` — лицензионный ключ для доступа в публичное хранилище образов контейнеров Stronghold CSE.

   С помощью команды `d8 tools gostsum` получите контрольные суммы файлов архивов с образами и сравните полученные результаты со значениями в соответствующих файлах с суффиксом `.gostsum`. Контрольные суммы должны совпадать.

   Загрузите данные в хранилище образов контейнеров, где размещён Stronghold CSE, выполнив следующую команду (измените путь к файлу архива или директории образов):

   ```shell
   d8 mirror push <PATH> <REGISTRY_URL>/<REGISTRY_PATH> \
     --registry-login=<USERNAME> --registry-password=<PASSWORD>
   ```

   Здесь:

   `<PATH>` — директория поставки, содержащая архивы с образами поставки Stronghold CSE;
   `<REGISTRY_URL>` — адрес хранилища образов контейнеров в локальной сети;
   `<REGISTRY_PATH>` — путь в хранилище образов контейнеров, в который будут загружаться образы Stronghold CSE. В примерах ниже будет использоваться путь `/stronghold/cse`;
   `<USERNAME>` — имя пользователя для аутентификации в хранилище образов контейнеров;
   `<PASSWORD>` — пароль пользователя для аутентификации в хранилище образов контейнеров.

4. Дождитесь перезапуска подов в неймспейсе `d8-stronghold`.

5. Если после перезапуска Stronghold CSE запечатан, выполните распечатку с помощью сохраненных unseal-ключей:

   ```shell
   d8 k -n d8-stronghold exec stronghold-0 -it -- stronghold operator unseal
   ```

6. После перезапуска подов проверьте версию Stronghold CSE следующей командой:

   ```shell
   d8 stronghold status
   ```

   Пример вывода:

   ```text
   Key                     Value
   ---                     -----
   Seal Type               shamir
   Initialized             true
   Sealed                  false
   Total Shares            1
   Threshold               1
   Version                 1.16.25+ee
   Build Date              2026-08-07T14:52:06Z
   Storage Type            raft
   Cluster Name            stronghold-cluster-8c677db0
   Cluster ID              849ab3e4-261c-b9ec-d857-c542dfadfe01
   HA Enabled              true
   HA Cluster              https://stronghold-0.stronghold-internal:8301
   HA Mode                 active
   Raft Committed Index    325470
   Raft Applied Index      325470
   Last WAL                122365
   ```
