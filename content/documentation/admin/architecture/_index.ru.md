---
title: "Архитектура"
description: "Компоненты Stronghold, путь запроса, схемы развёртывания в DKP и standalone, сетевые порты и модель угроз."
weight: 65
---

В разделе описано, из чего состоит Stronghold и как он развёртывается:

- [Обзор архитектуры](./overview/) — компоненты сервера, путь запроса от клиента до хранилища, seal и криптографический барьер.
- [Развёртывание в DP](./deployment-dkp/) — как модуль `stronghold` размещается в кластере Deckhouse Platform.
- [Standalone-развёртывание](./deployment-standalone/) — HA-кластер из нескольких серверов с хранилищем Raft.
- [Сетевые порты](./ports/) — какие соединения нужно разрешить между клиентами, узлами и внешними системами.
- [Модель угроз](./threat-model/) — границы доверия, от чего защищают барьер, seal, HSM и seal wrap, а от чего нет.
- [Функциональные характеристики](./functional-specifications/) — основные функции Stronghold.

Слои узла базового Stronghold и Stronghold EE, HA-кластер, performance standby и межкластерная репликация описаны на странице [«Архитектура: Stronghold и Stronghold EE»](../replication/architecture/).
