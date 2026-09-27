package mongodb

import (
	"context"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestSwitchModelPreservesTransactionContext(t *testing.T) {
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI("mongodb://127.0.0.1:27017"))
	if err != nil {
		t.Fatal(err)
	}
	defer client.Disconnect(context.Background())
	session, err := client.StartSession()
	if err != nil {
		t.Fatal(err)
	}
	defer session.EndSession(context.Background())

	type contextKey struct{}
	ctx := mongo.NewSessionContext(context.WithValue(context.Background(), contextKey{}, "original"), session)
	model := &SessionModel{session: session, Model: &Model{
		Tx:                 &DBConn{Client: client},
		Ctx:                ctx,
		WhereList:          bson.M{"account": "first"},
		transactionSession: session,
	}}
	switched := model.SwitchModel("other_collection")
	if mongo.SessionFromContext(switched.GetContext()) != session {
		t.Fatal("switched model lost the MongoDB session")
	}
	if got := switched.GetContext().Value(contextKey{}); got != "original" {
		t.Fatalf("switched model context value = %v", got)
	}
	if len(switched.(*Model).WhereList) != 0 {
		t.Fatal("switched model inherited the original collection's filter")
	}

	switched.SetContext(context.WithValue(context.Background(), contextKey{}, "updated"))
	if mongo.SessionFromContext(switched.GetContext()) != session {
		t.Fatal("SetContext detached a switched model from its transaction")
	}
	if got := switched.GetContext().Value(contextKey{}); got != "updated" {
		t.Fatalf("switched model updated context value = %v", got)
	}
	model.SetContext(context.Background())
	if mongo.SessionFromContext(model.GetContext()) != session {
		t.Fatal("SetContext detached the session model from its transaction")
	}
}

func TestCloneMongoValueCopiesNestedFilters(t *testing.T) {
	original := bson.M{
		"$or":  bson.A{bson.M{"state": "new"}, bson.M{"ids": bson.A{1, 2}}},
		"sort": bson.D{{Key: "time", Value: 1}},
	}
	copy := cloneMongoValue(original).(bson.M)
	copy["$or"].(bson.A)[0].(bson.M)["state"] = "done"
	copy["$or"].(bson.A)[1].(bson.M)["ids"].(bson.A)[0] = 9
	copy["sort"].(bson.D)[0].Value = -1
	if got := original["$or"].(bson.A)[0].(bson.M)["state"]; got != "new" {
		t.Fatalf("original filter mutated: %v", got)
	}
	if got := original["$or"].(bson.A)[1].(bson.M)["ids"].(bson.A)[0]; got != 1 {
		t.Fatalf("original array mutated: %v", got)
	}
	if got := original["sort"].(bson.D)[0].Value; got != 1 {
		t.Fatalf("original sort mutated: %v", got)
	}
}
