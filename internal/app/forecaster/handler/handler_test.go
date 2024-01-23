package handler

import (
	"github.com/golang/mock/gomock"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
	appErrors "identity-forecaster/internal/app/forecaster/app-errors"
	"identity-forecaster/internal/app/forecaster/domain"
	"identity-forecaster/internal/app/forecaster/domain/mocks"
	"identity-forecaster/internal/app/forecaster/service"
	"identity-forecaster/internal/pkg/logger"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

func testRouter(t *testing.T) *echo.Echo {
	e := echo.New()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockForecasterRepository(ctrl)

	mockRepo.EXPECT().CreatePerson(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).MaxTimes(1)

	mockRepo.EXPECT().DeletePersonByID(gomock.Any(), gomock.Any()).Return(nil).MaxTimes(1)
	mockRepo.EXPECT().DeletePersonByID(gomock.Any(), gomock.Any()).Return(appErrors.ErrNoRowsAffected).MaxTimes(1)

	mockRepo.EXPECT().UpdatePerson(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).MaxTimes(2)

	mockRepo.EXPECT().ReadPersons(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(make([]domain.PersonFromDB, 0), appErrors.ErrNoRowsFound).MaxTimes(1)
	mockRepo.EXPECT().ReadPersons(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(make([]domain.PersonFromDB, 0), nil).MaxTimes(2)

	s := service.New(mockRepo)

	var wg sync.WaitGroup

	h := New(s, &wg, 1, 1, []string{"http://localhost:8080/agify", "http://localhost:8080/genderize", "http://localhost:8080/nationalize"})

	e.POST("/create", h.CreatePerson)
	e.DELETE("/delete/:id", h.DeletePersonByID)
	e.PUT("/update/:id", h.UpdatePerson)
	e.GET("/read", h.ReadPersons)

	return e
}

func activateTestAPIs() {
	e := echo.New()

	e.GET("/agify", Age)
	e.GET("/genderize", Gender)
	e.GET("/nationalize", Nation)

	go e.Start(":8080")
}

func Age(c echo.Context) error {
	c.Response().Write([]byte("{\"age\": 20}"))
	c.Response().WriteHeader(http.StatusOK)
	return nil
}

func Gender(c echo.Context) error {
	c.Response().Write([]byte("{\"gender\": \"male\"}"))
	c.Response().WriteHeader(http.StatusOK)
	return nil
}

func Nation(c echo.Context) error {
	c.Response().Write([]byte("{\"country\": [{\"country_id\": \"QWE\",\"probability\": 0.2}]}"))
	c.Response().WriteHeader(http.StatusOK)
	return nil
}

func request(t *testing.T, ts *httptest.Server, code int, method, content, body, endpoint string) *http.Response {
	req, err := http.NewRequest(method, ts.URL+endpoint, strings.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", content)

	resp, err := ts.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	logger.Logger().Infoln(string(b))

	require.Equal(t, code, resp.StatusCode)

	return resp
}

func TestCreate(t *testing.T) {
	ts := httptest.NewServer(testRouter(t))
	activateTestAPIs()

	defer ts.Close()

	var testTable = []struct {
		endpoint string
		method   string
		content  string
		code     int
		body     string
	}{
		{
			"/create",
			http.MethodPost,
			"application/json",
			http.StatusBadRequest,
			"{}",
		},
		{
			"/create",
			http.MethodPost,
			"application/json",
			http.StatusBadRequest,
			"{\"name\": \"Dmitriy\"}",
		},
		{
			"/create",
			http.MethodPost,
			"application/json",
			http.StatusBadRequest,
			"{\"name\": \"Dmitriy\", \"surname\": \"Sidorov\", \"surname\": \"Sidorov\"}",
		},
		{
			"/create",
			http.MethodPost,
			"application/json",
			http.StatusAccepted,
			"{\"name\": \"Dmitriy\", \"surname\": \"Sidorov\"}",
		},
	}

	for _, testCase := range testTable {
		resp := request(t, ts, testCase.code, testCase.method, testCase.content, testCase.body, testCase.endpoint)
		resp.Body.Close()
	}
}

func TestDelete(t *testing.T) {
	ts := httptest.NewServer(testRouter(t))
	activateTestAPIs()

	defer ts.Close()

	var testTable = []struct {
		endpoint string
		method   string
		content  string
		code     int
		body     string
	}{
		{
			"/delete/1",
			http.MethodDelete,
			"",
			http.StatusOK,
			"",
		},
		{
			"/delete/abc",
			http.MethodDelete,
			"",
			http.StatusBadRequest,
			"",
		},
		{
			"/delete/1",
			http.MethodDelete,
			"",
			http.StatusNotFound,
			"",
		},
	}

	for _, testCase := range testTable {
		resp := request(t, ts, testCase.code, testCase.method, testCase.content, testCase.body, testCase.endpoint)
		resp.Body.Close()
	}
}

func TestUpdate(t *testing.T) {
	ts := httptest.NewServer(testRouter(t))
	activateTestAPIs()

	defer ts.Close()

	var testTable = []struct {
		endpoint string
		method   string
		content  string
		code     int
		body     string
	}{
		{
			"/update/1",
			http.MethodPut,
			"application/json",
			http.StatusOK,
			"{\"name\": \"Dmitriy\",\"surname\": \"тsщ\",\"patronymic\": \"\",\"gender\": \"helicopter\",\"nationality\": \"kitten\"}",
		},
		{
			"/update/1",
			http.MethodPut,
			"application/json",
			http.StatusOK,
			"{}",
		},
		{
			"/update/a",
			http.MethodPut,
			"application/json",
			http.StatusBadRequest,
			"{}",
		},
		{
			"/update/1",
			http.MethodPut,
			"",
			http.StatusBadRequest,
			"{}",
		},
	}

	for _, testCase := range testTable {
		resp := request(t, ts, testCase.code, testCase.method, testCase.content, testCase.body, testCase.endpoint)
		resp.Body.Close()
	}
}

func TestRead(t *testing.T) {
	ts := httptest.NewServer(testRouter(t))
	activateTestAPIs()

	defer ts.Close()

	var testTable = []struct {
		endpoint string
		method   string
		content  string
		code     int
		body     string
	}{
		{
			"/read",
			http.MethodGet,
			"",
			http.StatusNoContent,
			"",
		},
		{
			"/read",
			http.MethodGet,
			"",
			http.StatusOK,
			"",
		},
		{
			"/read?agegt=20",
			http.MethodGet,
			"",
			http.StatusOK,
			"",
		},
		{
			"/read?agegt=20&agelt=10",
			http.MethodGet,
			"",
			http.StatusBadRequest,
			"",
		},
	}

	for _, testCase := range testTable {
		resp := request(t, ts, testCase.code, testCase.method, testCase.content, testCase.body, testCase.endpoint)
		resp.Body.Close()
	}
}
