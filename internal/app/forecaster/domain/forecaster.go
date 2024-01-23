package domain

import "context"

type ForecasterService interface {
	CreatePerson(ctx context.Context, person Person, dataFromAPI DataFromAPI) error
	DeletePersonByID(ctx context.Context, id int) error
	UpdatePerson(ctx context.Context, id int, data PersonWithAPIData) error
	ReadPersons(ctx context.Context, page int, limit int, filters Filters) ([]PersonFromDB, error)
}

//go:generate mockgen -destination=mocks/forecaster_repo_mock.gen.go -package=mocks . ForecasterRepository
type ForecasterRepository interface {
	CreatePerson(ctx context.Context, person Person, dataFromAPI DataFromAPI) error
	DeletePersonByID(ctx context.Context, id int) error
	UpdatePerson(ctx context.Context, id int, data PersonWithAPIData) error
	ReadPersons(ctx context.Context, page int, limit int, filters Filters) ([]PersonFromDB, error)
}
