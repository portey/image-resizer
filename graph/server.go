package graph

import (
	"errors"
	"net/http"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	serviceerrors "github.com/portey/image-resizer/errors"
	"github.com/portey/image-resizer/graph/generated"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

import (
	"context"
	"fmt"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

type (
	Server struct {
		http      *http.Server
		runErr    error
		readiness bool
	}
)

func New(port int, resolver generated.ResolverRoot) *Server {
	srv := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{
		Resolvers: resolver,
	}))
	srv.SetErrorPresenter(func(ctx context.Context, err error) *gqlerror.Error {
		switch err.(type) {
		case serviceerrors.ServiceError:
			return &gqlerror.Error{
				Message: err.Error(),
				Extensions: map[string]interface{}{
					"type":     "service",
					"sub_type": err.Error(),
				},
			}
		case serviceerrors.InvalidParams:
			return &gqlerror.Error{
				Message: err.Error(),
				Extensions: map[string]interface{}{
					"type":     "service",
					"sub_type": "InvalidPayload",
					"details":  err,
				},
			}
		}

		return &gqlerror.Error{
			Message: err.Error(),
			Extensions: map[string]interface{}{
				"type":     "service",
				"sub_type": serviceerrors.Internal.Error(),
			},
		}
	})

	mux := http.NewServeMux()
	mux.HandleFunc("/", playground.Handler("GraphQL playground", "/query"))
	mux.HandleFunc("/query", srv.ServeHTTP)

	return &Server{
		http: &http.Server{
			Addr:    fmt.Sprintf(":%d", port),
			Handler: mux,
		},
	}
}

func (s *Server) Run(ctx context.Context, wg *sync.WaitGroup) {
	wg.Add(1)
	log.Info("graphql service: begin run")

	go func() {
		defer wg.Done()
		log.Debug("graphql service: addr=", s.http.Addr)
		err := s.http.ListenAndServe()
		s.runErr = err
		log.Info("graphql service: end run > ", err)
	}()

	go func() {
		<-ctx.Done()
		sdCtx, _ := context.WithTimeout(context.Background(), 5*time.Second) // nolint
		err := s.http.Shutdown(sdCtx)
		if err != nil {
			log.Info("graphql service shutdown (", err, ")")
		}
	}()

	s.readiness = true
}

func (s *Server) HealthCheck() error {
	if !s.readiness {
		return errors.New("http service is't ready yet")
	}
	if s.runErr != nil {
		return errors.New("http service: run issue")
	}
	return nil
}
