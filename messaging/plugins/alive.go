package plugins

import (
	"fmt"
	"runtime"
	"time"

	"bot/client"
	"bot/messaging"
	btypes "bot/types"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/types/events"
)

var startTime = time.Now()

func init() {
	messaging.RegisterCommand(&btypes.Command{
		Name:     "alive",
		Category: "System",
		FromMe:   false,
		IsGroup:  false,
		Handler:  Alive,
	})
}

func Alive(msg *events.Message, _ []string, sock *whatsmeow.Client) {
	uptime := time.Since(startTime).Truncate(time.Second)
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	response := fmt.Sprintf(
		"```\nBOT STATUS\n"+
			"Runtime      : %s\n"+
			"Go Version   : %s\n"+
			"Memory Usage : %.2f MB\n"+
			"Goroutines   : %d\n"+
			"CPU Cores    : %d\n"+
			"Platform     : %s/%s\n```",
		uptime,
		runtime.Version(),
		float64(memStats.Alloc)/1024/1024,
		runtime.NumGoroutine(),
		runtime.NumCPU(),
		runtime.GOOS,
		runtime.GOARCH,
	)

	_, _ = client.SendMessage(btypes.SendOptions{
		JID:  msg.Info.Chat,
		Type: btypes.MsgText,
		Text: response,
	})
}
