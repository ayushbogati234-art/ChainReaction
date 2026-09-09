// Package app owns startup and dependency wiring.
package app

import (
	"context"
	"errors"
	"github.com/ayushbogati234-art/ChainReaction/backend/internal/config"
	graph "github.com/ayushbogati234-art/ChainReaction/backend/internal/graph/neo4j"
	"github.com/ayushbogati234-art/ChainReaction/backend/internal/persistence/postgres"
	"github.com/ayushbogati234-art/ChainReaction/backend/internal/service"
	transport "github.com/ayushbogati234-art/ChainReaction/backend/internal/transport/http"
	"github.com/jackc/pgx/v5/pgxpool"
	neo "github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"net/http"
	"time"
)

func dependencies(ctx context.Context, c config.Config) (postgres.Store, graph.Repository, func(), error) {
	db, err := pgxpool.New(ctx, c.PostgresDSN)
	if err != nil {
		return postgres.Store{}, graph.Repository{}, nil, err
	}
	store := postgres.Store{DB: db}
	if err = store.Migrate(ctx); err != nil {
		db.Close()
		return store, graph.Repository{}, nil, err
	}
	driver, err := neo.NewDriverWithContext(c.Neo4jURI, neo.BasicAuth(c.Neo4jUser, c.Neo4jPassword, ""))
	if err != nil {
		db.Close()
		return store, graph.Repository{}, nil, err
	}
	closeAll := func() {
		closeCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = driver.Close(closeCtx)
		db.Close()
	}
	if err = driver.VerifyConnectivity(ctx); err != nil {
		closeAll()
		return store, graph.Repository{}, nil, err
	}
	return store, graph.Repository{Driver: driver}, closeAll, nil
}
func RunAPI(ctx context.Context, c config.Config) error {
	store, graph, closeAll, err := dependencies(ctx, c)
	if err != nil {
		return err
	}
	defer closeAll()
	handler := transport.Router(service.ProjectService{Store: store}, service.ScanService{Store: store}, service.GraphService{Graph: graph, Scans: store}, service.FindingService{Store: store, Scans: store})
	addr := c.HTTPAddr
	if addr == "" {
		addr = ":8080"
	}
	server := http.Server{Addr: addr, Handler: handler, ReadHeaderTimeout: 5 * time.Second}
	done := make(chan struct{})
	defer close(done)
	go func() {
		select {
		case <-done:
			return
		case <-ctx.Done():
		}
		stop, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = server.Shutdown(stop)
	}()
	err = server.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}
