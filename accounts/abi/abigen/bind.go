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

// Package abigen generates Ethereum contract Go bindings.
//
// Detailed usage document and tutorial available on the go-ethereum Wiki page:
// https://geth.ethereum.org/docs/developers/dapp-developer/native-bindings

// abigen 包用于生成以太坊合约的 Go 语言绑定。
//
// 详细的使用文档和教程可以在 go-ethereum Wiki 页面上找到：
// https://geth.ethereum.org/docs/developers/dapp-developer/native-bindings
package abigen

import (
	"bytes"
	"fmt"
	"go/format"
	"regexp"
	"strings"
	"text/template"
	"unicode"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/log"
)

var (
	// intRegex 用于匹配 (u)int<M> 类型的正则表达式
	intRegex = regexp.MustCompile(`(u)?int([0-9]*)`)
)

// isKeyWord 检查一个字符串是否是 Go 语言的关键字。
func isKeyWord(arg string) bool {
	switch arg {
	case "break":
	case "case":
	case "chan":
	case "const":
	case "continue":
	case "default":
	case "defer":
	case "else":
	case "fallthrough":
	case "for":
	case "func":
	case "go":
	case "goto":
	case "if":
	case "import":
	case "interface":
	case "iota":
	case "map":
	case "make":
	case "new":
	case "package":
	case "range":
	case "return":
	case "select":
	case "struct":
	case "switch":
	case "type":
	case "var":
	default:
		return false
	}

	return true
}

