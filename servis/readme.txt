//перед запуском main.go необходимо установить переменные окружения для файла конфигурации config/local.yaml
//в файле storage.go необходимо изменить путь к файлу миграции 
files := []string{"C:/dev/projects/time_tracker/servis/cmd/internal/storage/migrations/000001_create_people_and_tasks.up.sql"}
на ваш путь

//добавить нового пользователя с доп информацией из стороннего API
curl -X POST -H "Content-Type: application/json" -d "{\"passportNumber\":\"1234 567890\"}" http://localhost:8080/adduser

//начать отсчет времени, происходит одновременно с добавлением новой таски пользователю
curl -X POST -H "Content-Type: application/json" -d "{\"user_id\": 1, \"id_task\": 1}" http://localhost:8080/start_task


//остановить отсчет времени
curl -X POST -H "Content-Type: application/json" -d "{\"user_id\": 1, \"id_task\": 1}" http://localhost:8080/end_task

//получить все задачи пользователя за период с сортировкой
curl -X GET "http://localhost:8080/user_task?user_id=1&start_date=2024-07-01&end_date=2024-07-31"

//получить список пользователей с фильтрацией и пагинацией
curl -X GET "http://localhost:8080/users?passport_serie=1234&surname=Vadimov&page=1&limit=10"

//изменить личные данные пользователя
curl -X PUT -H "Content-Type: application/json" -d "{\"passport_serie\": 7777, \"passport_number\": 777777, \"surname\": \"Иванов\", \"name\": \"Иван\", \"patronymic\": \"Иванович\", \"address\": \"ул. Пушкина, дом Колотушкина\"}" http://localhost:8080/update_user/1

//удалить пользователя, вместе с этим и удаляются все задачи пользователя
curl -X DELETE "http://localhost:8080/delete_user?user_id=1"