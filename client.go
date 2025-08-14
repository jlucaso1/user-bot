package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"go.mau.fi/whatsmeow"
	es "go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/store/sqlstore"
	waLog "go.mau.fi/whatsmeow/util/log"

	"bot/client"
	"bot/config"
	ev "bot/events"
	_ "bot/messaging/plugins"
	sql "bot/sql"
	"bot/utils"
)

var sock *whatsmeow.Client
var store *sqlstore.Container

func cleanup() {
	client.StopPortServe()
	if sock != nil {
		sock.Disconnect()
	}
	if store != nil {
		store.Close()
	}
}

func main() {
	log.Println("Starting Client...")
	ctx := context.Background()

	store = sqlstore.NewWithDB(sql.Conn, "sqlite", waLog.Stdout("DB", "ERROR", true))

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		<-stop
		log.Println("shutting down...")
		cleanup()
		log.Println("shutdown complete")
		os.Exit(0)
	}()

	if err := store.Upgrade(ctx); err != nil {
		log.Fatal("Store upgrade failed:", err)
	}
	es.SetOSInfo("Xstro", [3]uint32{10, 15, 7})

	device, _ := store.GetFirstDevice(ctx)
	sock = whatsmeow.NewClient(device, waLog.Stdout("Client", "INFO", true))
	sock.AddEventHandler(func(evt any) {
		ev.EventHandler(sock, evt)
	})

	if err := sock.Connect(); err != nil {
		log.Fatal("Connection failed:", err)
	}

	client.RequestPairCode(ctx, sock, config.AppConfig)
	client.SetClient(sock)
	client.PortServe()

	utils.SetCleanupFunc(cleanup)

	select {}
}
