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

// 版权所有 2017 The go-ethereum Authors
// 此文件是 go-ethereum 库的一部分。
//
// go-ethereum 库是免费软件：您可以根据自由软件基金会发布的 GNU 宽通用公共许可证的条款重新分发和/或修改它，
// 可以是许可证的第 3 版，也可以是（由您选择）任何更高版本。
//
// go-ethereum 库的发布是希望它能有用，但没有任何保证；甚至没有对适销性或特定用途适用性的默示保证。
// 有关更多详细信息，请参阅 GNU 宽通用公共许可证。
//
// 您应该已经随 go-ethereum 库收到一份 GNU 宽通用公共许可证的副本。如果没有，请参阅 <http://www.gnu.org/licenses/>。

package accounts

import (
	"errors" // 导入 "errors" 包，用于创建和处理错误。
	"fmt"    // 导入 "fmt" 包，用于格式化字符串。
)

// ErrUnknownAccount is returned for any requested operation for which no backend
// provides the specified account.
// ErrUnknownAccount 在没有任何后端提供指定帐户的情况下返回。
var ErrUnknownAccount = errors.New("unknown account") // 定义一个名为 ErrUnknownAccount 的错误，表示未知账户。

// ErrUnknownWallet is returned for any requested operation for which no backend
// provides the specified wallet.
// ErrUnknownWallet 在没有任何后端提供指定钱包的情况下返回。
var ErrUnknownWallet = errors.New("unknown wallet") // 定义一个名为 ErrUnknownWallet 的错误，表示未知钱包。

// ErrNotSupported is returned when an operation is requested from an account
// backend that it does not support.
// ErrNotSupported 在请求帐户后端不支持的操作时返回。
var ErrNotSupported = errors.New("not supported") // 定义一个名为 ErrNotSupported 的错误，表示不支持的操作。

// ErrInvalidPassphrase is returned when a decryption operation receives a bad
// passphrase.
// ErrInvalidPassphrase 在解密操作接收到错误的密码时返回。
var ErrInvalidPassphrase = errors.New("invalid password") // 定义一个名为 ErrInvalidPassphrase 的错误，表示无效的密码。

// ErrWalletAlreadyOpen is returned if a wallet is attempted to be opened the
// second time.
// ErrWalletAlreadyOpen 在尝试第二次打开钱包时返回。
var ErrWalletAlreadyOpen = errors.New("wallet already open") // 定义一个名为 ErrWalletAlreadyOpen 的错误，表示钱包已打开。

// ErrWalletClosed is returned if a wallet is offline.
// ErrWalletClosed 在钱包离线时返回。
var ErrWalletClosed = errors.New("wallet closed") // 定义一个名为 ErrWalletClosed 的错误，表示钱包已关闭。

// AuthNeededError is returned by backends for signing requests where the user
// is required to provide further authentication before signing can succeed.
//
// This usually means either that a password needs to be supplied, or perhaps a
// one time PIN code displayed by some hardware device.
// AuthNeededError 由后端返回以用于签名请求，其中用户需要在签名成功之前提供进一步的身份验证。
//
// 这通常意味着要么需要提供密码，要么可能需要硬件设备显示的即时 PIN 码。
type AuthNeededError struct {
	Needed string // Extra authentication the user needs to provide // 用户需要提供的额外身份验证
}

// NewAuthNeededError creates a new authentication error with the extra details
// about the needed fields set.
// NewAuthNeededError 创建一个新的身份验证错误，其中设置了有关所需字段的额外详细信息。
func NewAuthNeededError(needed string) error {
	return &AuthNeededError{
		Needed: needed,
	}
}

// Error implements the standard error interface.
// Error 实现标准错误接口。
func (err *AuthNeededError) Error() string {
	return fmt.Sprintf("authentication needed: %s", err.Needed)
}
