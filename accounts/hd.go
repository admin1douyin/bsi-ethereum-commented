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

package accounts // 声明包名为 accounts

import (
	"encoding/json" // 导入 encoding/json 包，用于 JSON 数据的编码和解码。
	"errors"        // 导入 errors 包，用于创建和处理错误。
	"fmt"           // 导入 fmt 包，用于格式化字符串。
	"math"          // 导入 math 包，提供基本的数学常数和函数。
	"math/big"      // 导入 math/big 包，用于处理大数。
	"strings"       // 导入 strings 包，用于处理字符串。
)

// DefaultRootDerivationPath is the root path to which custom derivation endpoints
// are appended. As such, the first account will be at m/44'/60'/0'/0, the second
// at m/44'/60'/0'/1, etc.
// DefaultRootDerivationPath 是附加自定义派生端点的根路径。
// 因此，第一个账户将位于 m/44'/60'/0'/0，第二个账户位于 m/44'/60'/0'/1，依此类推。
var DefaultRootDerivationPath = DerivationPath{0x80000000 + 44, 0x80000000 + 60, 0x80000000 + 0, 0} // 定义默认的根派生路径 (m/44'/60'/0'/0)。

// DefaultBaseDerivationPath is the base path from which custom derivation endpoints
// are incremented. As such, the first account will be at m/44'/60'/0'/0/0, the second
// at m/44'/60'/0'/0/1, etc.
// DefaultBaseDerivationPath 是自定义派生端点递增的基础路径。
// 因此，第一个账户将位于 m/44'/60'/0'/0/0，第二个账户位于 m/44'/60'/0'/0/1，依此类推。
var DefaultBaseDerivationPath = DerivationPath{0x80000000 + 44, 0x80000000 + 60, 0x80000000 + 0, 0, 0} // 定义默认的基础派生路径 (m/44'/60'/0'/0/0)。

// LegacyLedgerBaseDerivationPath is the legacy base path from which custom derivation
// endpoints are incremented. As such, the first account will be at m/44'/60'/0'/0, the
// second at m/44'/60'/0'/1, etc.
// LegacyLedgerBaseDerivationPath 是自定义派生端点递增的旧版基础路径。
// 因此，第一个账户将位于 m/44'/60'/0'/0，第二个账户位于 m/44'/60'/0'/1，依此类推。
var LegacyLedgerBaseDerivationPath = DerivationPath{0x80000000 + 44, 0x80000000 + 60, 0x80000000 + 0, 0} // 定义旧版 Ledger 硬件钱包的基础派生路径 (m/44'/60'/0'/0)。

// DerivationPath represents the computer friendly version of a hierarchical
// deterministic wallet account derivation path.
//
// The BIP-32 spec https://github.com/bitcoin/bips/blob/master/bip-0032.mediawiki
// defines derivation paths to be of the form:
//
//	m / purpose' / coin_type' / account' / change / address_index
//
// The BIP-44 spec https://github.com/bitcoin/bips/blob/master/bip-0044.mediawiki
// defines that the `purpose` be 44' (or 0x8000002C) for crypto currencies, and
// SLIP-44 https://github.com/satoshilabs/slips/blob/master/slip-0044.md assigns
// the `coin_type` 60' (or 0x8000003C) to Ethereum.
//
// The root path for Ethereum is m/44'/60'/0'/0 according to the specification
// from https://github.com/ethereum/EIPs/issues/84, albeit it's not set in stone
// yet whether accounts should increment the last component or the children of
// that. We will go with the simpler approach of incrementing the last component.
// DerivationPath 代表分层确定性（HD）钱包账户派生路径的计算机友好版本。
//
// BIP-32 规范 (https://github.com/bitcoin/bips/blob/master/bip-0032.mediawiki)
// 定义派生路径的格式为：
//
//	m / purpose' / coin_type' / account' / change / address_index
//
// BIP-44 规范 (https://github.com/bitcoin/bips/blob/master/bip-0044.mediawiki)
// 为加密货币定义 `purpose` 为 44' (即 0x8000002C)，而
// SLIP-44 (https://github.com/satoshilabs/slips/blob/master/slip-0044.md) 将
// `coin_type` 60' (即 0x8000003C) 分配给以太坊。
//
// 根据 EIP-84 (https://github.com/ethereum/EIPs/issues/84) 的规范，以太坊的根路径为 m/44'/60'/0'/0，
// 尽管账户是应该递增最后一个组件还是其子组件尚未最终确定。
// 我们将采用递增最后一个组件的更简单方法。
type DerivationPath []uint32 // 定义 DerivationPath 类型，它是一个 uint32 的切片，用于存储派生路径的各个部分。

