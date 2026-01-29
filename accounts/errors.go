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

package accounts

import (
	"errors"
	"fmt"
)

// ErrUnknownAccount is returned for any requested operation for which no backend
// provides the specified account.
// ErrUnknownAccount 在没有任何后端提供指定帐户的情况下返回。
var ErrUnknownAccount = errors.New("unknown account")

// ErrUnknownWallet is returned for any requested operation for which no backend
// provides the specified wallet.
// ErrUnknownWallet 在没有任何后端提供指定钱包的情况下返回。
var ErrUnknownWallet = errors.New("unknown wallet")

// ErrNotSupported is returned when an operation is requested from an account
// backend that it does not support.
// ErrNotSupported 在请求帐户后端不支持的操作时返回。
var ErrNotSupported = errors.New("not supported")

// ErrInvalidPassphrase is returned when a decryption operation receives a bad
// passphrase.
// ErrInvalidPassphrase 在解密操作接收到错误的密码时返回。
var ErrInvalidPassphrase = errors.New("invalid password")

// ErrWalletAlreadyOpen is returned if a wallet is attempted to be opened the
// second time.
// ErrWalletAlreadyOpen 在尝试第二次打开钱包时返回。
var ErrWalletAlreadyOpen = errors.New("wallet already open")

// ErrWalletClosed is returned if a wallet is offline.
// ErrWalletClosed 在钱包离线时返回。
var ErrWalletClosed = errors.New("wallet closed")

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
