package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	mssql "github.com/denisenkom/go-mssqldb"
)

var db *sql.DB

//Server=A2NWPLSK14SQL-v05.shr.prod.iad2.secureserver.net;Database=GoLangDB;User Id=gluser;Password=<<REDACTED>>;
var server = "A2NWPLSK14SQL-v05.shr.prod.iad2.secureserver.net"
var port = 1433
var user = "gluser"
var password = "<<REDACTED>>"
var database = "GoLangDB"

func main() {
	// Build connection string
	connString := fmt.Sprintf("server=%s;user id=%s;password=%s;port=%d;database=%s;",
		server, user, password, port, database)

	var err error

	// Create connection pool
	db, err = sql.Open("sqlserver", connString)
	if err != nil {
		log.Fatal("Error creating connection pool: ", err.Error())
	}
	ctx := context.Background()
	err = db.PingContext(ctx)
	if err != nil {
		log.Fatal(err.Error())
	}
	fmt.Printf("Connected!\n")

	count, err := GetTasks(63)
	if err != nil {
		log.Fatal("Error Fetching Tasks: ", err.Error())
	}
	fmt.Printf("Read %d row(s).\n", count)

}

func GetTasks(completion int) (int, error) {
	ctx := context.Background()

	// Check if database is alive.
	err := db.PingContext(ctx)
	if err != nil {
		return -1, err
	}

	tsql := "SELECT * FROM Tasks where percentcomplete >@completion"

	fmt.Printf("Query: %s\n", tsql)

	// Execute query
	rows, err := db.QueryContext(
		ctx,
		tsql,
		sql.Named("completion", completion))

	if err != nil {
		return -1, err
	}

	defer rows.Close()

	var count int

	// Iterate through the result set.
	for rows.Next() {
		var id mssql.UniqueIdentifier
		var msg, status string
		var lastupdate time.Time
		var pctCmplt int
		// Get values from row.
		err := rows.Scan(&id, &msg, &lastupdate, &status, &pctCmplt)
		if err != nil {
			return -1, err
		}

		fmt.Printf("ID: %s, Message: %s, lastupdate:%s Status: %s PctComplete: %d \n", id, msg, lastupdate, status, pctCmplt)
		count++
	}

	return count, nil
}
