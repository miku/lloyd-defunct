// lloyd-permute takes a list of offset, length pairs from Stdin
// and outputs the parts in the order as they are read.
//
// w/ mmap
//
// * read lines from (offset, length) until a fixed offset (e.g. 100M, 1G)
// * mmap region for reading (R)
// * mmap a file for writing (W)
// * copy over slices from R to W
// * flush W and append to previously written parts
// * repeat until done

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

	"github.com/edsrzf/mmap-go"
	"github.com/miku/lloyd"
)

var Pagesize = int64(os.Getpagesize())

type SeekInfo struct {
	Offset int64
	Length int64
}

// lower returns the closest number that is <= val AND is a multiple of size
func lower(val, size int64) int64 {
	return val + (size - val%size) - size
}

func process(sis []SeekInfo, filename string, w io.Writer) {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		err := file.Close()
		if err != nil {
			log.Fatal(err)
		}
	}()

	if len(sis) == 0 {
		return
	}

	first := sis[0]
	last := sis[len(sis)-1]

	regionOffset := lower(first.Offset, Pagesize)
	shift := first.Offset - regionOffset
	regionLength := int(last.Offset+last.Length-first.Offset) + int(shift)

	mm, err := mmap.MapRegion(file, regionLength, mmap.RDONLY, 0, regionOffset)
	defer func() {
		if err := mm.Unmap(); err != nil {
			log.Fatal(err)
		}
	}()

	if err != nil {
		log.Fatal(err)
	}

	globalOffset := first.Offset

	bw := bufio.NewWriter(w)
	defer bw.Flush()

	for _, si := range sis {
		a := si.Offset - globalOffset + shift
		b := a + si.Length
		bw.WriteString(string(mm[a:b]))
	}
}

func main() {
	version := flag.Bool("v", false, "prints current program version")
	size := flag.Int("size", 32, "approximate size of mmap'd pieces in mb")

	flag.Parse()

	if *version {
		fmt.Println(lloyd.AppVersion)
		os.Exit(0)
	}

	if flag.NArg() < 1 {
		log.Fatal("input file required")
	}

	filename := flag.Arg(0)
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	reader := bufio.NewReader(os.Stdin)

	var slices []SeekInfo
	currentWindow := *size * 1048576

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
			log.Fatal("two columns required (offset, length)")
		}
		offset, err := strconv.Atoi(fields[0])
		if err != nil {
			log.Fatalf("invalid offset: %s", fields[0])
		}
		length, err := strconv.Atoi(fields[1])
		if err != nil {
			log.Fatalf("invalid length: %s", fields[1])
		}

		si := SeekInfo{Offset: int64(offset), Length: int64(length)}
		slices = append(slices, si)

		if offset > currentWindow {
			currentWindow = currentWindow + *size*1048576
			dst := make([]SeekInfo, len(slices))
			copy(dst, slices)
			process(dst, filename, os.Stdout)
			slices = slices[:0]
		}
	}

	dst := make([]SeekInfo, len(slices))
	copy(dst, slices)
	process(dst, filename, os.Stdout)
}
