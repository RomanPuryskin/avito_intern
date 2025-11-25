package utils

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

func ValidateStruct(str interface{}) error {
	validator := validator.New()
	err := validator.Struct(str)
	if err != nil {
		return fmt.Errorf("[ValidateStruct]: %w", err)
	}
	return nil
}
