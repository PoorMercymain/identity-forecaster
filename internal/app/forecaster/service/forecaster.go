package service

import (
	"context"

	"identity-forecaster/internal/app/forecaster/domain"
)

var _ domain.ForecasterService = (*forecaster)(nil)

type forecaster struct {
	repo domain.ForecasterRepository
}

func New(repo domain.ForecasterRepository) *forecaster {
	return &forecaster{repo: repo}
}

func (s *forecaster) CreatePerson(ctx context.Context, person domain.Person, apiData domain.DataFromAPI) error {
	return s.repo.CreatePerson(ctx, person, apiData)
}

func (s *forecaster) DeletePersonByID(ctx context.Context, id int) error {
	return s.repo.DeletePersonByID(ctx, id)
}

func (s *forecaster) UpdatePerson(ctx context.Context, id int, data domain.PersonWithAPIData) error {
	return s.repo.UpdatePerson(ctx, id, data)
}

func (s *forecaster) ReadPersons(ctx context.Context, page int, limit int, filters domain.Filters) ([]domain.PersonFromDB, error) {
	return s.repo.ReadPersons(ctx, page, limit, filters)
}
