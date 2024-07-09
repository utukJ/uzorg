package main

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
)

func makeUserDefaultOrg(u *User) Org {
	// create a default org for the user with name user firstName + "Organisation" and description "Default organisation for " + user firstName
	// generate unique orgId
	orgID := uuid.New().String()
	return Org{
		OrgID:       orgID,
		Name:        u.FirstName + "'s Organisation",
		Description: "Default organisation for " + u.FirstName,
	}
}

func writeBadRequestResponse(w http.ResponseWriter, statusCode int, message string) {
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(ErrorResponse{
		ResponseStatus: ResponseStatus{
			Status:  BadRequestStatus,
			Message: message,
		},
		Code: statusCode,
	})
}

func writeValidationErrorResponse(w http.ResponseWriter, errors []*ValidationError) {
	w.WriteHeader(http.StatusUnprocessableEntity)
	json.NewEncoder(w).Encode(ValidationErrorResponse{
		Errors: errors,
	})
}

func writeServerErrorResponse(w http.ResponseWriter, message string) {
	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(ErrorResponse{
		ResponseStatus: ResponseStatus{
			Status:  ServverErrorStatus,
			Message: message,
		},
		Code: http.StatusInternalServerError,
	})
}
