---
title: "Плагины в DKP"
weight: 30
description: "Подключение плагинов Stronghold в Deckhouse Kubernetes Platform."
---

В Deckhouse Kubernetes Platform порядок загрузки плагинов отличается от Standalone-установки: администратор не копирует бинарные файлы на сервер вручную, а описывает список плагинов в `ModuleConfig`.

После этого платформа:

- скачивает плагины по указанным URL;
- проверяет их контрольные суммы;
- помещает плагины в контейнер Stronghold;
- перезапускает Stronghold при изменении списка плагинов.

После загрузки плагин всё равно нужно зарегистрировать в Stronghold и затем включить по нужному пути.

## Настройка списка плагинов

Список загружаемых плагинов задаётся в `ModuleConfig`.

Пример:

```yaml
apiVersion: deckhouse.io/v1alpha1
kind: ModuleConfig
metadata:
  name: stronghold
spec:
  enabled: true
  version: 1
  settings:
    plugins:
      - name: "vault-plugin-secrets-github"
        url: "https://github.com/martinbaillie/vault-plugin-secrets-github/releases/download/v2.3.2/vault-plugin-secrets-github-linux-amd64"
        sha256: "72cb1f2775ee2abf12ffb725e469d0377fe7bbb93cd7aaa6921c141eddecab87"
      - name: "vault-plugin-auth-any"
        url: "https://plugins.example.local/myplugins/vault-plugin-auth-any-v1.0.0-linux-amd64"
        sha256: "c943b505b39b53e1f4cb07f2a3455b59eac523ebf600cb04813b9ad28a848b21"
        ignoreFailure: true
        insecureSkipVerify: false
        ca: |
          -----BEGIN CERTIFICATE-----
          MIIDDTCCAfWgAwIBAgIJAOb7PcmW8W9MMA0GCSqGSIb3DQEBCwUAMBQxEjAQBgNV
          BAMTCWxvY2FsaG9zdDAeFw0yNjA1MjAwMDAwMDBaFw0yNjA2MjAwMDAwMDBaMBQx
          EjAQBgNVBAMTCWxvY2FsaG9zdDCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoC
          ggEBAKHh4g5i1R+3+9XdG0RFLiX1x5T2PvQ92E/78vR6+Bn09+G0P+C6143+wLn
          j96/E8rHbHr4R6L0f62/OJZh8JnZ/qRqE1N8oNc06Vh9Y7X8EzF4nZ4KgX/3y6L
          vXD251Qm7g==
          -----END CERTIFICATE-----
```

## Параметры загрузки

- `name` — имя бинарного файла плагина;
- `url` — URL, откуда нужно скачать плагин;
- `sha256` — SHA256 контрольная сумма плагина;
- `ignoreFailure` — разрешает продолжить запуск Stronghold, даже если данный плагин не удалось скачать;
- `insecureSkipVerify` — отключает проверку TLS-сертификата удалённого сервера;
- `ca` — дополнительный CA-сертификат для проверки TLS-соединения.

## Что происходит при обновлении списка плагинов

Добавление или удаление плагинов вызывает перезапуск Stronghold.

Если плагин не удаётся скачать или провалидировать:

- перезапуск Stronghold блокируется;
- исключение составляют плагины с `ignoreFailure: true`;
- если не проходит проверка контрольной суммы, плагин считается не загруженным и удаляется.

## Закрытые контуры

В закрытых контурах, где у Stronghold нет доступа во внешний интернет, плагин можно разместить внутри самого кластера DKP.

Один из практических вариантов:

1. Подготовить контейнер с `nginx`, в который помещён бинарный файл плагина.
2. Запустить этот контейнер в Kubernetes.
3. Создать внутренний `Service`, который публикует `nginx` внутри кластера.
4. Указать в `ModuleConfig` URL вида `http://<service>.<namespace>.svc.cluster.local/...`, чтобы платформа скачивала плагин через внутренний Kubernetes-сервис.

Пример фрагмента `ModuleConfig`:

```yaml
spec:
  settings:
    plugins:
      - name: "vault-plugin-auth-any"
        url: "http://plugin-repo.plugins.svc.cluster.local/vault-plugin-auth-any"
        sha256: "c943b505b39b53e1f4cb07f2a3455b59eac523ebf600cb04813b9ad28a848b21"
```

Такой подход позволяет:

- не открывать исходящий доступ в интернет;
- хранить плагины внутри внутреннего контура;
- централизованно управлять версиями плагинов через внутренний сервис.

## Регистрация плагина

После доставки плагина в контейнер его нужно зарегистрировать через CLI:

```bash
PLUGIN_SHA=$(sha256sum <plugin_binary> | awk '{print $1;}')

d8 stronghold plugin register \
  -command <command_to_run_plugin_binary> \
  -sha256 "${PLUGIN_SHA}" \
  -version "<semantic_version>" \
  <plugin_type> \
  <plugin_name>
```

Пример регистрации secret-плагина `mykv`:

```bash
d8 stronghold plugin register \
  -command mykvplugin \
  -sha256 "${PLUGIN_SHA}" \
  -version "v1.0.1" \
  secret \
  mykv
```

## Включение плагина

После регистрации плагин можно включить как `secret` или `auth` engine:

```bash
d8 stronghold <secrets|auth> enable \
  -path <mount_path> \
  <plugin_name>
```

Пояснение:

- `secrets` — для плагинов типа `secret`;
- `auth` — для плагинов аутентификации;
- `-path` — путь монтирования;
- `plugin_name` — имя, под которым плагин зарегистрирован.

Пример:

```bash
d8 stronghold secrets enable -path test-kv mykv
```

## Отключение и удаление плагина

1. Отключите все `secret` и `auth` методы, использующие плагин.
1. Снимите плагин с регистрации:

```bash
d8 stronghold plugin deregister secret my-custom-plugin
```

1. Удалите плагин из конфигурации `ModuleConfig`.

## Практические рекомендации

- Для production желательно использовать внутренний репозиторий плагинов или доверенный артефактный storage.
- Всегда задавайте `sha256` и проверяйте, что он соответствует реальному бинарному файлу.
- Используйте `ignoreFailure` только для необязательных плагинов.
- Помните, что изменение списка плагинов приводит к перезапуску Stronghold.

## См. также

- [Плагины в Standalone](./standalone/)
