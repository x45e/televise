package televise

import (
	"context"
	"crypto/sha1"
	"database/sql"
	"errors"
	"net"
	"net/http"
	"time"
)

const (
	sqlCreateSessionTable = `
CREATE TABLE [Session]
(
	[Key] BINARY(20) NOT NULL,
	[Addr] VARCHAR(39) NOT NULL,
	[UserAgent] VARCHAR(1024),
	[Start] DATETIME NOT NULL DEFAULT GETDATE(),
	[LastSeen] DATETIME NOT NULL DEFAULT GETDATE()
);
ALTER TABLE [Session] ADD CONSTRAINT [PK_Session] PRIMARY KEY ([Key], [Start]);`
	sqlDropSessionTable = `DROP TABLE [Session];`

	sqlUpdateSession = `
IF NOT EXISTS (SELECT 1 FROM [Session] WHERE [Key] = @Key AND DATEDIFF(s, [LastSeen], GETDATE()) < @InactiveLimit)
	INSERT INTO [Session] ([Key], [Addr], [UserAgent])
		VALUES (@Key, @Addr, @UserAgent);
ELSE
	UPDATE TOP (1) [Session]
		SET [LastSeen] = GETDATE()
	WHERE [Key] IN
		(SELECT TOP (1) [Key] FROM [Session] WHERE [Key] = @Key ORDER BY [LastSeen] DESC);`

	sqlFetchSession = `
SELECT TOP (1) [Key], [Start], [LastSeen] FROM [Session]
WHERE [Key] = @Key ORDER BY [LastSeen] DESC;`

	InactiveSessionLimit = 25 * time.Second
)

type Identity struct {
	Key       []byte `json:"key,string"`
	Addr      string `json:"-"`
	UserAgent string `json:"-"`
}

func NewIdentity(r *http.Request) *Identity {
	addr := r.Header.Get("X-Forwarded-For")
	if addr == "" {
		addr = r.RemoteAddr
	}
	// ignore port
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		host = addr
	}
	h := sha1.New()
	h.Write([]byte(addr))
	h.Write([]byte(r.UserAgent()))
	return &Identity{
		Key:       h.Sum(nil),
		Addr:      host,
		UserAgent: r.UserAgent(),
	}
}

type Session struct {
	Identity
	Start    time.Time `json:"start"`
	LastSeen time.Time `json:"lastSeen"`
}

func CreateOrUpdateSession(db *sql.DB, id *Identity) (*Session, error) {
	if db == nil {
		return nil, errors.New("db nil")
	}
	if id == nil || id.Key == nil {
		return nil, errors.New("identity invalid")
	}
	ctx := context.Background()
	stmt, err := db.PrepareContext(ctx, sqlUpdateSession)
	if err != nil {
		return nil, err
	}
	_, err = stmt.ExecContext(
		ctx,
		sql.Named("Key", id.Key),
		sql.Named("Addr", id.Addr),
		sql.Named("UserAgent", id.UserAgent),
		sql.Named("InactiveLimit", InactiveSessionLimit/time.Second),
	)
	if err != nil {
		return nil, err
	}
	stmt, err = db.PrepareContext(ctx, sqlFetchSession)
	if err != nil {
		return nil, err
	}
	row := stmt.QueryRowContext(
		ctx,
		sql.Named("Key", id.Key),
	)
	s := &Session{}
	err = row.Scan(&s.Key, &s.Start, &s.LastSeen)
	if err != nil {
		return nil, err
	}
	return s, nil
}

// SessionCount returns the current number of active connections.
func SessionCount(db *sql.DB) (n int64, err error) {
	if db == nil {
		return -1, errors.New("db nil")
	}
	ctx := context.Background()
	tsql := `SELECT COUNT([Key]) FROM [Session] WHERE DATEDIFF(s, [LastSeen], GETDATE()) < @InactiveLimit;`
	stmt, err := db.PrepareContext(ctx, tsql)
	if err != nil {
		return -1, err
	}
	row := stmt.QueryRowContext(
		ctx,
		sql.Named("InactiveLimit", InactiveSessionLimit/time.Second),
	)
	err = row.Scan(&n)
	if err != nil {
		return -1, err
	}
	return n, nil
}
