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
	"sync"
)

// Batch contains lines and the offset from which they originate.
type Batch struct {
	Key        string
	Lines      []string
	BaseOffset int64
}

// RecordInfo contains the value and the document offset and length within the file.
type RecordInfo struct {
	Value  interface{}
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
			json.Unmarshal([]byte(line), &dst)
			value, ok := dst[batch.Key]
			if !ok {
				log.Fatal("missing key")
			}
			length := int64(len(line))
			ri := RecordInfo{Value: value, Offset: offset, Length: length}
			out <- ri
			offset += length
		}
	}
}

func sink(w io.Writer, out chan RecordInfo, done chan bool) {
	bw := bufio.NewWriter(w)
	defer bw.Flush()
	for ri := range out {
		bw.WriteString(fmt.Sprintf("%s\t%d\t%d\n", ri.Value, ri.Offset, ri.Length))
	}
	done <- true
}

func main() {
	key := flag.String("key", "", "key to extract (top-level only)")
	workers := flag.Int("w", runtime.NumCPU(), "number of workers")
	size := flag.Int("size", 10000, "number of lines in one batch")

	flag.Parse()

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
			batches <- Batch{Lines: l, Key: *key, BaseOffset: baseOffset}
			lines = lines[:0]
			baseOffset = offset
		}
	}
	l := make([]string, len(lines))
	copy(l, lines)
	batches <- Batch{Lines: l, Key: *key, BaseOffset: offset}
	lines = lines[:0]

	close(batches)
	wg.Wait()
	close(out)
	<-done
}
