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
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/ethereum/go-ethereum/common"
)

// Type enumerator
// 类型枚举器
const (
	IntTy        byte = iota // 有符号整型
	UintTy                     // 无符号整型
	BoolTy                     // 布尔型
	StringTy                   // 字符串
	SliceTy                    // 切片
	ArrayTy                    // 数组
	TupleTy                    // 元组 (结构体)
	AddressTy                  // 地址类型
	FixedBytesTy               // 固定长度字节数组
	BytesTy                    // 动态长度字节数组
	HashTy                     // 哈希类型
	FixedPointTy               // 定点数类型
	FunctionTy                 // 函数类型
)

// Type is the reflection of the supported argument type.
// Type 是支持的参数类型的反射表示。
type Type struct {
	Elem *Type // 嵌套元素类型（用于数组/切片）
	Size int    // 类型大小（例如 uint256 的 size 是 256，bytes32 的 size 是 32）
	T    byte   // 我们自己的类型检查，使用上面的枚举器

	stringKind string // 保存用于派生签名的未解析字符串

	// Tuple relative fields
	// 元组相关字段
	TupleRawName  string       // 源代码中定义的原始结构体名称，可能为空。
	TupleElems    []*Type      // 所有元组字段的类型信息
	TupleRawNames []string     // 所有元组字段的原始字段名称
	TupleType     reflect.Type // 元组的底层结构体类型
}

var (
	// typeRegex parses the abi sub types
	// typeRegex 解析 abi 子类型
	typeRegex = regexp.MustCompile("([a-zA-Z]+)(([0-9]+)(x([0-9]+))?)?")

	// sliceSizeRegex grab the slice size
	// sliceSizeRegex 获取切片/数组的大小
	sliceSizeRegex = regexp.MustCompile("[0-9]+")
)

