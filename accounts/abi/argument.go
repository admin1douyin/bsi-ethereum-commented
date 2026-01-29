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

package abi

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
)

// Argument holds the name of the argument and the corresponding type.
// Types are used when packing and testing arguments.
// Argument 结构体保存了参数的名称和对应的类型。
// 类型在打包和测试参数时使用。
type Argument struct {
	Name    string // 参数名称
	Type    Type   // 参数类型
	Indexed bool   // indexed 仅用于事件，表示该参数是否被索引
}

// Arguments 是 Argument 的切片。
type Arguments []Argument

// ArgumentMarshaling 用于辅助 Argument 的 JSON 解组。
type ArgumentMarshaling struct {
	Name         string               // 参数名称
	Type         string               // 参数的 ABI 类型字符串
	InternalType string               // 参数的内部类型（可选）
	Components   []ArgumentMarshaling // 用于元组类型的子组件
	Indexed      bool                 // 是否被索引
}

// UnmarshalJSON implements json.Unmarshaler interface.
// UnmarshalJSON 实现了 json.Unmarshaler 接口，用于自定义 JSON 反序列化。
func (argument *Argument) UnmarshalJSON(data []byte) error {
	var arg ArgumentMarshaling
	err := json.Unmarshal(data, &arg)
	if err != nil {
		return fmt.Errorf("argument json err: %v", err)
	}

	// 从解析出的字符串和组件创建 Type 对象
	argument.Type, err = NewType(arg.Type, arg.InternalType, arg.Components)
	if err != nil {
		return err
	}
	argument.Name = arg.Name
	argument.Indexed = arg.Indexed

	return nil
}

// NonIndexed returns the arguments with indexed arguments filtered out.
// NonIndexed 返回过滤掉索引参数后的参数列表。
func (arguments Arguments) NonIndexed() Arguments {
	var ret []Argument
	for _, arg := range arguments {
		if !arg.Indexed {
			ret = append(ret, arg)
		}
	}
	return ret
}

// isTuple returns true for non-atomic constructs, like (uint,uint) or uint[].
// isTuple 判断参数列表是否为非原子结构，例如元组 (uint,uint) 或数组 uint[]。
// 拥有多个参数即被认为是元组。
func (arguments Arguments) isTuple() bool {
	return len(arguments) > 1
}

// Unpack performs the operation hexdata -> Go format.
// Unpack 将 ABI 编码的十六进制数据解码为 Go 类型的值。
func (arguments Arguments) Unpack(data []byte) ([]any, error) {
	if len(data) == 0 {
		if len(arguments.NonIndexed()) != 0 {
			return nil, errors.New("abi: attempting to unmarshal an empty string while arguments are expected") // 错误：期望有参数但输入数据为空
		}
		return make([]any, 0), nil
	}
	return arguments.UnpackValues(data)
}

// UnpackIntoMap performs the operation hexdata -> mapping of argument name to argument value.
// UnpackIntoMap 将 ABI 编码的十六进制数据解码到一个 map[string]any 中，键为参数名，值为参数值。
func (arguments Arguments) UnpackIntoMap(v map[string]any, data []byte) error {
	// Make sure map is not nil
	if v == nil {
		return errors.New("abi: cannot unpack into a nil map") // 错误：不能解包到 nil map
	}
	if len(data) == 0 {
		if len(arguments.NonIndexed()) != 0 {
			return errors.New("abi: attempting to unmarshal an empty string while arguments are expected") // 错误：期望有参数但输入数据为空
		}
		return nil // Nothing to unmarshal, return
	}
	marshalledValues, err := arguments.UnpackValues(data)
	if err != nil {
		return err
	}
	for i, arg := range arguments.NonIndexed() {
		v[arg.Name] = marshalledValues[i]
	}
	return nil
}

