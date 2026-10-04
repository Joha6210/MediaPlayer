package config

import (
	"database/sql"
	"errors"
)

type Database struct {
	databaseFile string
	database     *sql.DB
}

func InitDb(dbFile string) (*Database, error) {

	db, err := sql.Open("sqllite", dbFile)
	if err != nil {
		return &Database{}, err
	}

	//Should we initalize the different tables here?

	return &Database{
		databaseFile: dbFile,
		database:     db,
	}, nil
}

// Should only run once at first startup
func (*Database) NewTable(name string) error {
	return errors.New("Not implemented yet")
}
