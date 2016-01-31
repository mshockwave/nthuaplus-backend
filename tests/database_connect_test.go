package tests

import (
	"testing"
	"../public"
)

func TestUserDbConnect(t *testing.T) {
	userDb := public.GetNewUserDatabase()
	if userDb == nil {
		t.Error("UserDb session nil")
		t.FailNow()
	}
	defer userDb.Session.Close()

	if names, err := userDb.CollectionNames(); err != nil {
		t.Error("Fail getting collections")
		t.FailNow()
	} else{
		t.Logf("Collection counts: %d", len(names))
		for _, name := range names {
			t.Log("Get collection: " + name)
		}
	}
}