// Bind generates a Go wrapper around a contract ABI. This wrapper isn't meant
// to be used as is in client code, but rather as an intermediate struct which
// enforces compile time type safety and naming convention as opposed to having to
// manually maintain hard coded strings that break on runtime.
// Bind 函数围绕合约 ABI 生成一个 Go 语言的包装器。这个包装器不打算直接在客户端代码中使用，
// 而是作为一个中间结构体，它强制执行编译时的类型安全和命名约定，
// 而不是手动维护在运行时才会中断的硬编码字符串。
//
// types: 合约类型名称列表。
// abis: JSON 格式的 ABI 字符串列表。
// bytecodes: 合约字节码的十六进制字符串列表。
// fsigs: 函数签名的映射列表。
// pkg: 生成的 Go 文件的包名。
// libs: 库链接的映射。
// aliases: 方法和事件名称的别名映射。
// return: 返回生成的 Go 代码字符串和可能的错误。
func Bind(types []string, abis []string, bytecodes []string, fsigs []map[string]string, pkg string, libs map[string]string, aliases map[string]string) (string, error) {
	var (
		// contracts 是为每个请求绑定的独立合约创建的映射。
		contracts = make(map[string]*tmplContract)

		// structs 是由传入的合约共享的所有重新声明的结构体的映射。
		structs = make(map[string]*tmplStruct)

		// isLib 用于标记遇到的每个库的映射。
		isLib = make(map[string]struct{})
	)
	for i := 0; i < len(types); i++ {
		// Parse the actual ABI to generate the binding for
		// 解析实际的 ABI 以生成绑定
		evmABI, err := abi.JSON(strings.NewReader(abis[i]))
		if err != nil {
			return "", err
		}
		// Strip any whitespace from the JSON ABI
		// 从 JSON ABI 中删除所有空白字符
		strippedABI := strings.Map(func(r rune) rune {
			if unicode.IsSpace(r) {
				return -1
			}
			return r
		}, abis[i])

		// Extract the call and transact methods; events, struct definitions; and sort them alphabetically
		// 提取调用和交易方法、事件、结构体定义，并按字母顺序对它们进行排序
		var (
			calls     = make(map[string]*tmplMethod)
			transacts = make(map[string]*tmplMethod)
			events    = make(map[string]*tmplEvent)
			fallback  *tmplMethod
			receive   *tmplMethod

			// identifiers are used to detect duplicated identifiers of functions
			// and events. For all calls, transacts and events, abigen will generate
			// corresponding bindings. However we have to ensure there is no
			// identifier collisions in the bindings of these categories.
			// identifiers 用于检测函数和事件的重复标识符。
			// 对于所有的调用、交易和事件，abigen 都会生成相应的绑定。
			// 但是，我们必须确保在这些类别的绑定中没有标识符冲突。
			callIdentifiers     = make(map[string]bool)
			transactIdentifiers = make(map[string]bool)
			eventIdentifiers    = make(map[string]bool)
		)

		// 处理构造函数的输入参数，提取其中包含的结构体类型
		for _, input := range evmABI.Constructor.Inputs {
			if hasStruct(input.Type) {
				bindStructType(input.Type, structs)
			}
		}
		
		// 遍历 ABI 中的所有方法
		for _, original := range evmABI.Methods {
			// Normalize the method for capital cases and non-anonymous inputs/outputs
			// 对方法进行规范化，以处理大写情况和非匿名输入/输出
			normalized := original
			// 将方法名转换为驼峰式，并应用别名
			normalizedName := abi.ToCamelCase(alias(aliases, original.Name))
			// Ensure there is no duplicated identifier
			// 确保没有重复的标识符
			var identifiers = callIdentifiers
			if !original.IsConstant() { // 如果是交易型方法
				identifiers = transactIdentifiers
			}
			// Name shouldn't start with a digit. It will make the generated code invalid.
			// 名称不应以数字开头。这会使生成的代码无效。
			if len(normalizedName) > 0 && unicode.IsDigit(rune(normalizedName[0])) {
				normalizedName = fmt.Sprintf("M%s", normalizedName)
				normalizedName = abi.ResolveNameConflict(normalizedName, func(name string) bool {
					_, ok := identifiers[name]
					return ok
				})
			}
			if identifiers[normalizedName] { // 检查是否有名称冲突
				return "", fmt.Errorf("duplicated identifier \"%s\"(normalized \"%s\"), use --alias for renaming", original.Name, normalizedName)
			}
			identifiers[normalizedName] = true

			normalized.Name = normalizedName
			normalized.Inputs = make([]abi.Argument, len(original.Inputs))
			copy(normalized.Inputs, original.Inputs)
			// 规范化输入参数
			for j, input := range normalized.Inputs {
				if input.Name == "" || isKeyWord(input.Name) { // 如果参数名为空或是关键字，则生成一个
					normalized.Inputs[j].Name = fmt.Sprintf("arg%d", j)
				}
				if hasStruct(input.Type) { // 提取结构体类型
					bindStructType(input.Type, structs)
				}
			}
			normalized.Outputs = make([]abi.Argument, len(original.Outputs))
			copy(normalized.Outputs, original.Outputs)
			// 规范化输出参数
			for j, output := range normalized.Outputs {
				if output.Name != "" {
					normalized.Outputs[j].Name = abi.ToCamelCase(output.Name)
				}
				if hasStruct(output.Type) { // 提取结构体类型
					bindStructType(output.Type, structs)
				}
			}
			// Append the methods to the call or transact lists
			// 将方法追加到调用或交易列表中
			if original.IsConstant() { // 如果是只读方法
				calls[original.Name] = &tmplMethod{Original: original, Normalized: normalized, Structured: structured(original.Outputs)}
			} else { // 如果是交易方法
				transacts[original.Name] = &tmplMethod{Original: original, Normalized: normalized, Structured: structured(original.Outputs)}
			}
		}
		// 遍历 ABI 中的所有事件
		for _, original := range evmABI.Events {
			// Skip anonymous events as they don't support explicit filtering
			// 跳过匿名事件，因为它们不支持显式过滤
			if original.Anonymous {
				continue
			}
			// Normalize the event for capital cases and non-anonymous outputs
			// 对事件进行规范化，以处理大写情况和非匿名输出
			normalized := original

			// Ensure there is no duplicated identifier
			// 确保没有重复的标识符
			normalizedName := abi.ToCamelCase(alias(aliases, original.Name))
			// Name shouldn't start with a digit. It will make the generated code invalid.
			// 名称不应以数字开头。这会使生成的代码无效。
			if len(normalizedName) > 0 && unicode.IsDigit(rune(normalizedName[0])) {
				normalizedName = fmt.Sprintf("E%s", normalizedName)
				normalizedName = abi.ResolveNameConflict(normalizedName, func(name string) bool {
					_, ok := eventIdentifiers[name]
					return ok
				})
			}
			if eventIdentifiers[normalizedName] {
				return "", fmt.Errorf("duplicated identifier \"%s\"(normalized \"%s\"), use --alias for renaming", original.Name, normalizedName)
			}
			eventIdentifiers[normalizedName] = true
			normalized.Name = normalizedName

			used := make(map[string]bool)
			normalized.Inputs = make([]abi.Argument, len(original.Inputs))
			copy(normalized.Inputs, original.Inputs)
			// 规范化事件的输入参数
			for j, input := range normalized.Inputs {
				if input.Name == "" || isKeyWord(input.Name) {
					normalized.Inputs[j].Name = fmt.Sprintf("arg%d", j)
				}
				// Event is a bit special, we need to define event struct in binding,
				// ensure there is no camel-case-style name conflict.
				// 事件有点特殊，我们需要在绑定中定义事件结构体，
				// 确保没有驼峰式命名冲突。
				for index := 0; ; index++ {
					if !used[abi.ToCamelCase(normalized.Inputs[j].Name)] {
						used[abi.ToCamelCase(normalized.Inputs[j].Name)] = true
						break
					}
					normalized.Inputs[j].Name = fmt.Sprintf("%s%d", normalized.Inputs[j].Name, index)
				}
				if hasStruct(input.Type) {
					bindStructType(input.Type, structs)
				}
			}
			// Append the event to the accumulator list
			// 将事件追加到累加器列表
			events[original.Name] = &tmplEvent{Original: original, Normalized: normalized}
		}
		// Add two special fallback functions if they exist
		// 如果存在，则添加两个特殊的回退函数
		if evmABI.HasFallback() {
			fallback = &tmplMethod{Original: evmABI.Fallback}
		}
		if evmABI.HasReceive() {
			receive = &tmplMethod{Original: evmABI.Receive}
		}
		
		// 组装合约的模板数据
		contracts[types[i]] = &tmplContract{
			Type:        abi.ToCamelCase(types[i]),
			InputABI:    strings.ReplaceAll(strippedABI, "\"", "\\\""),
			InputBin:    strings.TrimPrefix(strings.TrimSpace(bytecodes[i]), "0x"),
			Constructor: evmABI.Constructor,
			Calls:       calls,
			Transacts:   transacts,
			Fallback:    fallback,
			Receive:     receive,
			Events:      events,
			Libraries:   make(map[string]string),
		}

		// Function 4-byte signatures are stored in the same sequence
		// as types, if available.
		// 函数的 4 字节签名（如果可用）与类型以相同的顺序存储。
		if len(fsigs) > i {
			contracts[types[i]].FuncSigs = fsigs[i]
		}
		// Parse library references.
		// 解析库引用。
		for pattern, name := range libs {
			matched, err := regexp.MatchString("__\$"+pattern+"\$__", contracts[types[i]].InputBin)
			if err != nil {
				log.Error("Could not search for pattern", "pattern", pattern, "contract", contracts[types[i]], "err", err)
			}
			if matched {
				contracts[types[i]].Libraries[pattern] = name
				// keep track that this type is a library
				// 跟踪此类型是一个库
				if _, ok := isLib[name]; !ok {
					isLib[name] = struct{}{}
				}
			}
		}
	}
	// Check if that type has already been identified as a library
	// 检查该类型是否已被识别为库
	for i := 0; i < len(types); i++ {
		_, ok := isLib[types[i]]
		contracts[types[i]].Library = ok
	}

	// Generate the contract template data content and render it
	// 生成合约模板数据内容并进行渲染
	data := &tmplData{
		Package:   pkg,
		Contracts: contracts,
		Libraries: libs,
		Structs:   structs,
	}
	buffer := new(bytes.Buffer)
	
	// 定义模板函数
	funcs := map[string]interface{}{
		"bindtype":      bindType,
		"bindtopictype": bindTopicType,
		"capitalise":    abi.ToCamelCase,
		"decapitalise":  decapitalise,
	}
	tmpl := template.Must(template.New("").Funcs(funcs).Parse(tmplSource))
	if err := tmpl.Execute(buffer, data); err != nil {
		return "", err
	}
	// Pass the code through gofmt to clean it up
	// 通过 gofmt 来清理代码
	code, err := format.Source(buffer.Bytes())
	if err != nil {
		return "", fmt.Errorf("%v\n%s", err, buffer)
	}
	return string(code), nil
}

