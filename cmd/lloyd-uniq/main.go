// 1. Read n lines
// 2. Distribute batches to workers
// 3. Workers parse JSON and return {syno: int, key: string, payload: string}
// ...

package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"io"
	"log"
	"os"
	"strings"

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

	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			log.Fatal(err)
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
	}
}
