// Copyright 2017 The go-ethereum Authors
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

// Package accounts implements high level Ethereum account management.
// Package accounts 实现了高级的以太坊账户管理。
package accounts

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
	"golang.org/x/crypto/sha3"
)

// Account represents an Ethereum account located at a specific location defined
// by the optional URL field.
// Account 表示位于由可选 URL 字段定义的特定位置的以太坊帐户。
type Account struct {
	Address common.Address `json:"address"` // Ethereum account address derived from the key // 从密钥派生的以太坊帐户地址
	URL     URL            `json:"url"`     // Optional resource locator within a backend // 后端内的可选资源定位符
}

const (
	MimetypeDataWithValidator = "data/validator"
	MimetypeTypedData         = "data/typed"
	MimetypeClique            = "application/x-clique-header"
	MimetypeTextPlain         = "text/plain"
)

// Wallet represents a software or hardware wallet that might contain one or more
// accounts (derived from the same seed).
// Wallet 表示可能包含一个或多个帐户（派生自同一种子）的软件或硬件钱包。
type Wallet interface {
	// URL retrieves the canonical path under which this wallet is reachable. It is
	// used by upper layers to define a sorting order over all wallets from multiple
	// backends.
	// URL 检索此钱包可访问的规范路径。
	// 上层使用它来定义来自多个后端的所以钱包的排序顺序。
	URL() URL

	// Status returns a textual status to aid the user in the current state of the
	// wallet. It also returns an error indicating any failure the wallet might have
	// encountered.
	// Status 返回一个文本状态，以帮助用户了解钱包的当前状态。
	// 它还返回一个错误，指示钱包可能遇到的任何故障。
	Status() (string, error)

	// Open initializes access to a wallet instance. It is not meant to unlock or
	// decrypt account keys, rather simply to establish a connection to hardware
	// wallets and/or to access derivation seeds.
	//
	// The passphrase parameter may or may not be used by the implementation of a
	// particular wallet instance. The reason there is no passwordless open method
	// is to strive towards a uniform wallet handling, oblivious to the different
	// backend providers.
	//
	// Please note, if you open a wallet, you must close it to release any allocated
	// resources (especially important when working with hardware wallets).
	// Open 初始化对钱包实例的访问。它不是为了解锁或解密帐户密钥，
	// 而是为了建立与硬件钱包的连接和/或访问派生种子。
	//
	// 特定钱包实例的实现可能会也可能不会使用 passphrase 参数。
	// 没有无密码 open 方法的原因是为了争取统一的钱包处理，而无需关注不同的后端提供商。
	//
	// 请注意，如果打开钱包，必须关闭它以释放任何分配的资源（在使用硬件钱包时尤为重要）。
	Open(passphrase string) error

	// Close releases any resources held by an open wallet instance.
	// Close 释放已打开钱包实例持有的任何资源。
	Close() error

	// Accounts retrieves the list of signing accounts the wallet is currently aware
	// of. For hierarchical deterministic wallets, the list will not be exhaustive,
	// rather only contain the accounts explicitly pinned during account derivation.
	// Accounts 检索钱包当前感知的签名帐户列表。
	// 对于分层确定性钱包，列表不会是详尽的，而是只包含在帐户派生期间显式固定的帐户。
	Accounts() []Account

	// Contains returns whether an account is part of this particular wallet or not.
	// Contains 返回帐户是否属于此特定钱包。
	Contains(account Account) bool

	// Derive attempts to explicitly derive a hierarchical deterministic account at
	// the specified derivation path. If requested, the derived account will be added
	// to the wallet's tracked account list.
	// Derive 尝试在指定的派生路径上显式派生分层确定性帐户。
	// 如果请求，派生的帐户将添加到钱包的跟踪帐户列表中。
	Derive(path DerivationPath, pin bool) (Account, error)

	// SelfDerive sets a base account derivation path from which the wallet attempts
	// to discover non zero accounts and automatically add them to list of tracked
	// accounts.
	//
	// Note, self derivation will increment the last component of the specified path
	// opposed to descending into a child path to allow discovering accounts starting
	// from non zero components.
	//
	// Some hardware wallets switched derivation paths through their evolution, so
	// this method supports providing multiple bases to discover old user accounts
	// too. Only the last base will be used to derive the next empty account.
	//
	// You can disable automatic account discovery by calling SelfDerive with a nil
	// chain state reader.
	// SelfDerive 设置一个基础帐户派生路径，钱包尝试从中发现非零帐户并自动将它们添加到跟踪帐户列表中。
	//
	// 注意，自我派生将增加指定路径的最后一个组件，而不是下降到子路径，以允许从非零组件开始发现帐户。
	//
	// 一些硬件钱包在其演变过程中切换了派生路径，因此此方法支持提供多个基数组来发现旧的用户帐户。
	// 只有最后一个基数将用于派生下一个空帐户。
	//
	// 可以通过使用 nil 链状态读取器调用 SelfDerive 来禁用自动帐户发现。
	SelfDerive(bases []DerivationPath, chain ethereum.ChainStateReader)

	// SignData requests the wallet to sign the hash of the given data
	// It looks up the account specified either solely via its address contained within,
	// or optionally with the aid of any location metadata from the embedded URL field.
	//
	// If the wallet requires additional authentication to sign the request (e.g.
	// a password to decrypt the account, or a PIN code to verify the transaction),
	// an AuthNeededError instance will be returned, containing infos for the user
	// about which fields or actions are needed. The user may retry by providing
	// the needed details via SignDataWithPassphrase, or by other means (e.g. unlock
	// the account in a keystore).
	// SignData 请求钱包对给定数据的哈希进行签名
	// 它仅通过其中包含的地址，或者可选地借助嵌入式 URL 字段中的任何位置元数据来查找指定的帐户。
	//
	// 如果钱包需要额外的身份验证来签署请求（例如，解密帐户的密码或验证交易的 PIN 码），
	// 将返回 AuthNeededError 实例，其中包含有关需要哪些字段或操作的用户信息。
	// 用户可以通过 SignDataWithPassphrase 提供所需的详细信息来重试，或者通过其他方式（例如在密钥库中解锁帐户）。
	SignData(account Account, mimeType string, data []byte) ([]byte, error)

	// SignDataWithPassphrase is identical to SignData, but also takes a password
	// NOTE: there's a chance that an erroneous call might mistake the two strings, and
	// supply password in the mimetype field, or vice versa. Thus, an implementation
	// should never echo the mimetype or return the mimetype in the error-response
	// SignDataWithPassphrase 与 SignData 相同，但也接受密码
	// 注意：错误的调用可能会混淆这两个字符串，并在 mimetype 字段中提供密码，反之亦然。
	// 因此，实现绝不应回显 mimetype 或在错误响应中返回 mimetype
	SignDataWithPassphrase(account Account, passphrase, mimeType string, data []byte) ([]byte, error)

	// SignText requests the wallet to sign the hash of a given piece of data, prefixed
	// by the Ethereum prefix scheme
	// It looks up the account specified either solely via its address contained within,
	// or optionally with the aid of any location metadata from the embedded URL field.
	//
	// If the wallet requires additional authentication to sign the request (e.g.
	// a password to decrypt the account, or a PIN code to verify the transaction),
	// an AuthNeededError instance will be returned, containing infos for the user
	// about which fields or actions are needed. The user may retry by providing
	// the needed details via SignTextWithPassphrase, or by other means (e.g. unlock
	// the account in a keystore).
	//
	// This method should return the signature in 'canonical' format, with v 0 or 1.
	// SignText 请求钱包签署给定数据的哈希，前缀为以太坊前缀方案
	// 它仅通过其中包含的地址，或者可选地借助嵌入式 URL 字段中的任何位置元数据来查找指定的帐户。
	//
	// 如果钱包需要额外的身份验证来签署请求（例如，解密帐户的密码或验证交易的 PIN 码），
	// 将返回 AuthNeededError 实例，其中包含有关需要哪些字段或操作的用户信息。
	// 用户可以通过 SignTextWithPassphrase 提供所需的详细信息来重试，或者通过其他方式（例如在密钥库中解锁帐户）。
	//
	// 此方法应以 'canonical' 格式返回签名，其中 v 为 0 或 1。
	SignText(account Account, text []byte) ([]byte, error)

	// SignTextWithPassphrase is identical to Signtext, but also takes a password
	// SignTextWithPassphrase 与 Signtext 相同，但也接受密码
	SignTextWithPassphrase(account Account, passphrase string, hash []byte) ([]byte, error)

	// SignTx requests the wallet to sign the given transaction.
	//
	// It looks up the account specified either solely via its address contained within,
	// or optionally with the aid of any location metadata from the embedded URL field.
	//
	// If the wallet requires additional authentication to sign the request (e.g.
	// a password to decrypt the account, or a PIN code to verify the transaction),
	// an AuthNeededError instance will be returned, containing infos for the user
	// about which fields or actions are needed. The user may retry by providing
	// the needed details via SignTxWithPassphrase, or by other means (e.g. unlock
	// the account in a keystore).
	// SignTx 请求钱包签署给定的交易。
	//
	// 它仅通过其中包含的地址，或者可选地借助嵌入式 URL 字段中的任何位置元数据来查找指定的帐户。
	//
	// 如果钱包需要额外的身份验证来签署请求（例如，解密帐户的密码或验证交易的 PIN 码），
	// 将返回 AuthNeededError 实例，其中包含有关需要哪些字段或操作的用户信息。
	// 用户可以通过 SignTxWithPassphrase 提供所需的详细信息来重试，或者通过其他方式（例如在密钥库中解锁帐户）。
	SignTx(account Account, tx *types.Transaction, chainID *big.Int) (*types.Transaction, error)

	// SignTxWithPassphrase is identical to SignTx, but also takes a password
	// SignTxWithPassphrase 与 SignTx 相同，但也接受密码
	SignTxWithPassphrase(account Account, passphrase string, tx *types.Transaction, chainID *big.Int) (*types.Transaction, error)
}

