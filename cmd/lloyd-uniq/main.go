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
	"runtime/pprof"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/miku/lloyd"
)

func main() {
	key := flag.String("key", "", "key to deduplicate on")
	cpuprofile := flag.String("cpuprofile", "", "write cpu profile to file")

	flag.Parse()

	if flag.NArg() < 1 {
		log.Fatal("input file required")
	}

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	file, err := os.Open(flag.Arg(0))
	if err != nil {
		log.Fatal(err)
	}
	reader := bufio.NewReader(file)

	rand.Seed(time.Now().UTC().UnixNano())
	dbpath := fmt.Sprintf("./.lloyd.tmp-%d", rand.Int63n(10000000000))
	db, err := sql.Open("sqlite3", dbpath)
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		db.Close()
		os.Remove(dbpath)
	}()

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
