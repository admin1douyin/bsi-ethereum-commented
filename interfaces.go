// Copyright 2016 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

// Package ethereum defines interfaces for interacting with Ethereum.
// Package ethereum 定义了与以太坊交互的接口。
package ethereum

import (
	"context"
	"encoding/json"
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
)

// NotFound is returned by API methods if the requested item does not exist.
// NotFound 如果请求的项目不存在，API 方法将返回 NotFound。
var NotFound = errors.New("not found")

// Subscription represents an event subscription where events are
// delivered on a data channel.
// Subscription 表示一个事件订阅，其中事件通过数据通道传递。
type Subscription interface {
	// Unsubscribe cancels the sending of events to the data channel
	// and closes the error channel.
	// Unsubscribe 取消向数据通道发送事件并关闭错误通道。
	Unsubscribe()
	// Err returns the subscription error channel. The error channel receives
	// a value if there is an issue with the subscription (e.g. the network connection
	// delivering the events has been closed). Only one value will ever be sent.
	// The error channel is closed by Unsubscribe.
	// Err 返回订阅错误通道。如果订阅出现问题（例如传递事件的网络连接已关闭），
	// 错误通道将接收到一个值。只会发送一个值。
	// Unsubscribe 会关闭错误通道。
	Err() <-chan error
}

// ChainReader provides access to the blockchain. The methods in this interface access raw
// data from either the canonical chain (when requesting by block number) or any
// blockchain fork that was previously downloaded and processed by the node. The block
// number argument can be nil to select the latest canonical block. Reading block headers
// should be preferred over full blocks whenever possible.
//
// The returned error is NotFound if the requested item does not exist.
// ChainReader 提供对区块链的访问。此接口中的方法访问来自规范链（按块号请求时）
// 或节点先前下载并处理的任何区块链分叉的原始数据。块号参数可以为 nil 以选择最新的规范块。
// 尽可能优先读取块头而不是完整的块。
//
// 如果请求的项目不存在，返回的错误为 NotFound。
type ChainReader interface {
	BlockByHash(ctx context.Context, hash common.Hash) (*types.Block, error)
	BlockByNumber(ctx context.Context, number *big.Int) (*types.Block, error)
	HeaderByHash(ctx context.Context, hash common.Hash) (*types.Header, error)
	HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error)
	TransactionCount(ctx context.Context, blockHash common.Hash) (uint, error)
	TransactionInBlock(ctx context.Context, blockHash common.Hash, index uint) (*types.Transaction, error)

	// This method subscribes to notifications about changes of the head block of
	// the canonical chain.
	// 此方法订阅有关规范链头块更改的通知。
	SubscribeNewHead(ctx context.Context, ch chan<- *types.Header) (Subscription, error)
}

// TransactionReceiptsQuery defines criteria for transaction receipts subscription.
// If TransactionHashes is empty, receipts for all transactions included in new blocks will be delivered.
// Otherwise, only receipts for the specified transactions will be delivered.
// TransactionReceiptsQuery 定义了交易收据订阅的标准。
// 如果 TransactionHashes 为空，则将传递新块中包含的所有交易的收据。
// 否则，仅传递指定交易的收据。
type TransactionReceiptsQuery struct {
	TransactionHashes []common.Hash
}

