package client

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"bot/types"

	"github.com/joho/godotenv"
	"go.mau.fi/whatsmeow"
)

func RequestPairCode(ctx context.Context, client *whatsmeow.Client, config types.Config) {
	if client.Store.ID != nil {
		return
	}

	if config.UserPN == "" {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("enter your phone number (international format, e.g., 12345678900): ")
		input, _ := reader.ReadString('\n')
		config.UserPN = strings.TrimSpace(input)

		if len(config.UserPN) < 11 {
			log.Fatal("phone number too short")
		}

		envMap, _ := godotenv.Read()
		envMap["USER_PN"] = config.UserPN
		_ = godotenv.Write(envMap, ".env")
	}

	time.Sleep(2 * time.Second)
	log.Println("pairing new session with phone number:", config.UserPN)

	code, err := client.PairPhone(ctx, config.UserPN, true, whatsmeow.PairClientChrome, "Chrome (Windows)")
	if err != nil {
		log.Fatal("pairing failed:", err)
	}

	log.Println("pair code:", code)
}
