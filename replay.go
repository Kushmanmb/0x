package main

import (
"encoding/hex"
"fmt"
"math/big"
"strconv"
"strings"

"github.com/ethereum/go-ethereum/common"
"github.com/ethereum/go-ethereum/common/hexutil"
"github.com/ethereum/go-ethereum/consensus"
"github.com/ethereum/go-ethereum/consensus/ethash"
"github.com/ethereum/go-ethereum/core"
"github.com/ethereum/go-ethereum/core/rawdb"
"github.com/ethereum/go-ethereum/core/state"
"github.com/ethereum/go-ethereum/core/types"
"github.com/ethereum/go-ethereum/core/vm"
"github.com/ethereum/go-ethereum/params"
"github.com/ethereum/go-ethereum/rlp"
"github.com/holiman/uint256"
)

// ReplayTransaction executes the transaction described in the fixture against
// the provided genesis state and returns Parity-style call traces.
func ReplayTransaction(fixture *Fixture) ([]ParityTrace, error) {
chainCfg, err := buildChainConfig(&fixture.Genesis.Config)
if err != nil {
return nil, fmt.Errorf("chain config: %w", err)
}

stateDB, err := buildGenesisState(&fixture.Genesis)
if err != nil {
return nil, fmt.Errorf("genesis state: %w", err)
}

tx, err := decodeTransaction(fixture.Input)
if err != nil {
return nil, fmt.Errorf("decode tx: %w", err)
}

blockCtx, err := buildBlockContext(&fixture.Context)
if err != nil {
return nil, fmt.Errorf("block context: %w", err)
}

signer := types.MakeSigner(chainCfg, blockCtx.BlockNumber, blockCtx.Time)
_, err = types.Sender(signer, tx)
if err != nil {
return nil, fmt.Errorf("tx sender: %w", err)
}

tracer := &callTracer{}
vmCfg := vm.Config{Tracer: tracer}

gp := new(core.GasPool).AddGas(blockCtx.GasLimit)
header := &types.Header{
Number:     blockCtx.BlockNumber,
Time:       blockCtx.Time,
Difficulty: blockCtx.Difficulty,
GasLimit:   blockCtx.GasLimit,
Coinbase:   blockCtx.Coinbase,
}

engine := ethash.NewFaker()
_, err = core.ApplyTransaction(chainCfg, &noopChain{engine: engine}, &header.Coinbase, gp, stateDB, header, tx, new(uint64), vmCfg)
if err != nil {
return nil, fmt.Errorf("apply transaction: %w", err)
}

blockNum, _ := parseUint64(fixture.Context.Number)

traces := flattenTraces(tracer.root, tx, blockNum)
return traces, nil
}

// buildChainConfig converts a FixtureChainConfig into a params.ChainConfig.
func buildChainConfig(cfg *FixtureChainConfig) (*params.ChainConfig, error) {
toBlock := func(p *uint64) *big.Int {
if p == nil {
return nil
}
return new(big.Int).SetUint64(*p)
}

c := &params.ChainConfig{
ChainID:             new(big.Int).SetUint64(cfg.ChainID),
HomesteadBlock:      toBlock(cfg.HomesteadBlock),
DAOForkSupport:      cfg.DAOForkSupport,
EIP150Block:         toBlock(cfg.EIP150Block),
EIP155Block:         toBlock(cfg.EIP155Block),
EIP158Block:         toBlock(cfg.EIP158Block),
ByzantiumBlock:      toBlock(cfg.ByzantiumBlock),
ConstantinopleBlock: toBlock(cfg.ConstantinopleBlock),
PetersburgBlock:     toBlock(cfg.PetersburgBlock),
IstanbulBlock:       toBlock(cfg.IstanbulBlock),
MuirGlacierBlock:    toBlock(cfg.MuirGlacierBlock),
BerlinBlock:         toBlock(cfg.BerlinBlock),
LondonBlock:         toBlock(cfg.LondonBlock),
Ethash:              &params.EthashConfig{},
}

return c, nil
}

