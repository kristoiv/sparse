package main

import (
	"fmt"
	"io"
	"os"

	"github.com/kristoiv/sparse"
)

func main() {
	if len(os.Args) < 3 {
		usage()
	}

	in, err := os.Open(os.Args[1])
	if err != nil {
		fmt.Printf("Unable to open input sparse file %q with error: %q\n", os.Args[1], err)
		os.Exit(1)
	}
	defer in.Close()

	for _, file := range os.Args[2:] {
		_, err = in.Seek(0, 0) // we seek to the start of the file on each round
		if err != nil {
			fmt.Printf("Unable to seek to the start of our file %q with error: %q\n", os.Args[1], err)
			os.Exit(1)
		}

		out, err := os.Create(file)
		if err != nil {
			fmt.Printf("Unable to open output file %q with error: %q\n", file, err)
			os.Exit(1)
		}

		fmt.Printf("Writing file %q\n", file)
		decodedIn, err := sparse.Simg2imgReader(in)
		if err != nil {
			fmt.Printf("Unable parse sparse header for file %q with error: %q\n", os.Args[1], err)
			os.Exit(1)
		}
		if _, err = io.Copy(out, decodedIn); err != nil {
			fmt.Printf("Error decoding input file to output file %q with error: %q\n", file, err)
			os.Exit(1)
		}

		out.Close()
	}

	fmt.Println("All complete.")
}

func usage() {
	fmt.Printf("\nUsage: %s <sparse file> <raw files...>\n\n", os.Args[0])
	os.Exit(1)
}