// TransactionReader provides access to past transactions and their receipts.
// Implementations may impose arbitrary restrictions on the transactions and receipts that
// can be retrieved. Historic transactions may not be available.
//
// Avoid relying on this interface if possible. Contract logs (through the LogFilterer
// interface) are more reliable and usually safer in the presence of chain
// reorganisations.
//
// The returned error is NotFound if the requested item does not exist.
// TransactionReader 提供对过去交易及其收据的访问。
// 实现可能会对可检索的交易和收据施加任意限制。历史交易可能不可用。
//
// 尽可能避免依赖此接口。在存在链重组的情况下，合约日志（通过 LogFilterer 接口）更可靠且通常更安全。
//
// 如果请求的项目不存在，返回的错误为 NotFound。
type TransactionReader interface {
	// TransactionByHash checks the pool of pending transactions in addition to the
	// blockchain. The isPending return value indicates whether the transaction has been
	// mined yet. Note that the transaction may not be part of the canonical chain even if
	// it's not pending.
	// TransactionByHash 除了检查区块链外，还检查待处理交易池。isPending 返回值指示交易是否已被挖掘。
	// 请注意，即使交易不是待处理状态，它也可能不是规范链的一部分。
	TransactionByHash(ctx context.Context, txHash common.Hash) (tx *types.Transaction, isPending bool, err error)
	// TransactionReceipt returns the receipt of a mined transaction. Note that the
	// transaction may not be included in the current canonical chain even if a receipt
	// exists.
	// TransactionReceipt 返回已挖掘交易的收据。请注意，即使存在收据，交易也可能不包含在当前的规范链中。
	TransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error)
	// SubscribeTransactionReceipts subscribes to notifications about transaction receipts.
	// The receipts are delivered in batches when transactions are included in blocks.
	// If q is nil or has empty TransactionHashes, all receipts from new blocks will be delivered.
	// Otherwise, only receipts for the specified transaction hashes will be delivered.
	// SubscribeTransactionReceipts 订阅有关交易收据的通知。
	// 当交易包含在块中时，收据将分批传递。
	// 如果 q 为 nil 或 TransactionHashes 为空，则将传递新块中的所有收据。
	// 否则，仅传递指定交易哈希的收据。
	SubscribeTransactionReceipts(ctx context.Context, q *TransactionReceiptsQuery, ch chan<- []*types.Receipt) (Subscription, error)
}

// ChainStateReader wraps access to the state trie of the canonical blockchain. Note that
// implementations of the interface may be unable to return state values for old blocks.
// In many cases, using CallContract can be preferable to reading raw contract storage.
// ChainStateReader 封装了对规范区块链状态树的访问。请注意，接口的实现可能无法返回旧块的状态值。
// 在许多情况下，使用 CallContract 可能比读取原始合约存储更可取。
type ChainStateReader interface {
	BalanceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (*big.Int, error)
	StorageAt(ctx context.Context, account common.Address, key common.Hash, blockNumber *big.Int) ([]byte, error)
	CodeAt(ctx context.Context, account common.Address, blockNumber *big.Int) ([]byte, error)
	NonceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (uint64, error)
}

// SyncProgress gives progress indications when the node is synchronising with
// the Ethereum network.
// SyncProgress 在节点与以太坊网络同步时提供进度指示。
type SyncProgress struct {
	StartingBlock uint64 // Block number where sync began // 同步开始的块号
	CurrentBlock  uint64 // Current block number where sync is at // 同步当前所在的块号
	HighestBlock  uint64 // Highest alleged block number in the chain // 链中声称的最高块号

	// "fast sync" fields. These used to be sent by geth, but are no longer used
	// since version v1.10.
	// "快速同步" 字段。这些字段曾由 geth 发送，但自版本 v1.10 起不再使用。
	PulledStates uint64 // Number of state trie entries already downloaded // 已下载的状态树条目数
	KnownStates  uint64 // Total number of state trie entries known about // 已知的状态树条目总数

	// "snap sync" fields.
	// "快照同步" 字段。
	SyncedAccounts      uint64 // Number of accounts downloaded // 已下载的账户数
	SyncedAccountBytes  uint64 // Number of account trie bytes persisted to disk // 持久化到磁盘的账户树字节数
	SyncedBytecodes     uint64 // Number of bytecodes downloaded // 已下载的字节码数
	SyncedBytecodeBytes uint64 // Number of bytecode bytes downloaded // 已下载的字节码字节数
	SyncedStorage       uint64 // Number of storage slots downloaded // 已下载的存储槽数
	SyncedStorageBytes  uint64 // Number of storage trie bytes persisted to disk // 持久化到磁盘的存储树字节数

	HealedTrienodes     uint64 // Number of state trie nodes downloaded // 已下载的状态树节点数
	HealedTrienodeBytes uint64 // Number of state trie bytes persisted to disk // 持久化到磁盘的状态树字节数
	HealedBytecodes     uint64 // Number of bytecodes downloaded // 已下载的字节码数
	HealedBytecodeBytes uint64 // Number of bytecodes persisted to disk // 持久化到磁盘的字节码数

	HealingTrienodes uint64 // Number of state trie nodes pending // 待处理的状态树节点数
	HealingBytecode  uint64 // Number of bytecodes pending // 待处理的字节码数

	// "transaction indexing" fields
	// "交易索引" 字段
	TxIndexFinishedBlocks  uint64 // Number of blocks whose transactions are already indexed // 交易已索引的块数
	TxIndexRemainingBlocks uint64 // Number of blocks whose transactions are not indexed yet // 交易尚未索引的块数

	// "historical state indexing" fields
	// "历史状态索引" 字段
	StateIndexRemaining uint64 // Number of states remain unindexed // 剩余未索引的状态数
}

