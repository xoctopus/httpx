package confhttp

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/xoctopus/logx"
)

type Server struct {
	// TODO other configurations
	Addr string `url:""`

	srv     *http.Server
	handler http.Handler
}

func (s Server) Init(ctx context.Context) error {
	log := logx.From(ctx)
	srv := &http.Server{
		ReadHeaderTimeout: 30 * time.Second,
		Addr:              s.Addr,
		Handler:           s.handler,
	}

	go func() {
		log.Info("listen on %s", s.Addr)

		if err := srv.ListenAndServe(); err != nil {
			log.Error(err)

			if !errors.Is(err, http.ErrServerClosed) {
				panic(err)
			}
		}
	}()

	stopCh := make(chan os.Signal, 1)
	signal.Notify(stopCh, os.Interrupt, syscall.SIGTERM)
	<-stopCh

	timeout := 10 * time.Second

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	log.Info("shutdowning in %s", timeout)

	return srv.Shutdown(ctx)
}
