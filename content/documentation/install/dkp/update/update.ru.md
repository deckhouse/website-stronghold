---
title: "Обновление платформы"
description: "Как поступают обновления Deckhouse Kubernetes Platform и модуля stronghold, проверка текущей версии, подготовка к обновлению, проверка после обновления и ограничения отката."
weight: 10
---

Stronghold в Deckhouse Platform (DP) работает как модуль `stronghold` и обновляется механизмом обновлений DP. Новые версии модуля поступают по [каналам обновлений](../../release-channels/) и применяются в соответствии с настроенным режимом обновления.

## Как поступают обновления

- Версии модуля `stronghold` публикуются в каналах обновлений `Alpha`, `Beta`, `Early Access`, `Stable` и `Rock Solid` (см. [«Каналы обновлений»](../../release-channels/)).
- DP создаёт в кластере ресурс ModuleRelease для каждой доступной версии модуля и применяет её автоматически, в окно обновлений или после ручного подтверждения.
- Канал и режим обновления модуля по умолчанию определяются настройками модуля `deckhouse`. <!-- TODO(verify): наследует ли модуль stronghold канал и режим обновления модуля deckhouse, если ModuleUpdatePolicy не задан. --> Чтобы задать их отдельно для модуля `stronghold`, используйте ресурс ModuleUpdatePolicy (см. [«Ручное подтверждение обновлений модуля stronghold»](#ручное-подтверждение-обновлений-модуля-stronghold)).

Список версий модуля в кластере:

```shell
d8 k get modulereleases | grep stronghold
```

Изменения в каждой версии описаны в [истории изменений](../../../../release-notes/). Перед обновлением проверьте требования к версии DP: например, для Stronghold `1.19` требуется DP версии `1.76` или новее.

## Проверка текущей версии

Проверьте версию Stronghold одним из способов:

- Выполните команду:

  ```shell
  d8 stronghold status
  ```

  Версия указана в поле `Version`.

- Проверьте лейблы подов:

  ```shell
  d8 k -n d8-stronghold get pod -o yaml | grep version
  ```

- Проверьте стартовый баннер в логах:

  ```shell
  d8 k -n d8-stronghold logs stronghold-0 | head -20
  ```

  Пример строки с версией и редакцией:

  ```console
  Version: Stronghold v1.15.0+ee
  ```

Текущий канал обновлений модуля:

```shell
d8 k get module stronghold -o jsonpath='{.properties.releaseChannel}'
```

Если команда не выводит значение, канал обновлений для модуля отдельно не задан.

## Подготовка к обновлению

Перед обновлением выполните следующие действия:

1. Убедитесь, что в кластере отсутствуют алерты и очередь DP пуста:

   ```shell
   d8 system queue list
   ```

1. Убедитесь, что модуль `stronghold` находится в состоянии `Ready`:

   ```shell
   d8 k get modules stronghold
   ```

1. Проверьте состояние кластера Stronghold:

   ```shell
   d8 stronghold status
   d8 stronghold operator raft list-peers
   ```

   Значение `Sealed` должно быть равно `false`, все узлы Raft должны присутствовать в списке.

1. Сохраните ключи распечатывания и root-токен в защищённое хранилище за пределами кластера:

   ```shell
   d8 k -n d8-stronghold get secret stronghold-keys -o yaml > stronghold-keys.yaml
   chmod 600 stronghold-keys.yaml
   ```

1. Создайте снимок хранилища и проверьте его (см. [«Создание снимка»](../../../../admin/backups/save/) и [«Проверка снимка»](../../../../admin/backups/inspect/)):

   ```shell
   d8 stronghold operator raft snapshot save stronghold-$(date +%F_%H-%M).snap
   d8 stronghold operator raft snapshot inspect stronghold-<SNAPSHOT_DATE>.snap
   ```

   Храните снимок и файл с ключами за пределами кластера DP.

1. Ознакомьтесь с [историей изменений](../../../../release-notes/) для всех версий между текущей и целевой.

## Обновление платформы

Обновление платформы конфигурируется в ресурсе ModuleConfig [`deckhouse`](/modules/deckhouse/configuration.html).

Посмотреть текущую конфигурацию настроек обновления можно с помощью команды:

```shell
d8 k get mc deckhouse -oyaml
```

Пример вывода:

```yaml
# ...
spec:
  settings:
    releaseChannel: Stable
    update:
      windows:
        - days:
            - Mon
          from: "19:00"
          to: "20:00"
# ...
```

### Настройка режима обновления

Платформа поддерживает три режима обновления:

- **Автоматический + окна обновлений не заданы.** Кластер обновится сразу после появления новой версии на соответствующем [канале обновлений](../../release-channels/).
- **Автоматический + заданы окна обновлений.** Кластер обновится в ближайшее доступное окно после появления новой версии на канале обновлений.
- **Ручной режим.** Для применения обновления требуются [ручные действия](../manual-update-mode/).

Пример фрагмента конфигурации для включения автоматического обновления платформы:

```yaml
update:
  mode: Auto
```

Пример фрагмента конфигурации для включения автоматического обновления платформы с окнами обновлений:

```yaml
update:
  mode: Auto
  windows:
    - from: "8:00"
      to: "15:00"
      days:
        - Tue
        - Sat
```

Пример фрагмента конфигурации для включения ручного режима обновления платформы:

```yaml
update:
  mode: Manual
```

### Каналы обновлений

Платформа использует [пять каналов обновлений](../../release-channels/), предназначенных для использования в разных окружениях. Компоненты платформы могут обновляться автоматически, либо с ручным подтверждением по мере выхода обновлений в каналах обновления.

Информацию по версиям, доступным на каналах обновления, можно получить на сайте [https://releases.deckhouse.ru/](https://releases.deckhouse.ru/).

Чтобы перейти на другой канал обновлений, в конфигурации модуля `deckhouse` нужно установить параметр `.spec.settings.releaseChannel`.

Пример конфигурации модуля `deckhouse` с установленным каналом обновлений `Stable`:

```yaml
apiVersion: deckhouse.io/v1alpha1
kind: ModuleConfig
metadata:
  name: deckhouse
spec:
  version: 1
  settings:
    releaseChannel: Stable
```

- При смене канала обновлений на **более стабильный** (например, с `Alpha` на `EarlyAccess`) Deckhouse скачивает данные о релизе (в примере — из канала `EarlyAccess`) и сравнивает их с данными из существующих в кластере ресурсов `DeckhouseRelease`:
  - Более _поздние_ релизы, которые еще не были применены (в статусе `Pending`), удаляются.
  - Если более _поздние_ релизы уже применены (в статусе `Deployed`), смены релиза не происходит. В этом случае платформа останется на таком релизе до тех пор, пока на канале обновлений `EarlyAccess` не появится более поздний релиз.
- При смене канала обновлений на **менее стабильный** (например, с `EarlyAcess` на `Alpha`):
  - Deckhouse скачивает данные о релизе (в примере — из канала `Alpha`) и сравнивает их с данными из существующих в кластере ресурсов `DeckhouseRelease`.
  - Затем платформа выполняет обновление согласно установленным параметрам обновления.

Посмотреть список релизов платформы можно с использованием следующих команд:

```shell
d8 k get deckhouserelease
d8 k get modulereleases
```

{{% details summary="Схема использования параметра releaseChannel при установке и в процессе работы платформы" %}}
![Схема использования параметра releaseChannel при установке и в процессе работы платформы](/images/common/deckhouse-update-process.png)
{{% /details %}}

Для отключения механизма обновления платформы, удалите в конфигурации модуля `deckhouse` параметр `.spec.settings.releaseChannel`. В этом случае платформа не проверяет обновления и обновление на patch-релизы не выполняется.

{{< alert level="danger" >}}
Крайне не рекомендуется отключать автоматическое обновление! Это заблокирует обновления на patch-релизы, которые могут содержать исправления критических уязвимостей и ошибок.
{{< /alert >}}

### Немедленное применение обновлений

Чтобы применить обновление немедленно, установите в соответствующем ресурсе [DeckhouseRelease](/products/kubernetes-platform/documentation/v1/reference/api/cr.html#deckhouserelease) аннотацию `release.deckhouse.io/apply-now: "true"`.

{{< alert level="info" >}}
**Обратите внимание!** В этом случае будут проигнорированы окна обновления, настройки [canary-release](/products/kubernetes-platform/documentation/v1/reference/api/cr.html#deckhouserelease-v1alpha1-spec-applyafter) и режим [ручного обновления кластера](../manual-update-mode/). Обновление применится сразу после установки аннотации.
{{< /alert >}}

Пример команды установки аннотации пропуска окон обновлений для версии `v1.56.2`:

```shell
d8 k annotate deckhousereleases v1.56.2 release.deckhouse.io/apply-now="true"
```

Пример ресурса с установленной аннотацией пропуска окон обновлений:

```yaml
apiVersion: deckhouse.io/v1alpha1
kind: DeckhouseRelease
metadata:
  annotations:
    release.deckhouse.io/apply-now: "true"
```

## Ручное подтверждение обновлений модуля stronghold

Чтобы обновлять модуль `stronghold` только после ручного подтверждения, независимо от режима обновления платформы:

1. Создайте ресурс ModuleUpdatePolicy с ручным режимом обновления, указав в поле `releaseChannel` нужный канал обновлений:

   ```yaml
   apiVersion: deckhouse.io/v1alpha2
   kind: ModuleUpdatePolicy
   metadata:
     name: stronghold-update-policy
   spec:
     releaseChannel: Stable
     update:
       mode: Manual
   ```

1. Укажите созданный ресурс в ModuleConfig `stronghold`:

   ```shell
   d8 k patch moduleconfig stronghold --type merge --patch '{"spec":{"updatePolicy":"stronghold-update-policy"}}'
   ```

1. Убедитесь, что политика применилась:

   ```shell
   d8 k get module stronghold -o jsonpath='{.properties.updatePolicy}'
   ```

1. Когда появится новая версия, подтвердите соответствующий ModuleRelease. <!-- TODO(verify): команда подтверждения ModuleRelease (аннотация modules.deckhouse.io/approved="true" или поле approved) для текущей версии DKP. -->

Чтобы автоматически получать только patch-версии, переключите политику в режим `AutoPatch`:

```shell
d8 k patch mup stronghold-update-policy --type merge --patch '{"spec":{"update":{"mode":"AutoPatch"}}}'
```

## Проверка после обновления

1. Проверьте состояние модуля:

   ```shell
   d8 k get modules stronghold
   ```

   Значения `ENABLED` и `READY` должны быть равны `True`.

1. Проверьте поды:

   ```shell
   d8 k -n d8-stronghold get pods
   ```

   Поды `stronghold-*` должны находиться в состоянии `Running` и иметь готовность `2/2`. Состояния `ImagePullBackOff`, `ErrImagePull` и `CrashLoopBackOff` должны отсутствовать.

1. Проверьте версию и состояние Stronghold:

   ```shell
   d8 stronghold status
   d8 stronghold operator raft list-peers
   ```

   В поле `Version` должна быть указана целевая версия, значение `Sealed` должно быть равно `false`.

1. Проверьте вход и чтение тестового секрета, а также работу приложений, получающих секреты из Stronghold.

## Ограничения отката

- Механизм обновлений DP не откатывает модуль на предыдущую версию автоматически. <!-- TODO(verify): поддерживается ли понижение версии модуля stronghold и как его выполнить. -->
- Основной способ вернуться к состоянию до обновления — восстановление из снимка, созданного на этапе подготовки (см. [«Восстановление из снимка»](../../../../admin/backups/restore/)). <!-- TODO(verify): можно ли восстанавливать снимок, созданный на более старой версии, в кластер на более новой версии и наоборот. -->
- Данные, записанные после создания снимка, при восстановлении будут потеряны. Планируйте обновление так, чтобы между созданием снимка и обновлением прошло минимальное время.
- Если планируете переход между редакциями, выполняйте его отдельно от обновления версии (см. [«Переключение Stronghold с EE на CSE»](../../platform-management/switching-editions/ee-to-cse/)).