// Done returns the indicator if the initial sync is finished or not.
// Done 返回初始同步是否完成的指示符。
func (prog SyncProgress) Done() bool {
	if prog.CurrentBlock < prog.HighestBlock {
		return false
	}
	return prog.TxIndexRemainingBlocks == 0 && prog.StateIndexRemaining == 0
}

// ChainSyncReader wraps access to the node's current sync status. If there's no
// sync currently running, it returns nil.
// ChainSyncReader 封装了对节点当前同步状态的访问。如果当前没有正在运行的同步，则返回 nil。
type ChainSyncReader interface {
	SyncProgress(ctx context.Context) (*SyncProgress, error)
}

// CallMsg contains parameters for contract calls.
// CallMsg 包含合约调用的参数。
type CallMsg struct {
	From      common.Address  // the sender of the 'transaction' // '交易'的发送者
	To        *common.Address // the destination contract (nil for contract creation) // 目标合约（nil 表示合约创建）
	Gas       uint64          // if 0, the call executes with near-infinite gas // 如果为 0，则调用以近乎无限的 gas 执行
	GasPrice  *big.Int        // wei <-> gas exchange ratio // wei <-> gas 兑换比率
	GasFeeCap *big.Int        // EIP-1559 fee cap per gas. // EIP-1559 每 gas 费用上限。
	GasTipCap *big.Int        // EIP-1559 tip per gas. // EIP-1559 每 gas 小费。
	Value     *big.Int        // amount of wei sent along with the call // 随调用发送的 wei 数量
	Data      []byte          // input data, usually an ABI-encoded contract method invocation // 输入数据，通常是 ABI 编码的合约方法调用

	AccessList types.AccessList // EIP-2930 access list. // EIP-2930 访问列表。

	// For BlobTxType
	// 用于 BlobTxType
	BlobGasFeeCap *big.Int
	BlobHashes    []common.Hash

	// For SetCodeTxType
	// 用于 SetCodeTxType
	AuthorizationList []types.SetCodeAuthorization
}

// A ContractCaller provides contract calls, essentially transactions that are executed by
// the EVM but not mined into the blockchain. ContractCall is a low-level method to
// execute such calls. For applications which are structured around specific contracts,
// the abigen tool provides a nicer, properly typed way to perform calls.
// ContractCaller 提供合约调用，本质上是由 EVM 执行但未挖掘到区块链中的交易。
// ContractCall 是执行此类调用的低级方法。对于围绕特定合约构建的应用程序，
// abigen 工具提供了一种更好、类型正确的方式来执行调用。
type ContractCaller interface {
	CallContract(ctx context.Context, call CallMsg, blockNumber *big.Int) ([]byte, error)
}

