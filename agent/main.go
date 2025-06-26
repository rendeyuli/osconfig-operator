package main

import (
	 "context"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx,cancel := context.WithCancel(context.Background())
	
	//监听退出信号
	go func(){
		ch:= make(chan os.Signal,1) 
		signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
		<-ch
		cancel()
	}()

	// 启动ConfigMap监听
	if err := WatchConfigMap(ctx); err != nil {
		log.Fatalf("failed to start ConfigMap watcher: %v", err)
	}
}

