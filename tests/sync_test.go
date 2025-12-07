package tests

import (
	"context"
	"testing"

	"github.com/C5rogers/G-Synch/pkg/sync"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
)

func TestGetSynchAPI(t *testing.T) {
	ctx := context.Background()
	// GIVEN given database and target database postgres pools
	givenDBPool, err := pgxpool.New(ctx, "postgres://user:password@localhost:5432/givendb")
	if err != nil {
		t.Fatalf("failed to create givenDB pool: %v", err)
	}
	defer givenDBPool.Close()

	targetDBPool, err := pgxpool.New(ctx, "postgres://user:password@localhost:5432/targetdb")
	if err != nil {
		t.Fatalf("failed to create targetDB pool: %v", err)
	}
	defer targetDBPool.Close()

	// WHEN we create a new SyncAPI
	syncAPI, err := sync.NewSyncAPI(givenDBPool, targetDBPool)
	if err != nil {
		t.Fatalf("failed to create syncAPI: %v", err)
	}

	// THEN expect syncAPI is not nil
	if syncAPI == nil {
		t.Fatalf("syncAPI is nil")
	}
	assert.NotNil(t, syncAPI)
}
