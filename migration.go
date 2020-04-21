package televise

import (
	"log"

	"github.com/gocql/gocql"
)

type Migration struct {
	Name    string
	Forward string
	Reverse string
}

func (m Migration) Do(db *gocql.Session) error {
	return db.Query(m.Forward).Exec()
}

func (m Migration) Undo(db *gocql.Session) error {
	return db.Query(m.Reverse).Exec()
}

const (
	sqlMigrationCreateTable = `
		CREATE TABLE IF NOT EXISTS migration (
			name text PRIMARY KEY
		);`
	sqlMigrationCount  = `SELECT COUNT(name) FROM migration;`
	sqlMigrationCommit = `INSERT INTO migration (name) VALUES (?);`
)

func Migrate(db *gocql.Session) error {
	err := db.Query(sqlMigrationCreateTable).Exec()
	if err != nil {
		return err
	}
	var last int
	err = db.Query(sqlMigrationCount).Scan(&last)
	if err != nil {
		return err
	}
	if last == len(migrations) {
		// all migrations have been run
		return nil
	}
	for i := last; i < len(migrations); i++ {
		m := migrations[i]
		log.Println("Running migration", m.Name)
		err = m.Do(db)
		if err != nil {
			return err
		}
		err = db.Query(sqlMigrationCommit, m.Name).Exec()
		if err != nil {
			m.Undo(db)
			return err
		}
	}
	return nil
}

var migrations = []Migration{
	{"2020-04-20-CreateSession", querySessionCreateTable, querySessionDropTable},
	{"2020-04-20-CreateVisit", queryVisitCreateTable, queryVisitDropTable},
	{"2020-04-20-CreateMetadata", queryMetadataCreateTable, queryMetadataDropTable},
}
