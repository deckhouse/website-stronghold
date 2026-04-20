---
title: "SSH"
linkTitle: "SSH"
description: "Работа с механизмом секретов SSH в Deckhouse Stronghold"
weight: 40
---

Механизм секретов `SSH` в Deckhouse Stronghold позволяет подписывать SSH-ключи и использовать SSH-сертификаты вместо ручного распространения постоянных ключей доступа.

Это один из самых простых и практичных способов организовать SSH-доступ к серверам. Stronghold выступает как центр сертификации (CA), а OpenSSH использует выпущенные сертификаты для аутентификации.

На этой странице показан базовый сценарий работы с механизмом секретов `SSH`:

- подпись клиентских ключей;
- подпись ключей хостов;
- проверка сертификатов на стороне клиента;
- типовые проблемы и способы их устранения.

В этом разделе:
- **клиент** — пользователь или машина, которые выполняют SSH-подключение;
- **хост** — удалённая машина, к которой выполняется подключение.

## Когда использовать

Используйте механизм секретов `SSH`, если нужно:

- централизованно управлять SSH-доступом;
- отказаться от ручной раздачи постоянных SSH-ключей;
- выдавать короткоживущие SSH-сертификаты;
- ограничивать доступ через роли и параметры сертификатов;
- дополнительно проверять подлинность хостов через подпись host-ключей.

## Как это работает

Базовый сценарий выглядит так:

1. Администратор включает механизм секретов `SSH` и настраивает CA для подписи ключей.
2. Администратор создаёт роль, которая определяет параметры сертификатов.
3. Клиент отправляет свой публичный SSH-ключ в Stronghold.
4. Stronghold подписывает ключ и возвращает SSH-сертификат.
5. Клиент использует свой закрытый ключ и полученный сертификат для подключения к хосту.

При необходимости Stronghold также может подписывать ключи самих хостов. Тогда клиент сможет проверять, что подключается именно к доверенному серверу.

---

## Подпись ключей клиентов

### Что делает администратор

На этом этапе администратор включает механизм секретов, настраивает CA и создаёт роль для подписи клиентских ключей.

### Шаг 1. Включите механизм секретов SSH

Смонтируйте механизм секретов `SSH` по отдельному пути:

```text
$ stronghold secrets enable -path=ssh-client-signer ssh

Successfully mounted 'ssh' at 'ssh-client-signer'!
```

Эта команда включает механизм секретов `SSH` по пути `ssh-client-signer`.

Можно подключать один и тот же механизм секретов несколько раз с разными значениями `-path`. Имя `ssh-client-signer` не является специальным — вы можете выбрать любой другой `mount path`.

### Шаг 2. Настройте CA для подписи клиентских ключей

Если у вас нет собственной пары SSH-ключей CA, Stronghold может сгенерировать её автоматически:

```text
$ stronghold write ssh-client-signer/config/ca generate_signing_key=true

Key             Value
---             -----
public_key      ssh-rsa AAAAB3NzaC1yc2EA...
```

Если у вас уже есть собственная пара ключей, загрузите её вручную:

```text
$ stronghold write ssh-client-signer/config/ca \
  private_key="..." \
  public_key="..."
```

Механизм секретов `SSH` поддерживает несколько сертификатов доверенного центра сертификации в одном монтировании. Это упрощает ротацию CA.

При настройке CA один issuer назначается по умолчанию. Его операции будут использоваться, если в роли не указан конкретный issuer. Эмитент по умолчанию можно изменить позже — например, при ротации CA.

Открытый ключ CA доступен через API в методе `/public_key` или через CLI.

### Шаг 3. Добавьте открытый ключ CA на хосты

Открытый ключ CA нужно добавить в конфигурацию SSH на всех целевых хостах. Это можно сделать вручную или через инструмент управления конфигурацией.

Получить ключ можно так:

```text
curl -o /etc/ssh/trusted-user-ca-keys.pem http://127.0.0.1:8200/v1/ssh-client-signer/public_key
```

или так:

```text
stronghold read -field=public_key ssh-client-signer/config/ca > /etc/ssh/trusted-user-ca-keys.pem
```

Добавьте путь к файлу в конфигурацию SSH:

```text
# /etc/ssh/sshd_config
# ...
TrustedUserCAKeys /etc/ssh/trusted-user-ca-keys.pem
```

После этого перезапустите службу SSH, чтобы применить изменения.

### Шаг 4. Создайте роль для подписи клиентских ключей

Роль определяет, какие сертификаты Stronghold может выпускать и какие параметры в них разрешены.

Пример роли:

```text
$ stronghold write ssh-client-signer/roles/my-role -<<"EOH"
{
    "algorithm_signer": "rsa-sha2-256",
    "allow_user_certificates": true,
    "allowed_users": "*",
    "allowed_extensions": "permit-pty,permit-port-forwarding",
    "default_extensions": {
        "permit-pty": ""
    },
    "key_type": "ca",
    "default_user": "ubuntu",
    "ttl": "30m0s"
}
EOH
```

