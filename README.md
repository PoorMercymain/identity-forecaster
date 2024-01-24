# Тестовое задание
Формулировка задачи выглядела следующим образом
![задание](https://github.com/PoorMercymain/identity-forecaster/assets/67076111/e027baeb-03ff-4699-bef1-1a15a39116af)

# Как запустить
Простейшимй способ - переименовать [`.env.example`](https://github.com/PoorMercymain/identity-forecaster/blob/master/.env.example) в `.env` и использовать Docker, прописав в терминале корневой директории проекта `docker-compose up`
После этого сначала запустится Postgres, а после хэлсчека, и сам сервис, который будет доступен на `localhost:8787`

# Файл .env
В [`.env.example`](https://github.com/PoorMercymain/identity-forecaster/blob/master/.env.example) указаны возможные конфигурационные параметры с комментариями. Параметры для постгреса являются необходимыми, в свою очередь параметры сервиса (кроме `IN_CONTAINER`, в случае конфигурации по умолчанию) таковыми не являются, и при их отсутствии будут использованы параметры по умолчанию

# Swagger
После запуска сервиса, перейдя на `http://localhost:8787/swagger/` можно обнаружить Swagger-документацию к API. Часть параметров запросов там описана более подробно

# Postman коллекция
В файле [`identity-forecaster.postman_collection`](https://github.com/PoorMercymain/identity-forecaster/blob/master/identity-forecaster.postman_collection.json) находится коллекция Postman для данного API. При отправке запросов стоит учесть, что запросы к внешним API (при добавлении новой сущности) могут требовать времени, и, так как в данном сервисе это производится асинхронно, после успешного получения запроса будет сразу отправлен ответ, говорящий о том, что данные от пользователя получены успешно, но запросы к API все еще будут производиться (т.е. нужно задать небольшой `delay`)

# Миграции
Миграции применяются при подключении сервиса к постгресу, для этого используется пакет [goose](https://github.com/pressly/goose) и файл sql, лежащий в папке [migrations](https://github.com/PoorMercymain/identity-forecaster/tree/master/internal/app/forecaster/repository/migrations)

# Примеры запросов

`http://localhost:8787/create` - метод добавления новой сущности
<p align="center">
<img src=https://github.com/PoorMercymain/identity-forecaster/assets/67076111/62f396aa-c51b-47e5-bd72-e0ac2d16b9a1>
</p>
<br>

`http://localhost:8787/read` - метод для получения информации о сущностях с различными фильтрами (описаны в swagger), результат отсортирован по `id` (пагинация тоже есть, задается через query параметры `page` и `limit`, когда не задано - показывает первые 10 результатов)
<p align="center">
<img src=https://github.com/PoorMercymain/identity-forecaster/assets/67076111/87a071ca-fb0f-40e3-92f6-fad3ed290272>
</p>
<br>

`http://localhost:8787/update` - метод для обновления сущности, параметр пути (в данном случае 1) - это `id` сущности
<p align="center">
<img src=https://github.com/PoorMercymain/identity-forecaster/assets/67076111/40c4c307-8730-4fd7-9fef-44d8042a8efe>
</p>
<br>

`http://localhost:8787/delete` - метод удаления сущности, параметр пути, так же, как и в предыдущем случае - `id` сущности
<p align="center">
<img src=https://github.com/PoorMercymain/identity-forecaster/assets/67076111/f7303cbe-4080-4d12-80d8-f56860f48937>
</p>