// Backend is a "wallet provider" that may contain a batch of accounts they can
// sign transactions with and upon request, do so.
// Backend 是一个“钱包提供者”，可能包含一批帐户，可以使用这些帐户签署交易，并在请求时执行此操作。
type Backend interface {
	// Wallets retrieves the list of wallets the backend is currently aware of.
	//
	// The returned wallets are not opened by default. For software HD wallets this
	// means that no base seeds are decrypted, and for hardware wallets that no actual
	// connection is established.
	//
	// The resulting wallet list will be sorted alphabetically based on its internal
	// URL assigned by the backend. Since wallets (especially hardware) may come and
	// go, the same wallet might appear at a different positions in the list during
	// subsequent retrievals.
	// Wallets 检索后端当前感知的钱包列表。
	//
	// 返回的钱包默认未打开。对于软件 HD 钱包，这意味着没有解密基础种子，
	// 对于硬件钱包，这意味着没有建立实际连接。
	//
	// 结果钱包列表将根据后端分配的内部 URL 按字母顺序排序。
	// 由于钱包（尤其是硬件钱包）可能会出现和消失，同一钱包在随后的检索中可能会出现在列表中的不同位置。
	Wallets() []Wallet

	// Subscribe creates an async subscription to receive notifications when the
	// backend detects the arrival or departure of a wallet.
	// Subscribe 创建异步订阅，以便在后端检测到钱包到达或离开时接收通知。
	Subscribe(sink chan<- WalletEvent) event.Subscription
}