Что задаёт эта роль:

- разрешает выпуск пользовательских сертификатов;
- разрешает любых пользователей через `allowed_users: "*"`;
- позволяет использовать расширения `permit-pty` и `permit-port-forwarding`;
- по умолчанию включает `permit-pty`;
- задаёт пользователя по умолчанию `ubuntu`;
- ограничивает срок действия сертификата до 30 минут.

---

### Что делает клиент

После настройки механизма секретов клиент может подписать свой публичный SSH-ключ и использовать сертификат для подключения.

### Шаг 5. Найдите или создайте SSH-ключ

Обычно публичный ключ уже существует и расположен по пути `~/.ssh/id_rsa.pub`.

Если пары ключей нет, создайте её:

```text
ssh-keygen -t rsa -C "user@example.com"
```

### Шаг 6. Подпишите публичный ключ клиента

Передайте Stronghold публичный SSH-ключ:

```text
$ stronghold write ssh-client-signer/sign/my-role \
  public_key=@$HOME/.ssh/id_rsa.pub


 Key             Value
 ---             -----
 serial_number   c73f26d2340276aa
 signed_key      ssh-rsa-cert-v01@openssh.com AAAAHHNzaC1...
```

В ответе Stronghold вернёт:

- `serial_number` — уникальный идентификатор сертификата;
- `signed_key` — подписанный SSH-сертификат.

Если нужно явно задать параметры подписи, используйте JSON:

```text
$ stronghold write ssh-client-signer/sign/my-role -<<"EOH"
 {
   "public_key": "ssh-rsa AAA...",
   "valid_principals": "my-user",
   "key_id": "custom-prefix",
   "extensions": {
     "permit-pty": "",
     "permit-port-forwarding": ""
   }
 }
 EOH
```

Здесь можно задать:

- `valid_principals` — допустимые principal для сертификата;
- `key_id` — идентификатор сертификата;
- `extensions` — расширения, которые будут добавлены в сертификат.

### Шаг 7. Сохраните подписанный сертификат

Сохраните полученный сертификат на диск:

```text
$ stronghold write -field=signed_key ssh-client-signer/sign/my-role \
  public_key=@$HOME/.ssh/id_rsa.pub > signed-cert.pub
```

Если сохраняете сертификат рядом с основной парой ключей SSH, используйте суффикс `-cert.pub`, например `~/.ssh/id_rsa-cert.pub`.

При такой схеме именования OpenSSH сможет использовать сертификат автоматически.

### Шаг 8. При необходимости проверьте содержимое сертификата

Посмотреть параметры сертификата можно командой:

```text
ssh-keygen -Lf ~/.ssh/signed-cert.pub
```

Это удобно, чтобы проверить:

- срок действия сертификата;
- principal;
- расширения;
- серийный номер;
- дополнительные метаданные.

### Шаг 9. Подключитесь к хосту

Для подключения используйте и подписанный сертификат, и соответствующий закрытый ключ:

```text
ssh -i signed-cert.pub -i ~/.ssh/id_rsa username@10.0.23.5
```

Если сертификат сохранён рядом с ключом под именем `~/.ssh/id_rsa-cert.pub`, OpenSSH обычно подхватит его автоматически.

---

## Подпись ключей хоста

Подпись ключей хоста даёт дополнительный уровень защиты. В этом случае SSH-клиент сможет проверить, что удалённая машина действительно доверенная, а не подменена.

Этот режим стоит использовать вместе с подписью клиентских ключей.

### Шаг 1. Включите отдельный SSH mount для хостов

Используйте отдельный путь, отличный от пути для клиентских сертификатов:

```text
$ stronghold secrets enable -path=ssh-host-signer ssh

Successfully mounted 'ssh' at 'ssh-host-signer'!
```

### Шаг 2. Настройте CA для подписи ключей хоста

Если собственной пары SSH-ключей CA нет, Stronghold может сгенерировать её:

```text
$ stronghold write ssh-host-signer/config/ca generate_signing_key=true

Key             Value
---             -----
public_key      ssh-rsa AAAAB3NzaC1yc2EA...
```

Если у вас уже есть ключи, загрузите их вручную:

```text
$ stronghold write ssh-host-signer/config/ca \
  private_key="..." \
  public_key="..."
```

Открытый ключ подписывающего CA для хостов доступен через API в методе `/public_key`.

### Шаг 3. Увеличьте TTL для host certificates

Обычно сертификаты хостов живут дольше, чем клиентские сертификаты, поэтому увеличьте TTL:

```text
stronghold secrets tune -max-lease-ttl=87600h ssh-host-signer
```

