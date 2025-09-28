package models

type Pagination struct {
	Limit  int32 `json:"limit"`
	Offset int32 `json:"offset"`
}
