package mongodb

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/lfhy/morm/conf"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestClientTimeoutsDoNotCapDatabaseOperations(t *testing.T) {
	opts := options.Client()
	applyClientNetworkTimeouts(opts)
	if opts.Timeout != nil {
		t.Fatalf("database operation timeout unexpectedly configured: %v", *opts.Timeout)
	}
	if opts.ConnectTimeout == nil || *opts.ConnectTimeout != 30*time.Second {
		t.Fatalf("connect timeout = %v", opts.ConnectTimeout)
	}
	if opts.ServerSelectionTimeout == nil || *opts.ServerSelectionTimeout != 30*time.Second {
		t.Fatalf("server selection timeout = %v", opts.ServerSelectionTimeout)
	}
	explicit := options.Client().ApplyURI("mongodb://localhost/?connectTimeoutMS=17&serverSelectionTimeoutMS=23")
	applyClientNetworkTimeouts(explicit)
	if explicit.ConnectTimeout == nil || *explicit.ConnectTimeout != 17*time.Millisecond ||
		explicit.ServerSelectionTimeout == nil || *explicit.ServerSelectionTimeout != 23*time.Millisecond {
		t.Fatalf("explicit URI timeouts were overridden: connect=%v selection=%v",
			explicit.ConnectTimeout, explicit.ServerSelectionTimeout)
	}
}

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
