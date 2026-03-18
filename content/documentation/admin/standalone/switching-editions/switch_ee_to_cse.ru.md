---
title: "Переключение Stronghold с EE на CSE"
weight: 20
url: /documentation/admin/standalone/switching-editions/switch-ee-to-cse/
---

Stronghold Enterprise Edition (EE) можно обновить до Stronghold Certified Security Edition (CSE) одним из следующих способов:

- в исполнении Standalone;
- [в исполнении DKP](/products/stronghold/documentation/admin/platform-management/switching-editions/switch-ee-to-cse/).

> Поддерживается обновление с Stronghold EE 1.15.x до Stronghold CSE 1.16.0. Если используется версия Stronghold EE ниже 1.15.x, сначала [обновитесь до последней версии ветки](/products/stronghold/documentation/admin/update/update/) 1.15.x.
>
> При обновлении до Stronghold CSE возможна временная недоступность сервиса.

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
