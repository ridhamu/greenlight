package data

import "github.com/ridhamu/greenlight/internal/validator"

type Filters struct {
	Page         int
	PageSize     int
	Sort         string
	SortSafeList []string
}

func ValidateFilters(v *validator.Validator, filters Filters) {
	v.Check(filters.Page > 0, "page", "must be greater than zero")
	v.Check(filters.Page <= 10_000_000, "page", "must be less than 10 million")
	v.Check(filters.PageSize > 0, "page_size", "must be greater than zero")
	v.Check(filters.PageSize <= 100, "page_size", "must be a maximum of 100")

	v.Check(validator.PermittedValues(filters.Sort, filters.SortSafeList...), "sort", "invalid sort value")
}
