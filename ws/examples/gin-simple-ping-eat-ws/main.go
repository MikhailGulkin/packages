package main

import (
	"context"
	"fmt"
	"github.com/MikhailGulkin/packages/log"
	"github.com/MikhailGulkin/packages/ws"
	"github.com/gin-gonic/gin"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	logger := log.Default()

	manager := ws.NewManager(
		ws.WithProcessorFabric(&ws.PipeProcessorFabricImpl{}),
	)
	defer func() {
		err := manager.Close()
		if err != nil {
			logger.Errorw("error closing manager", "error", err)
		}
		logger.Infow("manager closed")
	}()
	app := gin.Default()
	go func() {
		manager.Run(context.Background())
	}()
	app.GET("/ws", func(c *gin.Context) {
		err := manager.Process(rand.Int(), c.Writer, c.Request, nil)
		if err != nil {
			fmt.Println("error", err)
		}
	})
	go func() {
		app.Run(":8000")
	}()

	exitCh := make(chan os.Signal)
	signal.Notify(exitCh, os.Interrupt, syscall.SIGTERM)
	<-exitCh

}
