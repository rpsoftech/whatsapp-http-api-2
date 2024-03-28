package utility

import (
	"net/http"

	"github.com/rpsoftech/whatsapp-http-api/interfaces"
	"github.com/rpsoftech/whatsapp-http-api/validator"
)

func ValidateReqInput(body interface{}) (err *interfaces.RequestError) {
	err = &interfaces.RequestError{
		StatusCode: http.StatusBadRequest,
		Code:       interfaces.ERROR_INVALID_INPUT,
		Message:    "",
		Name:       "INVALID_INPUT",
		Extra:      nil,
	}
	return ValidateStructAndReturnReqError(body, err)
}

func ValidateStructAndReturnReqError(data interface{}, err *interfaces.RequestError) *interfaces.RequestError {
	if errs := validator.Validator.Validate(data); len(errs) > 0 {
		err.Extra = errs
		err.AppendValidationErrors(errs)
		return err
	} else {
		return nil
	}
}
