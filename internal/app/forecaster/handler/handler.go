package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/avast/retry-go/v4"
	"github.com/labstack/echo/v4"

	appErrors "identity-forecaster/internal/app/forecaster/app-errors"
	"identity-forecaster/internal/app/forecaster/domain"
	"identity-forecaster/internal/pkg/logger"
	jsonDuplicateChecker "identity-forecaster/pkg/json-duplicate-checker"
	mimeChecker "identity-forecaster/pkg/json-mime-checker"
)

type forecaster struct {
	srv  domain.ForecasterService
	apis []string
	*sync.WaitGroup
	retriesAmount              uint
	millisecondsBetweenRetries uint
}

func New(srv domain.ForecasterService, wg *sync.WaitGroup, retriesAmount uint, millisecondsBetweenRetries uint, apis []string) *forecaster {
	return &forecaster{srv: srv, WaitGroup: wg, apis: apis, retriesAmount: retriesAmount, millisecondsBetweenRetries: millisecondsBetweenRetries}
}

// @Tags Persons
// @Summary Запрос добавления сущности
// @Description Запрос для добавления информации о новой сущности
// @Accept json
// @Param input body domain.Person true "person info"
// @Success 202
// @Failure 400
// @Failure 500
// @Router /create [post]
func (h *forecaster) CreatePerson(c echo.Context) error {
	if !mimeChecker.IsJSONContentTypeCorrect(c.Request()) {
		c.Response().WriteHeader(http.StatusBadRequest)
		logger.Logger().Debugln(appErrors.ErrWrongContentType)
		return appErrors.ErrWrongContentType
	}

	bytesToCheck, err := io.ReadAll(c.Request().Body)
	if err != nil {
		c.Response().WriteHeader(http.StatusBadRequest)
		logger.Logger().Debugln(err)
		return err
	}

	reader := bytes.NewReader(bytes.Clone(bytesToCheck))

	err = jsonDuplicateChecker.CheckDuplicatesInJSON(json.NewDecoder(reader), nil)
	if err != nil {
		c.Response().WriteHeader(http.StatusBadRequest)
		logger.Logger().Debugln(err)
		return err
	}

	c.Request().Body = io.NopCloser(bytes.NewBuffer(bytesToCheck))

	d := json.NewDecoder(c.Request().Body)
	d.DisallowUnknownFields()

	var person domain.Person

	if err = d.Decode(&person); err != nil {
		c.Response().WriteHeader(http.StatusBadRequest)
		logger.Logger().Debugln(err)
		return err
	}

	if person.Name == "" || person.Surname == "" {
		c.Response().WriteHeader(http.StatusBadRequest)
		logger.Logger().Debugln(appErrors.ErrRequiredFieldsNotProvided)
		return appErrors.ErrRequiredFieldsNotProvided
	}

	var data, resultData domain.DataFromAPI

	h.Add(1)
	go func() {
		defer h.Done()

		for _, api := range h.apis {
			if strings.Contains(api, "nationalize") {
				data, err = h.GetDataFromAPI(person.Surname, api, h.retriesAmount, h.millisecondsBetweenRetries)
			} else {
				data, err = h.GetDataFromAPI(person.Name, api, h.retriesAmount, h.millisecondsBetweenRetries)
			}

			if err != nil {
				logger.Logger().Debugln(err)
				return
			}

			if strings.Contains(api, "agify") {
				resultData.Age = data.Age
			} else if strings.Contains(api, "genderize") {
				resultData.Gender = data.Gender
			} else if strings.Contains(api, "nationalize") {
				logger.Logger().Infoln(data.CountrySlice)
				resultData.Nationality = data.CountrySlice[0].CountryID
			}
		}

		if err = h.srv.CreatePerson(context.Background(), person, resultData); err != nil {
			c.Response().WriteHeader(http.StatusInternalServerError)
			logger.Logger().Debugln(err)
			return
		}
	}()

	logger.Logger().Infoln("successfully got info to process")
	c.Response().WriteHeader(http.StatusAccepted)
	return nil
}

