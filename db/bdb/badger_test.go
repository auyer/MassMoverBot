package bdb

import (
	"os"
	"testing"
)

const (
	dbPath    = "./massmoverbot.db_test.go.db"
	testKey   = "TestKey"
	testValue = "TestValue"
)

func TestDatabase(t *testing.T) {
	if _, err := os.Stat(dbPath); !os.IsNotExist(err) {
		err = os.RemoveAll(dbPath)
		if err != nil {
			t.Fatal("Unable to clean Test Database Before testing. Check for permissions.")
		}
	}
	database, err := ConnectDB(dbPath)
	if err != nil {
		t.Errorf("Unable to Init Database")
	}
	err = UpdateDataTuple(database, testKey, testValue)
	if err != nil {
		t.Errorf("Unable to Insert Tuple")
	}
	value, err := GetDataTuple(database, testKey)
	if err != nil {
		t.Errorf("Unable to Fetch Tuple")
	}
	if value != testValue {
		t.Errorf("Received Value not mathing with what was inserted.")
	}
	values, err := GetDataTuples(database)
	if err != nil {
		t.Errorf("Unable to Fetch Tuple")
	}
	testValues := []DataTuple{{testKey, testValue}}
	if values[0] != testValues[0] {
		t.Errorf("Received Value not mathing with what was inserted.")
	}
	err = database.Close()
	if err != nil {
		t.Errorf("Failed at Closing Database")
	}
	err = os.RemoveAll(dbPath)
	if err != nil {
		t.Logf("Unable to clean Test Database Aftere test. Check for permissions, and remove foleder '%s' or Future Tests might Fail", dbPath)
	}
}
