package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"reflect"
	"regexp"
	"slices"
	"syscall"
	"time"

	// "github.com/golangWhatsappCustomSoftware/validator"
	socketio_client "github.com/zhouhui8915/go-socket.io-client"

	"github.com/go-playground/validator/v10"
	_ "github.com/mattn/go-sqlite3"
	"github.com/mdp/qrterminal/v3"
	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	"google.golang.org/protobuf/proto"

	// "go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
)

type Config struct {
	Numbers      []string `json:"numbers" validator:"required"`
	UplinkServer string   `json:"uplink_server" validator:"required"`
	Token        string   `json:"token" validator:"required"`
}

type FromRemoteData struct {
	Base64         string   `json:"base64"`
	Mime           string   `json:"mime"`
	ExtName        string   `json:"ext_name"`
	MediaName      string   `json:"media_name"`
	ImgDesc        string   `json:"img_desc"`
	Msg            string   `json:"msg"`
	WebMediaLink   string   `json:"web_media_link"`
	LocalMediaPath string   `json:"local_media_path"`
	To             []string `json:"to" validator:"required"`
}
type FromRemote struct {
	Channel string         `json:"channel" validator:"required"`
	From    string         `json:"from" validator:"required"`
	ReqId   string         `json:"req_id" validator:"required"`
	Data    FromRemoteData `json:"data" validator:"required"`
}

type WhatsappConnection struct {
	client *whatsmeow.Client
	number string
}

// disconnected: 2,
// not_logged_in: -1,
// logged_in: 1,
var NumberConnectionStatus = make(map[string]int)

var currentDir = ""
var OutPutFilePath = ""
var SqlContainer *sqlstore.Container
var ThisConfig = new(Config)
var NonNumber, _ = regexp.Compile(`/\D/g`)
var SocketIoConnection *socketio_client.Client
var ConnectionMap = make(map[string]*WhatsappConnection)

var Validator = validator.New()

var LoopStarted = false

func (connection *WhatsappConnection) eventHandler(evt interface{}) {
	switch v := evt.(type) {
	case *events.LoggedOut:
		// Send Status
		NumberConnectionStatus[connection.number] = -1
		EmitTheStatusToSever(connection.number, -1)
		println(connection.number, " Logged Out")
		connection.client.Store.Delete()
	case *events.Connected:
		// Send Status
		connection.client.Store.Save()
		NumberConnectionStatus[connection.number] = 1
		EmitTheStatusToSever(connection.number, 1)
		println(connection.number, " Logged In")
	default:
		fmt.Printf("Event Occurred%s\n", reflect.TypeOf(v))
	}
}
func (connection *WhatsappConnection) SendTextMessage(to []string, msg string, reqId string) {
	response := make(map[string]bool)
	for _, number := range to {
		IsOnWhatsappCheck, err := connection.client.IsOnWhatsApp([]string{"+" + number})
		if err != nil {
			AppendToOutPutFile(fmt.Sprintf("%s,false,Something Went Wrong %#v\n", number, err))
			// return
			response[number] = false
			continue
		}
		NumberOnWhatsapp := IsOnWhatsappCheck[0]
		if !NumberOnWhatsapp.IsIn {
			AppendToOutPutFile(fmt.Sprintf("%s,false,Number %s Not On Whatsapp\n", number, number))
			response[number] = false
			continue
			// return
		}
		targetJID := NumberOnWhatsapp.JID
		fmt.Printf("sending File To %s\n", number)
		response[number] = false
		if len(msg) > 0 {
			_, err := connection.client.SendMessage(context.Background(), targetJID, &waProto.Message{
				Conversation: proto.String(msg),
			})
			if err != nil {
				response[number] = true
			}
		}
	}

	SocketIoConnection.Emit("response", map[string]interface{}{
		"from":   connection.number,
		"req_id": reqId,
		"data":   response,
	})
}

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

