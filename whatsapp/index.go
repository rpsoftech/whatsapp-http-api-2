package whatsapp

import (
	"context"
	"fmt"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
	"github.com/rpsoftech/whatsapp-http-api/env"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	waLog "go.mau.fi/whatsmeow/util/log"
)

var SqlContainer *sqlstore.Container
var ctx = context.Background()

func InitSqlContainer() {
	dbLog := waLog.Stdout("Database", "WARN", true)
	// Make sure you add appropriate DB connector imports, e.g. github.com/mattn/go-sqlite3 for SQLite
	var err error
	SqlContainer, err = sqlstore.New(ctx, "sqlite3", fmt.Sprintf("file:%s?_foreign_keys=on", filepath.Join(env.CurrentDirectory, "WhatsappSuperSecrete.db")), dbLog)
	if err != nil {
		panic(err)
	}
}

func ConnectToNumber(jidString string, token string) {
	// SqlContainer.PutDevice()
	if deviceStores, _ := SqlContainer.GetAllDevices(ctx); true {
		for _, deviceStore := range deviceStores {
			println(deviceStore.ID.User)
		}
	}
	var JID types.JID
	if jidString != "" {
		JID, _ = types.ParseJID(jidString)
	}
	var deviceStore *store.Device
	if !JID.IsEmpty() {
		var err error
		deviceStore, err = SqlContainer.GetDevice(ctx, JID)
		if err != nil {
			println(err.Error())
		}
	}
	if deviceStore == nil {
		deviceStore = SqlContainer.NewDevice()
		// deviceStore = types.DEv(number, types.DefaultUserServer)
	}
	clientLog := waLog.Stdout("Client", "ERROR", true)
	client := whatsmeow.NewClient(deviceStore, clientLog)
	client.EnableAutoReconnect = true
	// client.
	println(client.LastSuccessfulConnect.String())

	// client.MessengerConfig = &whatsmeow.MessengerConfig{
	// 	UserAgent: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/123.0.0.0 Safari/537.36",
	// 	BaseURL:   "https://web.whatsapp.com",
	// }
	// client.PairPhone()
	connection := &WhatsappConnection{Client: client, ConnectionStatus: 0, SyncFinished: false, Token: token}
	ConnectionMap[token] = connection
	client.AddEventHandler(connection.eventHandler)

	connection.ConnectAndGetQRCode()
}
