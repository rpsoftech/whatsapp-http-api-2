package whatsapp

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"reflect"

	"github.com/rpsoftech/whatsapp-http-api/env"
	"github.com/rpsoftech/whatsapp-http-api/interfaces"
	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/types/events"
	"google.golang.org/protobuf/proto"
)

type (
	WhatsappConnection struct {
		Client           *whatsmeow.Client
		Number           string
		ConnectionStatus int
		QrCodeString     string
	}

	IWhatsappConnectionMap map[string]*WhatsappConnection
)

var OutPutFilePath = ""

func (connection *WhatsappConnection) ReturnStatusError() error {
	if connection.ConnectionStatus == 0 {
		return &interfaces.RequestError{
			StatusCode: http.StatusNotFound,
			Code:       interfaces.ERROR_CONNECTION_NOT_INITIALIZED,
			Message:    "Connection Not Initialized QR SCANNED",
			Name:       "ERROR_CONNECTION_NOT_INITIALIZED",
			Extra:      []string{connection.QrCodeString},
		}
	} else if connection.ConnectionStatus == -1 {
		return &interfaces.RequestError{
			StatusCode: http.StatusNotFound,
			Code:       interfaces.ERROR_CONNECTION_LOGGED_OUT,
			Message:    "Connection Logged Out",
			Name:       "ERROR_CONNECTION_LOGGED_OUT",
		}
	}
	return nil
}
func (connection *WhatsappConnection) eventHandler(evt interface{}) {
	switch v := evt.(type) {
	case *events.LoggedOut:
		// Send Status
		connection.ConnectionStatus = -1
		connection.Client.Logout()
		println(connection.Number, " Logged Out")
	case *events.Connected:
		// Send Status
		connection.Client.Store.Save()
		env.ServerConfig.JID[connection.Number] = connection.Client.Store.ID.String()
		env.ServerConfig.Save()
		connection.ConnectionStatus = 1
		println(connection.Number, " Logged In")
	default:
		fmt.Printf("Event Occurred%s\n", reflect.TypeOf(v))
	}
}
func (connection *WhatsappConnection) SendTextMessage(to []string, msg string) *map[string]bool {
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
			if err == nil {
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
