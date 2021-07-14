package main

/*
This is based off an article from here:
https://docs.microsoft.com/en-us/azure/azure-sql/database/connect-query-go

This is querying an MSSQL db with the following table schema:

CREATE TABLE [gluser].[Tasks](
	[id] [uniqueidentifier] NULL,
	[msg] [nvarchar](2048) NULL,
	[lastupdate] [datetime] NULL,
	[status] [nvarchar](50) NULL,
	[percentcomplete] [int] NULL
)


*/
import (
	"context"
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	mssql "github.com/denisenkom/go-mssqldb"
)

//Aliasing the go-mssqldb uniqueidentifer so I don't have to use the full namespace (uuid is more readable :D )
type uuid = mssql.UniqueIdentifier

var db *sql.DB

//Server=A2NWPLSK14SQL-v05.shr.prod.iad2.secureserver.net;Database=GoLangDB;User Id=gluser;Password=<<REDACTED>>;
var server = "A2NWPLSK14SQL-v05.shr.prod.iad2.secureserver.net"
var port = 1433
var user = "gluser"
var database = "GoLangDB"

func main() {
	var err error

	// This is good enough for this sort of throw away code--e.g. password in a file not checked in
	// as this DB instance is not long for this world
	// BUT need to check into how other teams/developers are doing for app config stuff
	// (would be nice if there is a lib that works somewhat like .net core--config json files and
	// the ability to override based off which environment you're in (DEV/TST/STG/PROD)

	password, err := ioutil.ReadFile("/home/mahldcat/pass")
	if err != nil {
		log.Fatal("Error fetching password file:", err.Error())
	}

	// Build connection string
	connString := fmt.Sprintf("server=%s;user id=%s;password=%s;port=%d;database=%s;",
		server, user, password, port, database)

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

	//the db I am querying has three rows in it
	/*
		7ff5f446-9afe-4b33-b9dc-e2c33f2452f4	working	2021-07-01 00:00:00.000	working	50
		999c5fee-715f-4466-b0a1-b7aaa965ba2e	some job	2021-07-01 00:00:00.000	working	60
		800ac31e-84fc-44c7-9368-e1dfbbcef995	some job	2021-07-01 00:00:00.000	working	65
	*/

	//This query should only pull 1 entry (the 65% completion)
	count, err := GetTasks(63)
	if err != nil {
		log.Fatal("Error Fetching Tasks: ", err.Error())
	}
	fmt.Printf("Read %d row(s).\n", count)

	db.Close() //lets close this......and then reopen

	//based off the semantics of this setup, I have to use the sql.Open call instead of being able to reopen from the underlying db object
	db, err = sql.Open("sqlserver", connString)
	if err != nil {
		log.Fatal("Error creating connection pool: ", err.Error())
	}

	//this one should fetch 2 rows....
	count, err = GetTasks(55)
	if err != nil {
		log.Fatal("Error Fetching Tasks: ", err.Error())
	}
	fmt.Printf("Read %d row(s).\n", count)

	//Lets play a bit more with the uuid
	gstr := "7ff5f446-9afe-4b33-b9dc-e2c33f2452f4"
	uuid, err := WorkWithUniqueIdentifierStruct(gstr)
	if err != nil {
		log.Fatal("Error parsing Guid: ", err.Error())
	}
	fmt.Printf("gstr: %s, uuid is %s\n", gstr, uuid)

	//throw in some casing
	gstr = "7FF5F446-9AFE-4B33-b9dc-e2c33f2452f4"
	uuid, err = WorkWithUniqueIdentifierStruct(gstr)
	if err != nil {
		log.Fatal("Error parsing Guid: ", err.Error())
	}
	//so based off what I'm seeing, this code's tostring puts this as all caps
	fmt.Printf("gstr: %s, uuid is %s\n", gstr, uuid)

	//<BUT> since this puts the string past the 36 character length
	//it's not exactly robust enough to handle curly braces
	gstr = "{7FF5F446-9AFE-4B33-b9dc-e2c33f2452f4}"
	uuid, err = WorkWithUniqueIdentifierStruct(gstr)
	if err != nil {
		log.Fatal("Error parsing Guid: ", err.Error())
	}
	fmt.Printf("gstr: %s, uuid is %s\n", gstr, uuid)

}

/*
Dumb function that I'm using to familiarize myself with
the UniqueIdentifier
*/
func WorkWithUniqueIdentifierStruct(toParse string) (uuid, error) {

	//I had a heckk of a time figuring this out
	/*
		tried things like var g uuid = new uuid()
		var g uuid= nil
		var g uuid(nil) etc


		Another discovery is go doesn't have traditional c'tors like java or C#?
		http://www.golangpatterns.info/object-oriented/constructors
	*/

	var g uuid
	err := g.Scan(toParse)
	return g, err
}

//Now I'm  curious---what is the difference between <this> function and the previous one
// beyond the fact that this looks exactly like the return is a pointer to the uuid
//
func WorkWithUniqueIdentifierStruct2(toParse string) *uuid {
	u := new(uuid)
	u.Scan(toParse)
	return u
}

//SO basic functionality is here for doing calls
// but I think I would tear what hair I have left out at the roots
// if this is the standard interface
//
// this is good for learning the basics (e.g. can hit MSSQL)
// but I would rather see if there might be any code generators out there
//   again .net Core [forn now] wins out with the entity framework
//   as I can point a tool at an existing db and with only a bit of pain have it generate
//       the OR mappings to handle Table CRUD as well as provide method to Sproc mappings so we can avoid having to
//       manually do it
func GetTasks(completion int) (int, error) {
	ctx := context.Background()

	// Check if database is alive.
	err := db.PingContext(ctx)
	if err != nil {
		return -1, err
	}

	//thank god this stuff has named param structures
	//
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

		//as a newbie I was not exactly sure how to reference this thing.....
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
