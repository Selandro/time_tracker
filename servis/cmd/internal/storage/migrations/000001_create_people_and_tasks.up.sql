CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    passport_serie INTEGER NOT NULL,
    passport_number INTEGER NOT NULL,
    surname VARCHAR(50) NOT NULL,
    name VARCHAR(50) NOT NULL,
    patronymic VARCHAR(50),
    address TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS tasks (
    id_task SERIAL PRIMARY KEY,
    task_name VARCHAR(100) NOT NULL
);

CREATE TABLE IF NOT EXISTS users_tasks (
    user_id INTEGER REFERENCES users(id),
    id_task INTEGER REFERENCES tasks(id_task),
    task_name VARCHAR(100) NOT NULL,
    start_time TIMESTAMP,
    end_time TIMESTAMP,
    total_minutes INTEGER,
    CONSTRAINT unique_user_task UNIQUE (user_id, id_task)
);

INSERT INTO tasks (id_task, task_name)
VALUES (1, 'работаю над таской 1');
INSERT INTO tasks (id_task, task_name)
VALUES (2, 'работаю над таской 2');
INSERT INTO tasks (id_task, task_name)
VALUES (3, 'работаю над таской 3');