// bindBasicType converts basic solidity types(except array, slice and tuple) to Go ones.
// bindBasicType 将基本的 Solidity 类型（数组、切片和元组除外）转换为 Go 类型。
func bindBasicType(kind abi.Type) string {
	switch kind.T {
	case abi.AddressTy:
		return "common.Address"
	case abi.IntTy, abi.UintTy:
		parts := intRegex.FindStringSubmatch(kind.String())
		switch parts[2] {
		case "8", "16", "32", "64": // 对于 8, 16, 32, 64 位的整数，直接映射
			return fmt.Sprintf("%sint%s", parts[1], parts[2])
		}
		return "*big.Int" // 其他大小的整数映射为 *big.Int
	case abi.FixedBytesTy:
		return fmt.Sprintf("[%d]byte", kind.Size)
	case abi.BytesTy:
		return "[]byte"
	case abi.FunctionTy:
		return "[24]byte"
	default:
		// string, bool types
		// 字符串、布尔类型
		return kind.String()
	}
}

// bindType converts solidity types to Go ones. Since there is no clear mapping
// from all Solidity types to Go ones (e.g. uint17), those that cannot be exactly
// mapped will use an upscaled type (e.g. BigDecimal).
// bindType 将 Solidity 类型转换为 Go 类型。由于并非所有 Solidity 类型都能清晰地映射到 Go 类型
// （例如 uint17），那些无法精确映射的类型将使用一个向上扩展的类型（例如 *big.Int）。
//
// kind: ABI 类型。
// structs: 结构体映射。
// return: 返回对应的 Go 类型字符串。
func bindType(kind abi.Type, structs map[string]*tmplStruct) string {
	switch kind.T {
	case abi.TupleTy: // 元组类型（结构体）
		return structs[kind.TupleRawName+kind.String()].Name
	case abi.ArrayTy: // 数组类型
		return fmt.Sprintf("[%d]", kind.Size) + bindType(*kind.Elem, structs)
	case abi.SliceTy: // 切片类型
		return "[]" + bindType(*kind.Elem, structs)
	default: // 其他基本类型
		return bindBasicType(kind)
	}
}