// FilterQuery contains options for contract log filtering.
// FilterQuery 包含合约日志过滤的选项。
type FilterQuery struct {
	BlockHash *common.Hash     // used by eth_getLogs, return logs only from block with this hash // 由 eth_getLogs 使用，仅返回具有此哈希的块中的日志
	FromBlock *big.Int         // beginning of the queried range, nil means genesis block // 查询范围的开始，nil 表示创世块
	ToBlock   *big.Int         // end of the range, nil means latest block // 范围的结束，nil 表示最新块
	Addresses []common.Address // restricts matches to events created by specific contracts // 限制匹配特定合约创建的事件

	// The Topic list restricts matches to particular event topics. Each event has a list
	// of topics. Topics matches a prefix of that list. An empty element slice matches any
	// topic. Non-empty elements represent an alternative that matches any of the
	// contained topics.
	//
	// Examples:
	// {} or nil          matches any topic list
	// {{A}}              matches topic A in first position
	// {{}, {B}}          matches any topic in first position AND B in second position
	// {{A}, {B}}         matches topic A in first position AND B in second position
	// {{A, B}, {C, D}}   matches topic (A OR B) in first position AND (C OR D) in second position
	// Topic 列表限制匹配特定事件主题。每个事件都有一个主题列表。
	// Topics 匹配该列表的前缀。空元素切片匹配任何主题。
	// 非空元素表示匹配包含的任何主题的替代方案。
	//
	// 示例:
	// {} 或 nil          匹配任何主题列表
	// {{A}}              匹配第一个位置的主题 A
	// {{}, {B}}          匹配第一个位置的任何主题 AND 第二个位置的主题 B
	// {{A}, {B}}         匹配第一个位置的主题 A AND 第二个位置的主题 B
	// {{A, B}, {C, D}}   匹配第一个位置的主题 (A OR B) AND 第二个位置的主题 (C OR D)
	Topics [][]common.Hash
}

// LogFilterer provides access to contract log events using a one-off query or continuous
// event subscription.
//
// Logs received through a streaming query subscription may have Removed set to true,
// indicating that the log was reverted due to a chain reorganisation.
// LogFilterer 使用一次性查询或连续事件订阅提供对合约日志事件的访问。
//
// 通过流式查询订阅接收的日志可能将 Removed 设置为 true，
// 表示由于链重组而撤消了该日志。
type LogFilterer interface {
	FilterLogs(ctx context.Context, q FilterQuery) ([]types.Log, error)
	SubscribeFilterLogs(ctx context.Context, q FilterQuery, ch chan<- types.Log) (Subscription, error)
}

// TransactionSender wraps transaction sending. The SendTransaction method injects a
// signed transaction into the pending transaction pool for execution. If the transaction
// was a contract creation, the TransactionReceipt method can be used to retrieve the
// contract address after the transaction has been mined.
//
// The transaction must be signed and have a valid nonce to be included. Consumers of the
// API can use package accounts to maintain local private keys and need can retrieve the
// next available nonce using PendingNonceAt.
// TransactionSender 封装交易发送。SendTransaction 方法将已签名的交易注入待处理交易池以执行。
// 如果交易是合约创建，则 TransactionReceipt 方法可用于在交易被挖掘后检索合约地址。
//
// 交易必须经过签名并具有有效的 nonce 才能包含在内。
// API 的使用者可以使用 accounts 包来维护本地私钥，并可以使用 PendingNonceAt 检索下一个可用的 nonce。
type TransactionSender interface {
	SendTransaction(ctx context.Context, tx *types.Transaction) error
}

// GasPricer wraps the gas price oracle, which monitors the blockchain to determine the
// optimal gas price given current fee market conditions.
// GasPricer 封装了 gas 价格预言机，该预言机监视区块链以根据当前费用市场状况确定最佳 gas 价格。
type GasPricer interface {
	SuggestGasPrice(ctx context.Context) (*big.Int, error)
}

// GasPricer1559 provides access to the EIP-1559 gas price oracle.
// GasPricer1559 提供对 EIP-1559 gas 价格预言机的访问。
type GasPricer1559 interface {
	SuggestGasTipCap(ctx context.Context) (*big.Int, error)
}

// FeeHistoryReader provides access to the fee history oracle.
// FeeHistoryReader 提供对费用历史预言机的访问。
type FeeHistoryReader interface {
	FeeHistory(ctx context.Context, blockCount uint64, lastBlock *big.Int, rewardPercentiles []float64) (*FeeHistory, error)
}

