package main

// Fixture represents the JSON test fixture format used for transaction replay.
type Fixture struct {
	Genesis FixtureGenesis `json:"genesis"`
	Context FixtureContext `json:"context"`
	Input   string         `json:"input"`
	Result  []ParityTrace  `json:"result"`
}

// FixtureGenesis holds the genesis block header fields and account allocations.
type FixtureGenesis struct {
	Difficulty string                     `json:"difficulty"`
	ExtraData  string                     `json:"extraData"`
	GasLimit   string                     `json:"gasLimit"`
	Hash       string                     `json:"hash"`
	Miner      string                     `json:"miner"`
	MixHash    string                     `json:"mixHash"`
	Nonce      string                     `json:"nonce"`
	Number     string                     `json:"number"`
	StateRoot  string                     `json:"stateRoot"`
	Timestamp  string                     `json:"timestamp"`
	Alloc      map[string]FixtureAccount  `json:"alloc"`
	Config     FixtureChainConfig         `json:"config"`
}

// FixtureAccount represents a pre-allocated account in the genesis state.
type FixtureAccount struct {
	Balance string            `json:"balance"`
	Nonce   string            `json:"nonce"`
	Code    string            `json:"code,omitempty"`
	Storage map[string]string `json:"storage,omitempty"`
}

// FixtureChainConfig holds the chain configuration fields used in the fixture.
type FixtureChainConfig struct {
	ChainID             uint64  `json:"chainId"`
	HomesteadBlock      *uint64 `json:"homesteadBlock,omitempty"`
	DAOForkSupport      bool    `json:"daoForkSupport,omitempty"`
	EIP150Block         *uint64 `json:"eip150Block,omitempty"`
	EIP150Hash          string  `json:"eip150Hash,omitempty"`
	EIP155Block         *uint64 `json:"eip155Block,omitempty"`
	EIP158Block         *uint64 `json:"eip158Block,omitempty"`
	ByzantiumBlock      *uint64 `json:"byzantiumBlock,omitempty"`
	ConstantinopleBlock *uint64 `json:"constantinopleBlock,omitempty"`
	PetersburgBlock     *uint64 `json:"petersburgBlock,omitempty"`
	IstanbulBlock       *uint64 `json:"istanbulBlock,omitempty"`
	MuirGlacierBlock    *uint64 `json:"muirGlacierBlock,omitempty"`
	BerlinBlock         *uint64 `json:"berlinBlock,omitempty"`
	LondonBlock         *uint64 `json:"londonBlock,omitempty"`
}

// FixtureContext holds the block header parameters for the block containing the
// transaction being replayed.
type FixtureContext struct {
	Number     string `json:"number"`
	Difficulty string `json:"difficulty"`
	Timestamp  string `json:"timestamp"`
	GasLimit   string `json:"gasLimit"`
	Miner      string `json:"miner"`
}

// ParityTrace is a single entry in a Parity-style transaction call trace.
type ParityTrace struct {
	Type                string      `json:"type"`
	Action              TraceAction `json:"action"`
	Result              TraceResult `json:"result"`
	TraceAddress        []int       `json:"traceAddress"`
	Subtraces           int         `json:"subtraces"`
	TransactionPosition int         `json:"transactionPosition"`
	TransactionHash     string      `json:"transactionHash"`
	BlockNumber         uint64      `json:"blockNumber"`
	BlockHash           string      `json:"blockHash"`
}

// TraceAction describes the call that was made.
type TraceAction struct {
	From     string `json:"from"`
	To       string `json:"to"`
	Value    string `json:"value"`
	Gas      string `json:"gas"`
	Input    string `json:"input"`
	CallType string `json:"callType"`
}

// TraceResult describes the outcome of a call.
type TraceResult struct {
	GasUsed string `json:"gasUsed"`
	Output  string `json:"output"`
}
