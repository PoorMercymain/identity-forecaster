package domain

import (
	"database/sql"
	"strconv"
)

type Filters struct {
	IDMoreThan         IntMoreFilter
	IDLessThan         IntLessFilter
	IDEqualTo          int
	AgeMoreThan        IntMoreFilter
	AgeLessThan        IntLessFilter
	AgeEqualTo         int
	NameEqualTo        StrFilter
	SurnameEqualTo     StrFilter
	PatronymicEqualTo  StrFilter
	GenderEqualTo      StrFilter
	NationalityEqualTo StrFilter
}

type IntMoreFilter struct {
	Value int
}

func (f *IntMoreFilter) Set(strValue string) {
	val, err := strconv.Atoi(strValue)
	if err != nil {
		f.Value = 0
	} else {
		f.Value = val
	}
}

type IntLessFilter struct {
	Value int
}

func (f *IntLessFilter) Set(strValue string, defaultValue int) {
	val, err := strconv.Atoi(strValue)
	if err != nil {
		f.Value = defaultValue
	} else {
		f.Value = val
	}
}

type StrFilter struct {
	Value sql.NullString
}

func (f *StrFilter) Set(value string) {
	if value == "" {
		f.Value = sql.NullString{String: "", Valid: false}
		return
	}

	f.Value = sql.NullString{String: value, Valid: true}
}
