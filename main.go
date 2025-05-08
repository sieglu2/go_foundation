package main

import (
	"log"
	"os"
)

func main() {
	logPath := "./xyz"
	if err := os.RemoveAll(logPath); err != nil {
		log.Fatalf("failed to remove WAL directory: %w", err)
	}
}
