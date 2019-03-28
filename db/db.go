// Package db manages data stored about the servers
// The storage is performed by a Key-Value community database called Badger.
package db

import (
	"os"
	"path/filepath"

	"github.com/dgraph-io/badger"
)

// DataTuple structure stores the DataTuples for the bot.
type DataTuple struct {
	Key   string `json:"Key"`
	Value string `json:"Value"`
}

// THIS SECTION REFERS TO MULTIPLE DBs. Unused atm
// // PointerDict structure stores pointers to the databases and a Mutex lock for preventing wrong usage of pointers.
// var PointerDict = struct {
// 	sync.RWMutex
// 	Dict map[string]*badger.DB
// }{Dict: make(map[string]*badger.DB)}

// // CloseDatabases close all databases in the PointerDict structure
// func CloseDatabases() {
// 	PointerDict.Lock()
// 	for _, i := range PointerDict.Dict {
// 		i.Lock()
// 		err := i.Close()
// 		if err != nil {
// 			log.Println("[DB] " + err.Error())
// 		}
// 	}
// 	PointerDict.Unlock()
// }

// RemoveDatabase function deletes the database in a folder
func RemoveDatabase(dir, id string) error {
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer d.Close()
	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}
	for _, name := range names {
		if name == id {
			err = os.RemoveAll(filepath.Join(dir, name))
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// // ConnectDB manages the database connection and configuration.
// func ConnectDB(databasePath string) (*badger.DB, error) {
// 	return connectDB(fmt.Sprintf(databasePath))
// }

// ConnectDB manages the database connection and configuration.
func ConnectDB(databasePath string) (*badger.DB, error) {
	opts := badger.DefaultOptions
	opts.Dir = databasePath
	opts.ValueDir = databasePath
	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}
	return db, nil
}

// UpdateDataTuple is a simple querry that inserts/updates the DataTuple tuple used by FastGate.
func UpdateDataTuple(database *badger.DB, key string, value string) error {
	return database.Update(func(txn *badger.Txn) error {
		err := txn.Set([]byte(key), []byte(value))
		return err
	})
}

// GetDataTuple finds an Key matching an key and returns it as a string.
func GetDataTuple(database *badger.DB, key string) (value string, err error) {
	var result []byte
	err = database.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}
		var val []byte
		val, err = item.ValueCopy(val)
		if err != nil {
			return err
		}
		result = val
		return err
	})
	return string(result), err
}

// DeleteDataTuple finds a matching Key and delets its data
func DeleteDataTuple(database *badger.DB, key string) error {
	return database.View(func(txn *badger.Txn) error {
		err := txn.Delete([]byte(key))
		return err
	})
}

// GetDataTuples function will read every entry in the database and return it as a list of DataTuples.
func GetDataTuples(database *badger.DB) (DataTuples []DataTuple, err error) {
	err = database.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 10
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			k := item.Key()
			var val []byte
			val, err = item.ValueCopy(val)
			if err != nil {
				return err
			}
			DataTuples = append(DataTuples, DataTuple{string(k), string(val)})
		}
		return nil
	})
	return
}
