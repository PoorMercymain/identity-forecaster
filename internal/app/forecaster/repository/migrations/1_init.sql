-- +goose Up
BEGIN TRANSACTION;
CREATE TABLE IF NOT EXISTS persons(id serial unique, name TEXT, surname TEXT, patronymic TEXT, age INTEGER, gender TEXT, nationality TEXT, is_deleted BOOLEAN, primary key (name, surname, patronymic));
COMMIT;

-- +goose Down
BEGIN TRANSACTION;
DROP TABLE IF EXISTS persons;
COMMIT;