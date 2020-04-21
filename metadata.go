package televise

import (
	"time"

	"github.com/gocql/gocql"
)

const (
	queryMetadataCreateTable = `
		CREATE TABLE metadata (
			name text PRIMARY KEY,
			value text,
			updated timestamp
		);`
	queryMetadataDropTable = `DROP TABLE metadata;`
	queryMetadataList      = `SELECT name, value, updated FROM metadata;`
	queryMetadataGet       = `SELECT value FROM metadata WHERE name = ?;`
	queryMetadataSet       = `UPDATE metadata SET value = ?, updated = now() WHERE name = ?;`
	queryMetadataDelete    = `DELETE FROM metadata WHERE name = ?;`
)

type MetadataValue struct {
	Value   string    `json:"value"`
	Updated time.Time `json:"updated"`
}

func MetadataDisplayList(db *gocql.Session) (list map[string]MetadataValue, err error) {
	it := db.Query(queryMetadataList).Iter()
	m := map[string]interface{}{}
	list = map[string]MetadataValue{}
	for it.MapScan(m) {
		v := MetadataValue{
			Value:   m["value"].(string),
			Updated: m["updated"].(time.Time),
		}
		list[m["name"].(string)] = v
		m = map[string]interface{}{}
	}
	return list, nil
}

func MetadataGet(db *gocql.Session, key string) (val string, err error) {
	err = db.Query(queryMetadataGet, key).Scan(&val)
	return val, err
}

func MetadataSet(db *gocql.Session, key string, value string) (err error) {
	return db.Query(queryMetadataSet, key, value).Exec()
}
