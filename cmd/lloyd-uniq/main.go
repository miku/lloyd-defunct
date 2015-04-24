package main

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"runtime/pprof"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"github.com/miku/lloyd"
)

func main() {
	key := flag.String("key", "", "key to deduplicate on")
	version := flag.Bool("v", false, "prints current program version")
	verbose := flag.Bool("verbose", false, "print debug information")
	cpuprofile := flag.String("cpuprofile", "", "write cpu profile to file")
	size := flag.Int("size", 200, "group insert size")
	commit := flag.Int("commit", 500000, "tx commit size")

	flag.Parse()

	if *version {
		fmt.Println(lloyd.AppVersion)
		os.Exit(0)
	}

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

	tmpfile, err := ioutil.TempFile("", "lloyd-")
	if err != nil {

	}

	if *verbose {
		log.Println(tmpfile.Name())
	}

	db, err := sql.Open("sqlite3", tmpfile.Name())
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		db.Close()
		os.Remove(tmpfile.Name())
		if *verbose {
			log.Println("cleanup done")
		}
	}()

	init := `CREATE TABLE IF NOT EXISTS store (key text UNIQUE, value text); CREATE INDEX store_key_idx ON store(key);`
	_, err = db.Exec(init)
	if err != nil {
		log.Fatalf("%q: %s\n", err, init)
	}

	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(`PRAGMA synchronous = OFF; PRAGMA journal_mode = OFF;`)
	if err != nil {
		log.Fatalf("%q: %s\n", err, init)
	}

	var vs []string
	for i := 0; i < *size; i++ {
		vs = append(vs, "(?, ?)")
	}

	stmt, err := tx.Prepare(fmt.Sprintf("INSERT OR REPLACE INTO store (key, value) VALUES %s", strings.Join(vs, ",")))
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	var i int
	var valueStrings []string
	var valueArgs []interface{}

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

		valueStrings = append(valueStrings, "(?, ?)")
		valueArgs = append(valueArgs, value)
		valueArgs = append(valueArgs, strings.TrimSpace(line))

		i++

		if i%*size == 0 {
			_, err = stmt.Exec(valueArgs...)
			if err != nil {
				log.Fatal(err)
			}
			valueStrings = valueStrings[:0]
			valueArgs = valueArgs[:0]
		}

		if i%*commit == 0 {
			err = tx.Commit()
			if err != nil {
				log.Fatal(err)
			}
			if *verbose {
				log.Printf("commit @%d", i)
			}
			tx, err = db.Begin()
			if err != nil {
				log.Fatal(err)
			}
			stmt, err = tx.Prepare(fmt.Sprintf("INSERT OR REPLACE INTO store (key, value) VALUES %s", strings.Join(vs, ",")))
			if err != nil {
				log.Fatal(err)
			}
		}
	}

	if len(valueArgs) > 0 {
		remaining := len(valueArgs) / 2

		vs = vs[:0]
		for j := 0; j < remaining; j++ {
			vs = append(vs, "(?, ?)")
		}

		stmt, err = tx.Prepare(fmt.Sprintf("INSERT OR REPLACE INTO store (key, value) VALUES %s", strings.Join(vs, ",")))
		if err != nil {
			log.Fatal(err)
		}
		defer stmt.Close()

		_, err = stmt.Exec(valueArgs...)
		if err != nil {
			log.Fatal(err)
		}

	}

	if *verbose {
		log.Println("committing...")
	}

	tx.Commit()

	if *verbose {
		log.Println("writing out...")
	}

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
