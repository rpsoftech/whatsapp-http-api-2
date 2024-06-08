package env

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/rpsoftech/whatsapp-http-api/validator"
)

type (
	EnvInterface struct {
		APP_ENV                  AppEnv `json:"APP_ENV" validate:"required,enum=AppEnv"`
		PORT                     int    `json:"PORT" validate:"required,port"`
		ALLOW_LOCAL_NO_AUTH      bool   `json:"ALLOW_LOCAL_NO_AUTH" validate:"boolean"`
		AUTO_CONNECT_TO_WHATSAPP bool   `json:"AUTO_CONNECT_TO_WHATSAPP" validate:"boolean"`
		OPEN_BROWSER_FOR_SCAN    bool   `json:"OPEN_BROWSER_FOR_SCAN" validate:"boolean"`
		// DB_URL                string `json:"DB_URL" validate:"required,url"`
		// DB_NAME               string `json:"DB_NAME_KEY" validate:"required,min=3"`
		// REDIS_DB_HOST         string `json:"REDIS_DB_HOST" validate:"required"`
		// REDIS_DB_PORT         int    `json:"REDIS_DB_PORT" validate:"required,port"`
		// REDIS_DB_PASSWORD     string `json:"REDIS_DB_PASSWORD" validate:"required"`
		// REDIS_DB_DATABASE     int    `json:"REDIS_DB_DATABASE" validate:"min=0,max=100"`
		// ACCESS_TOKEN_KEY      string `json:"ACCESS_TOKEN_KEY" validate:"required,min=100"`
		// REFRESH_TOKEN_KEY     string `json:"REFRESH_TOKEN_KEY" validate:"required,min=100"`
		// FIREBASE_JSON_STRING  string `json:"FIREBASE_JSON_STRING" validate:"required"`
		// FIREBASE_DATABASE_URL string `json:"FIREBASE_DATABASE_URL" validate:"required"`
	}
	IServerConfig struct {
		Tokens map[string]string `json:"tokens" validate:"required"`
		JID    map[string]string `json:"JID"`
	}
)

var (
	Env                  *EnvInterface
	ServerConfig         *IServerConfig
	CurrentDirectory     string = ""
	serverConfigFilePath string = ""
)

const ServerConfigFileName = "server.config.json"

func init() {
	godotenv.Load()
	PORT, err := strconv.Atoi(os.Getenv(port_KEY))
	if err != nil {
		panic("Please Pass Valid Port")
	}
	appEnv, _ := parseAppEnv(os.Getenv(app_ENV_KEY))
	allow_local_no_Auth, err := strconv.ParseBool(os.Getenv(allow_local_no_auth_KEY))
	if err != nil {
		log.Fatal(err)
	}

	auto_connect_to_whatsapp, err := strconv.ParseBool(os.Getenv(Auto_Connect_To_Whatsapp_KEY))
	if err != nil {
		log.Fatal(err)
	}
	open_browser_for_scan_KEY, err := strconv.ParseBool(os.Getenv(open_browser_for_scan_KEY))
	if err != nil {
		log.Fatal(err)
	}

	Env = &EnvInterface{
		APP_ENV:                  appEnv,
		PORT:                     PORT,
		ALLOW_LOCAL_NO_AUTH:      allow_local_no_Auth,
		AUTO_CONNECT_TO_WHATSAPP: auto_connect_to_whatsapp,
		OPEN_BROWSER_FOR_SCAN:    open_browser_for_scan_KEY,
	}
	errs := validator.Validator.Validate(Env)
	if len(errs) > 0 {
		println(errs)
		panic(errs[0])
	}
}

func (sc *IServerConfig) Save() {
	if serverConfigFilePath == "" {
		serverConfigFilePath = filepath.Join(CurrentDirectory, ServerConfigFileName)
	}
	byteJson, err := json.MarshalIndent(sc, "", "    ")
	if err != nil {
		return
	}
	f, err := os.OpenFile(serverConfigFilePath, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		fmt.Printf("%v \n", err)
		return
	}
	defer f.Close()
	if _, err = f.Write(byteJson); err != nil {
		fmt.Printf("%v \n", err)
	}
}
