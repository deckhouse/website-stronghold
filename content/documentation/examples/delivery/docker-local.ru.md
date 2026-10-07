---
title: "Локальный Stronghold для разработки"
linkTitle: "Локальный запуск в Docker"
description: "Запуск Stronghold в Docker на рабочей станции разработчика: режим разработки server -dev и одноузловая конфигурация с Raft в Docker Compose."
weight: 80
---

Для разработки и отладки интеграций можно запустить Stronghold локально в Docker. Используйте такой экземпляр только для разработки: не храните в нём реальные секреты.

## Сборка образа

Образ собирается командой [`stronghold bootstrap docker`](../../../install/standalone/bootstrap/#сборка-docker-образа) из бинарного файла Stronghold:

```bash
stronghold bootstrap docker
docker load -i stronghold-v1.19.0.tar
```

По умолчанию образ получает тег `stronghold:<версия>`, например `stronghold:1.19.0`. <!-- TODO(verify): наличие официального образа Stronghold в публичном registry -->

## Режим разработки

Команда контейнера по умолчанию — `stronghold server -dev`. В режиме разработки Stronghold:

- хранит данные в памяти — после остановки контейнера всё теряется;
- автоматически инициализируется и распечатывается;
- выводит в лог root-токен;
- работает без TLS.

Запустите контейнер:

```bash
docker run --rm --name stronghold-dev -p 8200:8200 stronghold:1.19.0
```

Найдите root-токен в логе контейнера (строка `Root Token`) и настройте клиент:

```bash
export STRONGHOLD_ADDR=http://127.0.0.1:8200
export STRONGHOLD_TOKEN=<root-токен из лога>
d8 stronghold status
```

Для клиентских библиотек Vault задайте те же значения в `VAULT_ADDR` и `VAULT_TOKEN` (см. [«Клиенты приложений»](../app-clients/)).

Чтобы задать предсказуемый root-токен и адрес прослушивания, передайте параметры режима разработки:

```bash
docker run --rm --name stronghold-dev -p 8200:8200 stronghold:1.19.0 \
  stronghold server -dev -dev-root-token-id=root -dev-listen-address=0.0.0.0:8200
```

## Одноузловая конфигурация с Raft в Docker Compose

Если данные должны сохраняться между перезапусками, запустите Stronghold с конфигурационным файлом и хранилищем Raft.

1. Создайте файл `config/stronghold.hcl` (формат описан в разделе [«Настройка»](../../../install/standalone/configuration/)):

   ```hcl
   ui            = true
   disable_mlock = true
   api_addr      = "http://127.0.0.1:8200"
   cluster_addr  = "http://127.0.0.1:8201"

   storage "raft" {
     path    = "/stronghold/data"
     node_id = "dev-node-1"
   }

   listener "tcp" {
     address     = "0.0.0.0:8200"
     tls_disable = true
   }
   ```

   {{< alert level="warning" >}}
   Параметр `tls_disable = true` допустим только на локальной машине разработчика.
   {{< /alert >}}

1. Создайте файл `compose.yaml`:

   ```yaml
   services:
     stronghold:
       image: stronghold:1.19.0
       command: ["stronghold", "server", "-config=/stronghold/config/stronghold.hcl"]
       ports:
         - "8200:8200"
       volumes:
         - ./config:/stronghold/config:ro
         - stronghold-data:/stronghold/data
       restart: unless-stopped

   volumes:
     stronghold-data:
   ```

   Образ запускается от root (пользователь в образе не задан), поэтому права на том `/stronghold/data` дополнительно настраивать не нужно. В образе нет ENTRYPOINT: команда задаётся целиком, как в примере.

1. Запустите контейнер:

   ```bash
   docker compose up -d
   ```

1. Инициализируйте и распечатайте Stronghold. Для локальной разработки достаточно одного ключа:

   ```bash
   export STRONGHOLD_ADDR=http://127.0.0.1:8200
   d8 stronghold operator init -key-shares=1 -key-threshold=1
   d8 stronghold operator unseal
   ```

   Сохраните ключ распечатывания и root-токен из вывода `operator init`. После каждого перезапуска контейнера выполняйте `d8 stronghold operator unseal`.

1. Войдите с root-токеном и включите нужные механизмы, например KV версии 2:

   ```bash
   export STRONGHOLD_TOKEN=<root-токен>
   d8 stronghold secrets enable -path=secret -version=2 kv
   d8 stronghold kv put -mount=secret myapp/db username=app password=S3cr3t
   ```

Веб-интерфейс доступен по адресу `http://127.0.0.1:8200/ui`.

## Запуск приложения рядом со Stronghold

Приложение в том же файле `compose.yaml` обращается к Stronghold по имени сервиса. В примере используется root-токен `root`, заданный параметром `-dev-root-token-id`:

```yaml
services:
  app:
    image: registry.example.com/myapp:1.0.0
    environment:
      VAULT_ADDR: http://stronghold:8200
      VAULT_TOKEN: root
    depends_on:
      - stronghold
```

Токен в переменной окружения допустим только для локальной разработки. В рабочих окружениях используйте методы аутентификации, описанные в разделах [«Доставка секретов в поды Kubernetes»](../kubernetes-workloads/) и [«Клиенты приложений»](../app-clients/).
