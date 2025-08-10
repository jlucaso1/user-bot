package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
	waLog "go.mau.fi/whatsmeow/util/log"

	"bot/client"
	"bot/config"
	ev "bot/events"
	_ "bot/messaging/plugins"
	sql "bot/sql"
)

var sock *whatsmeow.Client

func main() {
	ctx := context.Background()

	store := sqlstore.NewWithDB(sql.Conn, "sqlite", waLog.Stdout("DB", "ERROR", true))
	defer store.Close()

	if err := store.Upgrade(ctx); err != nil {
		log.Fatal("Store upgrade failed:", err)
	}

	device, _ := store.GetFirstDevice(ctx)
	sock = whatsmeow.NewClient(device, waLog.Stdout("Client", "INFO", true))
	sock.AddEventHandler(func(evt interface{}) {
		ev.EventHandler(sock, evt)
	})

	err := sock.Connect()
	if err != nil {
		log.Fatal("Connection failed:", err)
	}

	client.RequestPairCode(ctx, sock, config.AppConfig)
	client.SetClient(sock)
	client.PortServe()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	sock.Disconnect()
}
