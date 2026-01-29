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
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// Event is an event potentially triggered by the EVM's LOG mechanism. The Event
// holds type information (inputs) about the yielded output. Anonymous events
// don't get the signature canonical representation as the first LOG topic.
// Event 是一个可能由 EVM 的 LOG 机制触发的事件。Event
// 保存有关产生输出的类型信息（输入）。匿名事件
// 不会将签名的规范表示形式作为第一个 LOG 主题。
type Event struct {
	// Name is the event name used for internal representation. It's derived from
	// the raw name and a suffix will be added in the case of event overloading.
	//
	// e.g.
	// These are two events that have the same name:
	// * foo(int,int)
	// * foo(uint,uint)
	// The event name of the first one will be resolved as foo while the second one
	// will be resolved as foo0.
	// Name 是用于内部表示的事件名称。它源自
	// 原始名称，在事件重载的情况下会添加后缀。
	//
	// 例如
	// 这两个事件同名：
	// * foo(int,int)
	// * foo(uint,uint)
	// 第一个事件的名称将解析为 foo，而第二个事件
	// 将解析为 foo0。
	Name string

	// RawName is the raw event name parsed from ABI.
	// RawName 是从 ABI 解析的原始事件名称。
	RawName string
	// Anonymous 指示事件是否是匿名的。
	Anonymous bool
	// Inputs 是事件的参数列表。
	Inputs Arguments
	// str 是事件的缓存字符串表示形式。
	str string

	// Sig contains the string signature according to the ABI spec.
	// e.g.	 event foo(uint32 a, int b) = "foo(uint32,int256)"
	// Please note that "int" is substitute for its canonical representation "int256"
	// Sig 包含根据 ABI 规范的字符串签名。
	// 例如 event foo(uint32 a, int b) = "foo(uint32,int256)"
	// 请注意，"int" 会被替换为其规范表示 "int256"。
	Sig string

	// ID returns the canonical representation of the event's signature used by the
	// abi definition to identify event names and types.
	// ID 返回事件签名的规范表示形式，abi 定义使用它来识别事件名称和类型。
	// 它是事件签名的 Keccak256 哈希值。
	ID common.Hash
}

// NewEvent creates a new Event.
// It sanitizes the input arguments to remove unnamed arguments.
// It also precomputes the id, signature and string representation
// of the event.
// NewEvent 创建一个新的 Event 对象。
// 它会清理输入参数以删除未命名的参数。
// 它还预先计算事件的 id、签名和字符串表示形式。
func NewEvent(name, rawName string, anonymous bool, inputs Arguments) Event {
	// sanitize inputs to remove inputs without names
	// and precompute string and sig representation.
	// 清理输入以删除没有名称的输入
	// 并预计算字符串和签名表示。
	names := make([]string, len(inputs))
	types := make([]string, len(inputs))
	for i, input := range inputs {
		if input.Name == "" {
			// 如果参数没有名称，则为其生成一个默认名称，例如 arg0, arg1, ...
			inputs[i] = Argument{
				Name:    fmt.Sprintf("arg%d", i),
				Indexed: input.Indexed,
				Type:    input.Type,
			}
		} else {
			inputs[i] = input
		}
		// string representation
		// 字符串表示形式，例如 "uint256 value" 或 "uint256 indexed value"
		names[i] = fmt.Sprintf("%v %v", input.Type, inputs[i].Name)
		if input.Indexed {
			names[i] = fmt.Sprintf("%v indexed %v", input.Type, inputs[i].Name)
		}
		// sig representation
		// 签名表示形式，仅包含类型，例如 "uint256"
		types[i] = input.Type.String()
	}

	// 完整的事件字符串表示，例如 "event Transfer(address indexed from, address indexed to, uint256 value)"
	str := fmt.Sprintf("event %v(%v)", rawName, strings.Join(names, ", "))
	// 事件签名，例如 "Transfer(address,address,uint256)"
	sig := fmt.Sprintf("%v(%v)", rawName, strings.Join(types, ","))
	// 事件 ID 是事件签名的 Keccak256 哈希
	id := common.BytesToHash(crypto.Keccak256([]byte(sig)))

	return Event{
		Name:      name,
		RawName:   rawName,
		Anonymous: anonymous,
		Inputs:    inputs,
		str:       str,
		Sig:       sig,
		ID:        id,
	}
}

// String returns the string representation of the event.
// String 返回事件的字符串表示形式。
func (e Event) String() string {
	return e.str
}
