---
title: "Механизм секретов SSH"
description: "Сведения о разделе \"Механизм секретов SSH\" в Deckhouse Stronghold."
weight: 40
---

Механизм секретов SSH позволяет использовать SSH-сертификаты для доступа к серверам.
Это один из самых простых и удобных способов организовать SSH-доступ с минимальной зависимостью от платформы.

Stronghold может выступать в роли центра сертификации (CA) для SSH.
В сочетании со встроенными возможностями OpenSSH это позволяет клиентам подключаться к удалённым хостам по SSH, используя собственные локальные SSH-ключи.

На этой странице термин «клиент» обозначает пользователя или машину, которая инициирует SSH-подключение.
Термин «хост» обозначает удалённую машину.
На странице приведён краткий сценарий настройки и использования механизма секретов SSH.

## Подпись ключей клиентов

Сначала настройте механизм секретов SSH в Stronghold.
После этого клиенты смогут подписывать свои SSH-ключи.
Обычно эти действия выполняет администратор Stronghold, команда безопасности или система управления конфигурацией.

### Создание ключа подписи и настройка роли

Следующие шаги выполняет администратор Stronghold, команда безопасности или система управления конфигурацией.

1. Смонтируйте механизм секретов SSH.
   Без этого механизм секретов работать не будет:

   ```shell
   stronghold secrets enable -path=ssh-client-signer ssh
   ```

   Пример вывода:

   ```console
   Successfully mounted 'ssh' at 'ssh-client-signer'!
   ```

   Эта команда включает механизм секретов SSH по пути `ssh-client-signer`.
   Один и тот же механизм секретов можно смонтировать несколько раз, используя разные значения `-path`.
   Имя `ssh-client-signer` не является специальным.
   В примерах на этой странице используется именно оно.

1. Настройте Stronghold с CA для подписи клиентских ключей с помощью эндпоинта `/config/ca`.
   Если собственного CA нет, Stronghold может сгенерировать открытый и закрытый ключи:

   ```shell
   stronghold write ssh-client-signer/config/ca generate_signing_key=true
   ```

   Пример вывода:

   ```console
   Key             Value
   ---             -----
   public_key      ssh-rsa AAAAB3NzaC1yc2EA...
   ```

   Если у вас уже есть пара SSH-ключей, передайте открытый и закрытый ключи в запросе:

   ```shell
   stronghold write ssh-client-signer/config/ca \
     private_key="..." \
     public_key="..."
   ```

   Механизм секретов SSH поддерживает настройку нескольких CA-сертификатов в одном монтировании.
   Это удобно для ротации CA.
   При настройке один issuer назначается issuer по умолчанию.
   Он используется во всех операциях, если при создании роли не указан другой issuer.
   Issuer по умолчанию можно изменить в любой момент, создав новый CA или обновив существующий через конфигурационный эндпоинт.

   Независимо от того, был ключ сгенерирован или загружен, открытый ключ доступен через API по эндпоинту `/public_key` и через CLI.

1. Добавьте открытый ключ во все конфигурации SSH хостов.
   Это можно сделать вручную или автоматизировать через систему управления конфигурацией.
   Открытый ключ доступен по API и не требует аутентификации:

   ```shell
   curl -o /etc/ssh/trusted-user-ca-keys.pem http://127.0.0.1:8200/v1/ssh-client-signer/public_key
   ```

   Или через CLI:

   ```shell
   stronghold read -field=public_key ssh-client-signer/config/ca > /etc/ssh/trusted-user-ca-keys.pem
   ```

   Добавьте путь к файлу с открытым ключом в конфигурацию SSH как значение параметра `TrustedUserCAKeys`:

   ```text
   # /etc/ssh/sshd_config
   TrustedUserCAKeys /etc/ssh/trusted-user-ca-keys.pem
   ```

   Перезапустите SSH-службу, чтобы применить изменения.