// TextHash is a helper function that calculates a hash for the given message that can be
// safely used to calculate a signature from.
//
// The hash is calculated as
//
//	keccak256("\x19Ethereum Signed Message:\n"${message length}${message}).
//
// This gives context to the signed message and prevents signing of transactions.
// TextHash 是一个辅助函数，用于计算给定消息的哈希，该哈希可以安全地用于计算签名。
//
// 哈希计算如下：
//
//	keccak256("\x19Ethereum Signed Message:\n"${message length}${message}).
//
// 这为已签名的消息提供了上下文，并防止了交易的签名。
func TextHash(data []byte) []byte {
	hash, _ := TextAndHash(data)
	return hash
}

// TextAndHash is a helper function that calculates a hash for the given message that can be
// safely used to calculate a signature from.
//
// The hash is calculated as
//
//	keccak256("\x19Ethereum Signed Message:\n"${message length}${message}).
//
// This gives context to the signed message and prevents signing of transactions.
// TextAndHash 是一个辅助函数，用于计算给定消息的哈希，该哈希可以安全地用于计算签名。
//
// 哈希计算如下：
//
//	keccak256("\x19Ethereum Signed Message:\n"${message length}${message}).
//
// 这为已签名的消息提供了上下文，并防止了交易的签名。
func TextAndHash(data []byte) ([]byte, string) {
	msg := fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(data), data)
	hasher := sha3.NewLegacyKeccak256()
	hasher.Write([]byte(msg))
	return hasher.Sum(nil), msg
}

// WalletEventType represents the different event types that can be fired by
// the wallet subscription subsystem.
// WalletEventType 表示钱包订阅子系统可以触发的不同事件类型。
type WalletEventType int

const (
	// WalletArrived is fired when a new wallet is detected either via USB or via
	// a filesystem event in the keystore.
	// WalletArrived 在通过 USB 或密钥库中的文件系统事件检测到新钱包时触发。
	WalletArrived WalletEventType = iota

	// WalletOpened is fired when a wallet is successfully opened with the purpose
	// of starting any background processes such as automatic key derivation.
	// WalletOpened 在成功打开钱包以启动任何后台进程（如自动密钥派生）时触发。
	WalletOpened

	// WalletDropped is fired when a wallet is removed or disconnected, either via USB
	// or due to a filesystem event in the keystore. This event indicates that the wallet
	// is no longer available for operations.
	// WalletDropped 在通过 USB 或由于密钥库中的文件系统事件移除或断开钱包时触发。
	// 此事件表示钱包不再可用于操作。
	WalletDropped
)

// WalletEvent is an event fired by an account backend when a wallet arrival or
// departure is detected.
// WalletEvent 是在检测到钱包到达或离开时由帐户后端触发的事件。
type WalletEvent struct {
	Wallet Wallet          // Wallet instance arrived or departed // 到达或离开的钱包实例
	Kind   WalletEventType // Event type that happened in the system // 系统中发生的事件类型
}