// ParseDerivationPath converts a user specified derivation path string to the
// internal binary representation.
//
// Full derivation paths need to start with the `m/` prefix, relative derivation
// paths (which will get appended to the default root path) must not have prefixes
// in front of the first element. Whitespace is ignored.
// ParseDerivationPath 将用户指定的派生路径字符串转换为内部二进制表示形式。
//
// 完整的派生路径需要以 `m/` 前缀开头，而相对派生路径（将被附加到默认根路径）
// 在第一个元素前不能有前缀。字符串中的空格将被忽略。
// path: (string) 用户输入的派生路径字符串。
// return: (DerivationPath, error) 返回解析后的派生路径和可能发生的错误。
func ParseDerivationPath(path string) (DerivationPath, error) {
	var result DerivationPath // 声明一个 DerivationPath 类型的变量 result，用于存储解析结果。

	// Handle absolute or relative paths
	// 处理绝对路径或相对路径
	components := strings.Split(path, "/") // 按 "/" 分割路径字符串，得到路径的各个组成部分。
	switch {
	case len(components) == 0: // 如果分割后组件数量为 0，说明路径为空。
		return nil, errors.New("empty derivation path") // 返回错误，提示派生路径为空。

	case strings.TrimSpace(components[0]) == "": // 如果第一个组件是空字符串（例如路径以 "/" 开头）。
		return nil, errors.New("ambiguous path: use 'm/' prefix for absolute paths, or no leading '/' for relative ones") // 返回错误，提示路径格式不明确。

	case strings.TrimSpace(components[0]) == "m": // 如果第一个组件是 "m"，表示是绝对路径。
		components = components[1:] // 移除 "m" 组件，处理后面的部分。

	default: // 如果不是以 "m" 开头，也不是空，则认为是相对路径。
		result = append(result, DefaultRootDerivationPath...) // 将默认的根派生路径追加到结果中。
	}

	// All remaining components are relative, append one by one
	// 所有剩余的组件都是相对的，逐个追加
	if len(components) == 0 { // 如果移除 "m" 后没有其他组件了。
		// This can happen if the path was just "m". We treat that as a valid path
		// returning the root. However, a trailing / is not allowed
		// 如果路径仅为 "m"，这是允许的，表示根路径。但 "m/" 是不完整的。
		if strings.HasSuffix(path, "/") {
			return nil, errors.New("empty derivation path component")
		}
		return result, nil
	}

	for _, component := range components { // 遍历路径的每个数字组件。
		// Ignore any user added whitespace
		// 忽略用户添加的任何空格
		component = strings.TrimSpace(component) // 移除组件字符串两端的空格。
		var value uint32                   // 声明一个 uint32 类型的变量 value，用于存储组件的数值。

		// Handle hardened paths
		// 处理硬化路径
		if strings.HasSuffix(component, "'") { // 如果组件以 "'" 结尾，表示这是一个硬化派生路径。
			value = 0x80000000                                           // 为 value 设置硬化路径的标志位 (最高位为1)。
			component = strings.TrimSpace(strings.TrimSuffix(component, "'")) // 从组件字符串中移除 "'" 字符。
		}

		// Handle the non hardened component
		// 处理非硬化组件
		bigval, ok := new(big.Int).SetString(component, 0) // 将数字字符串部分转换为大整数类型。
		if !ok {                                          // 如果转换失败。
			return nil, fmt.Errorf("invalid component: %s", component) // 返回错误，提示组件无效。
		}

		max := math.MaxUint32 - value // 计算该组件允许的最大值。
		if bigval.Sign() < 0 || bigval.Cmp(big.NewInt(int64(max))) > 0 { // 检查数值是否在允许范围内。
			if value == 0 { // 如果是非硬化路径。
				return nil, fmt.Errorf("component %v out of allowed range [0, %d]", bigval, max) // 返回错误，提示非硬化组件越界。
			}
			return nil, fmt.Errorf("component %v out of allowed hardened range [0, %d]", bigval, max) // 返回错误，提示硬化组件越界。
		}
		value += uint32(bigval.Uint64()) // 将解析出的数值加到 value 上（如果是硬化路径，则加上硬化标志位）。

		// Append and repeat
		// 追加并重复
		result = append(result, value) // 将解析出的组件值追加到结果切片中。
	}
	return result, nil // 返回最终的派生路径和 nil 错误。
}

// String implements the stringer interface, converting a binary derivation path
// to its canonical representation.
// String 实现了 stringer 接口，将二进制派生路径转换为其规范的字符串表示形式。
// path: (DerivationPath) 接收者，一个 DerivationPath 类型的实例。
// return: (string) 返回路径的字符串表示。
func (path DerivationPath) String() string {
	result := "m" // 初始化结果字符串为 "m"，代表主密钥。
	for _, component := range path { // 遍历派生路径中的每个 uint32 组件。
		var hardened bool // 声明一个布尔变量 hardened，用于标记是否为硬化路径。
		if component >= 0x80000000 { // 如果组件值大于等于 0x80000000，说明是硬化路径。
			component -= 0x80000000 // 减去硬化标志位，得到原始数值。
			hardened = true         // 将 hardened 标记为 true。
		}
		result = fmt.Sprintf("%s/%d", result, component) // 将组件的数字部分格式化并追加到结果字符串中。
		if hardened { // 如果是硬化路径。
			result += "'" // 在数字后面追加 "'" 符号。
		}
	}
	return result // 返回最终的字符串表示。
}

