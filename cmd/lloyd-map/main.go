// lloyd-map extract a value per document and its offset and length.
// The result is written to Stdout, tab-separated, so it will lead
// to unexpected results, if the extracted values contain tabs.
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"strings"
	"sync"

	"github.com/miku/lloyd"
)

// Batch contains lines and the offset from which they originate.
type Batch struct {
	Keys       []string
	Lines      []string
	BaseOffset int64
}

// RecordInfo contains the value and the document offset and length within the file.
type RecordInfo struct {
	Values []string
	Offset int64
	Length int64
}

// worker turns batches into seek information about records
func worker(batches chan Batch, out chan RecordInfo, wg *sync.WaitGroup) {
	defer wg.Done()
	for batch := range batches {
		offset := batch.BaseOffset
		for _, line := range batch.Lines {
			dst := make(map[string]interface{})

			d := json.NewDecoder(strings.NewReader(line))
			d.UseNumber()
			if err := d.Decode(&dst); err != nil {
				log.Fatal(err)
			}

			var values []string
			for _, key := range batch.Keys {
				value, err := lloyd.StringValue(key, dst)
				if err != nil {
					log.Fatal(err)
				}
				values = append(values, value)
			}

			length := int64(len(line))
			ri := RecordInfo{Values: values, Offset: offset, Length: length}
			out <- ri
			offset += length
		}
	}
}

func sink(w io.Writer, out chan RecordInfo, done chan bool) {
	bw := bufio.NewWriter(w)
	for ri := range out {
		bw.WriteString(fmt.Sprintf("%s\t%d\t%d\n", strings.Join(ri.Values, "\t"), ri.Offset, ri.Length))
	}
	bw.Flush()
	done <- true
}

func parseList(s string) (items []string) {
	fields := strings.Split(s, ",")
	for _, f := range fields {
		items = append(items, strings.TrimSpace(f))
	}
	return
}

func main() {
	rawKeys := flag.String("keys", "", "key or keys to extract (top-level only)")
	workers := flag.Int("w", runtime.NumCPU(), "number of workers")
	size := flag.Int("size", 10000, "number of lines in one batch")
	version := flag.Bool("v", false, "prints current program version")

	flag.Parse()

	if *version {
		fmt.Println(lloyd.AppVersion)
		os.Exit(0)
	}

	runtime.GOMAXPROCS(*workers)

	if flag.NArg() < 1 {
		log.Fatal("input LDJ file required")
	}

	file, err := os.Open(flag.Arg(0))
	if err != nil {
		log.Fatal(err)
	}

	defer file.Close()
	reader := bufio.NewReader(file)

	var wg sync.WaitGroup
	batches := make(chan Batch)
	out := make(chan RecordInfo)
	done := make(chan bool)

	go sink(os.Stdout, out, done)

	for i := 0; i < *workers; i++ {
		wg.Add(1)
		go worker(batches, out, &wg)
	}

	var lines []string
	var baseOffset, offset int64

	keys := parseList(*rawKeys)

	for {
		s, err := reader.ReadString('\n')

		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		lines = append(lines, s)
		offset += int64(len(s))

		if len(lines)%*size == 0 {
			l := make([]string, len(lines))
			copy(l, lines)
			batches <- Batch{Lines: l, Keys: keys, BaseOffset: baseOffset}
			lines = lines[:0]
			baseOffset = offset
		}
	}

	l := make([]string, len(lines))
	copy(l, lines)
	batches <- Batch{Lines: l, Keys: keys, BaseOffset: baseOffset}
	lines = lines[:0]

	close(batches)
	wg.Wait()
	close(out)
	<-done
}
