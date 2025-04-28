package storage

import (
	"bytes"
	"database/sql"
	"encoding/binary"
	"errors"
	"sync"
	"go-learning/internal/model"

	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

type PersonStore interface {
	GetPersons() *[]model.Person
	AddPerson(p model.Person) error
	Count() int
}

type MemoryStore struct {
	db      *sql.DB
	lock    sync.RWMutex
	maxSize int
	Persons []model.Person // In-memory storage for persons
}

// NewMemoryStore creates a memory-backed store with SQL database persistence.
// It initializes the database, creates the necessary table if it doesn't exist, and loads existing persons.
func NewMemoryStore(dbPath string, maxSize int) (*MemoryStore, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	// Create the table if it doesn't exist
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS persons (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			features BLOB NOT NULL
		)
	`)
	if err != nil {
		return nil, err
	}

	ms := &MemoryStore{
		db:      db,
		maxSize: maxSize,
		Persons: []model.Person{}, // Initialize the in-memory slice
	}

	// Load existing persons into memory
	if err := ms.loadAllPersons(); err != nil {
		return nil, err
	}

	return ms, nil
}

// AddPerson adds a person to the database and the in-memory slice.
func (ms *MemoryStore) AddPerson(p model.Person) error {
	ms.lock.Lock()
	defer ms.lock.Unlock()

	// Check if the maximum size is reached
	if len(ms.Persons) >= ms.maxSize {
		return errors.New("maximum number of persons reached in the store")
	}

	// Serialize the features array to a byte slice
	featuresBlob, err := serializeFeatures(p.Features)
	if err != nil {
		return err
	}

	// Add to the database
	_, err = ms.db.Exec("INSERT INTO persons (name, features) VALUES (?, ?)", p.Name, featuresBlob)
	if err != nil {
		return err
	}

	// Add to the in-memory slice
	ms.Persons = append(ms.Persons, p)
	return nil
}

// Count returns the number of persons in the in-memory slice.
func (ms *MemoryStore) Count() int {
	ms.lock.RLock()
	defer ms.lock.RUnlock()

	return len(ms.Persons)
}

// 
func (ms *MemoryStore) GetPersons() *[]model.Person {
	ms.lock.RLock()
	defer ms.lock.RUnlock()
	return &ms.Persons
}

// loadAllPersons retrieves all persons from the database and loads them into the in-memory slice.
func (ms *MemoryStore) loadAllPersons() error {
	ms.lock.Lock()
	defer ms.lock.Unlock()

	rows, err := ms.db.Query("SELECT name, features FROM persons")
	if err != nil {
		return err
	}
	defer rows.Close()

	// Iterate through the rows and deserialize the features
	var name string
	var featuresBlob []byte

	for rows.Next() {
		if err := rows.Scan(&name, &featuresBlob); err != nil {
			return err
		}

		features, err := deserializeFeatures(featuresBlob)
		if err != nil {
			return err
		}

		ms.Persons = append(ms.Persons, model.Person{
			Name:     name,
			Features: features,
		})
	}

	// Check for errors from iterating over rows
	if err := rows.Err(); err != nil {
		return err
	}

	return nil
}

// serializeFeatures converts a [256]float64 array into a byte slice.
func serializeFeatures(features [256]float64) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, features)
	return buf.Bytes(), err
}

// deserializeFeatures converts a byte slice back into a [256]float64 array.
func deserializeFeatures(data []byte) ([256]float64, error) {
	var features [256]float64
	buf := bytes.NewReader(data)
	err := binary.Read(buf, binary.LittleEndian, &features)
	return features, err
}
