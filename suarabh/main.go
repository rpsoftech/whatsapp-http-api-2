package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/rpsoftech/whatsapp-http-api/validator"
)

const baseUrl = "http://103.174.103.159:4000/v1"
const token = "saura1"
const method = "POST"

var mediaUrl = fmt.Sprintf("%s/send_media", baseUrl)
var msgUrl = fmt.Sprintf("%s/send_message", baseUrl)

func SendTextMessage(number string, msg string) bool {

	postBody, _ := json.Marshal(map[string]interface{}{
		"to": []string{
			number,
		},
		"msg": msg,
	})
	payload := bytes.NewBuffer(postBody)

	client := &http.Client{}
	req, err := http.NewRequest(method, msgUrl, payload)

	if err != nil {
		fmt.Println(err)
		return false
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-Api-Token", token)

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return false
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return false
	}
	bodyString := string(body)
	println(bodyString)
	return !strings.Contains(bodyString, "false")
}
func SendMediaMessage(path string, msg string, number string) bool {
	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	file, _ := os.Open(path)
	defer file.Close()

	part1, _ := writer.CreateFormFile("file", filepath.Base(path))
	_, errFile1 := io.Copy(part1, file)
	if errFile1 != nil {
		fmt.Println(errFile1)
		return false
	}
	_ = writer.WriteField("to", fmt.Sprintf("[\"%s\"]", number))
	if msg != "" {
		_ = writer.WriteField("msg", msg)
	}
	err := writer.Close()
	if err != nil {
		fmt.Println(err)
		return false
	}

	client := &http.Client{}
	req, err := http.NewRequest(method, mediaUrl, payload)

	if err != nil {
		fmt.Println(err)
		return false
	}
	req.Header.Add("X-Api-Token", token)

	req.Header.Set("Content-Type", writer.FormDataContentType())
	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return false
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return false
	}
	bodyString := string(body)
	println(bodyString)
	return !strings.Contains(bodyString, "false")
}

type Config struct {
	UseTextMessage                                 bool   `json:"useTextMessage" validate:"boolean"`
	AppendMessageToMedia                           bool   `json:"appendMessageToMedia" validate:"boolean"`
	ReadMessageFromCsv                             bool   `json:"readMessageFromCsv" validate:"boolean"`
	Message                                        string `json:"message"`
	AddMinimumDelayInSecondsAfterSuccessfulMessage int    `json:"addMinimumDelayInSecondsAfterSuccessfulMessage" validate:"required"`
	BasePathForAssets                              string `json:"basePathForAssets"`
	InputFileName                                  string `json:"inputFileName"`
}

var currentDir = ""
var InputFilePath = ""
var OutPutFilePath = ""
var ThisConfig = new(Config)
var NonNumber, _ = regexp.Compile(`/\D/g`)

func AppendToOutPutFile(text string) {
	f, err := os.OpenFile(OutPutFilePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		panic(err)
	}

	defer f.Close()

	if _, err = f.WriteString(text); err != nil {
		panic(err)
	}
}

