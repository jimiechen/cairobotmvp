package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/jimiechen/mineplanet/go/tars/system/localhandler"
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

	// TODO: 启动真实 TarsGo server
	// 当前阶段 System 模块通过 localhandler 被 Gateway 调用
	handler := localhandler.NewHandler()
	_, _, _ = handler.Invoke(ctx, []byte{}, map[string]string{"method": "HealthCheck"})

	fmt.Println("system server started")
	<-ctx.Done()
}