// MarshalJSON turns a derivation path into its json-serialized string
// MarshalJSON 将派生路径转换为其 JSON 序列化的字符串形式。
// path: (DerivationPath) 接收者，一个 DerivationPath 类型的实例。
// return: ([]byte, error) 返回 JSON 编码后的字节切片和可能发生的错误。
func (path DerivationPath) MarshalJSON() ([]byte, error) {
	return json.Marshal(path.String()) // 调用 path.String() 获取字符串表示，然后将其进行 JSON 编码。
}

// UnmarshalJSON a json-serialized string back into a derivation path
// UnmarshalJSON 将 JSON 序列化的字符串反序列化回派生路径。
// path: (*DerivationPath) 接收者，一个指向 DerivationPath 类型的指针。
// b: ([]byte) 传入的 JSON 字节数据。
// return: (error) 返回可能发生的错误。
func (path *DerivationPath) UnmarshalJSON(b []byte) error {
	var dp string     // 声明一个字符串变量 dp，用于存储从 JSON 中解码出的字符串。
	var err error     // 声明一个错误变量 err。
	if err = json.Unmarshal(b, &dp); err != nil { // 将 JSON 数据解码到 dp 字符串。
		return err // 如果解码失败，返回错误。
	}
	*path, err = ParseDerivationPath(dp) // 使用 ParseDerivationPath 函数将字符串解析为派生路径。
	return err                             // 返回解析过程中可能发生的错误。
}

// DefaultIterator creates a BIP-32 path iterator, which progresses by increasing the last component:
// i.e. m/44'/60'/0'/0/0, m/44'/60'/0'/0/1, m/44'/60'/0'/0/2, ... m/44'/60'/0'/0/N.
// DefaultIterator 创建一个 BIP-32 路径迭代器，它通过递增最后一个组件来前进：
// 例如：m/44'/60'/0'/0/0, m/44'/60'/0'/0/1, m/44'/60'/0'/0/2, ... m/44'/60'/0'/0/N。
// base: (DerivationPath) 迭代器的基础路径。
// return: (func() DerivationPath) 返回一个函数，该函数每次被调用时返回下一个派生路径。
func DefaultIterator(base DerivationPath) func() DerivationPath {
	path := make(DerivationPath, len(base)) // 创建一个新的 DerivationPath 切片，长度与基础路径相同。
	copy(path[:], base[:])                   // 将基础路径的内容复制到新创建的切片中。
	// Set it back by one, so the first call gives the first result
	// 将其减一，以便第一次调用时能得到第一个结果
	path[len(path)-1]-- // 将路径的最后一个组件减一，为首次迭代做准备。
	return func() DerivationPath { // 返回一个闭包函数，即迭代器。
		path[len(path)-1]++ // 每次调用时，将路径的最后一个组件加一。
		return path          // 返回当前的派生路径。
	}
}

// LedgerLiveIterator creates a bip44 path iterator for Ledger Live.
// Ledger Live increments the third component rather than the fifth component
// i.e. m/44'/60'/0'/0/0, m/44'/60'/1'/0/0, m/44'/60'/2'/0/0, ... m/44'/60'/N'/0/0.
// LedgerLiveIterator 为 Ledger Live 创建一个 BIP-44 路径迭代器。
// Ledger Live 递增第三个组件而不是第五个组件，例如：
// m/44'/60'/0'/0/0, m/44'/60'/1'/0/0, m/44'/60'/2'/0/0, ... m/44'/60'/N'/0/0。
// base: (DerivationPath) 迭代器的基础路径。
// return: (func() DerivationPath) 返回一个函数，该函数每次被调用时返回下一个派生路径。
func LedgerLiveIterator(base DerivationPath) func() DerivationPath {
	path := make(DerivationPath, len(base)) // 创建一个新的 DerivationPath 切片，长度与基础路径相同。
	copy(path[:], base[:])                   // 将基础路径的内容复制到新创建的切片中。
	// Set it back by one, so the first call gives the first result
	// 将其减一，以便第一次调用时能得到第一个结果
	path[2]-- // 将路径的第三个组件减一，为首次迭代做准备。
	return func() DerivationPath { // 返回一个闭包函数，即迭代器。
		// ledgerLivePathIterator iterates on the third component
		// ledgerLivePathIterator 在第三个组件上进行迭代
		path[2]++ // 每次调用时，将路径的第三个组件加一。
		return path // 返回当前的派生路径。
	}
}