func (h *forecaster) GetDataFromAPI(name string, api string, retriesAmount uint, interval uint) (domain.DataFromAPI, error) {
	var dataToReturn domain.DataFromAPI

	url := api + "?name=" + name

	err := retry.Do(func() error {
		logger.Logger().Infoln("attempt to get info from external api...")
		resp, err := http.Get(url)
		logger.Logger().Infoln(resp.StatusCode)
		if err != nil {
			logger.Logger().Infoln(err)
			return err
		}
		defer resp.Body.Close()

		if !(resp.StatusCode > 199 && resp.StatusCode < 400) {
			return appErrors.ErrWrongStatusCode
		}

		var data domain.DataFromAPI
		d := json.NewDecoder(resp.Body)
		err = d.Decode(&data)
		if err != nil {
			logger.Logger().Infoln(err)
			return err
		}

		dataToReturn = data

		return nil
	}, retry.Attempts(retriesAmount), retry.Delay(time.Duration(interval)*time.Millisecond))

	if err != nil {
		logger.Logger().Debugln(err)
		return domain.DataFromAPI{}, err
	}

	logger.Logger().Infoln("successfully got info")
	return dataToReturn, nil
}

// @Tags Persons
// @Summary Запрос удаления сущности
// @Description Запрос для удаления сущности
// @Param id path int true "person id" Example(1)
// @Success 200
// @Failure 400
// @Failure 404
// @Failure 500
// @Router /delete/{id} [delete]
func (h *forecaster) DeletePersonByID(c echo.Context) error {
	idStr := c.Param("id")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.Response().WriteHeader(http.StatusBadRequest)
		logger.Logger().Debugln(err)
		return err
	}

	err = h.srv.DeletePersonByID(c.Request().Context(), id)
	if errors.Is(err, appErrors.ErrNoRowsAffected) {
		c.Response().WriteHeader(http.StatusNotFound)
		logger.Logger().Debugln(err)
		return err
	}

	if err != nil {
		c.Response().WriteHeader(http.StatusInternalServerError)
		logger.Logger().Debugln(err)
		return err
	}

	c.Response().WriteHeader(http.StatusOK)
	return nil
}

// @Tags Persons
// @Summary Запрос обновления информации о сущности
// @Description Запрос для обновления информации о сущности (кроме id)
// @Accept json
// @Param input body domain.PersonWithAPIData true "person description"
// @Param id path int true "person id" Example(1)
// @Success 200
// @Failure 400
// @Failure 404
// @Failure 500
// @Router /update/{id} [put]
func (h *forecaster) UpdatePerson(c echo.Context) error {
	idStr := c.Param("id")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.Response().WriteHeader(http.StatusBadRequest)
		logger.Logger().Debugln(err)
		return err
	}

	if !mimeChecker.IsJSONContentTypeCorrect(c.Request()) {
		c.Response().WriteHeader(http.StatusBadRequest)
		logger.Logger().Debugln(appErrors.ErrWrongContentType)
		return appErrors.ErrWrongContentType
	}

	bytesToCheck, err := io.ReadAll(c.Request().Body)
	if err != nil {
		c.Response().WriteHeader(http.StatusBadRequest)
		logger.Logger().Debugln(err)
		return err
	}

	reader := bytes.NewReader(bytes.Clone(bytesToCheck))

	err = jsonDuplicateChecker.CheckDuplicatesInJSON(json.NewDecoder(reader), nil)
	if err != nil {
		c.Response().WriteHeader(http.StatusBadRequest)
		logger.Logger().Debugln(err)
		return err
	}

	c.Request().Body = io.NopCloser(bytes.NewBuffer(bytesToCheck))

	d := json.NewDecoder(c.Request().Body)
	d.DisallowUnknownFields()

	var newPersonData domain.PersonWithAPIData

	if err = d.Decode(&newPersonData); err != nil {
		c.Response().WriteHeader(http.StatusBadRequest)
		logger.Logger().Debugln(err)
		return err
	}

	err = h.srv.UpdatePerson(c.Request().Context(), id, newPersonData)

	if errors.Is(err, appErrors.ErrNoRowsFound) || errors.Is(err, appErrors.ErrNoRowsAffected) {
		c.Response().WriteHeader(http.StatusNotFound)
		logger.Logger().Debugln(err)
		return err
	}

	if err != nil {
		c.Response().WriteHeader(http.StatusInternalServerError)
		logger.Logger().Debugln(err)
		return err
	}

	logger.Logger().Infoln("successfully updated")
	c.Response().WriteHeader(http.StatusOK)
	return nil
}

