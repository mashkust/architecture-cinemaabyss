```puml
@startuml C4_CinemaAbyss_ToBe
!include https://raw.githubusercontent.com/plantuml-stdlib/C4-PlantUML/master/C4_Container.puml

LAYOUT_WITH_LEGEND()
LAYOUT_TOP_DOWN()

title CinemaAbyss — To-Be Container Diagram

Person(user, "Пользователь")


System_Ext(payment_ext, "Payment Provider", "Внешняя платёжная система")
System_Ext(content_ext, "External Content Sources", "Партнёрские сервисы и открытые источники")
System_Ext(recommendation_ext, "Recommendation System", "Внешняя рекомендательная система")

System_Boundary(cinemaabyss, "CinemaAbyss") {

    Container(web_app, "Web App", "Client App", "Веб-клиент")
    Container(mobile_app, "Mobile App", "Client App", "Мобильный клиент")
    Container(tv_app, "Smart TV App", "Client App", "Клиент для Smart TV")

    Container(api_gateway, "API Gateway", "Go", "Единая точка входа для клиентов")

    Container(users_svc, "Users Service", "Go", "Пользователи, аутентификация, профили")
    Container(movies_svc, "Movies Service", "Go", "Фильмы, метаданные, жанры, рейтинги")
    Container(subscriptions_svc, "Subscriptions Service", "Go", "Подписки, тарифы, скидки")
    Container(payments_svc, "Payments Service", "Go", "Платежи и транзакции")
    Container(events_svc, "Events Service", "Go", "Обрабатывает коммуникацию между сервисами на основе событий")

    ContainerQueue(kafka, "Kafka", "Message Broker", "Внутренний обмен доменными событиями")
    ContainerQueue(rabbitmq, "RabbitMQ", "AMQP", "Интеграция с внешней рекомендательной системой")

    ContainerDb(users_db, "Users DB", "PostgreSQL", "Данные пользователей")
    ContainerDb(movies_db, "Movies DB", "PostgreSQL", "Метаданные фильмов")
    ContainerDb(subscriptions_db, "Subscriptions DB", "PostgreSQL", "Подписки и скидки")
    ContainerDb(payments_db, "Payments DB", "PostgreSQL", "Платежи")
}

Rel(user, web_app, "Использует", "HTTPS")
Rel(user, mobile_app, "Использует", "HTTPS")
Rel(user, tv_app, "Использует", "HTTPS")

Rel(web_app, api_gateway, "API-запросы", "REST")
Rel(mobile_app, api_gateway, "API-запросы", "REST")
Rel(tv_app, api_gateway, "API-запросы", "REST")

Rel(api_gateway, users_svc, "Маршрутизация", "REST")
Rel(api_gateway, movies_svc, "Маршрутизация", "REST")
Rel(api_gateway, subscriptions_svc, "Маршрутизация", "REST")
Rel(api_gateway, payments_svc, "Маршрутизация", "REST")

Rel(users_svc, users_db, "Читает/пишет")
Rel(movies_svc, movies_db, "Читает/пишет")
Rel(subscriptions_svc, subscriptions_db, "Читает/пишет")
Rel(payments_svc, payments_db, "Читает/пишет")

Rel(users_svc, kafka, "Публикует события", "Kafka")
Rel(movies_svc, kafka, "Публикует события", "Kafka")
Rel(subscriptions_svc, kafka, "Публикует события", "Kafka")
Rel(payments_svc, kafka, "Публикует события", "Kafka")

Rel(events_svc, kafka, "Читает и обрабатывает события", "Kafka")
Rel(events_svc, rabbitmq, "Публикует запросы / читает ответы", "AMQP")
Rel(recommendation_ext, rabbitmq, "Получает события / возвращает рекомендации", "AMQP")

Rel(payments_svc, payment_ext, "Проводит оплату", "HTTPS")
Rel(movies_svc, content_ext, "Получает данные о контенте", "REST")

@enduml

```
