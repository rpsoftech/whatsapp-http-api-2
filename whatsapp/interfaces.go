package whatsapp

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"reflect"
	"strings"

	"github.com/mdp/qrterminal/v3"
	"github.com/rpsoftech/whatsapp-http-api/env"
	"github.com/rpsoftech/whatsapp-http-api/interfaces"
	"github.com/rpsoftech/whatsapp-http-api/utility"
	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/types/events"
	"google.golang.org/protobuf/proto"
)

type (
	WhatsappConnection struct {
		Client           *whatsmeow.Client
		Number           string
		Token            string
		ConnectionStatus int
		QrCodeString     string
		SyncFinished     bool
	}

	IWhatsappConnectionMap map[string]*WhatsappConnection
)

var (
	OutPutFilePath = ""
	ConnectionMap  = make(IWhatsappConnectionMap)
)

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

func (connection *WhatsappConnection) ConnectAndGetQRCode() {
	if connection.Client.Store.ID == nil {
		// No ID stored, new login
		if env.Env.OPEN_BROWSER_FOR_SCAN {
			go func(token string) {
				utility.OpenBrowser(fmt.Sprintf("http://127.0.0.1:%d/scan/%s", env.Env.PORT, token))
			}(connection.Token)
		}
		qrChan, _ := connection.Client.GetQRChannel(context.Background())
		err := connection.Client.Connect()
		if err != nil {
			println(err.Error())
		}
		for evt := range qrChan {
			if evt.Event == "code" {
				fmt.Printf("QR code for %s\n", connection.Token)
				connection.QrCodeString = evt.Code
				// env.ServerConfig.Tokens[connection.Token] = "Something"
				if !env.Env.OPEN_BROWSER_FOR_SCAN {
					qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, os.Stdout)
				}
			} else {
				fmt.Println("Login event:", evt.Event)
			}
		}
	} else {
		// Already logged in, just connect
		println("Connected")
		err := connection.Client.Connect()
		if err != nil {
			println(err.Error())
		}
	}
}
func (connection *WhatsappConnection) eventHandler(evt interface{}) {
	switch v := evt.(type) {
	case *events.LoggedOut:
		// Send Status
		// connection.ConnectionStatus = -1
		connection.Client.Logout()
		connection.Client.Store.Delete()
		println(connection.Number, " Logged Out")
		connection.Client.Disconnect()
		delete(ConnectionMap, connection.Token)
		env.ServerConfig.Tokens[connection.Token] = ""
		delete(env.ServerConfig.JID, connection.Token)
		env.ServerConfig.Save()
		go connection.ConnectAndGetQRCode()
	case *events.Connected:
		// Send Status
		connection.Client.Store.Save()
		connection.Number = connection.Client.Store.ID.User
		go func() {
			env.ServerConfig.Tokens[connection.Token] = connection.Number
			env.ServerConfig.JID[connection.Token] = connection.Client.Store.ID.String()
			env.ServerConfig.Save()
		}()
		connection.ConnectionStatus = 1
		println(connection.Number, " Logged In")
	case *events.OfflineSyncPreview:
		connection.SyncFinished = false
	case *events.OfflineSyncCompleted:
		connection.SyncFinished = true
	default:
		fmt.Printf("Event Occurred%s\n", reflect.TypeOf(v))
	}
}
func (connection *WhatsappConnection) SendTextMessage(to []string, msg string) *map[string]interface{} {
	response := make(map[string]interface{})
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
		fmt.Printf("sending Text To %s\n", number)
		response[number] = false
		if len(msg) > 0 {
			resp, err := connection.Client.SendMessage(context.Background(), targetJID, &waProto.Message{
				Conversation: proto.String(msg),
			})
			if err == nil {
				response[number] = resp
			}
		}
	}
	return &response
}
func (connection *WhatsappConnection) SendMediaFileBase64(to []string, base64Data string, fileName string, msg string) *map[string]bool {
	// pdfBytes, err := os.ReadFile(filePath)
	bytesData, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		AppendToOutPutFile(fmt.Sprintf("false,Error While Reading File %#v\n", err))
		return nil
	}
	return connection.sendMediaFile(to, bytesData, fileName, msg)
}
func (connection *WhatsappConnection) SendMediaFileWithPath(to []string, filePath string, fileName string, msg string) *map[string]bool {
	pdfBytes, err := os.ReadFile(filePath)
	if err != nil {
		AppendToOutPutFile(fmt.Sprintf("false,Error While Reading File %#v\n", err))
		return nil
	}
	return connection.sendMediaFile(to, pdfBytes, fileName, msg)
}
func (connection *WhatsappConnection) sendMediaFile(to []string, fileByte []byte, fileName string, msg string) *map[string]bool {
	response := make(map[string]bool)
	var docProto *waProto.Message
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
		if docProto == nil {
			extensionName := utility.GetMime(fileName)
			if strings.Contains(extensionName, "image") {
				resp, err := connection.Client.Upload(context.Background(), fileByte, whatsmeow.MediaImage)
				if err != nil {
					AppendToOutPutFile(fmt.Sprintf("%s,false,Error While Uploading %#v\n", number, err))
					continue
				}
				docProto = &waProto.Message{
					ImageMessage: &waProto.ImageMessage{
						Caption:  proto.String(msg),
						Url:      &resp.URL,
						Mimetype: proto.String(extensionName),
						// FileName:      &fileName,
						DirectPath:    &resp.DirectPath,
						MediaKey:      resp.MediaKey,
						FileEncSha256: resp.FileEncSHA256,
						FileSha256:    resp.FileSHA256,
						FileLength:    &resp.FileLength,
					},
				}
			} else if strings.Contains(extensionName, "audio") {
				resp, err := connection.Client.Upload(context.Background(), fileByte, whatsmeow.MediaAudio)
				if err != nil {
					AppendToOutPutFile(fmt.Sprintf("%s,false,Error While Uploading %#v\n", number, err))
					continue
				}
				docProto = &waProto.Message{
					AudioMessage: &waProto.AudioMessage{
						// Caption:       proto.String(msg),
						Url:      &resp.URL,
						Mimetype: proto.String(extensionName),
						// FileName:      &fileName,
						DirectPath:    &resp.DirectPath,
						MediaKey:      resp.MediaKey,
						FileEncSha256: resp.FileEncSHA256,
						FileSha256:    resp.FileSHA256,
						FileLength:    &resp.FileLength,
					},
				}
			} else if strings.Contains(extensionName, "video") {
				// var thumbResp *whatsmeow.UploadResponse

				thumbBytes, _ := generateVideoThumbnail(fileByte, fileName)
				// thumbResp, _ := connection.Client.Upload(context.Background(), thumbBytes, whatsmeow.MediaImage)
				// }
				resp, err := connection.Client.Upload(context.Background(), fileByte, whatsmeow.MediaVideo)
				if err != nil {
					AppendToOutPutFile(fmt.Sprintf("%s,false,Error While Uploading %#v\n", number, err))
					continue
				}
				if len(thumbBytes) > 0 {
					docProto = &waProto.Message{
						VideoMessage: &waProto.VideoMessage{
							Caption:       proto.String(msg),
							Url:           &resp.URL,
							Mimetype:      proto.String(extensionName),
							JpegThumbnail: thumbBytes,
							DirectPath:    &resp.DirectPath,
							MediaKey:      resp.MediaKey,
							FileEncSha256: resp.FileEncSHA256,
							FileSha256:    resp.FileSHA256,
							FileLength:    &resp.FileLength,
						},
					}
				} else {
					docProto = &waProto.Message{
						VideoMessage: &waProto.VideoMessage{
							Caption:       proto.String(msg),
							Url:           &resp.URL,
							Mimetype:      proto.String(extensionName),
							DirectPath:    &resp.DirectPath,
							MediaKey:      resp.MediaKey,
							FileEncSha256: resp.FileEncSHA256,
							FileSha256:    resp.FileSHA256,
							FileLength:    &resp.FileLength,
						},
					}
				}
			} else {
				resp, err := connection.Client.Upload(context.Background(), fileByte, whatsmeow.MediaDocument)
				if err != nil {
					AppendToOutPutFile(fmt.Sprintf("%s,false,Error While Uploading %#v\n", number, err))
					continue
				}
				docProto = &waProto.Message{
					DocumentMessage: &waProto.DocumentMessage{
						Caption:       proto.String(msg),
						Url:           &resp.URL,
						Mimetype:      proto.String(extensionName),
						FileName:      &fileName,
						DirectPath:    &resp.DirectPath,
						MediaKey:      resp.MediaKey,
						FileEncSha256: resp.FileEncSHA256,
						FileSha256:    resp.FileSHA256,
						FileLength:    &resp.FileLength,
					},
				}
				println("finished uploading")
				if strings.Contains(extensionName, "pdf") {
					println("PDF to thumb")
					thumb, err := utility.ExtractFirstPage(fileByte)
					if err == nil && len(thumb) > 0 {
						docProto.DocumentMessage.JpegThumbnail = thumb
					} else {
						println(err.Error())
					}
				}
			}
		}
		response[number] = false
		if docProto != nil {
			_, err := connection.Client.SendMessage(context.Background(), targetJID, docProto)
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
