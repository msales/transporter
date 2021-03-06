package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/msales/transporter"
	"github.com/msales/transporter/server"
	"github.com/msales/transporter/server/middleware"
	"gopkg.in/urfave/cli.v1"
)

func runServer(c *cli.Context) {
	ctx, err := newContext(c)
	if err != nil {
		log.Fatal(err.Error())
	}

	app, err := newApplication(ctx)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer app.Close()

	port := c.String(FlagPort)
	srv := newServer(ctx, app)
	h := http.Server{Addr: ":" + port, Handler: srv}
	go func() {
		ctx.logger.Info(fmt.Sprintf("Starting server on port %s", port))
		if err := h.ListenAndServe(); err != nil {
			if err != http.ErrServerClosed {
				log.Fatal(err)
			}
		}
	}()

	quit := listenForSignals()
	<-quit

	ctxServer, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	h.Shutdown(ctxServer)
	ctx.logger.Info("Server stopped gracefully")
}

func newServer(ctx *Context, app *transporter.Application) http.Handler {
	s := server.New(app)

	h := middleware.Common(s)
	return middleware.WithContext(ctx, h)
}

// Wait for SIGTERM to end the application.
func listenForSignals() chan bool {
	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs

		done <- true
	}()

	return done
}