// FeeHistory provides recent fee market data that consumers can use to determine
// a reasonable maxPriorityFeePerGas value.
// FeeHistory 提供最近的费用市场数据，消费者可以使用这些数据来确定合理的 maxPriorityFeePerGas 值。
type FeeHistory struct {
	OldestBlock  *big.Int     // block corresponding to first response value // 对应于第一个响应值的块
	Reward       [][]*big.Int // list every txs priority fee per block // 每个块的每笔交易优先费列表
	BaseFee      []*big.Int   // list of each block's base fee // 每个块的基本费用列表
	GasUsedRatio []float64    // ratio of gas used out of the total available limit // 已用 gas 占总可用限额的比例
}

// A PendingStateReader provides access to the pending state, which is the result of all
// known executable transactions which have not yet been included in the blockchain. It is
// commonly used to display the result of ’unconfirmed’ actions (e.g. wallet value
// transfers) initiated by the user. The PendingNonceAt operation is a good way to
// retrieve the next available transaction nonce for a specific account.
// PendingStateReader 提供对待处理状态的访问，这是尚未包含在区块链中的所有已知可执行交易的结果。
// 它通常用于显示用户发起的“未确认”操作（例如钱包价值转移）的结果。
// PendingNonceAt 操作是检索特定帐户的下一个可用交易 nonce 的好方法。
type PendingStateReader interface {
	PendingBalanceAt(ctx context.Context, account common.Address) (*big.Int, error)
	PendingStorageAt(ctx context.Context, account common.Address, key common.Hash) ([]byte, error)
	PendingCodeAt(ctx context.Context, account common.Address) ([]byte, error)
	PendingNonceAt(ctx context.Context, account common.Address) (uint64, error)
	PendingTransactionCount(ctx context.Context) (uint, error)
}

// PendingContractCaller can be used to perform calls against the pending state.
// PendingContractCaller 可用于对待处理状态执行调用。
type PendingContractCaller interface {
	PendingCallContract(ctx context.Context, call CallMsg) ([]byte, error)
}

// GasEstimator wraps EstimateGas, which tries to estimate the gas needed to execute a
// specific transaction based on the pending state. There is no guarantee that this is the
// true gas limit requirement as other transactions may be added or removed by miners, but
// it should provide a basis for setting a reasonable default.
// GasEstimator 封装了 EstimateGas，它尝试根据待处理状态估算执行特定交易所需的 gas。
// 无法保证这是真实的 gas 限制要求，因为矿工可能会添加或删除其他交易，但它应该为设置合理的默认值提供基础。
type GasEstimator interface {
	EstimateGas(ctx context.Context, call CallMsg) (uint64, error)
}

// A PendingStateEventer provides access to real time notifications about changes to the
// pending state.
// PendingStateEventer 提供有关待处理状态更改的实时通知的访问。
type PendingStateEventer interface {
	SubscribePendingTransactions(ctx context.Context, ch chan<- *types.Transaction) (Subscription, error)
}

// BlockNumberReader provides access to the current block number.
// BlockNumberReader 提供对当前块号的访问。
type BlockNumberReader interface {
	BlockNumber(ctx context.Context) (uint64, error)
}

// ChainIDReader provides access to the chain ID.
// ChainIDReader 提供对链 ID 的访问。
type ChainIDReader interface {
	ChainID(ctx context.Context) (*big.Int, error)
}

