package app_errors

import "errors"

var (
	ErrWrongContentType          = errors.New("content type of the request is wrong")
	ErrWrongStatusCode           = errors.New("status code of a response of a required API is wrong")
	ErrNoRowsAffected            = errors.New("no rows were affected by the request")
	ErrNoRowsFound               = errors.New("no rows found")
	ErrIncorrectQueryParam       = errors.New("incorrect value of a query param provided")
	ErrRequiredFieldsNotProvided = errors.New("required fields not provided")
	ErrUniqueViolation           = errors.New("the entity already exists in table")
)
