package main

import (
	"encoding/json"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: 0x <fixture.json>")
		os.Exit(1)
	}

	data, err := os.ReadFile(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading %s: %v\n", os.Args[1], err)
		os.Exit(1)
	}

	var fixture Fixture
	if err := json.Unmarshal(data, &fixture); err != nil {
		fmt.Fprintf(os.Stderr, "error parsing fixture: %v\n", err)
		os.Exit(1)
	}

	traces, err := ReplayTransaction(&fixture)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error replaying transaction: %v\n", err)
		os.Exit(1)
	}

	out, err := json.MarshalIndent(traces, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error marshaling output: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(out))
}