### Шаг 4. Создайте роль для подписи host-ключей

Создайте роль, которая разрешает выпуск сертификатов хостов:

```text
$ stronghold write ssh-host-signer/roles/hostrole \
        key_type=ca \
        algorithm_signer=rsa-sha2-256 \
        ttl=87600h \
        allow_host_certificates=true \
        allowed_domains="localdomain,example.com" \
        allow_subdomains=true
```

При настройке роли обязательно укажите список разрешённых доменов и при необходимости включите `allow_subdomains`.

### Шаг 5. Подпишите публичный ключ хоста

Передайте публичный ключ хоста в Stronghold:

```text
$ stronghold write ssh-host-signer/sign/hostrole \
  cert_type=host \
  public_key=@/etc/ssh/ssh_host_rsa_key.pub

Key             Value
---             -----
serial_number   3746eb17371540d9
signed_key      ssh-rsa-cert-v01@openssh.com AAAAHHNzaC1y...
```

### Шаг 6. Сохраните сертификат хоста

Сохраните подписанный сертификат рядом с ключом хоста:

```text
$ stronghold write -field=signed_key ssh-host-signer/sign/hostrole \
  cert_type=host \
  public_key=@/etc/ssh/ssh_host_rsa_key.pub > /etc/ssh/ssh_host_rsa_key-cert.pub
```

Установите права доступа:

```text
chmod 0640 /etc/ssh/ssh_host_rsa_key-cert.pub
```

### Шаг 7. Обновите конфигурацию SSH на хосте

Добавьте ключ хоста и сертификат хоста в `/etc/ssh/sshd_config`:

```text
# /etc/ssh/sshd_config
# ...

# For client keys
TrustedUserCAKeys /etc/ssh/trusted-user-ca-keys.pem

# For host keys
HostKey /etc/ssh/ssh_host_rsa_key
HostCertificate /etc/ssh/ssh_host_rsa_key-cert.pub
```

После этого перезапустите службу SSH.

---

## Проверка хоста на стороне клиента

Чтобы клиент доверял сертификатам хостов, добавьте открытый ключ CA хоста в файл `known_hosts`.

### Шаг 1. Получите открытый ключ CA хоста

Через API:

```text
curl http://127.0.0.1:8200/v1/ssh-host-signer/public_key
```

или через CLI:

```text
stronghold read -field=public_key ssh-host-signer/config/ca
```

### Шаг 2. Добавьте ключ в `known_hosts`

Пример записи:

```text
# ~/.ssh/known_hosts

@cert-authority *.example.com ssh-rsa AAAAB3NzaC1yc2EAAA...
```

После этого SSH-клиент сможет проверять host-сертификаты, подписанные Stronghold.

---

## Устранение неполадок

Если SSH-подключение не работает, сначала включите подробное логирование SSH.

### Включите подробный уровень журналирования

Добавьте в конфигурацию SSH:

```text
# /etc/ssh/sshd_config
# ...
LogLevel VERBOSE
```

После изменения перезапустите SSH.

Обычно SSH пишет сообщения в `/var/log/auth.log`. Чтобы смотреть только сообщения `sshd`, используйте:

```shell-session
tail -f /var/log/auth.log | grep --line-buffered "sshd"
```

Это помогает быстро понять, на каком этапе возникает ошибка.

### Имя пользователя не входит в список principal

Если в журнале есть сообщения:

```text
# /var/log/auth.log
key_cert_check_authority: invalid certificate
Certificate invalid: name is not a listed principal
```

это значит, что сертификат не разрешает использовать текущее имя пользователя как principal для аутентификации.

Чаще всего проблема связана с тем, что OpenSSH не учитывает `allowed_users="*"` в некоторых сценариях.

Есть два способа это исправить.

#### Вариант 1. Задайте `default_user` в роли

Если пользователи всегда подключаются под одним и тем же именем, укажите его в роли:

```text
stronghold write ssh/roles/my-role -<<"EOH"
   {
     "default_user": "YOUR_USER",
     // ...
   }
EOH
```

#### Вариант 2. Явно передайте `valid_principals` при подписи

Если пользователи подключаются под разными учётными записями, задавайте principal во время подписи:

```text
$ stronghold write ssh-client-signer/sign/my-role -<<"EOH"
    {
      "valid_principals": "my-user"
      // ...
    }
EOH
```

### После входа нет приглашения командной строки

Если аутентификация проходит, но shell не появляется, в сертификате, скорее всего, отсутствует расширение `permit-pty`.

Добавьте его одним из двух способов.

#### Через настройки роли

```text
$ stronghold write ssh-client-signer/roles/my-role -<<"EOH"
  {
    "default_extensions": {
      "permit-pty": ""
    }
    // ...
  }
EOH
```

#### Во время подписи

