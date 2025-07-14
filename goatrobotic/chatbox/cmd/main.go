package main

import (
	"chatbox"
	"chatbox/service"
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

const (
	port            = 8080
	serverTimeoutMs = 10000
)

func main() {
	s := service.NewChatService()

	ep := chatbox.NewEndpoint(s)

	httpHandlers := chatbox.NewHTTPHandler(
		ep.JoinEndpoint,
		ep.SendEndpoint,
		ep.LeaveEndpoint,
		ep.MessageEndpoint,
	)

	httpServer := &http.Server{
		Addr: ":" + strconv.Itoa(port),
		Handler: http.TimeoutHandler(
			httpHandlers,
			time.Duration(serverTimeoutMs)*time.Millisecond,
			`{"error":"Server Execution Timeout"}`,
		),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
		BaseContext: func(_ net.Listener) context.Context {
			return context.Background()
		},
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		fmt.Println("Chat HTTP server started at http://localhost:" + strconv.Itoa(port))
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server failed: %v\n", err)
		}
	}()

	<-stop
	fmt.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Fatalf("Server Shutdown Failed: %v\n", err)
	}

	fmt.Println("Server gracefully stopped")
}
