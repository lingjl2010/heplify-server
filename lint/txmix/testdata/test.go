package testdata

import (
	"database/sql"
)

type DB struct {
	*sql.DB
}

func (db *DB) txWith(f func(tx *sql.Tx)) (*sql.Tx, error) {
	tx, err := db.Begin() // want `using db and tx in same scope, maybe bug`
	if err != nil {
		return nil, err
	}
	f(tx)
	return tx, tx.Commit()
}

func txWith(db *sql.DB, f func(tx *sql.Tx)) { // want `using db and tx in same scope, maybe bug`
	tx, _ := db.Begin() // want `using db and tx in same scope, maybe bug`
	f(tx)
	tx.Commit()
}

func connWith(db *sql.DB, f func(conn *sql.Conn)) { // want `using db and tx in same scope, maybe bug`
	conn, _ := db.Conn(nil) // want `using db and tx in same scope, maybe bug`
	f(conn)
	conn.Close()
}

func Test() {
	var db sql.DB
	db.Exec("select * from user")
	txWith(&db, func(tx *sql.Tx) {
		tx.Exec("select * from user2")
	})
	txWith(&db, func(tx *sql.Tx) {
		tx.Exec("select * from user3")
		db.Exec("select * from user4") // want `using db and tx in same scope, maybe bug`
	})
	connWith(&db, func(conn *sql.Conn) {
		conn.ExecContext(nil, "select * from user5")
	})
	connWith(&db, func(conn *sql.Conn) {
		conn.ExecContext(nil, "select * from user5")
		db.Exec("select * from user6") // want `using db and tx in same scope, maybe bug`
	})
}
