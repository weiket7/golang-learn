package errors

import (
	"net/http"
	"strings"

	z "github.com/Oudwins/zog"
)

type ErrorResponse struct {
	Message int                 `json:"message"`
	Errors  map[string][]string `json:"errors"`
}

func (errorResponse *ErrorResponse) NewValidationError(errs z.ZogIssueList) *ErrorResponse {
	errorResponse.Message = http.StatusUnprocessableEntity
	errorResponse.Errors = make(map[string][]string)

	// handle errors -> see Errors section
	for _, issue := range errs {
		key := strings.Join(issue.Path, ".")
		errorResponse.Errors[key] = append(errorResponse.Errors[key], strings.Title(key)+" "+issue.Message)
	}
	return errorResponse
}

func NewValidationError(errs z.ZogIssueList) *ErrorResponse {
	errorResponse := &ErrorResponse{
		Message: http.StatusUnprocessableEntity,
		Errors:  make(map[string][]string),
	}

	errorResponse.Message = http.StatusUnprocessableEntity
	errorResponse.Errors = make(map[string][]string)

	// handle errors -> see Errors section
	for _, issue := range errs {
		key := strings.Join(issue.Path, ".")
		errorResponse.Errors[key] = append(errorResponse.Errors[key], strings.Title(key)+" "+issue.Message)
	}
	return errorResponse
}