// bindTopicType converts a Solidity topic type to a Go one. It is almost the same
// functionality as for simple types, but dynamic types get converted to hashes.
// bindTopicType 将 Solidity 的 topic 类型转换为 Go 类型。它的功能与简单类型几乎相同，
// 但动态类型会被转换为哈希。
//
// kind: ABI 类型。
// structs: 结构体映射。
// return: 返回对应的 Go 类型字符串。
func bindTopicType(kind abi.Type, structs map[string]*tmplStruct) string {
	bound := bindType(kind, structs)

	// todo(rjl493456442) according solidity documentation, indexed event
	// parameters that are not value types i.e. arrays and structs are not
	// stored directly but instead a keccak256-hash of an encoding is stored.
	//
	// We only convert strings and bytes to hash, still need to deal with
	// array(both fixed-size and dynamic-size) and struct.
	// 根据 Solidity 文档，非值类型的索引事件参数（即数组和结构体）
	// 不会直接存储，而是存储其编码的 keccak256 哈希值。
	//
	// 我们只将字符串和字节转换为哈希，仍然需要处理数组（固定大小和动态大小）和结构体。
	if bound == "string" || bound == "[]byte" {
		bound = "common.Hash"
	}
	return bound
}

// bindStructType converts a Solidity tuple type to a Go one and records the mapping
// in the given map. Notably, this function will resolve and record nested struct
// recursively.
// bindStructType 将 Solidity 的元组类型转换为 Go 类型，并在给定的映射中记录该映射。
// 值得注意的是，此函数将递归地解析和记录嵌套的结构体。
//
// kind: ABI 类型。
// structs: 用于存储和查找已绑定结构体的映射。
// return: 返回生成的 Go 结构体名称。
func bindStructType(kind abi.Type, structs map[string]*tmplStruct) string {
	switch kind.T {
	case abi.TupleTy:
		// We compose a raw struct name and a canonical parameter expression
		// together here. The reason is before solidity v0.5.11, kind.TupleRawName
		// is empty, so we use canonical parameter expression to distinguish
		// different struct definition. From the consideration of backward
		// compatibility, we concat these two together so that if kind.TupleRawName
		// is not empty, it can have unique id.
		// 我们在这里将原始结构体名称和规范参数表达式组合在一起。
		// 原因是，在 solidity v0.5.11 之前，kind.TupleRawName 是空的，
		// 所以我们使用规范参数表达式来区分不同的结构体定义。
		// 出于向后兼容性的考虑，我们将这两者连接在一起，以便如果 kind.TupleRawName
		// 不为空，它也能有唯一的 id。
		id := kind.TupleRawName + kind.String()
		if s, exist := structs[id]; exist { // 如果已经处理过，直接返回
			return s.Name
		}
		var (
			names  = make(map[string]bool)
			fields []*tmplField
		)
		// 遍历元组的每个元素
		for i, elem := range kind.TupleElems {
			name := abi.ToCamelCase(kind.TupleRawNames[i])
			// 解决字段名冲突
			name = abi.ResolveNameConflict(name, func(s string) bool { return names[s] })
			names[name] = true
			fields = append(fields, &tmplField{
				Type:    bindStructType(*elem, structs), // 递归处理字段类型
				Name:    name,
				SolKind: *elem,
			})
		}
		name := kind.TupleRawName
		if name == "" { // 如果没有原始名称，则生成一个
			name = fmt.Sprintf("Struct%d", len(structs))
		}
		name = abi.ToCamelCase(name)

		// 存储新的结构体模板数据
		structs[id] = &tmplStruct{
			Name:   name,
			Fields: fields,
		}
		return name
	case abi.ArrayTy:
		return fmt.Sprintf("[%d]", kind.Size) + bindStructType(*kind.Elem, structs)
	case abi.SliceTy:
		return "[]" + bindStructType(*kind.Elem, structs)
	default:
		return bindBasicType(kind)
	}
}

