package apis

import (
	"github.com/gofiber/fiber/v2"
	"github.com/rpsoftech/whatsapp-http-api/env"
)

func AppendTokenInConfigJSON(c *fiber.Ctx) error {
	token := c.Query("token")
	if token == "" {
		return c.JSON(fiber.Map{
			"success": false,
		})
	}
	// check token exist in config
	if _, ok := env.ServerConfig.Tokens[token]; ok {
		return c.JSON(fiber.Map{
			"success": false,
		})
	}
	env.ServerConfig.Tokens[token] = ""
	env.ServerConfig.Save()
	return c.JSON(fiber.Map{
		"success": true,
	})
}
