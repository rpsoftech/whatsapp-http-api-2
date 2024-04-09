package whatsapp

import (
	_ "github.com/mattn/go-sqlite3"

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
	connection := &WhatsappConnection{Client: client, Number: number, ConnectionStatus: 0, SyncFinished: false}
	ConnectionMap[number] = connection
	client.AddEventHandler(connection.eventHandler)
	connection.ConnectAndGetQRCode()
}
