package repository

import (
	"context"
	"errors"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/jackc/pgx/v5"

	appErrors "identity-forecaster/internal/app/forecaster/app-errors"
	"identity-forecaster/internal/app/forecaster/domain"
	"identity-forecaster/internal/pkg/logger"
)

var (
	_ domain.ForecasterRepository = (*forecaster)(nil)
)

type forecaster struct {
	*postgres
}

func New(pg *postgres) *forecaster {
	return &forecaster{pg}
}
func (r *forecaster) CreatePerson(ctx context.Context, person domain.Person, apiData domain.DataFromAPI) error {
	return r.WithTransaction(ctx, func(ctx context.Context, tx pgx.Tx) error {
		logger.Logger().Debugln("CreatePerson with args:", person, apiData)
		_, err := tx.Exec(ctx, "INSERT INTO persons(name, surname, patronymic, age, gender, nationality, is_deleted)"+
			" VALUES ($1, $2, $3, $4, $5, $6, $7) ON CONFLICT ON CONSTRAINT persons_pkey DO UPDATE SET age ="+
			" EXCLUDED.age, gender = EXCLUDED.gender, nationality = EXCLUDED.nationality, is_deleted = "+
			"EXCLUDED.is_deleted WHERE persons.is_deleted = TRUE", person.Name, person.Surname, person.Patronymic, apiData.Age, apiData.Gender,
			apiData.Nationality, false)
		if err != nil {
			return err
		}

		return nil
	})
}

func (r *forecaster) DeletePersonByID(ctx context.Context, id int) error {
	return r.WithTransaction(ctx, func(ctx context.Context, tx pgx.Tx) error {
		logger.Logger().Debugln("DeletePersonByID with id:", id)
		tag, err := tx.Exec(ctx, "UPDATE persons SET is_deleted = true WHERE id = $1 AND is_deleted != true", id)
		if err != nil {
			return err
		}

		if tag.RowsAffected() == 0 {
			return appErrors.ErrNoRowsAffected
		}

		return nil
	})
}

func (r *forecaster) UpdatePerson(ctx context.Context, id int, data domain.PersonWithAPIData) error {
	return r.WithTransaction(ctx, func(ctx context.Context, tx pgx.Tx) error {
		logger.Logger().Debugln("UpdatePersons with args:", id, data)
		var previousValues domain.PersonWithAPIData
		err := tx.QueryRow(ctx, "SELECT name, surname, patronymic, age, gender, nationality, is_deleted FROM persons WHERE id = $1",
			id).Scan(&previousValues.Name, &previousValues.Surname, &previousValues.Patronymic,
			&previousValues.Age, &previousValues.Gender, &previousValues.Nationality, &previousValues.IsDeleted)

		if errors.Is(err, pgx.ErrNoRows) {
			return appErrors.ErrNoRowsFound
		}

		if err != nil {
			return err
		}

		data.ReplaceDefaultValuesWithFieldsOfStruct(previousValues)

		tag, err := tx.Exec(ctx, "UPDATE persons SET name = $1, surname = $2, patronymic = $3, age = $4, gender = $5,"+
			" nationality = $6, is_deleted = $7 WHERE id = $8", data.Name, data.Surname, data.Patronymic, data.Age,
			data.Gender, data.Nationality, *data.IsDeleted, id)

		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
				return appErrors.ErrUniqueViolation
			}

			return err
		}

		if tag.RowsAffected() == 0 {
			return appErrors.ErrNoRowsAffected
		}

		return nil
	})
}

func (r *forecaster) ReadPersons(ctx context.Context, page int, limit int, filters domain.Filters) ([]domain.PersonFromDB, error) {
	persons := make([]domain.PersonFromDB, 0)

	logger.Logger().Debugln("ReadPersons with args:", page, limit, filters)
	err := r.WithConnection(ctx, func(ctx context.Context, conn *pgxpool.Conn) error {
		rows, err := conn.Query(ctx, "SELECT id, name, surname, patronymic, age, gender, nationality FROM persons "+
			"WHERE (id >= $1 AND id < $2) AND (age >= $3 AND age < $4) AND ($5::TEXT IS NULL OR name = $5::TEXT) AND "+
			"($6::TEXT IS NULL OR surname = $6::TEXT) AND ($7::TEXT IS NULL OR patronymic = $7::TEXT) AND ($8::TEXT "+
			"IS NULL OR gender = $8::TEXT) AND ($9::TEXT IS NULL OR nationality = $9::TEXT) AND is_deleted != TRUE "+
			"ORDER BY id OFFSET $10 LIMIT $11", filters.IDMoreThan.Value, filters.IDLessThan.Value,
			filters.AgeMoreThan.Value, filters.AgeLessThan.Value, filters.NameEqualTo.Value, filters.SurnameEqualTo.Value,
			filters.PatronymicEqualTo.Value, filters.GenderEqualTo.Value, filters.NationalityEqualTo.Value, (page-1)*limit, limit)

		if err != nil {
			return err
		}

		var i int

		for rows.Next() {
			var person domain.PersonFromDB

			err = rows.Scan(&person.ID, &person.Name, &person.Surname, &person.Patronymic, &person.Age, &person.Gender, &person.Nationality)
			if err != nil {
				return err
			}

			persons = append(persons, person)
			i++
		}

		if i == 0 {
			return appErrors.ErrNoRowsFound
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return persons, nil
}
