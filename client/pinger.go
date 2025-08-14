package client

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"
)

var startTime = time.Now()
var httpServer *http.Server

func getLocalIP() string {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "error fetching interfaces"
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue
			}
			return ip.String()
		}
	}
	return "no ip found"
}

func PortServe() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		var memStats runtime.MemStats
		runtime.ReadMemStats(&memStats)
		uptime := time.Since(startTime)

		fmt.Fprintf(w, "PID: %d\n", os.Getpid())
		fmt.Fprintf(w, "Uptime: %s\n", uptime)
		fmt.Fprintf(w, "Go Version: %s\n", runtime.Version())
		fmt.Fprintf(w, "Num CPU: %d\n", runtime.NumCPU())
		fmt.Fprintf(w, "Num Goroutines: %d\n", runtime.NumGoroutine())
		fmt.Fprintf(w, "OS/Arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
		fmt.Fprintf(w, "Alloc: %d KB\n", memStats.Alloc/1024)
		fmt.Fprintf(w, "TotalAlloc: %d KB\n", memStats.TotalAlloc/1024)
		fmt.Fprintf(w, "Sys: %d KB\n", memStats.Sys/1024)
		fmt.Fprintf(w, "NumGC: %d\n", memStats.NumGC)

		env := os.Environ()
		for i, e := range env {
			if i >= 10 {
				fmt.Fprintf(w, "...and %d more\n", len(env)-10)
				break
			}
			pair := strings.SplitN(e, "=", 2)
			fmt.Fprintf(w, "%s = %s\n", pair[0], pair[1])
		}
	})

	httpServer = &http.Server{
		Addr:    ":8000",
		Handler: mux,
	}

	go func() {
		ip := getLocalIP()
		fmt.Printf("[Local Host 8000] IP %s\n", ip)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()
}

func StopPortServe() {
	if httpServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		httpServer.Shutdown(ctx)
	}
}
