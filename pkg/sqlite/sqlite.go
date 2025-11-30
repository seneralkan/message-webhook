package sqlite

import (
	"database/sql"
	"fmt"
	"log"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

type ISqliteInstance interface {
	Database() *sql.DB
	Close() error
	InitTables(schemas []string) error
}

type sqliteInstance struct {
	db *sql.DB
}

func NewSqliteInstance(dbName string) (ISqliteInstance, error) {
	instance := &sqliteInstance{}
	if err := instance.initDB(dbName); err != nil {
		log.Fatal("Failed to initialize database", err)
		return nil, err
	}
	return instance, nil
}

func NewSqliteInstanceWithSchemas(dbName string, schemas []string) (ISqliteInstance, error) {
	instance := &sqliteInstance{}
	if err := instance.initDB(dbName); err != nil {
		log.Fatal("Failed to initialize database", err)
		return nil, err
	}
	if err := instance.InitTables(schemas); err != nil {
		log.Fatal("Failed to initialize tables", err)
		return nil, err
	}
	return instance, nil
}

func (s *sqliteInstance) Database() *sql.DB {
	return s.db
}

func (s *sqliteInstance) Close() error {
	return s.db.Close()
}

func (s *sqliteInstance) initDB(dbName string) error {
	// Use the dbName as provided - it could be a full path or just a name
	dbPath := dbName
	// Only add .db extension if it's not already present and it's not a full path
	if filepath.Ext(dbName) == "" && !filepath.IsAbs(dbName) {
		dbPath = fmt.Sprintf("./%s.db", dbName)
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return err
	}

	s.db = db
	return nil
}

func (s *sqliteInstance) InitTables(schemas []string) error {
	if s.db == nil {
		return fmt.Errorf("database not initialized")
	}

	for _, schema := range schemas {
		if _, err := s.db.Exec(schema); err != nil {
			return fmt.Errorf("failed to execute schema: %v", err)
		}
	}

	return nil
}