```text
$ stronghold write ssh-client-signer/sign/my-role -<<"EOH"
  {
    "extensions": {
      "permit-pty": ""
    }
    // ...
  }
EOH
```

### Не работает переадресация портов

Если port forwarding не работает, в сертификате может отсутствовать расширение `permit-port-forwarding`.

Добавьте его в роль или передавайте во время подписи.

Пример:

```json
{
  "default_extensions": {
    "permit-port-forwarding": ""
  }
}
```

### Не работает X11 forwarding

Если не работает переадресация X11, в сертификате может отсутствовать расширение `permit-X11-forwarding`.

Пример:

```json
{
  "default_extensions": {
    "permit-X11-forwarding": ""
  }
}
```

### Не работает agent forwarding

Если не работает переадресация SSH-агента, в сертификате может отсутствовать расширение `permit-agent-forwarding`.

Пример:

```json
{
  "default_extensions": {
    "permit-agent-forwarding": ""
  }
}
```

### Как сохранить комментарии в ключе

Если нужно сохранить [атрибуты комментариев](https://www.rfc-editor.org/rfc/rfc4716#section-3.3.2) в ключах, для загрузки ключей в Stronghold могут потребоваться дополнительные шаги.

Пример генерации ключа с комментарием:

```shell-session
ssh-keygen -C "...Comments" -N "" -t rsa -b 4096 -f host-ca
```

Если ключи содержат комментарии, передавайте их вместе со значениями ключей.

#### Пример через CLI

```shell-extension
# Using CLI:
stronghold secrets enable -path=hosts-ca ssh
KEY_PRI=$(cat ~/.ssh/id_rsa | sed -z 's/\n/\\n/g')
KEY_PUB=$(cat ~/.ssh/id_rsa.pub | sed -z 's/\n/\\n/g')
# Create / update keypair in stronghold
stronghold write ssh-client-signer/config/ca \
  generate_signing_key=false \
  private_key="${KEY_PRI}" \
  public_key="${KEY_PUB}"
```

#### Пример через API

```shell-extension
# Using API:
curl -X POST -H "X-Vault-Token: ..." -d '{"type":"ssh"}' http://127.0.0.1:8200/v1/sys/mounts/hosts-ca
KEY_PRI=$(cat ~/.ssh/id_rsa | sed -z 's/\n/\\n/g')
KEY_PUB=$(cat ~/.ssh/id_rsa.pub | sed -z 's/\n/\\n/g')
tee payload.json <<EOF
{
  "generate_signing_key" : false,
  "private_key"          : "${KEY_PRI}",
  "public_key"           : "${KEY_PUB}"
}
EOF
# Create / update keypair in stronghold
curl -X POST -H "X-Vault-Token: ..." -d @payload.json http://127.0.0.1:8200/v1/hosts-ca/config/ca
```

> Важно  
> Не добавляйте пароль к закрытому ключу, потому что Stronghold не сможет его расшифровать.  
> После успешной загрузки удалите открытый и закрытый ключи, а также `payload.json` с хоста, если они больше не нужны.

---

## Известные проблемы

### SELinux может блокировать чтение сертификата

В системах с SELinux может потребоваться настроить правильные типы для файлов, чтобы демон SSH мог читать сертификат.

Например, для подписанного сертификата хоста может понадобиться тип `sshd_key_t`.

### Ошибка `no separate private key for certificate`

В некоторых версиях SSH может появляться ошибка:

```text
no separate private key for certificate
```

Эта ошибка появилась в OpenSSH 7.2 и была исправлена в OpenSSH 7.5.

### Ошибка `certificate signature algorithm ssh-rsa: signature algorithm not supported`

В некоторых версиях SSH на хосте может появляться ошибка:

```text
userauth_pubkey: certificate signature algorithm ssh-rsa: signature algorithm not supported [preauth]
```

Исправление — добавить в `/etc/ssh/sshd_config` строку:

```text
CASignatureAlgorithms ^ssh-rsa
```

Учтите, что алгоритм `ssh-rsa` больше не поддерживается в OpenSSH 8.2.

---

## Практические рекомендации

Чтобы механизм секретов `SSH` было проще поддерживать, используйте такие правила:

- разделяйте `mount path` для клиентских и host-сертификатов;
- задавайте короткий TTL для клиентских сертификатов;
- проверяйте роли и расширения на тестовом хосте перед массовым использованием;
- дополнительно включайте подпись ключей хостов, если хотите защититься от подключения к подменённой машине;
- автоматизируйте распространение открытых ключей CA и обновление `sshd_config` через инструмент управления конфигурацией.

## Что дальше

- Если вам нужен выпуск сертификатов X.509, используйте [PKI](../pki/).
- Если вам нужен сервис для шифрования, подписи и других криптографических операций без хранения передаваемых данных, используйте [Transit](../transit/).