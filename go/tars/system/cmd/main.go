package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/jimiechen/mineplanet/go/tars/system/adapter"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		cancel()
	}()

	sysAdapter := adapter.NewSystemAdapter()
	_, _, _ = sysAdapter.Invoke(ctx, []byte{}, map[string]string{"method": "HealthCheck"})

	fmt.Println("system server started")
	<-ctx.Done()
}
