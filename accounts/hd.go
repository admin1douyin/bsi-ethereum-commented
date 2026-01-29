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

package accounts

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
// DefaultRootDerivationPath 是自定义派生端点追加到的根路径。
// 因此，第一个帐户将位于 m/44'/60'/0'/0，第二个帐户将位于 m/44'/60'/0'/1，依此类推。
var DefaultRootDerivationPath = DerivationPath{0x80000000 + 44, 0x80000000 + 60, 0x80000000 + 0, 0} // 定义默认根派生路径。

// DefaultBaseDerivationPath is the base path from which custom derivation endpoints
// are incremented. As such, the first account will be at m/44'/60'/0'/0/0, the second
// at m/44'/60'/0'/0/1, etc.
// DefaultBaseDerivationPath 是自定义派生端点递增的基本路径。
// 因此，第一个帐户将位于 m/44'/60'/0'/0/0，第二个帐户将位于 m/44'/60'/0'/0/1，依此类推。
var DefaultBaseDerivationPath = DerivationPath{0x80000000 + 44, 0x80000000 + 60, 0x80000000 + 0, 0, 0} // 定义默认基础派生路径。

// LegacyLedgerBaseDerivationPath is the legacy base path from which custom derivation
// endpoints are incremented. As such, the first account will be at m/44'/60'/0'/0, the
// second at m/44'/60'/0'/1, etc.
// LegacyLedgerBaseDerivationPath 是自定义派生端点递增的旧版基本路径。
// 因此，第一个帐户将位于 m/44'/60'/0'/0，第二个帐户将位于 m/44'/60'/0'/1，依此类推。
var LegacyLedgerBaseDerivationPath = DerivationPath{0x80000000 + 44, 0x80000000 + 60, 0x80000000 + 0, 0} // 定义旧版 Ledger 基础派生路径。

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
// DerivationPath 表示分层确定性钱包帐户派生路径的计算机友好版本。
//
// BIP-32 规范 https://github.com/bitcoin/bips/blob/master/bip-0032.mediawiki
// 定义派生路径的形式为：
//
//	m / purpose' / coin_type' / account' / change / address_index
//
// BIP-44 规范 https://github.com/bitcoin/bips/blob/master/bip-0044.mediawiki
// 定义加密货币的 `purpose` 为 44' (或 0x8000002C)，
// SLIP-44 https://github.com/satoshilabs/slips/blob/master/slip-0044.md 分配
// `coin_type` 60' (或 0x8000003C) 给以太坊。
//
// 根据 https://github.com/ethereum/EIPs/issues/84 的规范，以太坊的根路径是 m/44'/60'/0'/0，
// 尽管关于帐户是否应该递增最后一个组件还是其子项尚未最终确定。
// 我们将采用递增最后一个组件的更简单方法。
type DerivationPath []uint32 // 定义 DerivationPath 类型，它是一个 uint32 的切片。

// ParseDerivationPath converts a user specified derivation path string to the
// internal binary representation.
//
// Full derivation paths need to start with the `m/` prefix, relative derivation
// paths (which will get appended to the default root path) must not have prefixes
// in front of the first element. Whitespace is ignored.
// ParseDerivationPath 将用户指定的派生路径字符串转换为内部二进制表示。
//
// 完整的派生路径需要以 `m/` 前缀开始，相对派生路径（将附加到默认根路径）
// 不能在第一个元素前面有前缀。忽略空格。
func ParseDerivationPath(path string) (DerivationPath, error) { // 定义 ParseDerivationPath 函数，用于解析派生路径字符串。
	var result DerivationPath // 声明一个 DerivationPath 类型的变量 result。

	// Handle absolute or relative paths
	// 处理绝对或相对路径
	components := strings.Split(path, "/") // 按 "/" 分割路径字符串。
	switch {
	case len(components) == 0: // 如果组件长度为 0，则返回错误。
		return nil, errors.New("empty derivation path")

	case strings.TrimSpace(components[0]) == "": // 如果第一个组件是空字符串，则返回错误。
		return nil, errors.New("ambiguous path: use 'm/' prefix for absolute paths, or no leading '/' for relative ones")

	case strings.TrimSpace(components[0]) == "m": // 如果第一个组件是 "m"，则为绝对路径。
		components = components[1:] // 去掉 "m"。

	default: // 否则为相对路径。
		result = append(result, DefaultRootDerivationPath...) // 将默认根路径追加到结果中。
	}
	// All remaining components are relative, append one by one
	// 所有剩余的组件都是相对的，逐个追加
	if len(components) == 0 { // 如果没有剩余组件，则返回错误。
		return nil, errors.New("empty derivation path") // Empty relative paths // 空的相对路径
	}
	for _, component := range components { // 遍历所有组件。
		// Ignore any user added whitespace
		// 忽略任何用户添加的空格
		component = strings.TrimSpace(component) // 去除组件两边的空格。
		var value uint32                   // 声明一个 uint32 类型的变量 value。

		// Handle hardened paths
		// 处理硬化路径
		if strings.HasSuffix(component, "'") { // 如果组件以 "'" 结尾，则为硬化路径。
			value = 0x80000000                                           // 设置硬化路径的标志位。
			component = strings.TrimSpace(strings.TrimSuffix(component, "'")) // 去掉 "'"。
		}
		// Handle the non hardened component
		// 处理非硬化组件
		bigval, ok := new(big.Int).SetString(component, 0) // 将组件字符串转换为大整数。
		if !ok {                                          // 如果转换失败，则返回错误。
			return nil, fmt.Errorf("invalid component: %s", component)
		}
		max := math.MaxUint32 - value // 计算允许的最大值。
		if bigval.Sign() < 0 || bigval.Cmp(big.NewInt(int64(max))) > 0 { // 如果值超出范围，则返回错误。
			if value == 0 { // 如果是非硬化路径。
				return nil, fmt.Errorf("component %v out of allowed range [0, %d]", bigval, max)
			}
			return nil, fmt.Errorf("component %v out of allowed hardened range [0, %d]", bigval, max) // 如果是硬化路径。
		}
		value += uint32(bigval.Uint64()) // 将组件值加到 value 上。

		// Append and repeat
		// 追加并重复
		result = append(result, value) // 将 value 追加到结果中。
	}
	return result, nil // 返回结果和 nil 错误。
}

