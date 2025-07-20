# kob-library-app

Статус: **в разработке**

Копирование репозитория для разработки

Выберите или создайте папку с проектами, например c:\terratensor,

Для создания наберите в консоли Git CMD
```
mkdir ./terratensor
```

Выбрать созданную папку, набрать в консоли Git CMD:
```
cd ./terratensor
```

Для клонирования репозитория terratensor/vpsssr_library_parser, наберите в консоли Git CMD:

```
git clone https://github.com/terratensor/kob-library-app.git
```

Запуск приложения на локальном компьютере для разработки

Все команды необходимо запускать из папки проекта
`terratensor/kob-library-app`

```
docker compose down --remove-orphans
```

```
docker compose pull
```

```
docker compose up -d
```

Проект можно открыть в браузере по ссылке:
http://localhost:8000/

```
REGISTRY=localhost IMAGE_TAG=0 APP_DB_PASSWORD_FILE-./docker/development/secrets/app_db_password docker compose -f docker-compose-production.yml up -d
```


Создание дампа БД:
```
docker exec -it kob-library-postgres bash
```

```
 pg_dump --dbname=vpsssr-library --username=app --host=postgres | gzip -9 > vpsssr-library_backup.gz
```
Enter password
`exit`

Копирование файла из контейнера:

```
docker cp kob-library-postgres:/vpsssr-library_backup.gz vpsssr-library_backup.gz
```

```
scp -P 22 /mnt/c/terratensor/backup/filename.sql root@host:/home/deploy/filname.sql
```

`psql -U app vpsssr-library`
`drop schema public CASCADE;`

остановить стек, удалить волум бд, перезапустить стек, скопировать файл дампа в контейнер, запустить

```
cp vpsssr-library_backup-filename.gz book-parser-postgres:app/vpsssr-library_backup-filename.gz

gzip -d vpsssr-library_backup-filename.gz

psql -U app -d lib < vpsssr-library_backup-filename.sql

```
