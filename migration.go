package televise

import (
	"context"
	"database/sql"
	"log"
)

type Migration struct {
	Name    string
	Forward string
	Reverse string
}

func (m Migration) Do(ctx context.Context, db *sql.DB) error { return execSQL(ctx, db, m.Forward) }

func (m Migration) Undo(ctx context.Context, db *sql.DB) error { return execSQL(ctx, db, m.Reverse) }

func execSQL(ctx context.Context, db *sql.DB, tsql string) error {
	stmt, err := db.PrepareContext(ctx, tsql)
	if err != nil {
		return err
	}
	_, err = stmt.ExecContext(ctx)
	return err
}

const (
	sqlCreateMigrationTable = `
IF NOT EXISTS (SELECT * FROM sysobjects WHERE name='Migration' and xtype='U')
	CREATE TABLE Migration (
		Name VARCHAR(255) PRIMARY KEY NOT NULL
	);`
	sqlLastMigration = `SELECT COUNT(Name) FROM Migration;`
	sqlLogMigration  = `INSERT INTO Migration (Name) VALUES (@Name);`
)

func Migrate(db *sql.DB) error {
	ctx := context.Background()
	stmt, err := db.PrepareContext(ctx, sqlCreateMigrationTable)
	if err != nil {
		return err
	}
	_, err = stmt.ExecContext(ctx)
	if err != nil {
		return err
	}
	stmt, err = db.PrepareContext(ctx, sqlLastMigration)
	if err != nil {
		return err
	}
	var last int
	err = stmt.QueryRowContext(ctx).Scan(&last)
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
		err = m.Do(ctx, db)
		if err != nil {
			return err
		}
		stmt, err = db.PrepareContext(ctx, sqlLogMigration)
		if err != nil {
			return err
		}
		_, err = stmt.ExecContext(ctx, sql.Named("Name", m.Name))
		if err != nil {
			// try to undo the migration, ignore any errors
			m.Undo(ctx, db)
			return err
		}
	}
	return nil
}

var migrations = []Migration{
	{"2020-04-11-CreateSession", sqlCreateSessionTable, sqlDropSessionTable},
	{"2020-04-13-CreateMetadata", sqlMetadataTable, sqlMetadataDropTable},
	{"2020-04-14-CreateOption", sqlCreateOptionTable, sqlDropOptionTable},
	{"2020-04-14-CreateVote", sqlCreateVoteTable, sqlDropVoteTable},
}
