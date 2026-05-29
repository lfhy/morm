package test

import (
	"fmt"
	"testing"
	"time"

	"github.com/lfhy/morm/db/mongodb"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Data struct {
	Str     string `bson:"str"`
	Int     int    `bson:"int"`
	Bool    bool   `bson:"bool"`
	BoolPtr *bool  `bson:"bool_ptr,must"`
	Empty   string `bson:"empty,must"`
}

func TestBSON(t *testing.T) {
	var data Data
	data.Str = "key"
	data.Int = 1
	data.Bool = true
	data.Empty = ""
	data.BoolPtr = &data.Bool
	m, err := mongodb.ConvertToBSONM(&data)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("m: %+v\n", m)
}

type oidStringData struct {
	ID string `bson:"_id"`
}

type oidObjectData struct {
	ID primitive.ObjectID `bson:"_id"`
}

type timeData struct {
	CreatedAt time.Time  `bson:"created_at"`
	UpdatedAt *time.Time `bson:"updated_at"`
}

func TestBSONOIDStringHexToObjectID(t *testing.T) {
	data := oidStringData{ID: "661a4faf9ecd66f803c400ea"}
	m, err := mongodb.ConvertToBSONM(data)
	if err != nil {
		t.Fatal(err)
	}

	if _, ok := m["_id"].(primitive.ObjectID); !ok {
		t.Fatalf("expected _id to be primitive.ObjectID, got %T", m["_id"])
	}
}

func TestBSONOIDStringFallbackToRawValue(t *testing.T) {
	data := oidStringData{ID: "role-super-admin"}
	m, err := mongodb.ConvertToBSONM(data)
	if err != nil {
		t.Fatal(err)
	}

	id, ok := m["_id"].(string)
	if !ok {
		t.Fatalf("expected _id to remain string, got %T", m["_id"])
	}
	if id != data.ID {
		t.Fatalf("expected _id to remain %q, got %q", data.ID, id)
	}
}

func TestBSONOIDObjectIDPreserved(t *testing.T) {
	oid := primitive.NewObjectID()
	data := oidObjectData{ID: oid}
	m, err := mongodb.ConvertToBSONM(data)
	if err != nil {
		t.Fatal(err)
	}

	got, ok := m["_id"].(primitive.ObjectID)
	if !ok {
		t.Fatalf("expected _id to remain primitive.ObjectID, got %T", m["_id"])
	}
	if got != oid {
		t.Fatalf("expected _id to remain %s, got %s", oid.Hex(), got.Hex())
	}
}

func TestBSONTimePreserved(t *testing.T) {
	now := time.Date(2026, time.May, 29, 8, 0, 0, 0, time.UTC)
	data := timeData{CreatedAt: now, UpdatedAt: &now}
	m, err := mongodb.ConvertToBSONM(data)
	if err != nil {
		t.Fatal(err)
	}

	createdAt, ok := m["created_at"].(time.Time)
	if !ok {
		t.Fatalf("expected created_at to remain time.Time, got %T", m["created_at"])
	}
	if !createdAt.Equal(now) {
		t.Fatalf("expected created_at to remain %v, got %v", now, createdAt)
	}

	updatedAt, ok := m["updated_at"].(time.Time)
	if !ok {
		t.Fatalf("expected updated_at to remain time.Time, got %T", m["updated_at"])
	}
	if !updatedAt.Equal(now) {
		t.Fatalf("expected updated_at to remain %v, got %v", now, updatedAt)
	}
}