// func AfterSuccessFullConnection() {
// 	if LoopStarted {
// 		println("Tried to Start Loop Again")
// 		return
// 	}
// 	LoopStarted = true
// 	time.Sleep(3 * time.Second)
// 	fmt.Printf("Reading File %s\n", InputFilePath)
// 	inputBytes, err := os.ReadFile(InputFilePath)
// 	check(err)
// 	input := string(inputBytes)
// 	RowsData := strings.Split(input, "\n")
// 	fmt.Printf("total %d Rows Found\n", len(RowsData))
// 	for _, row := range RowsData {
// 		func() {
// 			cols := strings.Split(row, ",")
// 			if len(cols) < 2 {
// 				AppendToOutPutFile("Cells Length < 2 Found\n")
// 				return
// 			}
// 			number := string(NonNumber.ReplaceAll([]byte(cols[0]), []byte("")))
// 			if len(number) < 10 {
// 				AppendToOutPutFile(fmt.Sprintf("%s,false,Length %d of Number is Less than 10\n", number, len(number)))
// 				return
// 			}
// 			defer func() {
// 				if r := recover(); r != nil {
// 					fmt.Println("panic occured: ", r)
// 					AppendToOutPutFile(fmt.Sprintf("%s,false,Something Went Wrong %#v\n", number, r))

// 				}
// 			}()
// 			fileName := fmt.Sprintf("%s.pdf", strings.TrimSpace(cols[1]))
// 			sendFilePath := filepath.Join(ThisConfig.BasePathForAssets, fileName)
// 			if _, err := os.Stat(sendFilePath); errors.Is(err, os.ErrNotExist) {
// 				AppendToOutPutFile(fmt.Sprintf("%s,false,File Path Not Exists %s\n", number, sendFilePath))
// 				return
// 			}
// 			IsOnWhatsappCheck, err := client.IsOnWhatsApp([]string{"+" + number})
// 			if err != nil {
// 				AppendToOutPutFile(fmt.Sprintf("%s,false,Something Went Wrong %#v\n", number, err))
// 				return
// 			}
// 			NumberOnWhatsapp := IsOnWhatsappCheck[0]
// 			if !NumberOnWhatsapp.IsIn {
// 				AppendToOutPutFile(fmt.Sprintf("%s,false,Number %s Not On Whatsapp\n", number, number))
// 				return
// 			}
// 			pdfBytes, err := os.ReadFile(sendFilePath)
// 			if err != nil {
// 				AppendToOutPutFile(fmt.Sprintf("%s,false,Error While Reading File %#v\n", number, err))
// 				return
// 			}
// 			println("Uploading File")
// 			resp, err := client.Upload(context.Background(), pdfBytes, whatsmeow.MediaDocument)
// 			if err != nil {
// 				AppendToOutPutFile(fmt.Sprintf("%s,false,Error While Uploading %#v\n", number, err))
// 				return
// 			}
// 			docProto := &waProto.DocumentMessage{
// 				Url:           &resp.URL,
// 				Mimetype:      proto.String("application/pdf"),
// 				FileName:      &fileName,
// 				DirectPath:    &resp.DirectPath,
// 				MediaKey:      resp.MediaKey,
// 				FileEncSha256: resp.FileEncSHA256,
// 				FileSha256:    resp.FileSHA256,
// 				FileLength:    &resp.FileLength,
// 			}

// 			if ThisConfig.AppendMessageToMedia {
// 				if !ThisConfig.ReadMessageFromCsv {
// 					docProto.Caption = &ThisConfig.Message
// 				} else if ThisConfig.ReadMessageFromCsv && len(cols) >= 3 && len(cols[2]) > 0 {
// 					docProto.Caption = &cols[2]
// 				}
// 			}
// 			// targetJID := types.NewJID("917016879936", types.DefaultUserServer)
// 			targetJID := NumberOnWhatsapp.JID
// 			fmt.Printf("sending File To %s\n", number)
// 			client.SendMessage(context.TODO(), targetJID, &waProto.Message{
// 				DocumentMessage: docProto,
// 			})
// 			if !ThisConfig.AppendMessageToMedia {
// 				message := new(string)
// 				if !ThisConfig.ReadMessageFromCsv {
// 					message = &ThisConfig.Message
// 				} else if ThisConfig.ReadMessageFromCsv && len(cols) >= 3 && len(cols[2]) > 0 {
// 					message = &cols[2]
// 				}
// 				if len(*message) > 0 {
// 					fmt.Printf("sending Message To %s\n", number)
// 					client.SendMessage(context.TODO(), targetJID, &waProto.Message{
// 						Conversation: proto.String(*message),
// 					})
// 				}
// 			}
// 			AppendToOutPutFile(fmt.Sprintf("%s,true\n", number))
// 			time.Sleep(time.Second * time.Duration(ThisConfig.AddMinimumDelayInSecondsAfterSuccessfulMessage))
// 		}()

// 	}