// alias returns an alias of the given string based on the aliasing rules
// or returns itself if no rule is matched.
// alias 根据别名规则返回给定字符串的别名，如果未匹配到规则，则返回其本身。
func alias(aliases map[string]string, n string) string {
	if alias, exist := aliases[n]; exist {
		return alias
	}
	return n
}

// decapitalise makes a camel-case string which starts with a lower case character.
// decapitalise 创建一个首字母小写的驼峰式字符串。
func decapitalise(input string) string {
	if len(input) == 0 {
		return input
	}
	goForm := abi.ToCamelCase(input)
	return strings.ToLower(goForm[:1]) + goForm[1:]
}

// structured checks whether a list of ABI data types has enough information to
// operate through a proper Go struct or if flat returns are needed.
// structured 检查 ABI 数据类型列表是否具有足够的信息以通过适当的 Go 结构体进行操作，
// 或者是否需要扁平化的返回。
func structured(args abi.Arguments) bool {
	if len(args) < 2 { // 少于 2 个返回值，没必要用结构体
		return false
	}
	exists := make(map[string]bool)
	for _, out := range args {
		// If the name is anonymous, we can't organize into a struct
		// 如果名称是匿名的，我们无法组织成结构体
		if out.Name == "" {
			return false
		}
		// If the field name is empty when normalized or collides (var, Var, _var, _Var),
		// we can't organize into a struct
		// 如果字段名在规范化后为空或发生冲突（var、Var、_var、_Var），
		// 我们无法组织成结构体
		field := abi.ToCamelCase(out.Name)
		if field == "" || exists[field] {
			return false
		}
		exists[field] = true
	}
	return true
}

// hasStruct returns an indicator whether the given type is struct, struct slice
// or struct array.
// hasStruct 返回一个指示符，指示给定类型是否是结构体、结构体切片或结构体数组。
func hasStruct(t abi.Type) bool {
	switch t.T {
	case abi.SliceTy: // 如果是切片，递归检查元素类型
		return hasStruct(*t.Elem)
	case abi.ArrayTy: // 如果是数组，递归检查元素类型
		return hasStruct(*t.Elem)
	case abi.TupleTy: // 如果是元组（结构体），返回 true
		return true
	default:
		return false
	}
}
