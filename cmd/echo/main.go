package main

import (
	"fmt"
	"os"
)

const (
	BUFSIZE = 1024
)

func main() {
	buffer := make([]byte, BUFSIZE)

	for {
		// Read from STDIN
		bytesRead, err := os.Stdin.Read(buffer)
		if bytesRead == 0 {
			break
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading from STDIN: %s\n", err)
			os.Exit(1)
		}

		// Write to STDOUT
		_, err = os.Stdout.Write(buffer[:bytesRead])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing to STDOUT: %s\n", err)
			os.Exit(1)
		}
	}
}
