package main

import (
	"net"
	"github.com/skaverat/gabber/connection"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"log"
)


func main() {

	listener, _ := net.Listen("tcp", "0.0.0.0:5222")

	db, err := sql.Open("mysql", "gabber:gabber@tcp(localhost:3306)/gabber")

	if(err != nil) {
		log.Fatal("Could not connect to database!")
	}

	for {
		conn, _ := listener.Accept()

		connChan := make(chan net.Conn)
		dbChan := make(chan *sql.DB)
		go connection.Run(connChan, dbChan)
		connChan<-conn
		dbChan<-db
	}
}
