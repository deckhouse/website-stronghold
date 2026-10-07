---
title: "Обновление"
description: "Порядок обновления кластера Stronghold в режиме HA с хранилищем Raft в Linux без простоя: резервная копия, обновление standby-узлов, переключение активного узла и проверка версий."
weight: 30
---

Кластер Stronghold в режиме HA с хранилищем Raft обновляется поочерёдно, узел за узлом. Сначала обновляются standby-узлы, затем активный узел передаёт роль одному из обновлённых узлов и обновляется последним. Такой порядок сохраняет кворум и доступность кластера на время обновления.

<!-- TODO(verify): поддерживаемые пути обновления (можно ли пропускать версии) и возможность отката на предыдущую версию после обновления. -->

## Подготовка

1. Ознакомьтесь с [примечаниями к выпуску](../../../release-notes/) для всех версий между текущей и целевой. Обратите внимание на изменения конфигурации и несовместимые изменения.
1. Проверьте версию и состояние кластера:

   ```shell
   d8 stronghold status
   d8 stronghold operator raft list-peers
   d8 stronghold operator raft autopilot state
   ```

   Все узлы должны быть в состоянии `voter`, Autopilot должен сообщать `Healthy: true`.

1. Создайте снимок хранилища и проверьте его (см. [«Создание снимка»](../../../admin/backups/save/) и [«Проверка снимка»](../../../admin/backups/inspect/)):

   ```shell
   d8 stronghold operator raft snapshot save pre-upgrade.snap
   d8 stronghold operator raft snapshot inspect pre-upgrade.snap
   ```

1. Сохраните копии конфигурационного файла, systemd-unit и текущего бинарного файла на каждом узле:

   ```shell
   cp -a /opt/stronghold/stronghold /opt/stronghold/stronghold.bak
   cp -a /opt/stronghold/config.hcl /opt/stronghold/config.hcl.bak
   ```

1. Убедитесь, что держатели ключей распечатывания доступны, если используется Shamir seal: каждый узел после перезапуска потребуется распечатать.
1. Скопируйте бинарный файл новой версии на все узлы, например в `/tmp/stronghold`, и проверьте его версию:

   ```shell
   /tmp/stronghold version
   ```

1. Определите активный узел: в выводе `d8 stronghold operator raft list-peers` у него состояние `leader`.

## Обновление standby-узлов

Выполните шаги на каждом standby-узле по очереди. Не переходите к следующему узлу, пока текущий не вернулся в кластер.

1. Замените бинарный файл:

   ```shell
   install -o root -g root -m 0755 /tmp/stronghold /opt/stronghold/stronghold
   ```

1. Перезапустите сервис:

   ```shell
   systemctl restart stronghold
   ```

1. Если используется Shamir seal, распечатайте узел. Укажите адрес обновляемого узла:

   ```shell
   export STRONGHOLD_ADDR=https://raft-node-2.demo.tld:8200
   d8 stronghold operator unseal
   ```

1. Проверьте, что узел вернулся в кластер и распечатан:

   ```shell
   d8 stronghold status
   d8 stronghold operator raft autopilot state
   ```

   Дождитесь, пока Autopilot покажет узел как `healthy`, а кластер — `Healthy: true`.

1. Проверьте [серверные журналы](../../../admin/operations/logs/) узла на ошибки:

   ```shell
   journalctl -u stronghold.service --since "10 minutes ago"
   ```

## Обновление активного узла

1. Передайте роль активного узла одному из обновлённых узлов. Выполните команду, указав адрес текущего активного узла:

   ```shell
   export STRONGHOLD_ADDR=https://raft-node-1.demo.tld:8200
   d8 stronghold operator step-down
   ```

1. Убедитесь, что лидером стал другой узел:

   ```shell
   d8 stronghold operator raft list-peers
   ```

1. Обновите бывший активный узел так же, как standby-узлы: замените бинарный файл, перезапустите сервис, распечатайте узел и дождитесь его возвращения в кластер.

## Проверка после обновления

1. Проверьте версию на каждом узле:

   ```shell
   for node in raft-node-1 raft-node-2 raft-node-3; do
     STRONGHOLD_ADDR=https://${node}.demo.tld:8200 d8 stronghold status | grep -E 'Version|HA Mode|Sealed'
   done
   ```

1. Проверьте состояние кластера:

   ```shell
   d8 stronghold operator raft list-peers
   d8 stronghold operator raft autopilot state
   ```

1. Проверьте основные сценарии: вход через используемые методы аутентификации, чтение и запись тестового секрета, выдачу динамических учётных данных.
1. Проверьте, что метрики и алерты работают (см. [«Мониторинг»](../../../admin/operations/monitoring/)).

В выводе `autopilot state` для каждого узла указана версия (`Version`), по ней можно убедиться, что все узлы обновлены. Автоматическое обновление Autopilot (upgrade migration) в исходном коде помечено как функция редакции Enterprise и в этой процедуре не используется.

## Откат

Если после обновления узел не запускается или кластер работает некорректно:

1. Остановите обновлённый узел и верните прежний бинарный файл из копии `/opt/stronghold/stronghold.bak`.
1. Если данные хранилища уже были изменены новой версией и прежняя версия не запускается, восстановите кластер из снимка `pre-upgrade.snap` на прежней версии (см. [«Восстановление из снимка»](../../../admin/backups/restore/)).

{{< alert level="warning" >}}
Восстановление из снимка возвращает данные к моменту его создания. Изменения, сделанные после создания снимка, будут потеряны.
{{< /alert >}}