// NewType creates a new reflection type of abi type given in t.
// NewType 根据给定的 t 创建一个新的 abi 类型的反射类型。
func NewType(t string, internalType string, components []ArgumentMarshaling) (typ Type, err error) {
	// check that array brackets are equal if they exist
	// 检查数组的方括号是否匹配
	if strings.Count(t, "[") != strings.Count(t, "]") {
		return Type{}, errors.New("invalid arg type in abi") // abi 中无效的参数类型
	}
	typ.stringKind = t

	// if there are brackets, get ready to go into slice/array mode and
	// recursively create the type
	// 如果有方括号，就进入切片/数组模式并递归地创建类型
	if strings.Count(t, "[") != 0 {
		// Note internalType can be empty here.
		// 注意 internalType 在这里可能为空。
		subInternal := internalType
		if i := strings.LastIndex(internalType, "["); i != -1 {
			subInternal = subInternal[:i]
		}
		// recursively embed the type
		// 递归地嵌入类型
		i := strings.LastIndex(t, "[")
		embeddedType, err := NewType(t[:i], subInternal, components)
		if err != nil {
			return Type{}, err
		}
		// grab the last cell and create a type from there
		// 获取最后一个单元并从中创建类型
		sliced := t[i:]
		// grab the slice size with regexp
		// 用正则表达式获取切片大小
		intz := sliceSizeRegex.FindAllString(sliced, -1)

		if len(intz) == 0 {
			// is a slice (e.g., "[]")
			// 是一个切片
			typ.T = SliceTy
			typ.Elem = &embeddedType
			typ.stringKind = embeddedType.stringKind + sliced
		} else if len(intz) == 1 {
			// is an array (e.g., "[3]")
			// 是一个数组
			typ.T = ArrayTy
			typ.Elem = &embeddedType
			typ.Size, err = strconv.Atoi(intz[0])
			if err != nil {
				return Type{}, fmt.Errorf("abi: error parsing variable size: %v", err) // abi：解析可变大小时出错
			}
			typ.stringKind = embeddedType.stringKind + sliced
		} else {
			return Type{}, errors.New("invalid formatting of array type") // 无效的数组类型格式
		}
		return typ, err
	}
	// parse the type and size of the abi-type.
	// 解析 abi 类型的类型和大小。
	matches := typeRegex.FindAllStringSubmatch(t, -1)
	if len(matches) == 0 {
		return Type{}, fmt.Errorf("invalid type '%v'", t) // 无效类型
	}
	parsedType := matches[0]

	// varSize is the size of the variable
	// varSize 是变量的大小
	var varSize int
	if len(parsedType[3]) > 0 {
		var err error
		varSize, err = strconv.Atoi(parsedType[2])
		if err != nil {
			return Type{}, fmt.Errorf("abi: error parsing variable size: %v", err) // abi: 解析变量大小时出错
		}
	} else {
		if parsedType[0] == "uint" || parsedType[0] == "int" {
			// this should fail because it means that there's something wrong with
			// the abi type (the compiler should always format it to the size...always)
			// 这应该会失败，因为这意味着 abi 类型有问题（编译器应该总是将其格式化为大小...总是）
			return Type{}, fmt.Errorf("unsupported arg type: %s", t) // 不支持的参数类型
		}
	}
	// varType is the parsed abi type
	// varType 是解析后的 abi 类型
	switch varType := parsedType[1]; varType {
	case "int":
		typ.Size = varSize
		typ.T = IntTy
	case "uint":
		typ.Size = varSize
		typ.T = UintTy
	case "bool":
		typ.T = BoolTy
	case "address":
		typ.Size = 20
		typ.T = AddressTy
	case "string":
		typ.T = StringTy
	case "bytes":
		if varSize == 0 {
			// 动态大小的 bytes
			typ.T = BytesTy
		} else {
			if varSize > 32 {
				return Type{}, fmt.Errorf("unsupported arg type: %s", t) // 不支持的参数类型
			}
			// 固定大小的 bytes, e.g., bytes32
			typ.T = FixedBytesTy
			typ.Size = varSize
		}
	case "tuple":
		var (
			fields     []reflect.StructField // Go 结构体字段
			elems      []*Type             // 元组元素的 ABI 类型
			names      []string            // 元组元素的原始名称
			expression string              // 规范的参数表达式, e.g., "(uint256,string)"
			used       = make(map[string]bool) // 用于处理字段名冲突
		)
		expression += "("
		for idx, c := range components {
			cType, err := NewType(c.Type, c.InternalType, c.Components)
			if err != nil {
				return Type{}, err
			}
			name := ToCamelCase(c.Name)
			if name == "" {
				return Type{}, errors.New("abi: purely anonymous or underscored field is not supported") // abi: 不支持纯匿名或下划线开头的字段
			}
			fieldName := ResolveNameConflict(name, func(s string) bool { return used[s] })
			used[fieldName] = true
			if !isValidFieldName(fieldName) {
				return Type{}, fmt.Errorf("field %d has invalid name", idx) // 字段 %d 的名称无效
			}
			fields = append(fields, reflect.StructField{
				Name: fieldName, // reflect.StructOf 对任何未导出的字段都会 panic。
				Type: cType.GetType(),
				Tag:  reflect.StructTag("json:\"" + c.Name + "\""),
			})
			elems = append(elems, &cType)
			names = append(names, c.Name)
			expression += cType.stringKind
			if idx != len(components)-1 {
				expression += ","
			}
		}
		expression += ")"

		typ.TupleType = reflect.StructOf(fields)
		typ.TupleElems = elems
		typ.TupleRawNames = names
		typ.T = TupleTy
		typ.stringKind = expression

		const structPrefix = "struct "
		// After solidity 0.5.10, a new field of abi "internalType"
		// is introduced. From that we can obtain the struct name
		// user defined in the source code.
		// 在 solidity 0.5.10 之后，引入了一个新的 abi 字段 "internalType"。
		// 从中我们可以获取用户在源代码中定义的结构体名称。
		if internalType != "" && strings.HasPrefix(internalType, structPrefix) {
			// Foo.Bar type definition is not allowed in golang,
			// convert the format to FooBar
			// Go 中不允许 Foo.Bar 这样的类型定义，
			// 将其格式转换为 FooBar
			typ.TupleRawName = strings.ReplaceAll(internalType[len(structPrefix):], ".", "")
		}

	case "function":
		typ.T = FunctionTy
		typ.Size = 24 // 20 字节地址 + 4 字节函数选择器
	default:
		if strings.HasPrefix(internalType, "contract ") {
			// 合约类型在 ABI 中表示为地址
			typ.Size = 20
			typ.T = AddressTy
		} else {
			return Type{}, fmt.Errorf("unsupported arg type: %s", t) // 不支持的参数类型
		}
	}

	return
}

