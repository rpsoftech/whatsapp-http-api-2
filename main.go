package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/rpsoftech/whatsapp-http-api/whatsapp"
)

var CurrentDirectory string = ""
var ServerConfig *whatsapp.IServerConfig
var Validator = validator.New()

func main() {
	CurrentDirectory = FindAndReturnCurrentDir()
	ServerConfig = ReadConfigFileAndReturnIt(CurrentDirectory)
	whatsapp.OutPutFilePath = ReturnOutPutFilePath(CurrentDirectory)
	whatsapp.InitSqlContainer()

	for _, number := range ServerConfig.Numbers {
		whatsapp.ConnectToNumber(number)
	}

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

func ReadConfigFileAndReturnIt(currentDir string) *whatsapp.IServerConfig {
	config := new(whatsapp.IServerConfig)
	configFilePAth := filepath.Join(currentDir, "server.config.json")
	if _, err := os.Stat(configFilePAth); errors.Is(err, os.ErrNotExist) {
		panic(fmt.Errorf("CONFIG_NOT_EXIST_ON_PATH %s", configFilePAth))
	}
	dat, err := os.ReadFile(configFilePAth)
	check(err)
	err = json.Unmarshal(dat, config)
	check(err)
	if errs := Validator.Struct(config); errs != nil {
		panic(fmt.Errorf("CONFIG_ERROR %#v", errs))
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