// String implements the stringer interface, converting a binary derivation path
// to its canonical representation.
// String 实现 stringer 接口，将二进制派生路径转换为其规范表示。
func (path DerivationPath) String() string { // 为 DerivationPath 类型定义 String 方法。
	result := "m" // 初始化结果字符串为 "m"。
	for _, component := range path { // 遍历路径中的每个组件。
		var hardened bool // 声明一个布尔变量 hardened。
		if component >= 0x80000000 { // 如果组件值大于等于 0x80000000，则为硬化路径。
			component -= 0x80000000 // 去掉硬化标志位。
			hardened = true         // 设置 hardened 为 true。
		}
		result = fmt.Sprintf("%s/%d", result, component) // 将组件格式化并追加到结果字符串中。
		if hardened { // 如果是硬化路径。
			result += "'" // 追加 "'"。
		}
	}
	return result // 返回结果字符串。
}

// MarshalJSON turns a derivation path into its json-serialized string
// MarshalJSON 将派生路径转换为其 json 序列化字符串
func (path DerivationPath) MarshalJSON() ([]byte, error) { // 为 DerivationPath 类型定义 MarshalJSON 方法。
	return json.Marshal(path.String()) // 将路径的字符串表示形式进行 JSON 编码。
}

// UnmarshalJSON a json-serialized string back into a derivation path
// UnmarshalJSON 将 json 序列化字符串转换回派生路径
func (path *DerivationPath) UnmarshalJSON(b []byte) error { // 为 DerivationPath 类型定义 UnmarshalJSON 方法。
	var dp string     // 声明一个字符串变量 dp。
	var err error     // 声明一个错误变量 err。
	if err = json.Unmarshal(b, &dp); err != nil { // 解码 JSON 数据到 dp。
		return err // 如果解码失败，则返回错误。
	}
	*path, err = ParseDerivationPath(dp) // 解析派生路径字符串。
	return err                             // 返回错误。
}

// DefaultIterator creates a BIP-32 path iterator, which progresses by increasing the last component:
// i.e. m/44'/60'/0'/0/0, m/44'/60'/0'/0/1, m/44'/60'/0'/0/2, ... m/44'/60'/0'/0/N.
// DefaultIterator 创建一个 BIP-32 路径迭代器，通过增加最后一个组件来推进：
// 即 m/44'/60'/0'/0/0, m/44'/60'/0'/0/1, m/44'/60'/0'/0/2, ... m/44'/60'/0'/0/N。
func DefaultIterator(base DerivationPath) func() DerivationPath { // 定义 DefaultIterator 函数，用于创建默认的派生路径迭代器。
	path := make(DerivationPath, len(base)) // 创建一个与 base 长度相同的 DerivationPath 切片。
	copy(path[:], base[:])                   // 复制 base 的内容到 path。
	// Set it back by one, so the first call gives the first result
	// 将其退回一步，以便第一次调用给出第一个结果
	path[len(path)-1]-- // 将最后一个组件减一。
	return func() DerivationPath { // 返回一个闭包函数。
		path[len(path)-1]++ // 每次调用时将最后一个组件加一。
		return path          // 返回路径。
	}
}

// LedgerLiveIterator creates a bip44 path iterator for Ledger Live.
// Ledger Live increments the third component rather than the fifth component
// i.e. m/44'/60'/0'/0/0, m/44'/60'/1'/0/0, m/44'/60'/2'/0/0, ... m/44'/60'/N'/0/0.
// LedgerLiveIterator 为 Ledger Live 创建 bip44 路径迭代器。
// Ledger Live 递增第三个组件而不是第五个组件
// 即 m/44'/60'/0'/0/0, m/44'/60'/1'/0/0, m/44'/60'/2'/0/0, ... m/44'/60'/N'/0/0。
func LedgerLiveIterator(base DerivationPath) func() DerivationPath { // 定义 LedgerLiveIterator 函数，用于创建 Ledger Live 的派生路径迭代器。
	path := make(DerivationPath, len(base)) // 创建一个与 base 长度相同的 DerivationPath 切片。
	copy(path[:], base[:])                   // 复制 base 的内容到 path。
	// Set it back by one, so the first call gives the first result
	// 将其退回一步，以便第一次调用给出第一个结果
	path[2]-- // 将第三个组件减一。
	return func() DerivationPath { // 返回一个闭包函数。
		// ledgerLivePathIterator iterates on the third component
		// ledgerLivePathIterator 在第三个组件上迭代
		path[2]++ // 每次调用时将第三个组件加一。
		return path // 返回路径。
	}
}
