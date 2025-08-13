# library

### Запуск парсера
Для запуска необходимо установить переменную окружения
`LIBRARY_CONFIG_PATH`

Windows:
```shell
SET LIBRARY_CONFIG_PATH=c:\library\local.yaml
```

```shell
library-parser.exe
```

Ubuntu:
```shell
LIBRARY_CONFIG_PATH=./library/local.yaml ./library-parser.linux.amd64
```

### Создание резервной копии

Пример команды `mysqldump` для создания резервной копии поисковой базы данных Manticore. Процесс создания резервной копии для базы размером 150 Гб занимает времени более часа. 
```shell
docker exec -it 41ff96f4a1a6 mysqldump library > library_backup.sql
```

```shell
docker exec -it 41ff96f4a1a6 mysql < library_backup.sql
```
Эта команда позволяет восстановить все данные из файла library_backup.sql.

Пересборка и запуск прокси
```
docker compose build manticore-proxy
docker compose up -d
```