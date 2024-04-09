package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/rpsoftech/whatsapp-http-api/apis"
	"github.com/rpsoftech/whatsapp-http-api/env"
	"github.com/rpsoftech/whatsapp-http-api/interfaces"
	"github.com/rpsoftech/whatsapp-http-api/middleware"
	"github.com/rpsoftech/whatsapp-http-api/validator"
	"github.com/rpsoftech/whatsapp-http-api/whatsapp"
)

var version string

func main() {
	println(version)
	// println(time.Now().Unix())
	if time.Now().Unix() > 1713262858 {
		println("Please Update The Binary From Keyur Shah")
		println("Press Any Key To Close")
		input := bufio.NewScanner(os.Stdin)
		input.Scan()
		return
	}
	env.CurrentDirectory = FindAndReturnCurrentDir()
	go func() {
		os.RemoveAll("./tmp")
		os.Mkdir("./tmp", os.ModeTemporary)
	}()
	env.ServerConfig = ReadConfigFileAndReturnIt(env.CurrentDirectory)
	whatsapp.OutPutFilePath = ReturnOutPutFilePath(env.CurrentDirectory)
	whatsapp.InitSqlContainer()
	go func() {
		for _, number := range env.ServerConfig.Numbers {
			jidString := env.ServerConfig.JID[number]
			whatsapp.ConnectToNumber(number, jidString)
		}
	}()
	InitFiberServer()

}

func InitFiberServer() {
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			mappedError, ok := err.(*interfaces.RequestError)
			if !ok {
				println(err.Error())
				return c.Status(500).JSON(interfaces.RequestError{
					Code:    interfaces.ERROR_INTERNAL_SERVER,
					Message: "Some Internal Error",
					Name:    "Global Error Handler Function",
				})
			}
			return c.Status(mappedError.StatusCode).JSON(mappedError)
		},
	})
	app.Use(logger.New())
	apis.AddApis(app.Group("/v1", middleware.TokenDecrypter, middleware.AllowOnlyValidTokenMiddleWare))
	app.Use(func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusNotFound).SendString("Sorry can't find that!")
	})
	hostAndPort := ""
	if env.Env.APP_ENV == env.APP_ENV_LOCAL || env.Env.APP_ENV == env.APP_ENV_DEVELOPE {
		hostAndPort = "127.0.0.1"
	}
	hostAndPort = hostAndPort + ":" + strconv.Itoa(env.Env.PORT)
	app.Listen(hostAndPort)
}

func FindAndReturnCurrentDir() string {
	currentDir := ""
	fmt.Println(len(os.Args), os.Args)
	if slices.Contains(os.Args, "--dev") {
		current, err := os.Getwd()
		check(err)
		currentDir = current
	} else {
		exePath, err := os.Executable()
		currentDir = filepath.Dir(exePath)
		check(err)
	}
	return currentDir
}

func ReadConfigFileAndReturnIt(currentDir string) *env.IServerConfig {
	config := new(env.IServerConfig)
	configFilePAth := filepath.Join(currentDir, env.ServerConfigFileName)
	if _, err := os.Stat(configFilePAth); errors.Is(err, os.ErrNotExist) {
		panic(fmt.Errorf("CONFIG_NOT_EXIST_ON_PATH %s", configFilePAth))
	}
	dat, err := os.ReadFile(configFilePAth)
	check(err)
	err = json.Unmarshal(dat, config)
	check(err)
	if errs := validator.Validator.Validate(config); len(errs) > 0 {
		panic(fmt.Errorf("CONFIG_ERROR %#v", errs))
	}
	if config.JID == nil {
		config.JID = make(map[string]string)
	}
	return config
}

func ReturnOutPutFilePath(currentDir string) string {
	t := time.Now()
	today := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, t.Nanosecond(), t.Location()).Unix()
	return filepath.Join(currentDir, "whatsapp_server_logs", fmt.Sprintf("%d.log.csv", today))
}
func check(e error) {
	if e != nil {
		panic(e)
	}
}