// buildGenesisState creates an in-memory state database populated with the
// accounts defined in the fixture genesis block.
func buildGenesisState(genesis *FixtureGenesis) (*state.StateDB, error) {
db := rawdb.NewMemoryDatabase()
sdb := state.NewDatabase(db)
stateDB, err := state.New(common.Hash{}, sdb, nil)
if err != nil {
return nil, err
}

for addrStr, acct := range genesis.Alloc {
addr := common.HexToAddress(addrStr)

bal, ok := new(big.Int).SetString(trimHexPrefix(acct.Balance), 16)
if !ok {
return nil, fmt.Errorf("invalid balance %q for %s", acct.Balance, addrStr)
}
u256bal, overflow := uint256.FromBig(bal)
if overflow {
return nil, fmt.Errorf("balance overflow for %s", addrStr)
}
stateDB.SetBalance(addr, u256bal)

nonce, err := parseUint64(acct.Nonce)
if err != nil {
return nil, fmt.Errorf("invalid nonce %q for %s: %w", acct.Nonce, addrStr, err)
}
stateDB.SetNonce(addr, nonce)

if acct.Code != "" {
code, err := hex.DecodeString(trimHexPrefix(acct.Code))
if err != nil {
return nil, fmt.Errorf("invalid code for %s: %w", addrStr, err)
}
stateDB.SetCode(addr, code)
}

for k, v := range acct.Storage {
key := common.HexToHash(k)
val := common.HexToHash(v)
stateDB.SetState(addr, key, val)
}
}

root, err := stateDB.Commit(0, false)
if err != nil {
return nil, err
}

// Flush trie nodes to the underlying raw database so the state can be
// re-opened from the committed root.
if err := sdb.TrieDB().Commit(root, false); err != nil {
return nil, err
}

// Re-open at committed root so the EVM state transitions apply cleanly.
stateDB, err = state.New(root, sdb, nil)
if err != nil {
return nil, err
}

return stateDB, nil
}

// buildBlockContext converts the fixture context fields into a vm.BlockContext.
func buildBlockContext(ctx *FixtureContext) (vm.BlockContext, error) {
number, err := parseUint64(ctx.Number)
if err != nil {
return vm.BlockContext{}, fmt.Errorf("block number: %w", err)
}

diff, ok := new(big.Int).SetString(ctx.Difficulty, 10)
if !ok {
return vm.BlockContext{}, fmt.Errorf("invalid difficulty %q", ctx.Difficulty)
}

ts, err := parseUint64(ctx.Timestamp)
if err != nil {
return vm.BlockContext{}, fmt.Errorf("timestamp: %w", err)
}

gasLimit, err := parseUint64(ctx.GasLimit)
if err != nil {
return vm.BlockContext{}, fmt.Errorf("gas limit: %w", err)
}

coinbase := common.HexToAddress(ctx.Miner)

return vm.BlockContext{
CanTransfer: core.CanTransfer,
Transfer:    core.Transfer,
GetHash:     func(n uint64) common.Hash { return common.Hash{} },
Coinbase:    coinbase,
BlockNumber: new(big.Int).SetUint64(number),
Time:        ts,
Difficulty:  diff,
GasLimit:    gasLimit,
}, nil
}

// decodeTransaction RLP-decodes a signed transaction from a hex string.
func decodeTransaction(input string) (*types.Transaction, error) {
b, err := hexutil.Decode(input)
if err != nil {
return nil, err
}
var tx types.Transaction
if err := rlp.DecodeBytes(b, &tx); err != nil {
return nil, err
}
return &tx, nil
}

// flattenTraces converts the callTracer tree into a flat list of ParityTraces
// using depth-first ordering.
func flattenTraces(root *callFrame, tx *types.Transaction, blockNum uint64) []ParityTrace {
if root == nil {
return nil
}
txHash := tx.Hash().Hex()
var traces []ParityTrace
var walk func(frame *callFrame, addr []int)
walk = func(frame *callFrame, addr []int) {
val := frame.value
if val == nil {
val = new(big.Int)
}
t := ParityTrace{
Type: "call",
Action: TraceAction{
From:     strings.ToLower(frame.from.Hex()),
To:       strings.ToLower(frame.to.Hex()),
Value:    hexutil.EncodeBig(val),
Gas:      hexutil.EncodeUint64(frame.gas),
Input:    hexutil.Encode(frame.input),
CallType: callTypeString(frame.typ),
},
Result: TraceResult{
GasUsed: hexutil.EncodeUint64(frame.gasUsed),
Output:  hexutil.Encode(frame.output),
},
TraceAddress:    addr,
Subtraces:       len(frame.children),
TransactionHash: txHash,
BlockNumber:     blockNum,
BlockHash:       common.Hash{}.Hex(),
}
traces = append(traces, t)
for i, child := range frame.children {
childAddr := make([]int, len(addr)+1)
copy(childAddr, addr)
childAddr[len(addr)] = i
walk(child, childAddr)
}
}
walk(root, []int{})
return traces
}

// noopChain satisfies core.ChainContext using a fake ethash engine.
type noopChain struct {
engine consensus.Engine
}

func (n *noopChain) Engine() consensus.Engine {
return n.engine
}

func (n *noopChain) GetHeader(hash common.Hash, number uint64) *types.Header {
return nil
}

// parseUint64 parses a decimal or hex string into a uint64.
func parseUint64(s string) (uint64, error) {
s = strings.TrimSpace(s)
if strings.HasPrefix(s, "0x") || strings.HasPrefix(s, "0X") {
return strconv.ParseUint(s[2:], 16, 64)
}
return strconv.ParseUint(s, 10, 64)
}

// trimHexPrefix removes a leading "0x" or "0X" prefix if present.
func trimHexPrefix(s string) string {
if strings.HasPrefix(s, "0x") || strings.HasPrefix(s, "0X") {
return s[2:]
}
return s
}
