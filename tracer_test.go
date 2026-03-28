package main

import (
	"encoding/json"
	"os"
	"testing"
)

func TestReplayTransaction(t *testing.T) {
	data, err := os.ReadFile("testdata/trace_call.json")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	var fixture Fixture
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatalf("parse fixture: %v", err)
	}

	traces, err := ReplayTransaction(&fixture)
	if err != nil {
		t.Fatalf("ReplayTransaction: %v", err)
	}

	if len(traces) == 0 {
		t.Fatal("expected at least one trace, got none")
	}

	// Verify the number of traces matches the fixture expectation.
	if got, want := len(traces), len(fixture.Result); got != want {
		t.Errorf("trace count: got %d, want %d", got, want)
	}

	// Validate structural properties of each trace.
	for i, tr := range traces {
		if tr.Type != "call" {
			t.Errorf("trace[%d].Type = %q, want \"call\"", i, tr.Type)
		}
		if tr.Action.From == "" {
			t.Errorf("trace[%d].Action.From is empty", i)
		}
		if tr.Action.To == "" {
			t.Errorf("trace[%d].Action.To is empty", i)
		}
		if tr.Action.CallType == "" {
			t.Errorf("trace[%d].Action.CallType is empty", i)
		}
	}

	// Validate that the top-level trace has the expected addresses.
	expected := fixture.Result[0]
	got := traces[0]
	if got.Action.From != expected.Action.From {
		t.Errorf("trace[0].from: got %q, want %q", got.Action.From, expected.Action.From)
	}
	if got.Action.To != expected.Action.To {
		t.Errorf("trace[0].to: got %q, want %q", got.Action.To, expected.Action.To)
	}
	if got.Action.Value != expected.Action.Value {
		t.Errorf("trace[0].value: got %q, want %q", got.Action.Value, expected.Action.Value)
	}
	if got.Subtraces != expected.Subtraces {
		t.Errorf("trace[0].subtraces: got %d, want %d", got.Subtraces, expected.Subtraces)
	}

	// Validate that the sub-trace has the expected addresses.
	if len(traces) > 1 {
		expected1 := fixture.Result[1]
		got1 := traces[1]
		if got1.Action.From != expected1.Action.From {
			t.Errorf("trace[1].from: got %q, want %q", got1.Action.From, expected1.Action.From)
		}
		if got1.Action.To != expected1.Action.To {
			t.Errorf("trace[1].to: got %q, want %q", got1.Action.To, expected1.Action.To)
		}
		if len(got1.TraceAddress) != 1 || got1.TraceAddress[0] != 0 {
			t.Errorf("trace[1].traceAddress: got %v, want [0]", got1.TraceAddress)
		}
	}
}