func AfterSuccessFullConnection() {
	time.Sleep(3 * time.Second)
	fmt.Printf("Reading File %s\n", InputFilePath)
	inputBytes, err := os.ReadFile(InputFilePath)
	check(err)
	input := string(inputBytes)
	RowsData := strings.Split(input, "\n")
	fmt.Printf("total %d Rows Found\n", len(RowsData))
	for _, row := range RowsData {
		func() {
			cols := strings.Split(row, ",")
			if len(cols) < 2 {
				AppendToOutPutFile("Cells Length < 2 Found\n")
				return
			}
			number := string(NonNumber.ReplaceAll([]byte(cols[0]), []byte("")))
			if len(number) < 10 {
				AppendToOutPutFile(fmt.Sprintf("%s,false,Length %d of Number is Less than 10\n", number, len(number)))
				return
			}
			defer func() {
				if r := recover(); r != nil {
					fmt.Println("panic occured: ", r)
					AppendToOutPutFile(fmt.Sprintf("%s,false,Something Went Wrong %#v\n", number, r))

				}
			}()
			sendFilePath := ""
			fileName := strings.TrimSpace(cols[1])
			FileNamesArray := []string{
				filepath.Join(ThisConfig.BasePathForAssets, fmt.Sprintf("%s.pdf", fileName)),
				filepath.Join(ThisConfig.BasePathForAssets, fmt.Sprintf("%s.png", fileName)),
				filepath.Join(ThisConfig.BasePathForAssets, fmt.Sprintf("%s.jpg", fileName)),
				filepath.Join(ThisConfig.BasePathForAssets, fmt.Sprintf("%s.jpeg", fileName)),
				filepath.Join(ThisConfig.BasePathForAssets, fmt.Sprintf("%s.webpp", fileName)),
				filepath.Join(ThisConfig.BasePathForAssets, fmt.Sprintf("%s.avi", fileName)),
				filepath.Join(ThisConfig.BasePathForAssets, fmt.Sprintf("%s.mkv", fileName)),
				filepath.Join(ThisConfig.BasePathForAssets, fmt.Sprintf("%s.dat", fileName)),
				filepath.Join(ThisConfig.BasePathForAssets, fmt.Sprintf("%s.mp4", fileName)),
				filepath.Join(ThisConfig.BasePathForAssets, fmt.Sprintf("%s.mp3", fileName)),
			}

			for _, fileName := range FileNamesArray {
				if _, err := os.Stat(fileName); !errors.Is(err, os.ErrNotExist) {
					sendFilePath = fileName
					break
				}
			}
			if sendFilePath == "" {
				AppendToOutPutFile(fmt.Sprintf("%s,false,File Path Not Exists for file %s\n", number, filepath.Join(ThisConfig.BasePathForAssets, fileName)))
				return
			}
			if err != nil {
				AppendToOutPutFile(fmt.Sprintf("%s,false,Error While Reading File %#v\n", number, err))
				return
			}
			println("Uploading File")
			message := ""
			if !ThisConfig.ReadMessageFromCsv {
				message = ThisConfig.Message
			} else if ThisConfig.ReadMessageFromCsv && len(cols) >= 3 && len(cols[2]) > 0 {
				for index, col := range cols {
					if index > 1 {
						if len(col) > 0 {
							message = strings.ReplaceAll(message, fmt.Sprintf("{{var%d}}", index-1), col)
						}
					}
				}
			}
			if message != "" && !ThisConfig.AppendMessageToMedia {
				if resp := SendTextMessage(number, message); !resp {
					AppendToOutPutFile(fmt.Sprintf("%s,false,Number Not Whatsapp\n", number))
				}
				message = ""
			}
			if resp := SendMediaMessage(sendFilePath, message, number); !resp {
				AppendToOutPutFile(fmt.Sprintf("%s,false,Number Not Whatsapp\n", number))
			}
			AppendToOutPutFile(fmt.Sprintf("%s,true\n", number))
			time.Sleep(time.Second * time.Duration(ThisConfig.AddMinimumDelayInSecondsAfterSuccessfulMessage))
		}()
	}
	println("It is Completed")
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	// using the function
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
	configFilePAth := filepath.Join(currentDir, "configs.json")
	if _, err := os.Stat(configFilePAth); errors.Is(err, os.ErrNotExist) {
		panic(fmt.Errorf("Config Not Exist on Path %s", configFilePAth))
	}
	dat, err := os.ReadFile(configFilePAth)
	check(err)
	json.Unmarshal(dat, ThisConfig)

	if errs := validator.Validator.Validate(ThisConfig); len(errs) > 0 {
		panic(fmt.Errorf("Config Error %#v", errs))
	}
	if ThisConfig.UseTextMessage {
		if !ThisConfig.ReadMessageFromCsv && len(ThisConfig.Message) == 0 {
			panic("Please Pass Message in Config File If you want to send Text Message Or Make useTextMessage to false")
		}
	}
	if ThisConfig.BasePathForAssets == "" {
		ThisConfig.BasePathForAssets = filepath.Join(currentDir, "assets")
	}

	if _, err := os.Stat(ThisConfig.BasePathForAssets); errors.Is(err, os.ErrNotExist) {
		panic(fmt.Errorf("base path for assets not exists %s", configFilePAth))
	}
	if len(ThisConfig.InputFileName) > 0 {
		InputFilePath = filepath.Join(currentDir, ThisConfig.InputFileName)
	} else {
		InputFilePath = filepath.Join(currentDir, "input.csv")
	}
	OutPutFilePath = filepath.Join(filepath.Dir(InputFilePath), "output.csv")
	if _, err := os.Stat(InputFilePath); errors.Is(err, os.ErrNotExist) {
		panic(fmt.Errorf("input File Not Exists at %s", InputFilePath))
	}
	AfterSuccessFullConnection()
}
