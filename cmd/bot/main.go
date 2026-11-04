package main

import (
	"fmt"
	"os"

	"github.com/swokeqq/tatusashbot/internal/config"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "fatal error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	_ = cfg

	return nil
}
