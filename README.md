# library

Пример команды `mysqldump` для создания резервной копии поисковой базы данных Manticore. Процесс создания резервной копии для базы размером 150 Гб занимает времени более часа. 
```shell
docker exec -it 41ff96f4a1a6 mysqldump library > library_backup.sql
```

```shell
docker exec -it 41ff96f4a1a6 mysql < library_backup.sql
```
Эта команда позволяет восстановить все данные из файла library_backup.sql.