// 	println("It is Completed")
// }

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
	configFilePAth := filepath.Join(currentDir, "service.config.json")
	if _, err := os.Stat(configFilePAth); errors.Is(err, os.ErrNotExist) {
		panic(fmt.Errorf("Config Not Exist on Path %s", configFilePAth))
	}
	dat, err := os.ReadFile(configFilePAth)
	check(err)
	json.Unmarshal(dat, ThisConfig)

	if errs := Validator.Struct(ThisConfig); errs != nil {
		panic(fmt.Errorf("Config Error %#v", errs))
	}
	t := time.Now()
	today := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, t.Nanosecond(), t.Location()).Unix()
	OutPutFilePath = filepath.Join(filepath.Dir(configFilePAth), fmt.Sprintf("%d.log.csv", today))
	ConnectToSocketIo()
	Whatsapp()
}

func Whatsapp() {
	dbLog := waLog.Stdout("Database", "DEBUG", true)
	// Make sure you add appropriate DB connector imports, e.g. github.com/mattn/go-sqlite3 for SQLite
	var err error
	SqlContainer, err = sqlstore.New("sqlite3", "file:WhatsappSuperSecrete.db?_foreign_keys=on", dbLog)
	if err != nil {
		panic(err)
	}
	// If you want multiple sessions, remember their JIDs and use .GetDevice(jid) or .GetAllDevices() instead.
	for _, number := range ThisConfig.Numbers {
		ConnectToNumber(number)
	}

	// Listen to Ctrl+C (you can also do something else that prevents the program from exiting)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	for _, connection := range ConnectionMap {
		connection.client.Disconnect()
		// delete(ConnectionMap, number)
	}
}

func ConnectToNumber(number string) {
	// SqlContainer.PutDevice()
	deviceStore, err := SqlContainer.GetDevice(types.NewJID(number, types.DefaultUserServer))
	if err != nil {
		panic(err)
	}
	if deviceStore == nil {
		deviceStore = SqlContainer.NewDevice()
		// deviceStore = types.DEv(number, types.DefaultUserServer)
	}
	clientLog := waLog.Stdout("Client", "ERROR", true)
	client := whatsmeow.NewClient(deviceStore, clientLog)
	connection := &WhatsappConnection{client: client, number: number}
	ConnectionMap[number] = connection
	client.AddEventHandler(connection.eventHandler)
	if client.Store.ID == nil {
		// No ID stored, new login
		qrChan, _ := client.GetQRChannel(context.Background())
		err = client.Connect()
		if err != nil {
			panic(err)
		}
		for evt := range qrChan {
			if evt.Event == "code" {
				fmt.Printf("QR code for %s\n", number)
				qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, os.Stdout)
			} else {
				fmt.Println("Login event:", evt.Event)
			}
		}
	} else {
		// Already logged in, just connect
		println("Connected")
		err = client.Connect()
		if err != nil {
			panic(err)
		}
	}
}

func ConnectToSocketIo() {
	opts := &socketio_client.Options{
		Transport: "websocket",
	}
	SocketIoConnection, err := socketio_client.NewClient(ThisConfig.UplinkServer, opts)
	if err != nil {
		log.Printf("NewClient error:%v\n", err)
		return
	}

	SocketIoConnection.On("from_remote", func(msg string) {
		data := &FromRemote{}
		err := json.Unmarshal([]byte(msg), data)
		if err != nil {
			log.Printf("on message:%v\n", msg)
			return
		}
		if data.Channel == "send_message" {
			ConnectionMap[data.From].SendTextMessage(data.Data.To, data.Data.Msg, data.ReqId)
		}
	})
	SocketIoConnection.On("reconnect", func() {
		SubscribeToNumbers()
	})
	SocketIoConnection.On("error", func() {
		log.Printf("Server  error\n")
	})
	SocketIoConnection.On("connection", func() {
		log.Printf("Server Connected connect\n")
		SubscribeToNumbers()
	})
	SocketIoConnection.On("message", func(msg string) {
		log.Printf("on message:%v\n", msg)
	})
	SocketIoConnection.On("disconnection", func() {
		log.Printf("Server disconnect\n")
	})
}

func SubscribeToNumbers() {
	for _, number := range ThisConfig.Numbers {
		SocketIoConnection.Emit("subscribe", number)
		status := NumberConnectionStatus[number]
		if status == 0 {
			status = -1
		} else if status == 2 {
			status = 0
		}
		EmitTheStatusToSever(number, status)
	}
}

func EmitTheStatusToSever(number string, status int) {
	SocketIoConnection.Emit("server_status", map[string]interface{}{
		"number": number,
		"status": status,
	})
}
