package mongodb

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/lfhy/morm/conf"
)

func TestInitWithConfigRejectsCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	conn, err := InitWithConfig(ctx, &conf.MongoDBConfig{Uri: "mongodb://127.0.0.1:1"})
	if conn != nil || !errors.Is(err, context.Canceled) {
		t.Fatalf("initialize canceled context: conn=%v err=%v", conn, err)
	}
}

func TestInitWithConfigHonorsDeadlineAndDoesNotPublishFailedConnection(t *testing.T) {
	previous := ORMConn
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	conn, err := InitWithConfig(ctx, &conf.MongoDBConfig{
		Uri: "mongodb://127.0.0.1:1", ReadMode: "nearest", OptionPoolSize: "1",
	})
	if conn != nil || !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("initialize with expired deadline: conn=%v err=%v", conn, err)
	}
	if ORMConn != previous {
		t.Fatal("failed initialization changed the legacy connection")
	}
}

func TestInitWithConfigDoesNotExposeCredentials(t *testing.T) {
	secret := "never-log-this-password"
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	conn, err := InitWithConfig(ctx, &conf.MongoDBConfig{
		Uri: "mongodb://username:" + secret + "@[::1",
	})
	if conn != nil || err == nil {
		t.Fatalf("invalid MongoDB URI: conn=%v err=%v", conn, err)
	}
	if strings.Contains(err.Error(), secret) {
		t.Fatal("initialization error disclosed the password")
	}
}

func TestInitWithConfigRejectsNilInputs(t *testing.T) {
	for _, tc := range []struct {
		name string
		ctx  context.Context
		cfg  *conf.MongoDBConfig
	}{
		{name: "context", cfg: &conf.MongoDBConfig{}},
		{name: "config", ctx: context.Background()},
	} {
		t.Run(tc.name, func(t *testing.T) {
			conn, err := InitWithConfig(tc.ctx, tc.cfg)
			if conn != nil || err == nil {
				t.Fatalf("nil %s: conn=%v err=%v", tc.name, conn, err)
			}
		})
	}
}
