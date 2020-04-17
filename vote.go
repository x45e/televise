package televise

import (
	"context"
	"database/sql"
)

const (
	sqlCreateOptionTable = `
CREATE TABLE [Option] (
	[Id] BIGINT IDENTITY(1, 1) PRIMARY KEY NOT NULL,
	[Title] VARCHAR(4096) NOT NULL
);`
	sqlDropOptionTable = `DROP TABLE [Option];`

	sqlOptionInsert = `INSERT INTO [Option] ([Title]) OUTPUT INSERTED.[Id] VALUES (@Title);`

	sqlCreateVoteTable = `
CREATE TABLE [Vote] (
	[Key] BINARY(20) NOT NULL,
	[OptionId] BIGINT NOT NULL FOREIGN KEY REFERENCES [Option]([Id]),
	[At] DATETIME NOT NULL DEFAULT GETDATE()
);
ALTER TABLE [Vote] ADD CONSTRAINT [PK_Vote] PRIMARY KEY ([Key], [OptionId]);`

	sqlDropVoteTable = `DROP TABLE [Vote];`

	sqlVoteInsert = `INSERT INTO [Vote] ([Key], [OptionId]) VALUES (@Key, @OptionId);`

	sqlLastVoteResults = `
	SELECT TOP 1 op.[Title], COUNT([Key])
	FROM [Vote] AS v
JOIN [Option] AS op
	ON op.[Id] = v.[OptionId]
	GROUP BY op.[Title], op.[Id]
	ORDER BY op.[Id] DESC;`
)

func InsertOption(db *sql.DB, title string) (id int64, err error) {
	ctx := context.Background()
	stmt, err := db.PrepareContext(ctx, sqlOptionInsert)
	if err != nil {
		return -1, err
	}
	row := stmt.QueryRowContext(
		ctx,
		sql.Named("Title", title),
	)
	err = row.Scan(&id)
	if err != nil {
		return -1, err
	}
	return id, nil
}

func CastVote(db *sql.DB, key []byte, optionId int64) error {
	ctx := context.Background()
	stmt, err := db.PrepareContext(ctx, sqlVoteInsert)
	if err != nil {
		return err
	}
	_, err = stmt.ExecContext(
		ctx,
		sql.Named("Key", key),
		sql.Named("OptionId", optionId),
	)
	if err != nil {
		return err
	}
	return nil
}

func LastVote(db *sql.DB) (title string, votes int64, err error) {
	ctx := context.Background()
	stmt, err := db.PrepareContext(ctx, sqlLastVoteResults)
	if err != nil {
		return "", -1, err
	}
	err = stmt.QueryRowContext(ctx).Scan(&title, &votes)
	return title, votes, err
}
