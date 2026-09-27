//go:build integration

package test

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/lfhy/morm"
	"github.com/lfhy/morm/db/mongodb"
	"go.mongodb.org/mongo-driver/bson"
)

func TestMongoInitContextTwoIsolatedDatabases(t *testing.T) {
	uri := os.Getenv("MONGO_TEST_URI")
	if uri == "" {
		t.Skip("set MONGO_TEST_URI for isolated replica-set integration test")
	}
	var suffix [8]byte
	if _, err := rand.Read(suffix[:]); err != nil {
		t.Fatal(err)
	}
	name := "morm_init_qa_" + hex.EncodeToString(suffix[:])
	legacy := mongodb.ORMConn
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	firstORM, err := morm.InitMongoDBWithDBConfigContext(ctx, &morm.MongoDBConfig{
		Uri: uri, Database: name + "_one", ReadMode: "master", OptionPoolSize: "2",
	})
	if err != nil {
		t.Fatal("initialize first test client failed")
	}
	first := firstORM.(*mongodb.DBConn)
	t.Cleanup(func() { disconnectMongoTestClient(t, first) })
	var topology struct {
		SetName string `bson:"setName"`
	}
	if err := first.Client.Database("admin").RunCommand(ctx, bson.D{{Key: "hello", Value: 1}}).Decode(&topology); err != nil || topology.SetName == "" {
		t.Fatal("test MongoDB must be an accessible replica set")
	}
	secondORM, err := morm.InitMongoDBWithDBConfigContext(ctx, &morm.MongoDBConfig{
		Uri: uri, Database: name + "_two", ReadMode: "nearest", OptionPoolSize: "2",
	})
	if err != nil {
		t.Fatal("initialize second test client failed")
	}
	second := secondORM.(*mongodb.DBConn)
	t.Cleanup(func() { disconnectMongoTestClient(t, second) })
	if first == second || first.Client == second.Client || first.Database == second.Database {
		t.Fatal("two database initializations shared a client or database")
	}
	if mongodb.ORMConn != legacy {
		t.Fatal("context initialization published a process-global connection")
	}
	if first.NearestClient != first.Client || second.NearestClient == second.Client {
		t.Fatal("read-mode client selection does not match the configuration")
	}
	if err := first.Client.Ping(ctx, nil); err != nil {
		t.Fatal("first connection lost its primary")
	}
	if err := second.NearestClient.Ping(ctx, nil); err != nil {
		t.Fatal("second connection lost its nearest client")
	}

	closed, closeCancel := context.WithCancel(ctx)
	closeCancel()
	orm, err := morm.InitMongoDBWithDBConfigContext(closed, &morm.MongoDBConfig{Uri: uri})
	if orm != nil || !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled initialization: orm=%v err=%v", orm, err)
	}
	deadline, deadlineCancel := context.WithTimeout(ctx, 10*time.Millisecond)
	defer deadlineCancel()
	orm, err = morm.InitMongoDBWithDBConfigContext(deadline, &morm.MongoDBConfig{
		Uri: "mongodb://127.0.0.1:1", Database: name + "_unreachable", ReadMode: "master",
	})
	if orm != nil || !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("deadline initialization: orm=%v err=%v", orm, err)
	}
}

func disconnectMongoTestClient(t *testing.T, conn *mongodb.DBConn) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := conn.Disconnect(ctx); err != nil {
		t.Errorf("disconnect test client: %v", err)
	}
}
