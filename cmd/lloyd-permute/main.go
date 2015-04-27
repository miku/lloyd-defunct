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

type SeekInfo struct {
	Offset int64
	Length int64
}

func lower(val, size int64) int64 {
	return val + (size - val%size) - size
}

func process(sis []SeekInfo, filename string) {

	pagesize := int64(os.Getpagesize())

	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	if len(sis) == 0 {
		return
	}

	first := sis[0]
	last := sis[len(sis)-1]
	globalDiff := first.Offset

	log.Println(len(sis), "items", first.Offset, last.Offset+last.Length, globalDiff)

	// region must start at some multiple of pagesize
	regionOffset := lower(first.Offset, pagesize)

	// shift is the amount you need to add to every seekinfo in order to be useful
	shift := regionOffset - first.Offset

	regionLength := int(first.Offset-regionOffset+last.Offset+last.Length) - int(globalDiff)
	log.Println("mmap region offset/length", regionOffset, regionLength)
	mm, err := mmap.MapRegion(file, regionLength, mmap.RDONLY, 0, regionOffset)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(string(mm[(first.Offset-globalDiff)-shift : first.Length-shift]))
	if err := mm.Unmap(); err != nil {
		log.Fatal(err)
	}
}

func main() {
	version := flag.Bool("v", false, "prints current program version")

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
	windowSize := 10000
	currentWindow := windowSize

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

		si := SeekInfo{Offset: int64(offset), Length: int64(length)}
		if offset > currentWindow {
			// hand over to mmap
			currentWindow = currentWindow + windowSize
			dst := make([]SeekInfo, len(slices))
			copy(dst, slices)
			process(dst, filename)
			slices = slices[:0]
		}
		slices = append(slices, si)

		// buf := make([]byte, length)
		// file.Seek(int64(offset), 0)
		// _, err = file.Read(buf)
		// if err != io.EOF && err != nil {
		// 	log.Fatal(err)
		// }
		// fmt.Printf(string(buf))
	}
}
