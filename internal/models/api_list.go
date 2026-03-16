package models

// APIListResult models standard paginated list responses used by V13 endpoints.
type APIListResult[T any] struct {
	Data       []T              `json:"data"`
	Pagination PaginationResult `json:"pagination"`
}
