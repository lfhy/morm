package test

import (
	"fmt"
	"testing"

	"github.com/lfhy/morm/db/mongodb"
)

type Data struct {
	Str     string `bson:"str"`
	Int     int    `bson:"int"`
	Bool    bool   `bson:"bool"`
	BoolPtr *bool  `bson:"bool_ptr,omitempty"`
	Empty   string `bson:"empty,omitempty"`
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
