package interfaces

import (
	"fmt"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/rpsoftech/whatsapp-http-api/validator"
)

const (
	REQ_LOCAL_KEY_ROLE           = "UserRole"
	REQ_LOCAL_ERROR_KEY          = "Error"
	REQ_LOCAL_NUMBER_KEY         = "Number"
	REQ_LOCAL_KEY_TOKEN_RAW_DATA = "TokenRawData"
)

type (
	RequestError struct {
		StatusCode int    `json:"-"`
		Code       int    `json:"code"`
		Message    string `json:"message"`
		Name       string `json:"name"`
		Extra      any    `json:"extra,omitempty"`
	}
)

func (r *RequestError) Error() string {
	return fmt.Sprintf("status %d: err %v", r.StatusCode, r.Message)
}
func (r *RequestError) AppendValidationErrors(errs []validator.ErrorResponse) *RequestError {
	// return fmt.Sprintf("status %d: err %v", r.StatusCode, r.Message)
	for index, element := range errs {
		if index != 0 {
			r.Message += "\n"
		}
		r.Message += fmt.Sprintf("FieldName:- %s,Passed Value:- %s,Failed Tag:- %s", element.FailedField, element.Value, element.Tag)
	}
	return r
}

func ExtractKeyFromHeader(c *fiber.Ctx, key string) string {
	reqHeaders := c.GetReqHeaders()
	if tokenString, foundToken := reqHeaders[key]; !foundToken || len(tokenString) != 1 || tokenString[0] == "" {
		return ""
	} else {
		return tokenString[0]
	}
}
func ExtractNumberFromCtx(c *fiber.Ctx) (string, error) {
	id, ok := c.Locals(REQ_LOCAL_NUMBER_KEY).(string)
	if !ok {
		return "", &RequestError{
			StatusCode: http.StatusForbidden,
			Code:       INVALID_NUMBER_FROM_TOKEN,
			Message:    "Invalid Number From Token",
			Name:       "INVALID_NUMBER_FROM_TOKEN",
		}
	}
	return id, nil
}