// OverrideAccount specifies the state of an account to be overridden.
// OverrideAccount 指定要覆盖的帐户状态。
type OverrideAccount struct {
	// Nonce sets nonce of the account. Note: the nonce override will only
	// be applied when it is set to a non-zero value.
	// Nonce 设置帐户的 nonce。注意：nonce 覆盖仅在设置为非零值时应用。
	Nonce uint64

	// Code sets the contract code. The override will be applied
	// when the code is non-nil, i.e. setting empty code is possible
	// using an empty slice.
	// Code 设置合约代码。当代码非 nil 时将应用覆盖，即可以使用空切片设置空代码。
	Code []byte

	// Balance sets the account balance.
	// Balance 设置帐户余额。
	Balance *big.Int

	// State sets the complete storage. The override will be applied
	// when the given map is non-nil. Using an empty map wipes the
	// entire contract storage during the call.
	// State 设置完整的存储。当给定的 map 非 nil 时将应用覆盖。在调用期间使用空 map 会擦除整个合约存储。
	State map[common.Hash]common.Hash

	// StateDiff allows overriding individual storage slots.
	// StateDiff 允许覆盖单个存储槽。
	StateDiff map[common.Hash]common.Hash
}

func (a OverrideAccount) MarshalJSON() ([]byte, error) {
	type acc struct {
		Nonce     hexutil.Uint64              `json:"nonce,omitempty"`
		Code      string                      `json:"code,omitempty"`
		Balance   *hexutil.Big                `json:"balance,omitempty"`
		State     interface{}                 `json:"state,omitempty"`
		StateDiff map[common.Hash]common.Hash `json:"stateDiff,omitempty"`
	}

	output := acc{
		Nonce:     hexutil.Uint64(a.Nonce),
		Balance:   (*hexutil.Big)(a.Balance),
		StateDiff: a.StateDiff,
	}
	if a.Code != nil {
		output.Code = hexutil.Encode(a.Code)
	}
	if a.State != nil {
		output.State = a.State
	}
	return json.Marshal(output)
}

// BlockOverrides specifies the set of header fields to override.
// BlockOverrides 指定要覆盖的头字段集。
type BlockOverrides struct {
	// Number overrides the block number.
	// Number 覆盖块号。
	Number *big.Int
	// Difficulty overrides the block difficulty.
	// Difficulty 覆盖块难度。
	Difficulty *big.Int
	// Time overrides the block timestamp. Time is applied only when
	// it is non-zero.
	// Time 覆盖块时间戳。仅当 Time 非零时应用。
	Time uint64
	// GasLimit overrides the block gas limit. GasLimit is applied only when
	// it is non-zero.
	// GasLimit 覆盖块 gas 限制。仅当 GasLimit 非零时应用。
	GasLimit uint64
	// Coinbase overrides the block coinbase. Coinbase is applied only when
	// it is different from the zero address.
	// Coinbase 覆盖块 coinbase。仅当 Coinbase 与零地址不同时应用。
	Coinbase common.Address
	// Random overrides the block extra data which feeds into the RANDOM opcode.
	// Random is applied only when it is a non-zero hash.
	// Random 覆盖输入 RANDOM 操作码的块额外数据。
	// 仅当 Random 为非零哈希时应用。
	Random common.Hash
	// BaseFee overrides the block base fee.
	// BaseFee 覆盖块基本费用。
	BaseFee *big.Int
}

func (o BlockOverrides) MarshalJSON() ([]byte, error) {
	type override struct {
		Number     *hexutil.Big    `json:"number,omitempty"`
		Difficulty *hexutil.Big    `json:"difficulty,omitempty"`
		Time       hexutil.Uint64  `json:"time,omitempty"`
		GasLimit   hexutil.Uint64  `json:"gasLimit,omitempty"`
		Coinbase   *common.Address `json:"feeRecipient,omitempty"`
		Random     *common.Hash    `json:"prevRandao,omitempty"`
		BaseFee    *hexutil.Big    `json:"baseFeePerGas,omitempty"`
	}

	output := override{
		Number:     (*hexutil.Big)(o.Number),
		Difficulty: (*hexutil.Big)(o.Difficulty),
		Time:       hexutil.Uint64(o.Time),
		GasLimit:   hexutil.Uint64(o.GasLimit),
		BaseFee:    (*hexutil.Big)(o.BaseFee),
	}
	if o.Coinbase != (common.Address{}) {
		output.Coinbase = &o.Coinbase
	}
	if o.Random != (common.Hash{}) {
		output.Random = &o.Random
	}
	return json.Marshal(output)
}