// Copy performs the operation go format -> provided struct.
// Copy 将解包后的 Go 值复制到提供的结构体或切片 v 中。
func (arguments Arguments) Copy(v any, values []any) error {
	// make sure the passed value is arguments pointer
	if reflect.Ptr != reflect.ValueOf(v).Kind() {
		return fmt.Errorf("abi: Unpack(non-pointer %T)", v) // 错误：v 必须是一个指针
	}
	if len(values) == 0 {
		if len(arguments.NonIndexed()) != 0 {
			return errors.New("abi: attempting to copy no values while arguments are expected") // 错误：期望有参数但没有值可以复制
		}
		return nil // Nothing to copy, return
	}
	if arguments.isTuple() {
		return arguments.copyTuple(v, values)
	}
	return arguments.copyAtomic(v, values[0])
}

// copyAtomic copies ( hexdata -> go ) a single value
// copyAtomic 复制单个原子值。
func (arguments Arguments) copyAtomic(v any, marshalledValues any) error {
	dst := reflect.ValueOf(v).Elem()
	src := reflect.ValueOf(marshalledValues)

	if dst.Kind() == reflect.Struct {
		return set(dst.Field(0), src) // 如果目标是结构体，设置其第一个字段
	}
	return set(dst, src)
}

// copyTuple copies a batch of values from marshalledValues to v.
// copyTuple 将一组值从 marshalledValues 复制到目标 v（结构体或切片）。
func (arguments Arguments) copyTuple(v any, marshalledValues []any) error {
	value := reflect.ValueOf(v).Elem()
	nonIndexedArgs := arguments.NonIndexed()

	switch value.Kind() {
	case reflect.Struct:
		argNames := make([]string, len(nonIndexedArgs))
		for i, arg := range nonIndexedArgs {
			argNames[i] = arg.Name
		}
		var err error
		abi2struct, err := mapArgNamesToStructFields(argNames, value) // 将 ABI 参数名映射到结构体字段名
		if err != nil {
			return err
		}
		for i, arg := range nonIndexedArgs {
			field := value.FieldByName(abi2struct[arg.Name])
			if !field.IsValid() {
				return fmt.Errorf("abi: field %s can't be found in the given value", arg.Name) // 错误：在目标值中找不到字段
			}
			if err := set(field, reflect.ValueOf(marshalledValues[i])); err != nil {
				return err
			}
		}
	case reflect.Slice, reflect.Array:
		if value.Len() < len(marshalledValues) {
			return fmt.Errorf("abi: insufficient number of arguments for unpack, want %d, got %d", len(arguments), value.Len()) // 错误：目标切片/数组长度不足
		}
		for i := range nonIndexedArgs {
			if err := set(value.Index(i), reflect.ValueOf(marshalledValues[i])); err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("abi:[2] cannot unmarshal tuple in to %v", value.Type()) // 错误：无法将元组解组到指定类型
	}
	return nil
}

// UnpackValues can be used to unpack ABI-encoded hexdata according to the ABI-specification,
// without supplying a struct to unpack into. Instead, this method returns a list containing the
// values. An atomic argument will be a list with one element.
// UnpackValues 根据 ABI 规范解包 ABI 编码的十六进制数据，
// 无需提供用于解包的结构体。相反，此方法返回一个包含值的列表。
// 单个原子参数将是带有一个元素的列表。
func (arguments Arguments) UnpackValues(data []byte) ([]any, error) {
	var (
		retval      = make([]any, 0) // 返回值列表
		virtualArgs = 0              // 用于计算静态数组和元组的偏移量
		index       = 0              // 当前处理的非索引参数的索引
	)

	for _, arg := range arguments {
		if arg.Indexed {
			continue // 跳过索引参数
		}
		marshalledValue, err := toGoType((index+virtualArgs)*32, arg.Type, data)
		if err != nil {
			return nil, err
		}
		if arg.Type.T == ArrayTy && !isDynamicType(arg.Type) {
			// If we have a static array, like [3]uint256, these are coded as
			// just like uint256,uint256,uint256.
			// This means that we need to add two 'virtual' arguments when
			// we count the index from now on.
			//
			// Array values nested multiple levels deep are also encoded inline:
			// [2][3]uint256: uint256,uint256,uint256,uint256,uint256,uint256
			//
			// Calculate the full array size to get the correct offset for the next argument.
			// Decrement it by 1, as the normal index increment is still applied.
			// 如果我们有一个静态数组，例如 [3]uint256，它们的编码方式
			// 就如同 uint256,uint256,uint256。
			// 这意味着从现在开始计算索引时，我们需要添加两个“虚拟”参数。
			//
			// 嵌套多层的数组值也是内联编码的：
			// [2][3]uint256: uint256,uint256,uint256,uint256,uint256,uint256
			//
			// 计算完整的数组大小以获取下一个参数的正确偏移量。
			// 将其减 1，因为正常的索引增量仍然适用。
			virtualArgs += getTypeSize(arg.Type)/32 - 1
		} else if arg.Type.T == TupleTy && !isDynamicType(arg.Type) {
			// If we have a static tuple, like (uint256, bool, uint256), these are
			// coded as just like uint256,bool,uint256
			// 如果我们有一个静态元组，例如 (uint256, bool, uint256)，它们的编码
			// 就如同 uint256,bool,uint256
			virtualArgs += getTypeSize(arg.Type)/32 - 1
		}
		retval = append(retval, marshalledValue)
		index++
	}
	return retval, nil
}

// PackValues performs the operation Go format -> Hexdata.
// It is the semantic opposite of UnpackValues.
// PackValues 执行 Go 类型 -> 十六进制数据的操作。
// 它是 UnpackValues 的语义逆操作。
func (arguments Arguments) PackValues(args []any) ([]byte, error) {
	return arguments.Pack(args...)
}

// Pack performs the operation Go format -> Hexdata.
// Pack 执行 Go 类型 -> 十六进制数据的操作。
func (arguments Arguments) Pack(args ...any) ([]byte, error) {
	// Make sure arguments match up and pack them
	abiArgs := arguments
	if len(args) != len(abiArgs) {
		return nil, fmt.Errorf("argument count mismatch: got %d for %d", len(args), len(abiArgs)) // 错误：参数数量不匹配
	}
	// variableInput 是追加在打包输出末尾的输出。
	// 这用于字符串和字节类型的输入。
	var variableInput []byte

	// inputOffset 是打包输出的字节偏移量
	inputOffset := 0
	for _, abiArg := range abiArgs {
		inputOffset += getTypeSize(abiArg.Type)
	}
	var ret []byte
	for i, a := range args {
		input := abiArgs[i]
		// pack the input
		// 打包输入
		packed, err := input.Type.pack(reflect.ValueOf(a))
		if err != nil {
			return nil, err
		}
		// check for dynamic types
		// 检查动态类型
		if isDynamicType(input.Type) {
			// set the offset
			// 设置偏移量
			ret = append(ret, packNum(reflect.ValueOf(inputOffset))...)
			// calculate next offset
			// 计算下一个偏移量
			inputOffset += len(packed)
			// append to variable input
			// 追加到可变输入
			variableInput = append(variableInput, packed...)
		} else {
			// append the packed value to the input
			// 将打包后的值追加到输入
			ret = append(ret, packed...)
		}
	}
	// append the variable input at the end of the packed input
	// 将可变输入追加到打包输入的末尾
	ret = append(ret, variableInput...)

	return ret, nil
}

// ToCamelCase converts an under-score string to a camel-case string
// ToCamelCase 将下划线分隔的字符串转换为驼峰式字符串。
func ToCamelCase(input string) string {
	parts := strings.Split(input, "_")
	for i, s := range parts {
		if len(s) > 0 {
			parts[i] = strings.ToUpper(s[:1]) + s[1:]
		}
	}
	return strings.Join(parts, "")
}
