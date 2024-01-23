package domain

type Person struct {
	Name       string `json:"name" example:"Dmitriy"`
	Surname    string `json:"surname" example:"Smirnov"`
	Patronymic string `json:"patronymic,omitempty" example:"Petrovich"`
}

type PersonFromDB struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Surname     string `json:"surname"`
	Patronymic  string `json:"patronymic,omitempty"`
	Age         int    `json:"age"`
	Gender      string `json:"gender"`
	Nationality string `json:"nationality"`
}

type PersonWithAPIData struct {
	Name        string `json:"name,omitempty" example:"Dmitriy"`
	Surname     string `json:"surname,omitempty" example:"Smirnov"`
	Patronymic  string `json:"patronymic,omitempty" example:"Petrovich"`
	Age         int    `json:"age,omitempty" example:"25"`
	Gender      string `json:"gender,omitempty" example:"male"`
	Nationality string `json:"nationality,omitempty" example:"RU"`
	IsDeleted   *bool  `json:"is_deleted,omitempty" example:"false"`
}

func (p *PersonWithAPIData) ReplaceDefaultValuesWithFieldsOfStruct(newValues PersonWithAPIData) {
	if p.Name == "" {
		p.Name = newValues.Name
	}

	if p.Surname == "" {
		p.Surname = newValues.Surname
	}

	if p.Patronymic == "" {
		p.Patronymic = newValues.Patronymic
	}

	if p.Age == 0 {
		p.Age = newValues.Age
	}

	if p.Gender == "" {
		p.Gender = newValues.Gender
	}

	if p.Nationality == "" {
		p.Nationality = newValues.Nationality
	}

	if p.IsDeleted == nil && newValues.IsDeleted != nil {
		p.IsDeleted = newValues.IsDeleted
		return
	}

	if p.IsDeleted == nil {
		isDeleted := true
		p.IsDeleted = &isDeleted
	}
}
