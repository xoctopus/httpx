package confhttp

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/xoctopus/logx"
)

type Server struct {
	Addr string `url:""`
	Name string `url:""`

	srv     *http.Server
	handler http.Handler
}

func (s *Server) Run(ctx context.Context) error {
	log := logx.From(ctx)

	s.srv = &http.Server{
		ReadHeaderTimeout: 30 * time.Second,
		Addr:              s.Addr,
		Handler:           s.handler,
	}

	log.Info("%s started and listenning on %s", s.Name, s.Addr)
	if err := s.srv.ListenAndServe(); err != nil {
		log.Error(err)

		if !errors.Is(err, http.ErrServerClosed) {
			panic(err)
		}
		return err
	}
	return nil
}

func (s *Server) ApplyHandler(h http.Handler) {
	s.handler = h
}

func (s *Server) Shutdown(ctx context.Context) error {
	if s.srv != nil {
		timeout := 10 * time.Second

		logx.From(ctx).Info("%s shutdown in %s", s.Name, timeout)

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		return s.srv.Shutdown(ctx)
	}
	return nil
}

func (s Server) Close(ctx context.Context) error {
	if s.srv != nil {
		return s.srv.Close()
	}
	return nil
}

func ListenAndServe(ctx context.Context, addr string, h http.Handler) error {
	return (&Server{Addr: addr, handler: h}).Run(ctx)
}
