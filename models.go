package main

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

type User struct {
	UserID    string `json:"userId"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
	Password  string `json:"-"`
	Phone     string `json:"phone"`
}

type RegisterUserRequest struct {
	FirstName string `json:"firstName" validate:"required"`
	LastName  string `json:"lastName"  validate:"required"`
	Email     string `json:"email"     validate:"required,email"`
	Password  string `json:"password"  validate:"required,min=8"`
	Phone     string `json:"phone"     validate:"required,e164"`
}

// Validate is a method of RegisterUserRequest that validates its fields.
func (r *RegisterUserRequest) Validate() []*ValidationError {
	validate := validator.New()
	err := validate.Struct(r)
	if err != nil {
		if _, ok := err.(*validator.InvalidValidationError); ok {
			return nil // or handle the error
		}

		var errors []*ValidationError
		for _, err := range err.(validator.ValidationErrors) {
			var element ValidationError
			element.Field = err.StructNamespace()
			// Construct a descriptive message for the validation error
			element.Message = fmt.Sprintf(
				"Field '%s' failed validation for '%s' condition",
				err.StructNamespace(),
				err.Tag(),
			)
			errors = append(errors, &element)
		}
		return errors
	}
	return nil
}

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type ValidationErrorResponse struct {
	Errors []*ValidationError `json:"errors"`
}

type UserData struct {
	Token string `json:"accessToken"`
	User  *User  `json:"user"`
}

type ResponseStatus struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type ErrorResponse struct {
	ResponseStatus
	Code int `json:"statusCode"`
}

const SuccessStatus = "success"
const BadRequestStatus = "Bad Request"
const ServverErrorStatus = "Server Error"

type RegisterUserResponse struct {
	ResponseStatus
	Data *UserData `json:"data"`
}

type Org struct {
	OrgID       string `json:"orgId"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type LoginRequest struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type LoginResponse struct {
	ResponseStatus
	Data *UserData `json:"data"`
}

type GetUserResponse struct {
	ResponseStatus
	Data *User `json:"data"`
}

type Organizations struct {
	Orgs []*Org `json:"organizations"`
}

type GetOrgsResponse struct {
	ResponseStatus
	Data *Organizations `json:"data"`
}

type CreateOrgRequest struct {
	Name        string `json:"name"        validate:"required"`
	Description string `json:"description" validate:"required"`
}

// validate is a method of CreateOrgRequest that validates its fields.
func (r *CreateOrgRequest) Validate() []*ValidationError {
	validate := validator.New()
	err := validate.Struct(r)
	if err != nil {
		if _, ok := err.(*validator.InvalidValidationError); ok {
			return nil // or handle the error
		}

		var errors []*ValidationError
		for _, err := range err.(validator.ValidationErrors) {
			var element ValidationError
			element.Field = err.StructNamespace()
			// Construct a descriptive message for the validation error
			element.Message = fmt.Sprintf(
				"Field '%s' failed validation for '%s' condition",
				err.StructNamespace(),
				err.Tag(),
			)
			errors = append(errors, &element)
		}
		return errors
	}
	return nil
}

type CreateOrgResponse struct {
	ResponseStatus
	Data *Org `json:"data"`
}

type GetOrgResponse struct {
	ResponseStatus
	Data *Org `json:"data"`
}

type GetOrgUsersResponse struct {
	ResponseStatus
	Data []*User `json:"data"`
}

type AddUserToOrgRequest struct {
	UserID string `json:"userId" validate:"required"`
}

// validate is a method of AddUserToOrgRequest that validates its fields.
func (r *AddUserToOrgRequest) Validate() []*ValidationError {
	validate := validator.New()
	err := validate.Struct(r)
	if err != nil {
		if _, ok := err.(*validator.InvalidValidationError); ok {
			return nil // or handle the error
		}

		var errors []*ValidationError
		for _, err := range err.(validator.ValidationErrors) {
			var element ValidationError
			element.Field = err.StructNamespace()
			// Construct a descriptive message for the validation error
			element.Message = fmt.Sprintf(
				"Field '%s' failed validation for '%s' condition",
				err.StructNamespace(),
				err.Tag(),
			)
			errors = append(errors, &element)
		}
		return errors
	}
	return nil
}

type AddUserToOrgResponse struct {
	ResponseStatus
}
