package domain

type DataFromAPI struct {
	Age          int           `json:"age,omitempty"`
	Gender       string        `json:"gender,omitempty"`
	Nationality  string        `json:"nationality,omitempty"`
	CountrySlice []CountryInfo `json:"country"`
}

type CountryInfo struct {
	CountryID   string  `json:"country_id"`
	Probability float32 `json:"probability"`
}
