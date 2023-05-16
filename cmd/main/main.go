package main

import (
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func main() {
	e := echo.New()
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	s := http.Server{
		Addr:    ":8080",
		Handler: e,
	}
	if err := s.ListenAndServe(); err != http.ErrServerClosed {
		logger.DPanic("Error occurred", zap.Error(err))
	}
}