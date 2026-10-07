---
title: "TLS-сертификат для веб-сервера на ВМ из PKI Stronghold"
linkTitle: "TLS для веб-сервера на ВМ"
description: "Автоматический выпуск и обновление TLS-сертификата Nginx на виртуальной машине: роль PKI, политика, шаблон Stronghold Agent с pki/issue и перезагрузка Nginx после обновления; замечания для Apache."
weight: 60
params:
  relatedLinks:
    - title: "Механизм секретов PKI"
      url: ../../../user/secrets-engines/pki/
    - title: "Внутренний PKI на Stronghold"
      url: ../../certificates/internal-pki/
    - title: "Stronghold Agent: основные возможности"
      url: ../../../user/agent/key-features/
    - title: "Приложение на ВМ со Stronghold Agent"
      url: ../legacy-app-on-vm/
    - title: "Запуск и управление Agent"
      url: ../../../user/agent/launch-and-control/
---

Веб-серверу на виртуальной машине не нужен сертификат, выпущенный вручную на год. Stronghold Agent запрашивает сертификат в механизме секретов PKI, записывает его на диск, перевыпускает до истечения срока действия и перезагружает Nginx. Закрытый ключ создаётся в Stronghold и передаётся только Agent на ВМ.

![Схема выпуска и обновления TLS-сертификата веб-сервера](../../../images/ex-web-server-tls.png)

## Цель

Настроить на ВМ Stronghold Agent, который выпускает для Nginx сертификат `www.example.com` сроком на 72 часа из промежуточного CA `pki_int`, автоматически перевыпускает его и выполняет перезагрузку Nginx без разрыва соединений.

## Предварительные требования

- Промежуточный CA в Stronghold, подключённый по пути `pki_int`, как в руководстве [«Внутренний PKI на Stronghold»](../../certificates/internal-pki/).
- Виртуальная машина Linux с systemd и Nginx.
- Stronghold Agent, установленный на ВМ и входящий в Stronghold по AppRole, как в руководстве [«Приложение на ВМ со Stronghold Agent»](../legacy-app-on-vm/).
- Токен Stronghold с правами на настройку PKI, AppRole и политик.

## Шаг 1. Создайте роль PKI

Роль ограничивает имена и срок жизни сертификатов, которые может получить веб-сервер:

```bash
d8 stronghold write pki_int/roles/web-server \
  allowed_domains="example.com" \
  allow_subdomains=true \
  max_ttl=720h
```

Короткий срок жизни сертификата снижает последствия компрометации ключа: Agent обновляет сертификат автоматически, поэтому ручной перевыпуск не нужен.

## Шаг 2. Создайте политику и роль AppRole

1. Создайте политику, которая разрешает выпуск сертификатов только по роли `web-server`:

   ```bash
   d8 stronghold policy write web-server-tls - <<'POLICY'
   path "pki_int/issue/web-server" {
     capabilities = ["update"]
   }
   POLICY
   ```

1. Создайте роль AppRole для ВМ:

   ```bash
   d8 stronghold write auth/approle/role/web-server \
     token_policies=web-server-tls \
     token_ttl=1h \
     token_max_ttl=24h \
     secret_id_ttl=720h \
     secret_id_bound_cidrs="10.0.10.20/32"
   ```

