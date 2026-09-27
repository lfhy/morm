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
	"go.mongodb.org/mongo-driver/mongo"
)

type transactionAccount struct {
	ID      string `bson:"_id"`
	Balance int64  `bson:"balance"`
}

func (transactionAccount) TableName() string { return "transaction_accounts" }

type transactionEntry struct {
	ID     string `bson:"_id"`
	Amount int64  `bson:"amount"`
}

func (transactionEntry) TableName() string { return "transaction_entries" }

func TestMongoSessionSpansCollectionsAndRollsBack(t *testing.T) {
	uri := os.Getenv("MORM_MONGO_TEST_URI")
	if uri == "" {
		t.Skip("set MORM_MONGO_TEST_URI for an isolated MongoDB replica-set integration test")
	}
	var suffix [8]byte
	if _, err := rand.Read(suffix[:]); err != nil {
		t.Fatal(err)
	}
	database := "morm_session_qa_" + hex.EncodeToString(suffix[:])
	orm, err := morm.InitMongoDBWithDBConfigWithError(&morm.MongoDBConfig{
		Uri: uri, Database: database, OptionPoolSize: "4", ReadMode: "master", W: "majority",
	})
	if err != nil {
		t.Fatal("cannot initialize test MongoDB connection")
	}
	conn, ok := orm.(*mongodb.DBConn)
	if !ok {
		t.Fatal("morm did not return a MongoDB connection")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	defer func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cleanupCancel()
		// This database name was generated only for this test; never drop an existing database.
		if err := conn.Client.Database(database).Drop(cleanupCtx); err != nil {
			t.Errorf("drop isolated test database: %v", err)
		}
		if err := conn.Disconnect(cleanupCtx); err != nil {
			t.Errorf("disconnect test client: %v", err)
		}
	}()
	if err := conn.Ping(ctx, nil); err != nil {
		t.Fatal("cannot authenticate and ping test MongoDB")
	}
	var hello struct {
		SetName string `bson:"setName"`
	}
	if err := conn.Client.Database("admin").RunCommand(ctx, bson.D{{Key: "hello", Value: 1}}).Decode(&hello); err != nil {
		t.Fatal("cannot inspect test MongoDB topology")
	}
	if hello.SetName == "" {
		t.Fatal("MongoDB must be a replica set for multi-collection transactions")
	}

	base := orm.Model(&transactionAccount{}).SetContext(ctx)
	rollback := errors.New("roll back synthetic transfer")
	err = base.Session(func(tx morm.Session) error {
		if mongo.SessionFromContext(tx.GetContext()) == nil {
			t.Error("session model lost the driver session context")
		}
		if _, err := tx.Create(&transactionAccount{ID: "rollback-account", Balance: 10}); err != nil {
			return err
		}
		entries := tx.SwitchModel(&transactionEntry{}).SetContext(ctx)
		if mongo.SessionFromContext(entries.GetContext()) == nil {
			t.Error("SwitchModel lost the driver session context")
		}
		if _, err := entries.Create(&transactionEntry{ID: "rollback-entry", Amount: 10}); err != nil {
			return err
		}
		if err := entries.BulkWrite([]mongo.WriteModel{
			mongo.NewInsertOneModel().SetDocument(&transactionEntry{ID: "rollback-bulk", Amount: 10}),
		}, true); err != nil {
			return err
		}
		var accounts []transactionAccount
		if err := tx.Find().All(&accounts); err != nil {
			return err
		}
		if len(accounts) != 1 {
			t.Errorf("transaction read saw %d accounts, want 1", len(accounts))
		}
		return rollback
	})
	if !errors.Is(err, rollback) {
		t.Fatalf("rollback error = %v, want sentinel", err)
	}
	if mongo.SessionFromContext(base.GetContext()) != nil {
		t.Fatal("transaction session context leaked into the parent model")
	}
	for _, collection := range []string{"transaction_accounts", "transaction_entries"} {
		count, err := conn.Client.Database(database).Collection(collection).CountDocuments(ctx, bson.D{})
		if err != nil || count != 0 {
			t.Fatalf("%s after rollback: count=%d err=%v", collection, count, err)
		}
	}

	err = base.Session(func(tx morm.Session) error {
		if _, err := tx.Create(&transactionAccount{ID: "committed-account", Balance: 20}); err != nil {
			return err
		}
		_, err := tx.SwitchModel(&transactionEntry{}).Create(&transactionEntry{ID: "committed-entry", Amount: 20})
		return err
	})
	if err != nil {
		t.Fatalf("commit synthetic transfer: %v", err)
	}
	for _, collection := range []string{"transaction_accounts", "transaction_entries"} {
		count, err := conn.Client.Database(database).Collection(collection).CountDocuments(ctx, bson.D{})
		if err != nil || count != 1 {
			t.Fatalf("%s after commit: count=%d err=%v", collection, count, err)
		}
	}

	err = base.Session(func(tx morm.Session) error {
		if _, err := tx.Create(&transactionAccount{ID: "manual-rollback", Balance: 30}); err != nil {
			return err
		}
		return tx.Rollback()
	})
	if err != nil {
		t.Fatalf("explicit rollback: %v", err)
	}
	count, err := conn.Client.Database(database).Collection("transaction_accounts").CountDocuments(ctx, bson.D{})
	if err != nil || count != 1 {
		t.Fatalf("accounts after explicit rollback: count=%d err=%v", count, err)
	}

	err = base.Session(func(tx morm.Session) error {
		if _, err := tx.Create(&transactionAccount{ID: "manual-commit", Balance: 40}); err != nil {
			return err
		}
		return tx.Commit()
	})
	if err != nil {
		t.Fatalf("explicit commit: %v", err)
	}
	count, err = conn.Client.Database(database).Collection("transaction_accounts").CountDocuments(ctx, bson.D{})
	if err != nil || count != 2 {
		t.Fatalf("accounts after explicit commit: count=%d err=%v", count, err)
	}
}
