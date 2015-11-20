package main

import (
	"fmt"
	"io"
	"os"

	"github.com/gholt/store"
	"gopkg.in/gholt/brimtime.v1"
)

func main() {
	if len(os.Args) != 4 {
		fmt.Printf("Syntax: %s <value|group> file filetoc\n", os.Args[0])
		os.Exit(1)
	}
	var errs []error
	switch os.Args[1] {
	case "value":
		errs = valueAudit(os.Args[2], os.Args[3])
	case "group":
		errs = groupAudit(os.Args[2], os.Args[3])
	default:
		fmt.Printf("Syntax: %s <value|group> file filetoc\n", os.Args[0])
		os.Exit(1)
	}
	if len(errs) != 0 {
		for _, err := range errs {
			fmt.Println(err)
		}
		os.Exit(1)
	}
}

func valueAudit(path string, pathtoc string) []error {
	var errs []error
	df := store.NewValueDirectFile(path, pathtoc, openReadSeeker, openWriteSeeker)
	if ok, verrs := df.VerifyHeadersAndTrailers(); !ok {
		return append(errs, verrs...)
	} else if len(verrs) > 0 {
		errs = append(errs, verrs...)
	}
	dataSize, err := df.DataSize()
	if err != nil {
		return append(errs, err)
	}
	fmt.Printf("Data: %d bytes\n", dataSize)

	entryCount, err := df.EntryCount()
	if err != nil {
		return append(errs, err)
	}
	fmt.Printf("Entry Count: %d\n", entryCount)

	_, _, timestamp, _, length, err := df.FirstEntry()
	if err != nil {
		errs = append(errs, err)
	}
	oldest := timestamp
	newest := timestamp
	smallest := length
	biggest := length
	for i := int64(1); i < entryCount; i++ {
		if timestamp < oldest {
			oldest = timestamp
		}
		if timestamp > newest {
			newest = timestamp
		}
		if length < smallest {
			smallest = length
		}
		if length > biggest {
			biggest = length
		}
		_, _, timestamp, _, length, err = df.NextEntry()
		if err != nil {
			errs = append(errs, err)
		}
	}
	fmt.Println(brimtime.UnixMicroToTime(int64(oldest>>8)), "to", brimtime.UnixMicroToTime(int64(newest>>8)))
	fmt.Println(smallest, "to", biggest, "bytes")
	return errs
}

func groupAudit(path string, pathtoc string) []error {
	return nil
}

func openReadSeeker(name string) (io.ReadSeeker, error) {
	return os.Open(name)
}

func openWriteSeeker(name string) (io.WriteSeeker, error) {
	return os.OpenFile(name, os.O_RDWR, 0666)
}
