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
	"errors"
	"fmt"
	"math/big"
	"reflect"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
)

// packBytesSlice packs the given bytes as [L, V] as the canonical representation
// bytes slice.
// packBytesSlice 将给定的字节打包为 [L, V] 格式，作为字节切片的规范表示。
// L 是长度，V 是内容。
func packBytesSlice(bytes []byte, l int) []byte {
	// 打包长度
	len := packNum(reflect.ValueOf(l))
	// 拼接长度和内容。内容向右填充到 32 字节的倍数。
	return append(len, common.RightPadBytes(bytes, (l+31)/32*32)...)
}

// packElement packs the given reflect value according to the abi specification in
// t.
// packElement 根据 t 中的 abi 规范打包给定的反射值。
func packElement(t Type, reflectValue reflect.Value) ([]byte, error) {
	switch t.T {
	case UintTy:
		// make sure to not pack a negative value into a uint type.
		// 确保不要将负值打包到 uint 类型中。
		if reflectValue.Kind() == reflect.Ptr {
			val := new(big.Int).Set(reflectValue.Interface().(*big.Int))
			if val.Sign() == -1 {
				return nil, errInvalidSign
			}
		}
		return packNum(reflectValue), nil
	case IntTy:
		return packNum(reflectValue), nil
	case StringTy:
		// 字符串作为动态字节切片处理
		return packBytesSlice([]byte(reflectValue.String()), reflectValue.Len()), nil
	case AddressTy:
		// 地址类型
		if reflectValue.Kind() == reflect.Array {
			// 如果是数组，则转换为字节切片
			reflectValue = mustArrayToByteSlice(reflectValue)
		}

		// 向左填充到 32 字节
		return common.LeftPadBytes(reflectValue.Bytes(), 32), nil
	case BoolTy:
		// 布尔类型
		if reflectValue.Bool() {
			// true 表示为 1
			return math.PaddedBigBytes(common.Big1, 32), nil
		}
		// false 表示为 0
		return math.PaddedBigBytes(common.Big0, 32), nil
	case BytesTy:
		// 动态字节数组
		if reflectValue.Kind() == reflect.Array {
			reflectValue = mustArrayToByteSlice(reflectValue)
		}
		if reflectValue.Type() != reflect.TypeOf([]byte{}) {
			return []byte{}, errors.New("bytes type is neither slice nor array") // 字节类型既不是切片也不是数组
		}
		return packBytesSlice(reflectValue.Bytes(), reflectValue.Len()), nil
	case FixedBytesTy, FunctionTy:
		// 固定大小的字节数组或函数类型
		if reflectValue.Kind() == reflect.Array {
			reflectValue = mustArrayToByteSlice(reflectValue)
		}
		// 向右填充到 32 字节
		return common.RightPadBytes(reflectValue.Bytes(), 32), nil
	default:
		return []byte{}, fmt.Errorf("could not pack element, unknown type: %v", t.T) // 无法打包元素，未知类型
	}
}

// packNum packs the given number (using the reflect value) and will cast it to appropriate number representation.
// packNum 打包给定的数字（使用反射值），并将其转换为适当的数字表示形式。
// 所有整数类型都打包为 256 位的 big-endian 数。
func packNum(value reflect.Value) []byte {
	switch kind := value.Kind(); kind {
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return math.U256Bytes(new(big.Int).SetUint64(value.Uint()))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return math.U256Bytes(big.NewInt(value.Int()))
	case reflect.Ptr:
		// 指针类型，假定为 *big.Int
		return math.U256Bytes(new(big.Int).Set(value.Interface().(*big.Int)))
	default:
		panic("abi: fatal error") // abi：致命错误
	}
}