// GetType returns the reflection type of the ABI type.
// GetType 返回 ABI 类型的 Go 反射类型。
func (t Type) GetType() reflect.Type {
	switch t.T {
	case IntTy:
		return reflectIntType(false, t.Size)
	case UintTy:
		return reflectIntType(true, t.Size)
	case BoolTy:
		return reflect.TypeFor[bool]()
	case StringTy:
		return reflect.TypeFor[string]()
	case SliceTy:
		return reflect.SliceOf(t.Elem.GetType())
	case ArrayTy:
		return reflect.ArrayOf(t.Size, t.Elem.GetType())
	case TupleTy:
		return t.TupleType
	case AddressTy:
		return reflect.TypeFor[common.Address]()
	case FixedBytesTy:
		return reflect.ArrayOf(t.Size, reflect.TypeFor[byte]())
	case BytesTy:
		return reflect.TypeFor[[]byte]()
	case HashTy, FixedPointTy: // currently not used (当前未使用)
		return reflect.TypeFor[[32]byte]()
	case FunctionTy:
		return reflect.TypeFor[[24]byte]()
	default:
		panic("Invalid type") // 无效类型
	}
}

// String implements Stringer.
// String 实现了 Stringer 接口，返回类型的字符串表示形式。
func (t Type) String() (out string) {
	return t.stringKind
}

// pack 对值 v 进行 ABI 编码
func (t Type) pack(v reflect.Value) ([]byte, error) {
	// dereference pointer first if it's a pointer
	// 如果是指针，首先解引用
	v = indirect(v)
	if err := typeCheck(t, v); err != nil {
		return nil, err
	}

	switch t.T {
	case SliceTy, ArrayTy:
		var ret []byte

		if t.requiresLengthPrefix() {
			// append length
			// 对动态数组/切片，前缀添加长度
			ret = append(ret, packNum(reflect.ValueOf(v.Len()))...)
		}

		// calculate offset if any
		// 计算偏移量（如果需要的话）
		offset := 0
		// 如果元素是动态类型，则需要计算偏移量
		offsetReq := isDynamicType(*t.Elem)
		if offsetReq {
			offset = getTypeSize(*t.Elem) * v.Len()
		}
		var tail []byte // 用于存储动态类型的数据
		for i := 0; i < v.Len(); i++ {
			val, err := t.Elem.pack(v.Index(i))
			if err != nil {
				return nil, err
			}
			if !offsetReq {
				// 如果是静态类型，直接追加编码后的数据
				ret = append(ret, val...)
				continue
			}
			// 如果是动态类型，先追加偏移量，然后将数据暂存到 tail
			ret = append(ret, packNum(reflect.ValueOf(offset))...)
			offset += len(val)
			tail = append(tail, val...)
		}
		return append(ret, tail...), nil
	case TupleTy:
		// (T1,...,Tk) for k >= 0 and any types T1, …, Tk
		// enc(X) = head(X(1)) ... head(X(k)) tail(X(1)) ... tail(X(k))
		// where X = (X(1), ..., X(k)) and head and tail are defined for Ti being a static
		// type as
		//     head(X(i)) = enc(X(i)) and tail(X(i)) = "" (the empty string)
		// and as
		//     head(X(i)) = enc(len(head(X(1)) ... head(X(k)) tail(X(1)) ... tail(X(i-1))))
		//     tail(X(i)) = enc(X(i))
		// otherwise, i.e. if Ti is a dynamic type.
		// 元组编码规则：
		// 编码结果是 head(X(1))...head(X(k))tail(X(1))...tail(X(k))
		// 对于静态类型 Ti，head(X(i)) = enc(X(i))，tail(X(i)) 为空
		// 对于动态类型 Ti，head(X(i)) 是一个偏移量，指向 tail(X(i)) 的起始位置，而 tail(X(i)) = enc(X(i))
		fieldmap, err := mapArgNamesToStructFields(t.TupleRawNames, v)
		if err != nil {
			return nil, err
		}
		// Calculate prefix occupied size.
		// 计算头部占用的总大小。
		offset := 0
		for _, elem := range t.TupleElems {
			offset += getTypeSize(*elem)
		}
		var ret, tail []byte
		for i, elem := range t.TupleElems {
			field := v.FieldByName(fieldmap[t.TupleRawNames[i]])
			if !field.IsValid() {
				return nil, fmt.Errorf("field %s for tuple not found in the given struct", t.TupleRawNames[i]) // 在给定的结构体中找不到元组的字段 %s
			}
			val, err := elem.pack(field)
			if err != nil {
				return nil, err
			}
			if isDynamicType(*elem) {
				// 如果是动态类型，在头部放入偏移量，将实际数据追加到尾部
				ret = append(ret, packNum(reflect.ValueOf(offset))...)
				tail = append(tail, val...)
				offset += len(val)
			} else {
				// 如果是静态类型，直接将数据放入头部
				ret = append(ret, val...)
			}
		}
		return append(ret, tail...), nil

	default:
		return packElement(t, v)
	}
}

