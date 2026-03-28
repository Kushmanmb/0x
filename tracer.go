package main

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
)

// callFrame captures the data for a single call within the call tree.
type callFrame struct {
	typ     vm.OpCode
	from    common.Address
	to      common.Address
	value   *big.Int
	gas     uint64
	gasUsed uint64
	input   []byte
	output  []byte
	err     error
	depth   int
	children []*callFrame
}

// callTracer implements vm.EVMLogger and collects a tree of call frames that
// can later be flattened into Parity-style traces.
type callTracer struct {
	callStack []*callFrame
	root      *callFrame
}

// CaptureStart is called once at the start of the top-level call.
func (t *callTracer) CaptureStart(env *vm.EVM, from common.Address, to common.Address, create bool, input []byte, gas uint64, value *big.Int) {
	typ := vm.CALL
	if create {
		typ = vm.CREATE
	}
	inputCopy := make([]byte, len(input))
	copy(inputCopy, input)

	var val *big.Int
	if value != nil {
		val = new(big.Int).Set(value)
	}

	t.root = &callFrame{
		typ:   typ,
		from:  from,
		to:    to,
		value: val,
		gas:   gas,
		input: inputCopy,
		depth: 0,
	}
	t.callStack = []*callFrame{t.root}
}

// CaptureEnd is called at the end of the top-level call.
func (t *callTracer) CaptureEnd(output []byte, gasUsed uint64, err error) {
	t.root.gasUsed = gasUsed
	t.root.output = make([]byte, len(output))
	copy(t.root.output, output)
	t.root.err = err
}

// CaptureEnter is called at the start of each sub-call (CALL, DELEGATECALL, etc.).
func (t *callTracer) CaptureEnter(typ vm.OpCode, from common.Address, to common.Address, input []byte, gas uint64, value *big.Int) {
	inputCopy := make([]byte, len(input))
	copy(inputCopy, input)

	var val *big.Int
	if value != nil {
		val = new(big.Int).Set(value)
	}

	depth := len(t.callStack)
	frame := &callFrame{
		typ:   typ,
		from:  from,
		to:    to,
		value: val,
		gas:   gas,
		input: inputCopy,
		depth: depth,
	}

	parent := t.callStack[len(t.callStack)-1]
	parent.children = append(parent.children, frame)
	t.callStack = append(t.callStack, frame)
}

// CaptureExit is called at the end of each sub-call.
func (t *callTracer) CaptureExit(output []byte, gasUsed uint64, err error) {
	frame := t.callStack[len(t.callStack)-1]
	frame.gasUsed = gasUsed
	frame.output = make([]byte, len(output))
	copy(frame.output, output)
	frame.err = err
	t.callStack = t.callStack[:len(t.callStack)-1]
}

// CaptureState is called for every EVM opcode; we don't need opcode-level tracing.
func (t *callTracer) CaptureState(pc uint64, op vm.OpCode, gas, cost uint64, scope *vm.ScopeContext, rData []byte, depth int, err error) {
}

// CaptureFault is called when an opcode faults.
func (t *callTracer) CaptureFault(pc uint64, op vm.OpCode, gas, cost uint64, scope *vm.ScopeContext, depth int, err error) {
}

// CaptureTxStart is called before transaction execution starts.
func (t *callTracer) CaptureTxStart(gasLimit uint64) {}

// CaptureTxEnd is called after transaction execution ends.
func (t *callTracer) CaptureTxEnd(restGas uint64) {}

// callTypeString converts an EVM opcode into the Parity callType string.
func callTypeString(op vm.OpCode) string {
	switch op {
	case vm.CALL:
		return "call"
	case vm.CALLCODE:
		return "callcode"
	case vm.DELEGATECALL:
		return "delegatecall"
	case vm.STATICCALL:
		return "staticcall"
	case vm.CREATE, vm.CREATE2:
		return "create"
	default:
		return "call"
	}
}
