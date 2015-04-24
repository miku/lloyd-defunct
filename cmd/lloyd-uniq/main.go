// 1. Read n lines
// 2. Distribute batches to workers
// 3. Workers parse JSON and return {syno: int, key: string, payload: string}
// ...

package main

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/miku/lloyd"
)

func main() {
	key := flag.String("key", "", "key to deduplicate on")

	flag.Parse()

	if flag.NArg() < 1 {
		log.Fatal("input file required")
	}

	file, err := os.Open(flag.Arg(0))
	if err != nil {
		log.Fatal(err)
	}
	reader := bufio.NewReader(file)

	rand.Seed(time.Now().UTC().UnixNano())
	db, err := sql.Open("sqlite3", fmt.Sprintf("./.lloyd.tmp-%d", rand.Int63n(10000000000)))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	init := `CREATE TABLE IF NOT EXISTS store (key text UNIQUE, value text)`
	_, err = db.Exec(init)
	if err != nil {
		log.Fatalf("%q: %s\n", err, init)
	}

	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	stmt, err := tx.Prepare("INSERT OR REPLACE INTO store VALUES (?, ?)")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		dst := make(map[string]interface{})
		d := json.NewDecoder(strings.NewReader(line))
		d.UseNumber()
		if err := d.Decode(&dst); err != nil {
			log.Fatal(err)
		}

		value, err := lloyd.StringValue(*key, dst)
		if err != nil {
			log.Fatal(err)
		}

		_, err = stmt.Exec(value, strings.TrimSpace(line))
		if err != nil {
			log.Fatal(err)
		}
	}

	tx.Commit()

	// query
	rows, err := db.Query("SELECT value FROM store")
	if err != nil {
		log.Fatal(err)
	}
	
	for rows.Next() {
		var value string
		err = rows.Scan(&value)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(value)
	}
}