// requiresLengthPrefix returns whether the type requires any sort of length
// prefixing.
// requiresLengthPrefix 返回该类型是否需要任何形式的长度前缀。
func (t Type) requiresLengthPrefix() bool {
	return t.T == StringTy || t.T == BytesTy || t.T == SliceTy
}

// isDynamicType returns true if the type is dynamic.
// The following types are called “dynamic”:
// * bytes
// * string
// * T[] for any T
// * T[k] for any dynamic T and any k >= 0
// * (T1,...,Tk) if Ti is dynamic for some 1 <= i <= k
// isDynamicType 检查一个类型是否是动态的。
// 以下类型被称为“动态类型”：
// * bytes
// * string
// * 任意 T 的 T[]
// * 任意动态 T 和任意 k >= 0 的 T[k]
// * (T1,...,Tk)，如果对于某个 1 <= i <= k，Ti 是动态的
func isDynamicType(t Type) bool {
	if t.T == TupleTy {
		for _, elem := range t.TupleElems {
			if isDynamicType(*elem) {
				return true
			}
		}
		return false
	}
	return t.T == StringTy || t.T == BytesTy || t.T == SliceTy || (t.T == ArrayTy && isDynamicType(*t.Elem))
}

// getTypeSize returns the size that this type needs to occupy.
// We distinguish static and dynamic types. Static types are encoded in-place
// and dynamic types are encoded at a separately allocated location after the
// current block.
// So for a static variable, the size returned represents the size that the
// variable actually occupies.
// For a dynamic variable, the returned size is fixed 32 bytes, which is used
// to store the location reference for actual value storage.
// getTypeSize 返回此类型在 ABI 编码的头部所占用的字节大小。
// 我们区分静态类型和动态类型。静态类型是原地编码的，
// 而动态类型是在当前块之后的一个单独分配的位置进行编码的。
// 因此，对于静态变量，返回的大小表示变量实际占用的空间。
// 对于动态变量，返回的大小是固定的 32 字节，用于存储指向实际值存储位置的引用（偏移量）。
func getTypeSize(t Type) int {
	if t.T == ArrayTy && !isDynamicType(*t.Elem) {
		// Recursively calculate type size if it is a nested array
		// 如果是嵌套数组，则递归计算类型大小
		if t.Elem.T == ArrayTy || t.Elem.T == TupleTy {
			return t.Size * getTypeSize(*t.Elem)
		}
		return t.Size * 32
	} else if t.T == TupleTy && !isDynamicType(t) {
		// 静态元组的大小是其所有元素大小的总和
		total := 0
		for _, elem := range t.TupleElems {
			total += getTypeSize(*elem)
		}
		return total
	}
	// 动态类型和基本静态类型在头部都占用 32 字节
	return 32
}

// isLetter reports whether a given 'rune' is classified as a Letter.
// This method is copied from reflect/type.go
// isLetter 报告给定的 'rune' 是否被归类为字母。
// 此方法从 reflect/type.go 复制而来。
func isLetter(ch rune) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_' || ch >= utf8.RuneSelf && unicode.IsLetter(ch)
}

// isValidFieldName checks if a string is a valid (struct) field name or not.
//
// According to the language spec, a field name should be an identifier.
//
// identifier = letter { letter | unicode_digit } .
// letter = unicode_letter | "_" .
// This method is copied from reflect/type.go
// isValidFieldName 检查一个字符串是否是有效的（结构体）字段名。
//
// 根据语言规范，字段名应该是一个标识符。
//
// identifier = letter { letter | unicode_digit } .
// letter = unicode_letter | "_" .
// 此方法从 reflect/type.go 复制而来。
func isValidFieldName(fieldName string) bool {
	for i, c := range fieldName {
		if i == 0 && !isLetter(c) {
			return false
		}

		if !(isLetter(c) || unicode.IsDigit(c)) {
			return false
		}
	}

	return len(fieldName) > 0
}
