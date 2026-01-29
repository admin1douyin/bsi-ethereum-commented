// Copyright 2015 The go-ethereum Authors
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

// 版权所有 2015 The go-ethereum Authors
// 此文件是 go-ethereum 库的一部分。
//
// go-ethereum 库是免费软件：您可以根据自由软件基金会发布的 GNU 宽通用公共许可证的条款重新分发和/或修改它，
// 可以是许可证的第 3 版，也可以是（由您选择）任何更高版本。
//
// go-ethereum 库的发布是希望它能有用，但没有任何保证；甚至没有对适销性或特定用途适用性的默示保证。
// 有关更多详细信息，请参阅 GNU 宽通用公共许可证。
//
// 您应该已经随 go-ethereum 库收到一份 GNU 宽通用公共许可证的副本。如果没有，请参阅 <http://www.gnu.org/licenses/>。

// Package abi implements the Ethereum ABI (Application Binary
// Interface).
//
// The Ethereum ABI is strongly typed, known at compile time
// and static. This ABI will handle basic type casting; unsigned
// to signed and visa versa. It does not handle slice casting such
// as unsigned slice to signed slice. Bit size type casting is also
// handled. ints with a bit size of 32 will be properly cast to int256,
// etc.
// abi 包实现了以太坊 ABI（应用程序二进制接口）。
//
// 以太坊 ABI 是强类型的，在编译时已知并且是静态的。
// 此 ABI 将处理基本的类型转换；无符号到有符号，反之亦然。
// 它不处理切片转换，例如无符号切片到有符号切片。
// 位大小类型转换也会处理。位大小为 32 的 int 将被正确转换为 int256，等等。
package abi
