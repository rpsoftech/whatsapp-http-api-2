package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/rpsoftech/whatsapp-http-api/env"
	"github.com/rpsoftech/whatsapp-http-api/interfaces"
)

// fiber middleware for jwt
func TokenDecrypter(c *fiber.Ctx) error {
	reqHeaders := c.GetReqHeaders()
	println(c.IP())
	tokenString, foundToken := reqHeaders[env.RequestTokenHeaderKey]
	if !foundToken || len(tokenString) != 1 || tokenString[0] == "" {
		if c.IP() == "127.0.0.1" || env.Env.ALLOW_LOCAL_NO_AUTH {
			tokenString = []string{"default"}
		} else {

			c.Locals(interfaces.REQ_LOCAL_ERROR_KEY, &interfaces.RequestError{
				StatusCode: 403,
				Code:       interfaces.ERROR_TOKEN_NOT_PASSED,
				Message:    "Please Pass Valid Token",
				Name:       "ERROR_TOKEN_NOT_PASSED",
			})
			return c.Next()
		}
	}
	if value, ok := env.ServerConfig.Tokens[tokenString[0]]; !ok || value == "" {
		c.Locals(interfaces.REQ_LOCAL_ERROR_KEY, &interfaces.RequestError{
			StatusCode: 401,
			Code:       interfaces.ERROR_INVALID_TOKEN,
			Message:    "Invalid Token",
			Name:       "ERROR_INVALID_TOKEN",
		})
	} else {
		c.Locals(interfaces.REQ_LOCAL_NUMBER_KEY, value)
	}
	return c.Next()
}

func AllowOnlyValidTokenMiddleWare(c *fiber.Ctx) error {
	jwtRawFromLocal := c.Locals(interfaces.REQ_LOCAL_NUMBER_KEY)
	localError := c.Locals(interfaces.REQ_LOCAL_ERROR_KEY)

	if jwtRawFromLocal == nil {
		if localError != nil {
			err, ok := localError.(*interfaces.RequestError)
			if !ok {
				return &interfaces.RequestError{
					StatusCode: 403,
					Code:       interfaces.ERROR_INTERNAL_SERVER,
					Message:    "Cannot Cast Error Token",
					Name:       "NOT_VALID_DECRYPTED_TOKEN",
				}
			}
			return err
		}
		return &interfaces.RequestError{
			StatusCode: 403,
			Code:       interfaces.ERROR_TOKEN_NOT_PASSED,
			Message:    "Invalid Token or token expired",
			Name:       "ERROR_TOKEN_NOT_PASSED",
		}
	}
	return c.Next()
}
