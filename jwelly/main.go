package main

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/rpsoftech/whatsapp-http-api/validator"
)

type JwellyWhatsappConfig struct {
	ServerUrl            string `json:"serverUrl"`
	DoNotWaitForResponse bool   `json:"doNotWaitForResponse"`
}

var config *JwellyWhatsappConfig

func CreateLogsFile() {
	// os.OpenFile("./output.whatsapp.logs",os.)
	f, err := os.OpenFile("./rps_whatsapp.logs.txt", os.O_TRUNC, 0600)
	if err != nil {
		panic(err)
	}
	f.WriteString(time.Now().String() + "\n")
}

func main() {
	// args := os.Args
	// if !slices.Contains(args, "--dev") && !slices.Contains(args, "--prod") {
	// 	cmd := exec.Command(filepath.Join(FindAndReturnCurrentDir(), os.Args[0]), "--prod")
	// 	cmd.Stdout = os.Stdout
	// 	err := cmd.Start()
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}
	// 	log.Printf("Just ran subprocess %d, exiting\n", cmd.Process.Pid)
	// 	// time.Sleep(5 * time.Second)
	// 	return
	// }
	if _, err := os.Stat("./whatsapp.config"); err == nil {
		// path/to/whatever exists
		if res, err := os.ReadFile("./whatsapp.config"); err == nil {
			config = ReadConfigFileAndReturnIt(FindAndReturnCurrentDir())
			AfterWhatsappConfigFile(string(res))
		}
		// time.Sleep(8 * time.Second)
	} else if errors.Is(err, os.ErrNotExist) {
		// path/to/whatever does *not* exist
		AppendToOutPutFile("File Does Not Exist")
	} else {
		AppendToOutPutFile("Something Went Wrong")
		// Schrodinger: file may or may not exist. See err for details.
		// Therefore, do *NOT* use !os.IsNotExist(err) to test for file existence
	}

	AppendToOutPutFile(fmt.Sprintln(len(os.Args), os.Args))
}

func ReadConfigFileAndReturnIt(currentDir string) *JwellyWhatsappConfig {
	configFilePAth := filepath.Join(currentDir, "jwelly.whatsapp.config")
	dat, err := os.ReadFile(configFilePAth)
	if err != nil {
		panic(fmt.Errorf("failed to read config file: %w", err))
	}
	config := &JwellyWhatsappConfig{}
	dataArray := strings.Split(string(dat), "\n")
	if len(dataArray) < 2 {
		panic(fmt.Errorf("expected dataArray to have length 2, got %d", len(dataArray)))
	}
	config.ServerUrl = strings.TrimSpace(dataArray[0])
	config.DoNotWaitForResponse = strings.TrimSpace(dataArray[1]) == "true"

	if errs := validator.Validator.Validate(config); len(errs) > 0 {
		panic(fmt.Errorf("CONFIG_ERROR %#v", errs))
	}
	return config
}

func AfterWhatsappConfigFile(data string) {
	// Check that config is not null
	if config == nil {
		go AppendToOutPutFile("Config is null")
		return
	}

	dataToBeSend := strings.SplitN(data, "\n", 5)

	// Check that dataToBeSend is the expected length
	if len(dataToBeSend) < 5 {
		go AppendToOutPutFile(fmt.Sprintf("Expected dataToBeSend to have length 5, got %d", len(dataToBeSend)))
		return
	}

	number, key, filePathToBeSend, messageToSend := dataToBeSend[1], dataToBeSend[0], dataToBeSend[3], dataToBeSend[2]
	number = strings.TrimSpace(number)
	key = strings.TrimSpace(key)
	messageToSend = strings.TrimSpace(messageToSend)
	filePathToBeSend = strings.TrimSpace(filePathToBeSend)
	// Check that filePathToBeSend is not empty
	if filePathToBeSend == "" {
		go AppendToOutPutFile("filePathToBeSend is empty")
		return
	}

	// Check that key is not empty
	if key == "" {
		go AppendToOutPutFile("key is empty")
		return
	}

	reqUrl := config.ServerUrl + "/send_media_64"
	if _, err := os.Stat(filePathToBeSend); err != nil {
		for i := 0; i < 15; i++ {
			_, err := os.Stat(filePathToBeSend)
			if err == nil {
				break
			}
			if errors.Is(err, os.ErrNotExist) {
				go AppendToOutPutFile("Waiting For File")
				time.Sleep(1 * time.Second)
			}
		}
		go AppendToOutPutFile(err.Error())
		return
	}
	go AppendToOutPutFile("File Read Successfully")
	fileBytes, err := os.ReadFile(filePathToBeSend)
	if err != nil {
		go AppendToOutPutFile(err.Error())
		return
	}

	if len(fileBytes) <= 10 {
		for i := 0; i < 3; i++ {
			fileBytes, _ = os.ReadFile(filePathToBeSend)
			if len(fileBytes) <= 10 {
				go AppendToOutPutFile("Waiting For File")
				time.Sleep(5 * time.Second)
			} else {
				break
			}
		}
		if len(fileBytes) <= 10 {
			go AppendToOutPutFile("File Length <= 10")
			return
		}
	}
	base64File := base64.StdEncoding.EncodeToString(fileBytes)

	postBody := fmt.Sprintf(`{"msg":"%s","fileName":"%s","to":["%s"],"base64":"%s"}`, messageToSend,
		filepath.Base(filePathToBeSend), number, base64File)

	// Send the POST request
	go AppendToOutPutFile("Sending Request To " + reqUrl)
	req, err := http.NewRequest("POST", reqUrl, strings.NewReader(postBody))
	if err != nil {
		go AppendToOutPutFile(err.Error())
		return
	}

	if config.DoNotWaitForResponse {
		req.Header.Add("headless", "true")
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-Api-Token", key)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		go AppendToOutPutFile(err.Error())
		return
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		go AppendToOutPutFile(err.Error())
		return
	}
	go AppendToOutPutFile("Response " + string(body))
}

func AppendToOutPutFile(text string) {
	text = text + "\n"
	println(text)

	f, err := os.OpenFile("./rps_whatsapp.logs.txt", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		panic(err)
	}

	defer f.Close()

	if _, err = f.WriteString(text); err != nil {
		panic(err)
	}
}

func FindAndReturnCurrentDir() string {
	currentDir := ""
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

func check(e error) {
	if e != nil {
		panic(e)
	}
}
