package bdb

import (
	"os"
	"testing"
)

func TestModel(t *testing.T) {
	if _, err := os.Stat(dbPath); !os.IsNotExist(err) {
		err = os.RemoveAll(dbPath)
		if err != nil {
			t.Fatal("Unable to clean Test Database Before testing. Check for permissions.")
		}
	}
	b, err := NewBadgerDB(dbPath)
	if err != nil {
		t.Fatalf(err.Error())
	}

	t.Run("insersions", func(t *testing.T) {
		t.Run("statistics", func(t *testing.T) {
			t.Parallel()
			stats := map[string]int{}
			stats[testKey] = 1
			stats[testValue] = 2
			err := b.SetStatistics(stats)
			if err != nil {
				t.Log("failed setting statistics")
				t.Fail()
			}
		})
		t.Run("welcomeMessage", func(t *testing.T) {
			t.Parallel()
			err := b.SetWelcomeMessageSent(testKey, true)
			if err != nil {
				t.Fail()
			}
		})
		t.Run("guildLang", func(t *testing.T) {
			t.Parallel()
			err := b.SetGuildLang(testKey, testValue)
			if err != nil {
				t.Fail()
			}
		})
	})

	t.Run("reads", func(t *testing.T) {
		t.Run("statistics", func(t *testing.T) {
			t.Parallel()
			stats, err := b.GetStatistics()
			if err != nil {
				t.Errorf(err.Error())
				t.Fail()
			}
			if len(stats) <= 0 {
				t.Log("Database was empty")
				t.Fail()
			}
			if 1 != stats[testKey] && 2 != stats[testValue] {
				t.Log("Statistics values are not the same.", stats[testKey], stats[testValue])
				t.Fail()
			}
		})
		t.Run("welcomeMessage", func(t *testing.T) {
			t.Parallel()
			wasWelcome, err := b.WasWelcomeMessageSent(testKey)
			if !wasWelcome {
				t.Fail()
			}
			if err != nil {
				t.Fail()
			}
		})
		t.Run("guildLang", func(t *testing.T) {
			t.Parallel()
			lang, err := b.GetGuildLang(testKey)
			if err != nil {
				t.Fail()
			}
			if lang != testValue {
				t.Fail()
			}
		})
	})

	err = b.Close()
	if err != nil {
		t.Log("Failed to close db conn")
		t.Fail()
	}

	err = os.RemoveAll(dbPath)
	if err != nil {
		t.Logf("Unable to clean Test Database Aftere test. Check for permissions, and remove foleder '%s' or Future Tests might Fail", dbPath)
	}
}