1. Создайте в Stronghold роль для подписи клиентских ключей.
   Из-за особенностей SSH-сертификатов некоторые параметры передаются в формате «ключ–значение».
   Следующий пример добавляет расширение `permit-pty` в сертификат и разрешает пользователю указывать собственные значения для `permit-pty` и `permit-port-forwarding` при запросе сертификата:

   ```shell
   stronghold write ssh-client-signer/roles/my-role -<<"EOH"
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

### Аутентификация клиента по SSH

Следующие шаги выполняет клиент, который хочет аутентифицироваться на машинах, настроенных для работы со Stronghold.
Обычно эти команды запускаются на локальной рабочей станции клиента.

1. Найдите или сгенерируйте открытый SSH-ключ.
   Обычно он расположен по пути `~/.ssh/id_rsa.pub`.
   Если пары SSH-ключей ещё нет, создайте её:

   ```shell
   ssh-keygen -t rsa -C "user@example.com"
   ```

1. Попросите Stronghold подписать ваш открытый ключ.
   Обычно это файл с расширением `.pub`, а его содержимое начинается с `ssh-rsa ...`:

   ```shell
   stronghold write ssh-client-signer/sign/my-role \
     public_key=@$HOME/.ssh/id_rsa.pub
   ```

   Пример вывода:

   ```console
   Key             Value
   ---             -----
   serial_number   c73f26d2340276aa
   signed_key      ssh-rsa-cert-v01@openssh.com AAAAHHNzaC1...
   ```

   В ответе возвращаются серийный номер, то есть уникальный идентификатор сертификата, и подписанный ключ.
   Подписанный ключ также является открытым ключом.

   Чтобы дополнительно настроить параметры подписи, используйте JSON-запрос:

   ```shell
   stronghold write ssh-client-signer/sign/my-role -<<"EOH"
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

1. Сохраните полученный подписанный открытый ключ на диск.
   При необходимости ограничьте права доступа к файлу:

   ```shell
   stronghold write -field=signed_key ssh-client-signer/sign/my-role \
     public_key=@$HOME/.ssh/id_rsa.pub > signed-cert.pub
   ```

   Если сохраняете сертификат рядом с парой SSH-ключей, используйте суффикс `-cert.pub`, например `~/.ssh/id_rsa-cert.pub`.
   В этом случае OpenSSH автоматически использует сертификат при аутентификации.

1. При необходимости просмотрите расширения, список пользователей, хостов и метаданные подписанного ключа:

   ```shell
   ssh-keygen -Lf ~/.ssh/signed-cert.pub
   ```

1. Выполните SSH-подключение с использованием подписанного открытого ключа и соответствующего закрытого ключа:

   ```shell
   ssh -i signed-cert.pub -i ~/.ssh/id_rsa username@10.0.23.5
   ```

## Подпись ключей хостов

Для дополнительной защиты рекомендуется включить подпись ключей хостов.
Эта возможность дополняет подпись клиентских ключей и позволяет повысить целостность SSH-подключений.

Если подпись ключей хостов настроена, SSH-клиент сможет проверить, что удалённый хост действительно является доверенным, до установления соединения.
Это снижает риск случайного подключения к вредоносной машине.

### Настройка подписи ключей хостов

1. Смонтируйте механизм секретов SSH по другому пути, отличному от пути подписи клиентских ключей:

   ```shell
   stronghold secrets enable -path=ssh-host-signer ssh
   ```

   Пример вывода:

   ```console
   Successfully mounted 'ssh' at 'ssh-host-signer'!
   ```

1. Настройте Stronghold с CA для подписи ключей хостов через эндпоинт `/config/ca`.
   Если собственного CA нет, Stronghold может сгенерировать ключевую пару:

   ```shell
   stronghold write ssh-host-signer/config/ca generate_signing_key=true
   ```

   Пример вывода:

   ```console
   Key             Value
   ---             -----
   public_key      ssh-rsa AAAAB3NzaC1yc2EA...
   ```

   Если у вас уже есть пара SSH-ключей, передайте её в запросе:

   ```shell
   stronghold write ssh-host-signer/config/ca \
     private_key="..." \
     public_key="..."
   ```

   Открытый ключ CA для подписи ключей хостов доступен через API по эндпоинту `/public_key`.

1. Увеличьте TTL сертификата ключа хоста:

   ```shell
   stronghold secrets tune -max-lease-ttl=87600h ssh-host-signer
   ```

