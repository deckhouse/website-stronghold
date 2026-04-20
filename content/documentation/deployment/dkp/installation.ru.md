---
title: "Установка Stronghold в DKP"
linkTitle: "Установка Stronghold в DKP"
description: "Установка Deckhouse Stronghold в существующий кластер Deckhouse Kubernetes Platform"
weight: 20
---

В этом разделе описана установка **Deckhouse Stronghold** в **существующий кластер Deckhouse Kubernetes Platform**.

Установка выполняется через включение модуля `stronghold` с помощью ресурса `ModuleConfig`.

> Примечание  
> В этом разделе предполагается, что кластер **Deckhouse Kubernetes Platform** уже установлен и работает корректно. Если кластер DKP ещё не развернут, используйте [документация Deckhouse Kubernetes Platform](/products/kubernetes-platform/documentation/) или воспользуйтесь [быстрым стартом](/products/kubernetes-platform/gs/).

## Перед началом

Перед установкой убедитесь, что:

- выполнены требования из раздела [Требования к окружению](../environment-requirements/);
- у вас есть доступ к кластеру через `d8` и `kubectl`;
- кластер находится в работоспособном состоянии;
- узлы имеют доступ к хранилищу образов контейнеров.

Проверьте доступ к кластеру:

```bash
d8 k get nodes
```

Если команда выполняется успешно и возвращает список узлов, можно переходить к установке Stronghold.

## Шаг 1. При необходимости проверьте Ingress-контроллер и DNS

Этот шаг особенно важен, если вы планируете использовать **веб-интерфейс Stronghold**.

Проверьте, что Ingress-контроллер запущен:

```bash
d8 k -n d8-ingress-nginx get po
```

Дождитесь перехода подов в состояние `Ready`.

Если Stronghold будет публиковаться через внешний адрес, также проверьте сервис балансировщика:

```bash
d8 k -n d8-ingress-nginx get svc nginx-load-balancer
```

Убедитесь, что значение `EXTERNAL-IP` заполнено.

Если в кластере используется доступ по DNS-именам, заранее проверьте корректность настройки `publicDomainTemplate`.

> Примечание  
> Если Stronghold будет использоваться только через внутренние сервисные интеграции, CLI или API, внешний доступ через Ingress может не потребоваться.

## Шаг 2. Включите модуль Stronghold

Для установки Stronghold создайте ресурс `ModuleConfig` с именем `stronghold`.

Минимальный рабочий пример:

```bash
d8 k apply -f - <<EOF
apiVersion: deckhouse.io/v1alpha1
kind: ModuleConfig
metadata:
  name: stronghold
spec:
  enabled: true
  version: 1
  settings:
    management:
      mode: Automatic
EOF
```

После применения ресурса Deckhouse Kubernetes Platform начнёт установку компонентов Stronghold в кластер.

## Шаг 3. Дождитесь применения конфигурации

После включения модуля дождитесь завершения обработки заданий:

```bash
d8 system queue main
```

Пример ожидаемого состояния:

```text
Queue 'main': length 0, status: 'waiting for task 1m1s'
```

Если очередь пуста и новые задачи не появляются, можно переходить к проверке результата.

## Шаг 4. Проверьте создание namespace и компонентов Stronghold

После успешного включения модуля должен появиться namespace `d8-stronghold`.

Проверьте это командой:

```bash
d8 k get ns d8-stronghold
```

Пример ожидаемого результата:

```text
NAME            STATUS   AGE
d8-stronghold   Active   1m
```

Затем проверьте поды Stronghold:

```bash
d8 k -n d8-stronghold get po
```

Убедитесь, что поды созданы и переходят в состояние `Running` или `Ready`.

## Шаг 5. При необходимости настройте доступ к веб-интерфейсу

Если планируется работа через веб-интерфейс, после установки убедитесь, что:

- опубликован маршрут доступа к Stronghold;
- корректно работает Ingress;
- DNS-имя, если оно используется, разрешается в нужный адрес;
- сертификаты и TLS-конфигурация соответствуют политике безопасности вашей инфраструктуры.

Конкретные параметры доступа и первичной настройки Stronghold в DKP приведены в следующем разделе.

## Интеграция с приложениями через secrets-store-integration

Модуль `secrets-store-integration` **не является строго обязательным** для работы **Deckhouse Kubernetes Platform** и **Deckhouse Stronghold**. Stronghold можно установить и использовать в DKP без него.

Однако модуль `secrets-store-integration` **обязателен**, если ваша задача — безопасная и автоматизированная доставка секретов из Stronghold в приложения, работающие в Kubernetes.

Этот модуль выступает связующим звеном между Stronghold и потребителями секретов в кластере.

Без него вы всё равно сможете:
- использовать Stronghold как хранилище секретов;
- работать с ним через API и CLI;
- интегрировать приложения с Stronghold напрямую.

Но без `secrets-store-integration` будут недоступны ключевые сценарии автоматизации в Kubernetes:

- **доставка через CSI-драйвер** — монтирование секретов в под как файлов без сохранения в `etcd` и без публикации через API Kubernetes;
- **инъекция через Mutating Webhook** — автоматическая подстановка секретов в переменные окружения контейнеров при старте;
- **синхронизация с native Secrets** — автоматическое создание стандартных секретов Kubernetes на основе данных из Stronghold.

> Важно  
> Если ваша цель — безопасная доставка секретов в приложения без их хранения в стандартных секретах Kubernetes, установка `secrets-store-integration` обязательна.

Если этот сценарий требуется, включите модуль отдельно:

```bash
d8 k apply -f - <<EOF
apiVersion: deckhouse.io/v1alpha1
kind: ModuleConfig
metadata:
  name: secrets-store-integration
spec:
  enabled: true
  version: 1
EOF
```

После этого дождитесь завершения обработки очереди:

```bash
d8 system queue main
```

## Результат установки

Если установка выполнена успешно:

- в кластере включён модуль `stronghold`;
- создан namespace `d8-stronghold`;
- компоненты Stronghold запущены;
- кластер готов к дальнейшей настройке;
- при необходимости можно дополнительно включить `secrets-store-integration` для интеграции с приложениями.

## Что дальше

После установки перейдите к разделу [Первичная настройка](../configuration/).
