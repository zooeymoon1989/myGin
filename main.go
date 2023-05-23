package main

import (
	"context"
	"errors"
	"log"
	"myGin/bootstrap"
	"myGin/common"
	"myGin/routes"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	if err := bootstrap.Bootstrap(common.YamlFile); err != nil {
		panic(err)
	}
	srv := &http.Server{
		Addr:    common.Addr,
		Handler: routes.Routes(),
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil && errors.Is(err, http.ErrServerClosed) {
			log.Printf("listen: %s\n", err)
		}
	}()
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}
	log.Println("Server exiting")
}
