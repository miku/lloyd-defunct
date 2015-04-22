// lloyd-uniq is like the standard unix tool uniq, but for line-delimited JSON files.
// The last occurence of a given keys' value will be kept.
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
)

func main() {
	key := flag.String("key", "", "value of this key is used to deduplicate")

	flag.Parse()

	var reader *bufio.Reader
	if flag.NArg() < 1 {
		reader = bufio.NewReader(os.Stdin)
	} else {
		file, err := os.Open(flag.Arg(0))
		if err != nil {
			log.Fatal(err)
		}
		reader = bufio.NewReader(file)
	}

	var lastLine string
	var lastValue interface{}
	var i int

	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		dst := make(map[string]interface{})
		json.Unmarshal([]byte(line), &dst)
		value, ok := dst[*key]
		if !ok {
			log.Fatal("missing key")
		}
		if i == 0 {
			lastLine = line
			lastValue = value
			i++
			continue
		}

		if lastValue != value {
			fmt.Printf(lastLine)
		}

		lastLine = line
		lastValue = value
		i++
	}
}
