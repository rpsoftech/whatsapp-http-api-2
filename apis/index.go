package apis

import (
	"fmt"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/rpsoftech/whatsapp-http-api/interfaces"
	"github.com/rpsoftech/whatsapp-http-api/utility"
	"github.com/rpsoftech/whatsapp-http-api/whatsapp"
)

type (
	apiSendMessage struct {
		To  []string `json:"to" validate:"required,dive,min=1"`
		Msg string   `json:"msg" validate:"required,min=3"`
	}
)

func AddApis(app fiber.Router) {
	app.Get("/qr_code", GetQrCode)
	app.Post("/send_message", SendMessage)
	// auth.AddAuthPackages(app.Group("/auth"))
	// data.AddDataPackage(app.Group("/data"))
	// order.AddOrderPackage(app.Group("/order"))
}

func SendMessage(c *fiber.Ctx) error {
	body := new(apiSendMessage)
	c.BodyParser(body)
	if err := utility.ValidateReqInput(body); err != nil {
		return err
	}
	number, err := interfaces.ExtractNumberFromCtx(c)
	if err != nil {
		return err
	}
	connection, ok := whatsapp.ConnectionMap[number]
	if !ok || connection == nil {
		return &interfaces.RequestError{
			StatusCode: http.StatusNotFound,
			Code:       interfaces.ERROR_CONNECTION_NOT_FOUND,
			Message:    fmt.Sprintf("Number %s Not Found", number),
			Name:       "ERROR_CONNECTION_NOT_FOUND",
		}
	}
	err = connection.ReturnStatusError()
	if err != nil {
		return err
	}
	return c.JSON(connection.SendTextMessage(body.To, body.Msg))
}

func GetQrCode(c *fiber.Ctx) error {
	number, err := interfaces.ExtractNumberFromCtx(c)
	if err != nil {
		return err
	}
	connection, ok := whatsapp.ConnectionMap[number]
	if !ok || connection == nil {
		return &interfaces.RequestError{
			StatusCode: http.StatusNotFound,
			Code:       interfaces.ERROR_CONNECTION_NOT_FOUND,
			Message:    fmt.Sprintf("Number %s Not Found", number),
			Name:       "ERROR_CONNECTION_NOT_FOUND",
		}
	}
	err = connection.ReturnStatusError()
	if err != nil {
		return c.JSON(fiber.Map{
			"qrCode": connection.QrCodeString,
		})
	}
	return c.JSON(fiber.Map{
		"success": true,
	})
}
