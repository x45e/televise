package televise

import (
	"github.com/gocql/gocql"
)

const (
	queryOptionCreateTable = `
		CREATE TABLE option (
			id bigint PRIMARY KEY,
			title text
		);`
	queryOptionDropTable = `DROP TABLE option;`
	queryOptionList      = `SELECT id, title FROM option;`
	queryOptionInsert    = `INSERT INTO option (id, title) VALUES (?, ?);`

	queryVoteCreateTable = `
		CREATE TABLE vote (
			id bigint,
			option_id bigint,
			at timeuuid,
			PRIMARY KEY (id, option_id)
		);`
	queryVoteDropTable = `DROP TABLE vote;`
	queryVoteInsert    = `INSERT INTO vote (id, option_id, at) VALUES (?, ?, now());`
)

func ListOptions(db *gocql.Session) (list map[Snowflake]string, err error) {
	it := db.Query(queryOptionList).Iter()
	m := map[string]interface{}{}
	list = map[Snowflake]string{}
	for it.MapScan(m) {
		list[m["id"].(Snowflake)] = m["title"].(string)
		m = map[string]interface{}{}
	}
	return list, nil
}

func InsertOption(sf *snowflaker, db *gocql.Session, title string) (id Snowflake, err error) {
	id = sf.next()
	err = db.Query(queryOptionInsert, id, title).Exec()
	if err != nil {
		return NilSnowflake, err
	}
	return id, nil
}

func CastVote(db *gocql.Session, id *Identity, optionId Snowflake) error {
	err := db.Query(queryVoteInsert, id.ID, optionId).Exec()
	if err != nil {
		return err
	}
	return nil
}
