package utils

import (
	"auth-golang-cookies/models"
	"errors"
	"regexp"
)

type ValidationResult struct {
	IsValid bool
	Error   error
}

var emailRegex = regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

func NewValidationResult(isValid bool, err error) *ValidationResult {
	return &ValidationResult{
		IsValid: isValid,
		Error:   err,
	}
}

func ValidateEmail(email string) *ValidationResult {
	if email == "" {
		return NewValidationResult(false, errors.New("email is required"))
	}

	if !emailRegex.MatchString(email) {
		return NewValidationResult(false, errors.New("email is not valid"))
	}

	return NewValidationResult(true, nil)
}

func ValidatePassword(password string) *ValidationResult {
	if len(password) < 6 {
		return NewValidationResult(true, errors.New("password must be at least 6 characters"))
	}

	return NewValidationResult(true, nil)
}

func ValidateUserToAuth(userToAuth models.UserToAuth) []string {
	var _error []string

	if v := ValidateEmail(userToAuth.Email); !v.IsValid {
		_error = append(_error, v.Error.Error())
	}

	if v := ValidatePassword(userToAuth.Password); !v.IsValid {
		_error = append(_error, v.Error.Error())
	}
	return _error
}
