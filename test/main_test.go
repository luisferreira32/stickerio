package test

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	ctx, _ := signal.NotifyContext(context.Background(), os.Interrupt)

	ticker := time.NewTicker(1 * time.Second)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}

		resp, err := http.DefaultClient.Get("http://localhost:8080")
		if err == nil && resp.Body != nil {
			break
		}
		log.Printf("server not running yet: %v", err)
	}
	os.Exit(m.Run())
}
