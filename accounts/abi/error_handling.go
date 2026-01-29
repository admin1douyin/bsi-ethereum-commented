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
	"reflect"
)

// 定义了一系列关于 ABI 编码的错误变量。
var (
	errBadBool     = errors.New("abi: improperly encoded boolean value") // abi: 布尔值编码不当
	errBadUint8    = errors.New("abi: improperly encoded uint8 value")   // abi: uint8 值编码不当
	errBadUint16   = errors.New("abi: improperly encoded uint16 value")  // abi: uint16 值编码不当
	errBadUint32   = errors.New("abi: improperly encoded uint32 value")  // abi: uint32 值编码不当
	errBadUint64   = errors.New("abi: improperly encoded uint64 value")  // abi: uint64 值编码不当
	errBadInt8     = errors.New("abi: improperly encoded int8 value")    // abi: int8 值编码不当
	errBadInt16    = errors.New("abi: improperly encoded int16 value")   // abi: int16 值编码不当
	errBadInt32    = errors.New("abi: improperly encoded int32 value")   // abi: int32 值编码不当
	errBadInt64    = errors.New("abi: improperly encoded int64 value")   // abi: int64 值编码不当
	errInvalidSign = errors.New("abi: negatively-signed value cannot be packed into uint parameter") // abi: 负数值不能打包到 uint 参数中
)

// formatSliceString formats the reflection kind with the given slice size
// and returns a formatted string representation.
// formatSliceString 使用给定的切片大小格式化反射类型，
// 并返回格式化的字符串表示形式。
func formatSliceString(kind reflect.Kind, sliceSize int) string {
	if sliceSize == -1 {
		// 动态大小的切片表示为 "[]T"
		return fmt.Sprintf("[]%v", kind)
	}
	// 固定大小的数组表示为 "[N]T"
	return fmt.Sprintf("[%d]%v", sliceSize, kind)
}

// sliceTypeCheck checks that the given slice can by assigned to the reflection
// type in t.
// sliceTypeCheck 检查给定的切片是否可以赋值给 t 中的反射类型。
func sliceTypeCheck(t Type, val reflect.Value) error {
	if val.Kind() != reflect.Slice && val.Kind() != reflect.Array {
		// 值必须是切片或数组
		return typeErr(formatSliceString(t.GetType().Kind(), t.Size), val.Type())
	}

	if t.T == ArrayTy && val.Len() != t.Size {
		// 对于固定大小的数组，长度必须匹配
		return typeErr(formatSliceString(t.Elem.GetType().Kind(), t.Size), formatSliceString(val.Type().Elem().Kind(), val.Len()))
	}

	if t.Elem.T == SliceTy || t.Elem.T == ArrayTy {
		// 递归检查多维切片/数组的元素类型
		if val.Len() > 0 {
			return sliceTypeCheck(*t.Elem, val.Index(0))
		}
	}

	if val.Type().Elem().Kind() != t.Elem.GetType().Kind() {
		// 检查元素的基础类型是否匹配
		return typeErr(formatSliceString(t.Elem.GetType().Kind(), t.Size), val.Type())
	}
	return nil
}

// typeCheck checks that the given reflection value can be assigned to the reflection
// type in t.
// typeCheck 检查给定的反射值是否可以赋值给 t 中的反射类型。
func typeCheck(t Type, value reflect.Value) error {
	if t.T == SliceTy || t.T == ArrayTy {
		// 如果是切片或数组类型，调用 sliceTypeCheck
		return sliceTypeCheck(t, value)
	}

	// Check base type validity. Element types will be checked later on.
	// 检查基本类型的有效性。元素类型将在稍后检查。
	if t.GetType().Kind() != value.Kind() {
		// 检查 Go 的 Kind 是否匹配
		return typeErr(t.GetType().Kind(), value.Kind())
	} else if t.T == FixedBytesTy && t.Size != value.Len() {
		// 对于固定字节数组，长度必须匹配
		return typeErr(t.GetType(), value.Type())
	} else {
		return nil
	}
}

// typeErr returns a formatted type casting error.
// typeErr 返回一个格式化的类型转换错误。
func typeErr(expected, got interface{}) error {
	return fmt.Errorf("abi: cannot use %v as type %v as argument", got, expected) // abi: 无法将 %v 用作 %v 类型的参数
}
