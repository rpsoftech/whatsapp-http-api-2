package whatsapp

import (
	"context"
	"fmt"
	"os"
	"reflect"

	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/types/events"
	"google.golang.org/protobuf/proto"
)

type (
	IServerConfig struct {
		Tokens  map[string]string `json:"tokens" validator:"required"`
		Numbers []string          `json:"numbers" validator:"required"`
	}

	WhatsappConnection struct {
		Client           *whatsmeow.Client
		Number           string
		ConnectionStatus int
	}

	IWhatsappConnectionMap map[string]*WhatsappConnection
)

var OutPutFilePath = ""

func (connection *WhatsappConnection) eventHandler(evt interface{}) {
	switch v := evt.(type) {
	case *events.LoggedOut:
		// Send Status
		connection.ConnectionStatus = -1
		// EmitTheStatusToSever(connection.Number, -1)
		println(connection.Number, " Logged Out")
		connection.Client.Store.Delete()
	case *events.Connected:
		// Send Status
		connection.Client.Store.Save()
		connection.ConnectionStatus = 1
		println(connection.Number, " Logged In")
	default:
		fmt.Printf("Event Occurred%s\n", reflect.TypeOf(v))
	}
}
func (connection *WhatsappConnection) SendTextMessage(to []string, msg string, reqId string) *map[string]bool {
	response := make(map[string]bool)
	for _, number := range to {
		IsOnWhatsappCheck, err := connection.Client.IsOnWhatsApp([]string{"+" + number})
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
			_, err := connection.Client.SendMessage(context.Background(), targetJID, &waProto.Message{
				Conversation: proto.String(msg),
			})
			if err != nil {
				response[number] = true
			}
		}
	}
	return &response
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
