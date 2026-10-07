---
title: "TLS-сертификаты"
description: "Замена TLS-сертификатов слушателей Stronghold в Linux без перезапуска сервиса и обновление сертификатов Stronghold в DKP."
weight: 60
---

Страница описывает замену собственных TLS-сертификатов Stronghold: сертификатов слушателя API и сертификатов, используемых узлами для взаимодействия в кластере Raft. Выпуск сертификатов для приложений описан в разделе [PKI](../../../user/secrets-engines/pki/).

Контролируйте срок действия сертификатов с помощью алертов из раздела [«Мониторинг»](../monitoring/#примеры-правил-алертинга) и заменяйте их заранее.

## Stronghold в Linux

Сертификат и ключ слушателя задаются параметрами `tls_cert_file` и `tls_key_file` блока `listener "tcp"` (см. [«Настройка»](../../../install/standalone/configuration/#listener)). Файл сертификата должен содержать полную цепочку: сначала сертификат сервера, затем промежуточные сертификаты CA.

### Проверка текущего сертификата

```shell
openssl x509 -in /opt/stronghold/tls/node-1-cert.pem -noout -subject -issuer -dates -ext subjectAltName
```

### Замена сертификата без перезапуска

Stronghold перечитывает файлы сертификата и ключа слушателя при получении сигнала `SIGHUP`. Изменение путей к файлам при этом не применяется, поэтому новые файлы размещайте по тем же путям.

Выполните на каждом узле по очереди:

1. Выпустите новый сертификат с теми же SAN (DNS-имена и IP-адреса узла, адрес балансировщика, если он используется).
1. Проверьте, что сертификат соответствует ключу:

   ```shell
   openssl x509 -in node-1-cert.pem -noout -pubkey | sha256sum
   openssl pkey -in node-1-key.pem -pubout | sha256sum
   ```

   Хеши должны совпадать.

1. Сохраните резервную копию текущих файлов и замените их новыми, сохранив владельца и права доступа:

   ```shell
   cp -a /opt/stronghold/tls /opt/stronghold/tls.bak-$(date +%F)
   install -o stronghold -g stronghold -m 0600 node-1-key.pem /opt/stronghold/tls/node-1-key.pem
   install -o stronghold -g stronghold -m 0644 node-1-cert.pem /opt/stronghold/tls/node-1-cert.pem
   ```

1. Отправьте процессу сигнал `SIGHUP`:

   ```shell
   systemctl reload stronghold
   ```

1. Проверьте, что сервер отдаёт новый сертификат:

   ```shell
   openssl s_client -connect raft-node-1.demo.tld:8200 </dev/null 2>/dev/null \
     | openssl x509 -noout -dates
   ```

1. Проверьте журналы на ошибки перезагрузки сертификата:

   ```shell
   journalctl -u stronghold.service --since "5 minutes ago"
   ```

Если ключ зашифрован паролем, при перезагрузке по `SIGHUP` пароль должен совпадать с исходным.

Сертификаты `retry_join` (`leader_client_cert_file`, `leader_ca_cert_file`, `leader_client_key_file`) по `SIGHUP` не перечитываются: они загружаются при попытке присоединения к кластеру, поэтому новые файлы применяются после перезапуска узла. Перечитывание по `SIGHUP` поддерживают только сертификаты TLS-листенеров.

### Замена CA

При смене CA, выпустившего сертификаты узлов, соблюдайте порядок, чтобы узлы и клиенты не потеряли доверие друг к другу:

1. Добавьте новый CA в доверенные на всех клиентах и в файлы `leader_ca_cert_file` на всех узлах (файл может содержать несколько сертификатов CA).
1. Перезапустите узлы по одному, начиная со standby-узлов, и дождитесь, пока каждый узел вернётся в кластер (`d8 stronghold operator raft list-peers`). Узлы с Shamir seal после перезапуска нужно распечатать.
1. Замените сертификаты узлов на выпущенные новым CA, как описано выше.
1. После замены на всех узлах удалите старый CA из доверенных.

Перезапуск нужен: файлы `leader_ca_cert_file` читаются только при присоединении узла.

## Stronghold в DP

В DP поддерживаются следующие способы внешнего доступа (инлеты): `Ingress`, `GatewayAPI`, `LoadBalancer`, `NodePort` и `None`. Для `LoadBalancer`, `NodePort` и `None` обязателен режим `https.mode: CustomCertificate`. Сертификат для домена `stronghold.<домен>` выпускается и хранится согласно настройкам `https` (см. [«Настройка Stronghold»](../../../install/dkp/configuration/#способы-организации-доступа-через-инлет-ingress)).

Stronghold использует два вида сертификатов:

- **Внешний** сертификат: секрет `ingress-tls` в режиме `CertManager` или секрет `ingress-tls-customcertificate` в режиме `CustomCertificate` в пространстве имён `d8-stronghold`. Монтируется в Pod в `/stronghold/tls` и используется listener'ом API на порту `8200`.
- **Внутренний** сертификат: секрет `stronghold-tls` (`ca.crt`, `tls.crt`, `tls.key`). Выпускается хуком модуля как самоподписанный и монтируется в `/stronghold/tls-internal`. Используется на портах `8300`, `8301` и `8500` (взаимодействие между Pod'ами и с прокси). В SAN входят `127.0.0.1`, `stronghold`, `*.stronghold-internal`, `stronghold.d8-stronghold` и `stronghold.d8-stronghold.svc`. Им управляет модуль, заменять его вручную не требуется.

| Способ | Обновление сертификата |
| --- | --- |
| `CertManager` с ClusterIssuer (Let's Encrypt или собственный CA) | Автоматически, средствами cert-manager |
| `CustomCertificate` | Вручную, обновлением секрета в пространстве имён `d8-system` |

Ресурс Certificate `stronghold` существует только при инлете `Ingress` и режиме `CertManager`.

### Проверка сертификата

```shell
d8 k -n d8-stronghold get certificate
d8 k -n d8-stronghold get secret ingress-tls -o jsonpath='{.data.tls\.crt}' \
  | base64 -d | openssl x509 -noout -subject -issuer -dates
```

Команды приведены для режима `CertManager`. В режиме `CustomCertificate` используйте секрет `ingress-tls-customcertificate` (ресурса Certificate при этом нет).

### Сертификаты cert-manager

cert-manager продлевает сертификат автоматически до истечения срока действия. Если сертификат не обновился, проверьте ресурсы CertificateRequest:

```shell
d8 k -n d8-stronghold get certificaterequest
d8 k -n d8-stronghold describe certificaterequest <NAME>
```

Чтобы принудительно перевыпустить сертификат, удалите секрет `ingress-tls` — cert-manager выпустит его заново.

<!-- TODO(verify): безопасно ли удалять ingress-tls для принудительного перевыпуска (под ContainerCreating при отсутствии секрета), либо рекомендуется cmctl renew. -->

### Собственный сертификат

При режиме `CustomCertificate` замените содержимое секрета, указанного в `settings.modules.https.customCertificate.secretName`:

```shell
d8 k -n d8-system create secret tls mycompany-wildcard-tls \
  --cert=kubernetes_fullchain.crt --key=kubernetes.key \
  --dry-run=client -o yaml | d8 k apply -f -
```

Deckhouse скопирует обновлённый сертификат в пространства имён модулей. Проверьте, что Stronghold отдаёт новый сертификат:

```shell
openssl s_client -connect stronghold.mycompany.tld:443 </dev/null 2>/dev/null \
  | openssl x509 -noout -dates
```

Если CA изменился и в `user-authn` используется `dexCAMode: FromIngressSecret`, убедитесь, что Dex получил новый CA, иначе вход через OIDC перестанет работать.

<!-- TODO(verify): время и механизм распространения CustomCertificate в d8-stronghold; перезапускаются ли поды Stronghold; как ротируются внутренние сертификаты между подами (порт 8300). -->
