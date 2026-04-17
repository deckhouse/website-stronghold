---
title: "Переключение с EE на CSE"
linkTitle: "Переключение с EE на CSE"
weight: 10
---

**Deckhouse Stronghold Enterprise Edition (EE)** можно переключить на **Deckhouse Stronghold Certified Security Edition (CSE)** одним из следующих способов:

- в исполнении **Standalone**;
- [в исполнении DKP](/products/stronghold/documentation/Руководство по развертыванию/Развертывание в DKP/Переключение редакций/Переключение с EE на CSE/).

> Предупреждение  
> Поддерживается обновление с **Stronghold EE 1.15.x** до **Stronghold CSE 1.16.0**. Если используется версия Stronghold EE ниже `1.15.x`, сначала [обновитесь до последней версии ветки](/products/stronghold/documentation/admin/update/stanalone-update/).

> Предупреждение  
> При переключении на Stronghold CSE возможна временная недоступность сервиса.

## Переключение в Standalone

Перед началом переключения выполните следующие действия:

1. Проверьте текущую версию Stronghold:

   ```bash
   stronghold version
   ```

1. Сохраните unseal-ключи и root-токен в защищённое хранилище.

1. Создайте снимок кластера Stronghold:

   ```bash
   stronghold operator raft snapshot save stronghold-$(date +%F_%H-%M).snap
   ls -lh stronghold-*.snap
   ```

1. Убедитесь, что снимок создан:

   ```bash
   ls -lh ./stronghold-*.snap
   ```

   > Важно  
   > Храните файлы снимков вне кластера Stronghold и вне узлов, на которых он запущен.

1. Подготовьте пакет или бинарный файл **Stronghold CSE 1.16.0** на каждом узле.

## Переключение кластера с одним узлом

Для переключения кластера с одним узлом выполните следующие действия:

1. Проверьте состояние кластера перед переключением:

   ```bash
   stronghold status
   ```

   Убедитесь, что в выводе присутствуют следующие значения:

   ```text
   Initialized             true
   Sealed                  false
   Version                 1.15.0+ee
   HA Mode                 active
   ```

1. Остановите сервис `stronghold`:

   ```bash
   sudo systemctl stop stronghold
   ```

1. Замените бинарный файл Stronghold EE на бинарный файл Stronghold CSE `1.16.0`.

1. Убедитесь, что права на бинарный файл сохранились. При необходимости восстановите их:

   ```bash
   sudo chmod 511 /opt/stronghold/stronghold
   sudo chown stronghold:stronghold /opt/stronghold/stronghold
   ```

1. Запустите сервис `stronghold`:

   ```bash
   sudo systemctl start stronghold
   sudo systemctl status stronghold --no-pager
   ```

1. Проверьте, что узел запустился корректно:

   ```bash
   stronghold status
   stronghold version
   ```

   Убедитесь, что:
   - сервис `stronghold` находится в состоянии `running`;
   - ошибки запуска отсутствуют;
   - версия соответствует `1.16.0` (`CSE`).

1. При необходимости [распечатайте Stronghold](/products/stronghold/documentation/Руководство администратора/Надёжность и восстановление/Восстановление после потери кворума/).

## Переключение кластера с несколькими узлами

Переключение кластера с несколькими узлами выполняется поочерёдно, чтобы кластер оставался доступным.

1. Определите текущий active-узел:

   ```bash
   stronghold status
   ```

   В выводе будет указан адрес текущего active-узла в поле `HA Cluster`:

   ```text
   ...
   HA Cluster              https://10.241.32.36:8201
   ...
   ```

   Зафиксируйте active-узел.

1. Поочерёдно переключите все standby-узлы. На каждом standby-узле:

   - Остановите сервис:

     ```bash
     sudo systemctl stop stronghold
     ```

   - Замените бинарный файл Stronghold EE на Stronghold CSE `1.16.0`.

   - При необходимости восстановите права:

     ```bash
     sudo chmod 511 /opt/stronghold/stronghold
     sudo chown stronghold:stronghold /opt/stronghold/stronghold
     ```

   - Запустите сервис:

     ```bash
     sudo systemctl start stronghold
     sudo systemctl status stronghold --no-pager
     ```

   - При необходимости [распечатайте узел Stronghold](/products/stronghold/documentation/Руководство администратора/Надёжность и восстановление/Восстановление после потери кворума/).

   - Проверьте состояние узла:

     ```bash
     stronghold status
     stronghold version
     ```

   - Убедитесь, что:
     - сервис `stronghold` находится в состоянии `running`;
     - ошибки запуска отсутствуют;
     - версия соответствует `1.16.0` (`CSE`).

   - Переходите к следующему standby-узлу только после успешной проверки текущего.

1. После переключения всех standby-узлов выполните смену active-узла:

   ```bash
   stronghold operator step-down
   stronghold status
   ```

   Ожидаемый результат: active-узел сменился, в поле `HA Cluster` указано другое значение.

1. Переключите прежний active-узел:

   ```bash
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

   Убедитесь, что:
   - сервис `stronghold` находится в состоянии `running`;
   - ошибки запуска отсутствуют;
   - версия соответствует `1.16.0` (`CSE`).

1. При необходимости [распечатайте последний узел Stronghold](/products/stronghold/documentation/Руководство администратора/Надёжность и восстановление/Восстановление после потери кворума/).

1. Выполните финальную проверку кластера:

   ```bash
   stronghold status
   stronghold version
   ```

   Убедитесь, что:
   - все узлы кластера доступны;
   - отсутствуют ошибки репликации и проблемы с кворумом;
   - на всех узлах установлена версия `1.16.0` (`CSE`).

## Результат

После успешного выполнения шагов Stronghold в исполнении Standalone будет переключён с **Enterprise Edition** на **Certified Security Edition**.

Если требуется дополнительная проверка состояния сервиса после переключения, перейдите в раздел [Проверка работоспособности](../../../deployment/functionality-check/).
