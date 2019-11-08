package bdb

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/dgraph-io/badger"
)

// BadgerDB struct stores a connection to a Badger embedded database
type BadgerDB struct {
	conn *badger.DB
}

func initDB(DatabasePath string) (*badger.DB, error) {
	err := os.Mkdir(DatabasePath, os.ModePerm)
	if err != nil && !os.IsExist(err) {
		log.Println("Error creating Databases folder: ", err)
		return nil, err
	}

	conn, err := ConnectDB(DatabasePath + "/db")
	if err != nil {
		log.Println("Error creating guildDB " + err.Error())
		return nil, err
	}
	return conn, nil
}

// NewBadgerDB creates connection to a Badger Database
func NewBadgerDB(path string) (*BadgerDB, error) {
	conn, err := initDB(path)
	bDB := &BadgerDB{
		conn: conn,
	}

	stats, err := bDB.GetStatistics()
	if err != nil {
		if err.Error() != "Failed to get Statistics: Key not found" {
			log.Println("Error reading guildDB " + err.Error())
			return bDB, err
		}
		log.Println("Failed to get Statistics")
		stats = map[string]int{}
		stats["usrs"] = 0
		stats["movs"] = 0
		err = bDB.SetStatistics(stats)
		return bDB, err
	}
	log.Println(fmt.Sprintf("Moved %d players in %d actions", stats["usrs"], stats["movs"]))
	return bDB, err
}

// GetStatistics retrieves a statistics object in the Database
func (b *BadgerDB) GetStatistics() (map[string]int, error) {
	bytesStats, err := GetDataTupleBytes(b.conn, "statistics")
	if err != nil {
		return nil, fmt.Errorf("Failed to get Statistics: %s", err)
	}
	stats := make(map[string]int)
	err = json.Unmarshal(bytesStats, &stats)
	if err != nil {
		return nil, fmt.Errorf("Failed to decode Statistics: %s", err)
	}
	return stats, nil
}

// SetStatistics creates/updates a statistics object in the Database
func (b *BadgerDB) SetStatistics(stats map[string]int) error {
	bytesStats, err := json.Marshal(stats)
	if err != nil {
		return err
	}
	return UpdateDataTupleBytes(b.conn, "statistics", bytesStats)

}

// WasWelcomeMessageSent retrieves the status for a sent message
func (b *BadgerDB) WasWelcomeMessageSent(id string) (bool, error) {
	val, err := GetDataTuple(b.conn, "M:"+id)
	if val == "1" {
		return true, err
	}
	return false, err
}

// SetWelcomeMessageSent sets a status for a "was message sent" to True
func (b *BadgerDB) SetWelcomeMessageSent(id string, value bool) error {
	binValue := "1"
	if !value {
		binValue = "0"
	}
	return UpdateDataTuple(b.conn, "M:"+id, binValue)
}

// GetGuildLang retrieves a Language entry in the database
func (b *BadgerDB) GetGuildLang(id string) (string, error) {
	return GetDataTuple(b.conn, id)
}

// SetGuildLang creates/updates a Language entry in the database
func (b *BadgerDB) SetGuildLang(id, value string) error {
	return UpdateDataTuple(b.conn, id, value)
}

// DeleteGuildLang deletes a Language entry in the database
func (b *BadgerDB) DeleteGuildLang(id string) error {
	return DeleteDataTuple(b.conn, id)
}

// Close the DB connection
func (b BadgerDB) Close() error {
	return b.conn.Close()
}
