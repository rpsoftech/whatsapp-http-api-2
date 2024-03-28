package whatsapp

import (
	"context"
	"fmt"
	"os"

	_ "github.com/mattn/go-sqlite3"

	"github.com/mdp/qrterminal/v3"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	waLog "go.mau.fi/whatsmeow/util/log"
)

var (
	SqlContainer *sqlstore.Container

	ConnectionMap = make(IWhatsappConnectionMap)
)

func InitSqlContainer() {

	dbLog := waLog.Stdout("Database", "WARN", true)
	// Make sure you add appropriate DB connector imports, e.g. github.com/mattn/go-sqlite3 for SQLite
	var err error
	SqlContainer, err = sqlstore.New("sqlite3", "file:WhatsappSuperSecrete.db?_foreign_keys=on", dbLog)
	if err != nil {
		panic(err)
	}
}

func ConnectToNumber(number string, jidString string) {
	// SqlContainer.PutDevice()
	if deviceStores, _ := SqlContainer.GetAllDevices(); true {
		for _, deviceStore := range deviceStores {
			println(deviceStore.ID.User)
		}
	}
	var JID types.JID
	if jidString != "" {
		JID, _ = types.ParseJID(jidString)
	} else {
		JID = types.NewJID(number, types.DefaultUserServer)
	}

	deviceStore, err := SqlContainer.GetDevice(JID)
	if err != nil {
		panic(err)
	}
	if deviceStore == nil {
		deviceStore = SqlContainer.NewDevice()
		// deviceStore = types.DEv(number, types.DefaultUserServer)
	}
	clientLog := waLog.Stdout("Client", "ERROR", true)
	client := whatsmeow.NewClient(deviceStore, clientLog)
	connection := &WhatsappConnection{Client: client, Number: number, ConnectionStatus: 0}
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
				connection.QrCodeString = evt.Code
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
