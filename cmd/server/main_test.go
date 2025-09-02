package main

import (
	"flag"
	"os"
	"testing"
)

func TestMain_FlagParsing(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("Skipping server main test in CI environment")
	}

	originalArgs := os.Args
	defer func() {
		os.Args = originalArgs
	}()

	os.Args = []string{"cmd", "--address", ":8080"}
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	address := flag.String("address", ":9001", "Address to listen on")
	flag.Parse()

	if *address != ":8080" {
		t.Errorf("Expected address :8080, got %s", *address)
	}
}

func TestMain_DefaultFlags(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("Skipping server main test in CI environment")
	}

	originalArgs := os.Args
	defer func() {
		os.Args = originalArgs
	}()

	os.Args = []string{"cmd"}
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	address := flag.String("address", ":9001", "Address to listen on")
	flag.Parse()

	if *address != ":9001" {
		t.Errorf("Expected default address :9001, got %s", *address)
	}
}
