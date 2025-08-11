package sql

import (
	"database/sql"
	"log"

	_ "modernc.org/sqlite"
)

var Conn = sqlite()

func sqlite() *sql.DB {
	dsn := "file:database.db" +
		"?_pragma=foreign_keys=ON" +
		"&_pragma=journal_mode=WAL" +
		"&_pragma=synchronous=OFF" +
		"&_pragma=cache_size=100000" +
		"&_pragma=busy_timeout=15000" +
		"&_pragma=temp_store=MEMORY" +
		"&_pragma=mmap_size=30000000000" +
		"&cache=shared"

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		log.Fatal(err)
	}

	db.SetMaxOpenConns(50)
	db.SetMaxIdleConns(25)

	return db
}