// @Tags Persons
// @Summary Запрос чтения информации о сущностях
// @Description Запрос для получения сохраненной информации о сущностях с возможностью применения фильтров и пагинацией
// @Produce json
// @Param page query int false "номер страницы (1 и больше)" Example(1)
// @Param limit query int false "максимальное число записей на странице (1 и больше)" Example(1)
// @Param agegt query int false "нижняя граница возраста (включительно)" Example(1)
// @Param agelt query int false "верхняя граница возраста (не включительно)" Example(1)
// @Param age query int false "конкретный возраст (если заданы границы - перезаписывает их)" Example(1)
// @Param idgt query int false "нижняя граница id (включительно)" Example(1)
// @Param idlt query int false "верхняя граница id (не включительно)" Example(1)
// @Param id query int false "конкретный id (если заданы границы - перезаписывает их)" Example(1)
// @Param name query string false "конкретное имя" Example("Dmitriy")
// @Param surname query string false "конкретная фамилия" Example("Smirnov")
// @Param patronymic query string false "конкретное отчество" Example("Petrovich")
// @Param gender query string false "конкретный гендер" Example("male")
// @Param nationality query string false "конкретная национальность" Example("RU")
// @Success 200
// @Success 204
// @Failure 400
// @Failure 500
// @Router /read [get]
func (h *forecaster) ReadPersons(c echo.Context) error {
	pageStr := c.QueryParam("page")
	if pageStr == "" {
		pageStr = "1"
	}

	page, err := strconv.Atoi(pageStr)
	if err != nil {
		c.Response().WriteHeader(http.StatusBadRequest)
		logger.Logger().Debugln(err)
		return err
	}

	limitStr := c.QueryParam("limit")
	if limitStr == "" {
		limitStr = "10"
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		c.Response().WriteHeader(http.StatusBadRequest)
		logger.Logger().Debugln(err)
		return err
	}

	if page < 1 || limit < 1 || limit > 50 {
		c.Response().WriteHeader(http.StatusBadRequest)
		logger.Logger().Debugln(appErrors.ErrIncorrectQueryParam)
		return appErrors.ErrIncorrectQueryParam
	}

	var filters domain.Filters

	idGreaterThanStr := c.QueryParam("idgt")
	filters.IDMoreThan.Set(idGreaterThanStr)

	idLessThanStr := c.QueryParam("idlt")
	filters.IDLessThan.Set(idLessThanStr, (page*limit)+1)

	idEqualToStr := c.QueryParam("id")
	idEqualTo, err := strconv.Atoi(idEqualToStr)
	if err != nil {
		idEqualTo = -1
	} else {
		filters.IDMoreThan.Set(idEqualToStr)
		filters.IDLessThan.Set(idEqualToStr, (page*limit)+1)
		filters.IDLessThan.Value++
		filters.IDEqualTo = idEqualTo
	}

	ageGreaterThanStr := c.QueryParam("agegt")
	filters.AgeMoreThan.Set(ageGreaterThanStr)

	ageLessThanStr := c.QueryParam("agelt")
	filters.AgeLessThan.Set(ageLessThanStr, 200)

	ageEqualToStr := c.QueryParam("age")
	ageEqualTo, err := strconv.Atoi(ageEqualToStr)
	if err != nil {
		ageEqualTo = -1
	} else {
		filters.AgeMoreThan.Set(ageEqualToStr)
		filters.AgeLessThan.Set(ageEqualToStr, 200)
		filters.AgeLessThan.Value++
		filters.AgeEqualTo = ageEqualTo
	}

	nameStr := c.QueryParam("name")
	filters.NameEqualTo.Set(nameStr)

	surnameStr := c.QueryParam("name")
	filters.SurnameEqualTo.Set(surnameStr)

	patronymicStr := c.QueryParam("patronymic")
	filters.PatronymicEqualTo.Set(patronymicStr)

	genderStr := c.QueryParam("gender")
	filters.GenderEqualTo.Set(genderStr)

	nationalityStr := c.QueryParam("nationality")
	filters.NationalityEqualTo.Set(nationalityStr)

	if filters.IDLessThan.Value < filters.IDMoreThan.Value || filters.AgeLessThan.Value < filters.AgeMoreThan.Value {
		c.Response().WriteHeader(http.StatusBadRequest)
		logger.Logger().Infoln(appErrors.ErrIncorrectQueryParam)
		return appErrors.ErrIncorrectQueryParam
	}

	persons, err := h.srv.ReadPersons(c.Request().Context(), page, limit, filters)

	if errors.Is(err, appErrors.ErrNoRowsFound) {
		c.Response().WriteHeader(http.StatusNoContent)
		logger.Logger().Debugln(err)
		return err
	}

	if err != nil {
		c.Response().WriteHeader(http.StatusInternalServerError)
		logger.Logger().Debugln(err)
		return err
	}

	c.Response().Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(c.Response()).Encode(persons)
	if err != nil {
		c.Response().WriteHeader(http.StatusInternalServerError)
		logger.Logger().Debugln(err)
		return err
	}

	logger.Logger().Infoln("successfully sent info about persons")
	c.Response().WriteHeader(http.StatusOK)
	return nil
}
