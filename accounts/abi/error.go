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

// 版权所有 2016 The go-ethereum Authors
// 此文件是 go-ethereum 库的一部分。
//
// go-ethereum 库是免费软件：您可以根据自由软件基金会发布的 GNU 宽通用公共许可证的条款重新分发和/或修改它，
// 可以是许可证的第 3 版，也可以是（由您选择）任何更高版本。
//
// go-ethereum 库的发布是希望它能有用，但没有任何保证；甚至没有对适销性或特定用途适用性的默示保证。
// 有关更多详细信息，请参阅 GNU 宽通用公共许可证。
//
// 您应该已经随 go-ethereum 库收到一份 GNU 宽通用公共许可证的副本。如果没有，请参阅 <http://www.gnu.org/licenses/>。

package abi

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// Error 结构体表示 ABI 中的自定义错误。
type Error struct {
	Name   string
	Inputs Arguments
	str    string // 缓存的字符串表示形式

	// Sig 包含根据 ABI 规范的字符串签名。
	// 例如: error foo(uint32 a, int b) = "foo(uint32,int256)"
	// 请注意，“int”会替换为其规范表示“int256”。
	Sig string

	// ID 返回错误的规范表示形式，abi 定义使用它来识别事件名称和类型。
	ID common.Hash
}

// NewError 创建一个新的 Error 对象。
func NewError(name string, inputs Arguments) Error {
	// 清理输入，以删除没有名称的输入，
	// 并预先计算字符串和签名表示形式。
	names := make([]string, len(inputs))
	types := make([]string, len(inputs))
	for i, input := range inputs {
		if input.Name == "" {
			inputs[i] = Argument{
				Name:    fmt.Sprintf("arg%d", i),
				Indexed: input.Indexed,
				Type:    input.Type,
			}
		} else {
			inputs[i] = input
		}
		// 字符串表示形式
		names[i] = fmt.Sprintf("%v %v", input.Type, inputs[i].Name)
		if input.Indexed {
			names[i] = fmt.Sprintf("%v indexed %v", input.Type, inputs[i].Name)
		}
		// 签名表示形式
		types[i] = input.Type.String()
	}

	str := fmt.Sprintf("error %v(%v)", name, strings.Join(names, ", "))
	sig := fmt.Sprintf("%v(%v)", name, strings.Join(types, ","))
	// 错误签名的 Keccak256 哈希值作为 ID
	id := common.BytesToHash(crypto.Keccak256([]byte(sig)))

	return Error{
		Name:   name,
		Inputs: inputs,
		str:    str,
		Sig:    sig,
		ID:     id,
	}
}

// String 返回错误的字符串表示形式。
func (e Error) String() string {
	return e.str
}

// Unpack 将给定的数据解包为合适的值。
// 它首先检查 4 字节的错误 ID，然后解包参数。
func (e *Error) Unpack(data []byte) (interface{}, error) {
	if len(data) < 4 {
		return "", fmt.Errorf("insufficient data for unpacking: have %d, want at least 4", len(data))
	}
	if !bytes.Equal(data[:4], e.ID[:4]) {
		return "", fmt.Errorf("invalid identifier, have %#x want %#x", data[:4], e.ID[:4])
	}
	return e.Inputs.Unpack(data[4:])
}