1. Создайте роль для подписи ключей хостов.
   Обязательно укажите список разрешённых доменов, параметр `allow_bare_domains` или оба параметра сразу:

   ```shell
   stronghold write ssh-host-signer/roles/hostrole \
     key_type=ca \
     algorithm_signer=rsa-sha2-256 \
     ttl=87600h \
     allow_host_certificates=true \
     allowed_domains="localdomain,example.com" \
     allow_subdomains=true
   ```

1. Подпишите открытый SSH-ключ хоста:

   ```shell
   stronghold write ssh-host-signer/sign/hostrole \
     cert_type=host \
     public_key=@/etc/ssh/ssh_host_rsa_key.pub
   ```

   Пример вывода:

   ```console
   Key             Value
   ---             -----
   serial_number   3746eb17371540d9
   signed_key      ssh-rsa-cert-v01@openssh.com AAAAHHNzaC1y...
   ```

1. Сохраните подписанный сертификат и настройте его как значение параметра `HostCertificate` в конфигурации SSH на хосте:

   ```shell
   stronghold write -field=signed_key ssh-host-signer/sign/hostrole \
     cert_type=host \
     public_key=@/etc/ssh/ssh_host_rsa_key.pub > /etc/ssh/ssh_host_rsa_key-cert.pub
   ```

   Установите права доступа `0640`:

   ```shell
   chmod 0640 /etc/ssh/ssh_host_rsa_key-cert.pub
   ```

   Добавьте ключ хоста и сертификат хоста в конфигурационный файл SSH:

   ```text
   # /etc/ssh/sshd_config
   TrustedUserCAKeys /etc/ssh/trusted-user-ca-keys.pem
   HostKey /etc/ssh/ssh_host_rsa_key
   HostCertificate /etc/ssh/ssh_host_rsa_key-cert.pub
   ```

   Перезапустите SSH-службу, чтобы применить изменения.

### Проверка хоста на стороне клиента

1. Получите открытый ключ CA хоста для проверки подписи:

   ```shell
   curl http://127.0.0.1:8200/v1/ssh-host-signer/public_key
   ```

   Или через CLI:

   ```shell
   stronghold read -field=public_key ssh-host-signer/config/ca
   ```

1. Добавьте полученный открытый ключ в файл `known_hosts`:

   ```text
   # ~/.ssh/known_hosts
   @cert-authority *.example.com ssh-rsa AAAAB3NzaC1yc2EAAA...
   ```

1. После этого можно выполнять SSH-подключение к удалённым машинам.

## Устранение неполадок

Чтобы упростить настройку и отладку процесса подписи ключей, включите уровень логирования `VERBOSE` в конфигурации SSH:

```text
# /etc/ssh/sshd_config
LogLevel VERBOSE
```

После изменения конфигурации перезапустите SSH-службу.

По умолчанию SSH пишет журналы в файл `/var/log/auth.log`.
Так как в этом файле могут быть записи от других служб, для фильтрации только SSH-журналов используйте следующую команду:

```shell
tail -f /var/log/auth.log | grep --line-buffered "sshd"
```

Если не удаётся установить соединение с хостом, журналы SSH-сервера помогут найти причину.

### Имя пользователя отсутствует в списке valid principals

Если в `/var/log/auth.log` отображаются следующие сообщения:

```text
key_cert_check_authority: invalid certificate
Certificate invalid: name is not a listed principal
```

