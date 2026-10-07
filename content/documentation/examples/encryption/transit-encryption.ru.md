---
title: "Шифрование полей в базе данных с помощью Transit"
linkTitle: "Шифрование через Transit"
description: "Шифрование отдельных полей в базе данных приложения механизмом Transit, конвертное шифрование через datakey, ротация ключа, rewrap и min_decryption_version."
weight: 10
params:
  relatedLinks:
    - title: "Механизм секретов Transit"
      url: ../../../user/secrets-engines/transit/
    - title: "API механизмов секретов"
      url: ../../../reference/api/secrets/
    - title: "Политики"
      url: ../../../concepts/policy/
---

Приложение шифрует чувствительные поля (номера документов, телефоны, токены) через API Stronghold и сохраняет в базе данных только шифртекст. Ключ шифрования не покидает Stronghold, поэтому утечка дампа базы данных не раскрывает данные.

## Цель

Настроить ключ Transit для приложения, зашифровать и расшифровать поле, применить конвертное шифрование для больших объектов, выполнить ротацию ключа, перешифровать данные (rewrap) и запретить расшифровку старыми версиями ключа.

## Предварительные требования

- Токен Stronghold с правами на включение механизмов секретов и создание политик.
- Утилиты `base64`, `jq` и `openssl` для примеров.

## Шаг 1. Создайте ключ

1. Включите механизм секретов:

   ```bash
   d8 stronghold secrets enable transit
   ```

1. Создайте именованный ключ для приложения:

   ```bash
   d8 stronghold write -f transit/keys/orders type=aes256-gcm96
   ```

   Используйте отдельный ключ для каждого приложения или набора данных.

## Шаг 2. Разделите права

1. Политика приложения — только шифрование и расшифровка:

   ```bash
   d8 stronghold policy write orders-app - <<'POLICY'
   path "transit/encrypt/orders" {
     capabilities = ["update"]
   }
   path "transit/decrypt/orders" {
     capabilities = ["update"]
   }
   path "transit/datakey/plaintext/orders" {
     capabilities = ["update"]
   }
   POLICY
   ```

1. Политика задачи перешифрования — только rewrap, без доступа к открытым данным:

   ```bash
   d8 stronghold policy write orders-rewrap - <<'POLICY'
   path "transit/rewrap/orders" {
     capabilities = ["update"]
   }
   path "transit/keys/orders" {
     capabilities = ["read"]
   }
   POLICY
   ```

Управление ключом (`transit/keys/orders/*`) оставьте администраторам.

## Шаг 3. Зашифруйте и расшифруйте поле

Данные передаются в Stronghold в кодировке base64.

1. Зашифруйте значение:

   ```bash
   d8 stronghold write -field=ciphertext transit/encrypt/orders \
     plaintext=$(echo -n "4509 123456" | base64)
   ```

   Результат вида `vault:v1:...` сохраните в столбце таблицы. Префикс `v1` — версия ключа, которой выполнено шифрование.

1. Расшифруйте значение:

   ```bash
   d8 stronghold write -field=plaintext transit/decrypt/orders \
     ciphertext="vault:v1:..." | base64 --decode
   ```

Для обработки множества записей за один запрос используйте параметр `batch_input` эндпоинтов `encrypt`, `decrypt` и `rewrap`.

## Шаг 4. Используйте конвертное шифрование для больших данных

Передавать через API файлы и большие объекты неэффективно. Вместо этого запросите ключ данных (data key): Stronghold возвращает его в открытом виде для локального шифрования и в зашифрованном виде для хранения.

1. Получите ключ данных:

   ```bash
   d8 stronghold write -format=json -f transit/datakey/plaintext/orders > datakey.json
   jq -r '.data.plaintext' datakey.json | base64 --decode > dek.bin
   jq -r '.data.ciphertext' datakey.json > dek.enc
   ```

1. Зашифруйте файл локально и удалите открытый ключ:

   ```bash
   openssl enc -aes-256-cbc -pbkdf2 -in report.pdf -out report.pdf.enc -pass file:dek.bin
   shred -u dek.bin
   ```

   Сохраните `report.pdf.enc` вместе с `dek.enc`.

1. Для расшифровки восстановите ключ данных через Transit:

   ```bash
   d8 stronghold write -field=plaintext transit/decrypt/orders \
     ciphertext="$(cat dek.enc)" | base64 --decode > dek.bin
   openssl enc -d -aes-256-cbc -pbkdf2 -in report.pdf.enc -out report.pdf -pass file:dek.bin
   shred -u dek.bin
   ```

Если ключ данных нужен только для хранения (например, его расшифрует другой сервис), используйте `transit/datakey/wrapped/orders` — ответ не содержит открытого ключа.

## Шаг 5. Выполните ротацию ключа

1. Создайте новую версию ключа:

   ```bash
   d8 stronghold write -f transit/keys/orders/rotate
   ```

   Новые данные шифруются последней версией, старые по-прежнему расшифровываются.

1. Чтобы ротация выполнялась автоматически, задайте период:

   ```bash
   d8 stronghold write transit/keys/orders/config auto_rotate_period=720h
   ```

## Шаг 6. Перешифруйте данные (rewrap)

Rewrap перешифровывает шифртекст последней версией ключа без раскрытия открытых данных. Выполните его для всех записей таблицы, например фоновой задачей с политикой `orders-rewrap`:

```bash
d8 stronghold write -field=ciphertext transit/rewrap/orders ciphertext="vault:v1:..."
```

Результат `vault:v2:...` запишите обратно в столбец.

## Шаг 7. Запретите расшифровку старыми версиями

Когда в базе данных не осталось шифртекста со старыми версиями, поднимите минимальную версию для расшифровки:

```bash
d8 stronghold write transit/keys/orders/config min_decryption_version=2
```

Версии ниже `min_decryption_version` переносятся в архив и не используются для расшифровки. В экстренном случае значение можно вернуть назад.

Если старые версии нужно удалить окончательно, используйте `transit/keys/orders/trim` с параметром `min_available_version`. Это действие необратимо.

## Проверка

1. Просмотрите состояние ключа:

   ```bash
   d8 stronghold read transit/keys/orders
   ```

   Проверьте значения `latest_version`, `min_decryption_version` и `auto_rotate_period`.

1. Попробуйте расшифровать шифртекст `vault:v1:...` после шага 7 — запрос должен завершиться ошибкой.
1. Убедитесь, что токен с политикой `orders-rewrap` не может выполнить `transit/decrypt/orders`.

## Очистка

1. Удалите политики: `d8 stronghold policy delete orders-app` и `d8 stronghold policy delete orders-rewrap`.
1. Удалите локальные файлы `datakey.json`, `dek.enc`, `report.pdf.enc`.
1. Для удаления тестового ключа разрешите удаление и удалите ключ:

   ```bash
   d8 stronghold write transit/keys/orders/config deletion_allowed=true
   d8 stronghold delete transit/keys/orders
   ```

   Данные, зашифрованные этим ключом, станут недоступны.
