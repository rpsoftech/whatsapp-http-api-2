package apis

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/rpsoftech/whatsapp-http-api/env"
	"github.com/rpsoftech/whatsapp-http-api/interfaces"
	"github.com/rpsoftech/whatsapp-http-api/whatsapp"
)

func StartNumber(c *fiber.Ctx) error {
	token, err := interfaces.ExtractNumberFromCtx(c)
	if err != nil {
		return err
	}
	_, ok := whatsapp.ConnectionMap[token]
	if ok {
		return c.JSON(fiber.Map{
			"success": false,
			"reason":  fmt.Sprintf("Number %s is already connected", token),
		})
	}
	jidString := env.ServerConfig.JID[token]
	whatsapp.ConnectToNumber(jidString, token)
	return c.JSON(fiber.Map{
		"success": true,
	})
}
