package televise

import (
	"errors"
	"net"
	"net/http"

	"github.com/gocql/gocql"
)

const (
	querySessionCreateTable = `
		CREATE TABLE session (
			id bigint PRIMARY KEY,
			addr text,
			user_agent text
		);`
	querySessionDropTable = `DROP TABLE session;`
	querySessionInsert    = `INSERT INTO session (id, addr, user_agent) VALUES (?, ?, ?)`

	queryVisitCreateTable = `
		CREATE TABLE visit (
			id bigint,
			ts timeuuid,
			PRIMARY KEY(id, ts)
		) WITH CLUSTERING ORDER BY (ts DESC);`
	queryVisitDropTable = `DROP TABLE visit;`
	queryVisitInsert    = `INSERT INTO visit (id, ts) VALUES (?, now());`
	queryVisitCount     = `SELECT COUNT(id) FROM visit WHERE ts > minTimeuuid(currentTime() - 25s)`
)

type Identity struct {
	ID        Snowflake `json:"id,string"`
	Addr      string    `json:"-"`
	UserAgent string    `json:"-"`
}

func (n *Identity) fromRequest(r *http.Request) {
	n.Addr = r.Header.Get("X-Forwarded-For")
	if n.Addr == "" {
		n.Addr = r.RemoteAddr
	}
	addr, _, err := net.SplitHostPort(n.Addr)
	if err == nil {
		n.Addr = addr
	}
	n.UserAgent = r.UserAgent()
}

func NewIdentity(sf *snowflaker, r *http.Request) *Identity {
	n := &Identity{}
	n.fromRequest(r)
	n.ID = sf.next()
	return n
}

func FindIdentity(r *http.Request) *Identity {
	id := ParseSnowflake(r.URL.Query().Get("id"))
	if id == NilSnowflake {
		return nil
	}
	n := &Identity{
		ID: id,
	}
	n.fromRequest(r)
	return n
}

func CreateSession(db *gocql.Session, id *Identity) error {
	if db == nil {
		return errors.New("db nil")
	}
	if id == nil || id.ID == NilSnowflake {
		return errors.New("identity invalid")
	}
	return db.Query(querySessionInsert, id.ID, id.Addr, id.UserAgent).Exec()
}

func LogVisit(db *gocql.Session, id *Identity) error {
	if db == nil {
		return errors.New("db nil")
	}
	if id == nil || id.ID == NilSnowflake {
		return errors.New("identity invalid")
	}
	return db.Query(queryVisitInsert, id.ID).Exec()
}

// VisitorCount returns the current number of active visitors.
func VisitorCount(db *gocql.Session) (n int64, err error) {
	if db == nil {
		return -1, errors.New("db nil")
	}
	err = db.Query(queryVisitCount).Scan(&n)
	return n, err
}