Это означает, что сертификат не разрешает использовать указанное имя пользователя в качестве valid principal для аутентификации в системе.
Скорее всего, причина связана с особенностью OpenSSH.
Подробнее см. в разделе [Известные проблемы](#известные-проблемы).

Эта ошибка не учитывает значение `allowed_users="*"`.
Используйте один из следующих способов обхода:

- Настройте `default_user`, если вы всегда аутентифицируетесь под одним и тем же пользователем:

  ```shell
  stronghold write ssh/roles/my-role -<<"EOH"
  {
    "default_user": "YOUR_USER"
  }
  EOH
  ```

- Укажите `valid_principals` при подписи ключа, если SSH-аутентификацию через Stronghold проходят разные пользователи:

  ```shell
  stronghold write ssh-client-signer/sign/my-role -<<"EOH"
  {
    "valid_principals": "my-user"
  }
  EOH
  ```

### Нет приглашения командной строки после входа

Если после аутентификации на хосте не отображается приглашение командной строки, в подписанном сертификате может отсутствовать расширение `permit-pty`.

Добавить это расширение можно двумя способами:

- При создании роли:

  ```shell
  stronghold write ssh-client-signer/roles/my-role -<<"EOH"
  {
    "default_extensions": {
      "permit-pty": ""
    }
  }
  EOH
  ```

- Во время подписи ключа:

  ```shell
  stronghold write ssh-client-signer/sign/my-role -<<"EOH"
  {
    "extensions": {
      "permit-pty": ""
    }
  }
  EOH
  ```

### Не работает переадресация портов

Если не работает переадресация портов, в сертификате может отсутствовать расширение `permit-port-forwarding`.

Добавьте его при создании роли или во время подписи ключа:

```json
{
  "default_extensions": {
    "permit-port-forwarding": ""
  }
}
```

### Не работает переадресация X11

Если не работает переадресация X11, в сертификате может отсутствовать расширение `permit-X11-forwarding`.

Добавьте его при создании роли или во время подписи ключа:

```json
{
  "default_extensions": {
    "permit-X11-forwarding": ""
  }
}
```

### Не работает переадресация SSH-агента

Если не работает переадресация SSH-агента, в сертификате может отсутствовать расширение `permit-agent-forwarding`.

Добавьте его при создании роли или во время подписи ключа:

```json
{
  "default_extensions": {
    "permit-agent-forwarding": ""
  }
}
```

### Комментарии в ключах

Если требуется сохранить [атрибуты комментариев](https://www.rfc-editor.org/rfc/rfc4716#section-3.3.2) в ключах, для этой операции могут потребоваться дополнительные шаги.

Закрытый и открытый ключи могут содержать комментарии.
Например, их можно задать через `ssh-keygen` с параметром `-C`:

```shell
ssh-keygen -C "...Comments" -N "" -t rsa -b 4096 -f host-ca
```

Значения ключей с комментариями нужно передавать вместе с параметрами, связанными с этим ключом.
Ниже приведены примеры для Stronghold CLI и API.

Пример с CLI:

```shell
stronghold secrets enable -path=hosts-ca ssh
KEY_PRI=$(cat ~/.ssh/id_rsa | sed -z 's/\n/\\n/g')
KEY_PUB=$(cat ~/.ssh/id_rsa.pub | sed -z 's/\n/\\n/g')
stronghold write ssh-client-signer/config/ca \
  generate_signing_key=false \
  private_key="${KEY_PRI}" \
  public_key="${KEY_PUB}"
```

Пример с API:

```shell
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
curl -X POST -H "X-Vault-Token: ..." -d @payload.json http://127.0.0.1:8200/v1/hosts-ca/config/ca
```

{{< alert level="warning" >}}
Не добавляйте пароль к закрытому ключу, так как Stronghold не сможет его расшифровать.
Удалите открытый и закрытый ключи, а также файл `payload.json` с хоста сразу после подтверждения успешной загрузки.
{{< /alert >}}

### Известные проблемы

- В системах с SELinux может потребоваться настроить соответствующие типы, чтобы SSH-демон мог читать нужные файлы.
  Например, для подписанного сертификата хоста может потребоваться тип `sshd_key_t`

- В некоторых версиях SSH может возникать следующая ошибка:

  ```text
  no separate private key for certificate
  ```

  Эта ошибка появилась в OpenSSH 7.2 и была исправлена в OpenSSH 7.5.
  Дополнительную информацию см. в [OpenSSH bug 2617](https://bugzilla.mindrot.org/show_bug.cgi?id=2617)

- В некоторых версиях SSH на хосте может возникать следующая ошибка:

  ```text
  userauth_pubkey: certificate signature algorithm ssh-rsa: signature algorithm not supported [preauth]
  ```

  Чтобы исправить её, добавьте следующую строку в файл `/etc/ssh/sshd_config`:

  ```text
  CASignatureAlgorithms ^ssh-rsa
  ```

  Алгоритм `ssh-rsa` больше не поддерживается в [OpenSSH 8.2](https://www.openssh.com/txt/release-8.2)
