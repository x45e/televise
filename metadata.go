package televise

import (
	"context"
	"database/sql"
	"time"
)

var displayKeys = []string{
	"movie",
	"movie_id",
}

const (
	sqlMetadataTable = `
CREATE TABLE [Metadata] (
	[Key] VARCHAR(255) PRIMARY KEY NOT NULL,
	[Value] VARCHAR(4096),
	[Created] DATETIME NOT NULL DEFAULT GETDATE(),
	[Updated] DATETIME NOT NULL DEFAULT GETDATE(),
	[Display] BIT DEFAULT 'FALSE'
);`
	sqlMetadataDropTable = `DROP TABLE [Metadata];`

	sqlMetadataGet = `SELECT [Value] FROM [Metadata] WHERE [Key] = @Key;`

	sqlMetadataUpsert = `
IF NOT EXISTS (SELECT 1 FROM [Metadata] WHERE [Key] = @Key)
	INSERT INTO [Metadata] ([Key], [Value], [Display]) VALUES (@Key, @Value, @Display)
ELSE
	UPDATE [Metadata] SET [Value] = @Value, [Updated] = GETDATE() WHERE [Key] = @Key`
	sqlMetadataDisplayValues = `SELECT [Key], [Value], [Updated] FROM [Metadata] WHERE [Display] = 1;`
)

type MetadataValue struct {
	Value   *string   `json:"value"`
	Updated time.Time `json:"updated"`
}

func MetadataDisplayList(db *sql.DB) (m map[string]MetadataValue, err error) {
	ctx := context.Background()
	stmt, err := db.PrepareContext(ctx, sqlMetadataDisplayValues)
	if err != nil {
		return nil, err
	}
	rows, err := stmt.QueryContext(ctx)
	if err != nil {
		return nil, err
	}
	m = make(map[string]MetadataValue)
	for rows.Next() {
		var key string
		var val MetadataValue
		err = rows.Scan(&key, &val.Value, &val.Updated)
		if err != nil {
			return nil, err
		}
		m[key] = val
	}
	return m, nil
}

func MetadataGet(db *sql.DB, key string) (val *string, err error) {
	ctx := context.Background()
	stmt, err := db.PrepareContext(ctx, sqlMetadataGet)
	if err != nil {
		return nil, err
	}
	err = stmt.QueryRowContext(ctx, sql.Named("Key", key)).Scan(&val)
	if err != nil {
		return nil, err
	}
	return val, nil
}

func MetadataSet(db *sql.DB, key string, value *string) (err error) {
	display := false
	for _, k := range displayKeys {
		if k == key {
			display = true
			break
		}
	}
	ctx := context.Background()
	stmt, err := db.PrepareContext(ctx, sqlMetadataUpsert)
	if err != nil {
		return err
	}
	_, err = stmt.ExecContext(
		ctx,
		sql.Named("Key", key),
		sql.Named("Value", value),
		sql.Named("Display", display),
	)
	if err != nil {
		return err
	}
	return nil
}
