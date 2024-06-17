package main

// import (
// 	"bytes"
// 	"encoding/json"
// 	"fmt"
// 	"io"
// 	"net/http"
// 	"strings"
// )

// const baseUrl = "http://103.174.103.159:4000/v1"
// const token = "saura"
// const method = "POST"

// var mediaUrl = fmt.Sprintf("%s/send_media", baseUrl)
// var msgUrl = fmt.Sprintf("%s/send_message", baseUrl)

// func SendTextMessage(number string, msg string) bool {

// 	postBody, _ := json.Marshal(map[string]interface{}{
// 		"to": []string{
// 			number,
// 		},
// 		"msg": msg,
// 	})

// 	responseBody := bytes.NewBuffer(postBody)
// 	client := &http.Client{}
// 	req, err := http.NewRequest(method, msgUrl, responseBody)
// 	req.Header.Set("Content-Type", "application/json")
// 	req.Header.Set("X-Api-Token", token)
// 	if err != nil {
// 		fmt.Println(err)
// 		return false
// 	}
// 	res, err := client.Do(req)
// 	if err != nil {
// 		fmt.Println(err)
// 		return false
// 	}
// 	defer res.Body.Close()

// 	body, err := io.ReadAll(res.Body)
// 	if err != nil {
// 		fmt.Println(err)
// 		return false
// 	}
// 	bodyString := string(body)
// 	println(bodyString)
// 	return !strings.Contains(bodyString, "false")
// }
// func main() {
// 	SendTextMessage("919428393489", "DEMO TEST MESSAGE")
// }