1. Доставьте `role_id` и обернутый `secret_id` на ВМ в каталог `/etc/stronghold-agent/`, как описано в шаге 3 руководства [«Приложение на ВМ со Stronghold Agent»](../legacy-app-on-vm/#шаг-3-доставьте-role_id-и-обернутый-secret_id-на-вм).

## Шаг 3. Создайте шаблон

Создайте файл `/etc/stronghold-agent/templates/www.pem.ctmpl`. Шаблон записывает закрытый ключ, сертификат и цепочку промежуточных CA в один файл, поэтому ключ и сертификат всегда соответствуют друг другу:

```text
{{ with secret "pki_int/issue/web-server" "common_name=www.example.com" "alt_names=example.com" "ttl=72h" }}
{{ .Data.private_key }}
{{ .Data.certificate }}
{{ range .Data.ca_chain }}{{ . }}
{{ end }}{{ end }}
```

Каждый вызов `secret "pki_int/issue/..."` выпускает сертификат с новым ключом. Если нужны отдельные файлы для сертификата и ключа, как в примере из раздела [«Основные возможности»](../../../user/agent/key-features/), используйте в обоих шаблонах вызов с одинаковыми аргументами.

Одинаковые вызовы (тот же путь и те же аргументы) Agent выполняет один раз и использует результат во всех шаблонах, поэтому ключ и сертификат в разных файлах будут парой.

Для выпуска сертификатов также доступна функция шаблона `pkiCert`: она сама записывает сертификат, ключ и цепочку в файлы и не требует ручного разбора ответа `secret`.

## Шаг 4. Настройте Agent

Добавьте в `/etc/stronghold-agent/agent.hcl` блок `template`:

```hcl
template {
  source          = "/etc/stronghold-agent/templates/www.pem.ctmpl"
  destination     = "/etc/nginx/tls/www.pem"
  perms           = "0600"
  command         = "sudo /usr/sbin/nginx -t && sudo /usr/sbin/nginx -s reload"
  command_timeout = "30s"
  error_on_missing_key = true
}
```

- `perms = "0600"` — файл содержит закрытый ключ; мастер-процесс Nginx читает его с правами root;
- `command` — выполняется после каждой записи файла: `nginx -t` проверяет конфигурацию, `nginx -s reload` перечитывает сертификат без разрыва текущих соединений.

Если `command` задана одной строкой с пробелами, Agent запускает её через `sh -c`, поэтому `&&` и другие возможности оболочки работают.

Разрешите пользователю `stronghold-agent` выполнять только эти команды, например через sudoers:

```text
stronghold-agent ALL=(root) NOPASSWD: /usr/sbin/nginx -t, /usr/sbin/nginx -s reload
```

Создайте каталог для сертификата и укажите его в `ReadWritePaths` unit-файла Agent:

```bash
sudo install -d -o stronghold-agent -g root -m 0750 /etc/nginx/tls
```

## Шаг 5. Настройте Nginx

Укажите один и тот же файл в `ssl_certificate` и `ssl_certificate_key`:

```nginx
server {
    listen 443 ssl;
    server_name www.example.com example.com;

    ssl_certificate     /etc/nginx/tls/www.pem;
    ssl_certificate_key /etc/nginx/tls/www.pem;
    ssl_protocols       TLSv1.2 TLSv1.3;

    location / {
        root /var/www/html;
    }
}
```

Запустите Agent до первого запуска Nginx с этой конфигурацией, чтобы файл сертификата уже существовал:

```bash
sudo systemctl restart stronghold-agent
sudo systemctl reload nginx
```

## Шаг 6. Как работает обновление

Сертификат PKI выдаётся без продлеваемой аренды. Agent отслеживает срок действия сертификата и до его истечения повторно выполняет `pki_int/issue/web-server`, перезаписывает файл и выполняет `command`. Выбирайте `ttl` в шаблоне так, чтобы перевыпуск происходил не реже, чем вы готовы перезагружать Nginx, и не позже, чем вы успеете заметить сбой.

Agent перевыпускает сертификат, когда проходит около 90 % оставшегося срока его действия (со случайным разбросом ±5 %, чтобы клиенты не приходили одновременно). Для `ttl = 24h` это примерно через 21–22 часа.

## Apache HTTP Server

Для Apache используйте тот же шаблон и тот же блок `template` со следующими отличиями:

- в конфигурации виртуального хоста укажите файл в директивах `SSLCertificateFile` и `SSLCertificateKeyFile`. Apache 2.4.8 и выше читает промежуточные сертификаты из `SSLCertificateFile`:

  ```apacheconf
  <VirtualHost *:443>
      ServerName www.example.com
      SSLEngine on
      SSLCertificateFile    /etc/apache2/tls/www.pem
      SSLCertificateKeyFile /etc/apache2/tls/www.pem
  </VirtualHost>
  ```

- в `destination` укажите `/etc/apache2/tls/www.pem`, а в `command` — плавную перезагрузку: `sudo /usr/sbin/apachectl -t && sudo /usr/sbin/apachectl -k graceful` (или `systemctl reload apache2`/`systemctl reload httpd` в зависимости от дистрибутива).

## Проверка

1. Проверьте журнал Agent — должны быть сообщения о рендеринге шаблона и выполнении команды:

   ```bash
   sudo journalctl -u stronghold-agent -n 50
   ```

1. Проверьте сертификат, который отдаёт Nginx:

   ```bash
   openssl s_client -connect www.example.com:443 -servername www.example.com </dev/null 2>/dev/null \
     | openssl x509 -noout -subject -issuer -serial -dates
   ```

   Издатель должен совпадать с промежуточным CA `pki_int`, а срок действия — с `ttl` из шаблона.

1. Проверьте цепочку доверия с корневым сертификатом CA:

   ```bash
   curl --cacert root-ca.pem -sSI https://www.example.com
   ```

1. Проверьте перевыпуск: временно задайте в шаблоне `ttl=10m`, перезапустите Agent и убедитесь, что серийный номер сертификата из шага 2 проверки меняется без ручных действий.

## Очистка

1. Остановите Agent и удалите блок `template` и файл шаблона.
1. Удалите файл `/etc/nginx/tls/www.pem` и уберите директивы `ssl_certificate` из конфигурации Nginx.
1. При необходимости отзовите выпущенный сертификат и удалите роли и политику:

   ```bash
   d8 stronghold write pki_int/revoke serial_number=<serial_number>
   d8 stronghold delete pki_int/roles/web-server
   d8 stronghold delete auth/approle/role/web-server
   d8 stronghold policy delete web-server-tls
   ```
