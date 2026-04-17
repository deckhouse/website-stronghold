---
title: "Ссылки на документацию DKP"
linkTitle: "Ссылки на документацию DKP"
description: "Полезные ссылки на документацию Deckhouse Kubernetes Platform для работы с Deckhouse Stronghold"
weight: 50
---

При развёртывании и эксплуатации **Deckhouse Stronghold** в составе **Deckhouse Kubernetes Platform** часть задач решается средствами самой платформы. В таких случаях следует обращаться к документации DKP.

Ниже приведены разделы документации DKP, которые чаще всего используются при установке, настройке и сопровождении Stronghold в кластере.

## Установка и обновление платформы

Используйте эти разделы, если требуется установить DKP, обновить платформу или проверить требования к целевой версии:

- [Установка](/products/kubernetes-platform/documentation/v1/installing/)
- [Обновление](/products/kubernetes-platform/documentation/v1/admin/configuration/update/)

## Управление доступом

Если необходимо настроить права доступа пользователей и групп в DKP, используйте документацию по RBAC:

- [Управление доступом](/modules/commander/stable/rbac.html)

> Примечание  
> Ссылка ведёт на документацию модуля, который отвечает за настройку управления доступом в DKP.

## DNS и сетевой доступ

Эти материалы полезны, если требуется настроить DNS-разрешение, доступ к Stronghold по доменному имени или сетевую инфраструктуру для работы веб-интерфейса и Ingress:

- [Управление DNS в кластере Kubernetes](/products/kubernetes-platform/documentation/v1/admin/configuration/network/other/dns.html)
- [Модуль ingress-nginx](/modules/ingress-nginx/)

> Примечание  
> Документация по `ingress-nginx` пригодится, если Stronghold публикуется через Ingress и требуется проверить маршрутизацию, точки входа или параметры публикации сервиса.

## Сертификаты и HTTPS

Если для доступа к Stronghold требуется корректная настройка TLS, сертификатов или цепочек доверия, используйте следующий раздел:

- [Управление сертификатами](/products/kubernetes-platform/documentation/v1/admin/configuration/security/certificates.html)

## Разработка и расширение платформы

Если Stronghold используется в расширенных платформенных сценариях, связанных с модульной архитектурой DKP, обратитесь к разделу:

- [Разработка модуля Deckhouse Kubernetes Platform](/products/kubernetes-platform/documentation/v1/architecture/module-development/)

## Резервное копирование и восстановление

Если требуется учесть платформенные сценарии резервного копирования и восстановления кластера DKP, используйте:

- [Резервное копирование и восстановление](/products/kubernetes-platform/documentation/v1/admin/configuration/backup/backup-and-restore.html)

## Когда обращаться к документации DKP

Документация DKP особенно полезна в следующих случаях:

- Stronghold разворачивается в уже существующем кластере DKP;
- необходимо настроить публикацию Stronghold через Ingress;
- требуется настроить DNS, сертификаты или доступ пользователей;
- нужно выполнить обновление платформы или учесть совместимость версий;
- требуется диагностика проблем, связанных не с самим Stronghold, а с инфраструктурой DKP.

## Что дальше

Вернитесь к следующему разделу руководства по развёртыванию или перейдите к разделу [Проверка работоспособности](/products/stronghold/documentation/deployment/functionality_check/).
