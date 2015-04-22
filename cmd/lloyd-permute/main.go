// lloyd-permute takes a list of offset, length pairs from Stdin
// and outputs the parts in the order as they are read.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
)

func main() {
	flag.Parse()
	if flag.NArg() < 1 {
		log.Fatal("input file required")
	}

	file, err := os.Open(flag.Arg(0))
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	reader := bufio.NewReader(os.Stdin)

	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		fields := strings.Fields(line)
		if len(fields) != 2 {
			log.Fatal("two columns required")
		}
		offset, err := strconv.Atoi(fields[0])
		if err != nil {
			log.Fatalf("invalid offset: %s", fields[0])
		}
		length, err := strconv.Atoi(fields[1])
		if err != nil {
			log.Fatalf("invalid length: %s", fields[1])
		}
		buf := make([]byte, length)
		file.Seek(int64(offset), 0)
		_, err = file.Read(buf)
		if err != io.EOF && err != nil {
			log.Fatal(err)
		}
		fmt.Printf(string(buf))
	}
}
