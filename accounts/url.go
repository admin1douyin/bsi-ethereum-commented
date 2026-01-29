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
	"strings"       // 导入 strings 包，用于处理字符串。
)

// URL represents the canonical identification URL of a wallet or account.
//
// It is a simplified version of url.URL, with the important limitations (which
// are considered features here) that it contains value-copyable components only,
// as well as that it doesn't do any URL encoding/decoding of special characters.
//
// The former is important to allow an account to be copied without leaving live
// references to the original version, whereas the latter is important to ensure
// one single canonical form opposed to many allowed ones by the RFC 3986 spec.
//
// As such, these URLs should not be used outside of the scope of an Ethereum
// wallet or account.
// URL 表示钱包或帐户的规范标识 URL。
//
// 它是 url.URL 的简化版本，具有重要的限制（在此视为功能）：
// 它仅包含可值复制的组件，并且不对特殊字符进行任何 URL 编码/解码。
//
// 前者对于允许在不留下对原始版本的实时引用情况下复制帐户非常重要，
// 而后者对于确保单一的规范形式而不是 RFC 3986 规范允许多种形式非常重要。
//
// 因此，这些 URL 不应在以太坊钱包或帐户范围之外使用。
type URL struct {
	Scheme string // Protocol scheme to identify a capable account backend // 协议方案以标识有能力的帐户后端
	Path   string // Path for the backend to identify a unique entity // 后端标识唯一实体的路径
}

// parseURL converts a user supplied URL into the accounts specific structure.
// parseURL 将用户提供的 URL 转换为特定于帐户的结构。
func parseURL(url string) (URL, error) { // 定义 parseURL 函数，用于解析 URL 字符串。
	parts := strings.Split(url, "://") // 按 "://" 分割 URL 字符串。
	if len(parts) != 2 || parts[0] == "" { // 如果分割后的部分不等于 2 或协议方案为空。
		return URL{}, errors.New("protocol scheme missing") // 返回错误。
	}
	return URL{ // 返回解析后的 URL 结构体。
		Scheme: parts[0], // 设置协议方案。
		Path:   parts[1], // 设置路径。
	}, nil // 返回 nil 错误。
}

// String implements the stringer interface.
// String 实现 stringer 接口。
func (u URL) String() string { // 为 URL 类型定义 String 方法。
	if u.Scheme != "" { // 如果协议方案不为空。
		return fmt.Sprintf("%s://%s", u.Scheme, u.Path) // 返回格式化后的 URL 字符串。
	}
	return u.Path // 否则只返回路径。
}

// TerminalString implements the log.TerminalStringer interface.
// TerminalString 实现 log.TerminalStringer 接口。
func (u URL) TerminalString() string { // 为 URL 类型定义 TerminalString 方法。
	url := u.String() // 获取 URL 的字符串表示形式。
	if len(url) > 32 { // 如果 URL 字符串长度超过 32。
		return url[:31] + ".." // 截断并添加 ".."。
	}
	return url // 返回 URL 字符串。
}

// MarshalJSON implements the json.Marshaller interface.
// MarshalJSON 实现 json.Marshaller 接口。
func (u URL) MarshalJSON() ([]byte, error) { // 为 URL 类型定义 MarshalJSON 方法。
	return json.Marshal(u.String()) // 将 URL 的字符串表示形式进行 JSON 编码。
}

// UnmarshalJSON parses url.
// UnmarshalJSON 解析 url。
func (u *URL) UnmarshalJSON(input []byte) error { // 为 URL 类型定义 UnmarshalJSON 方法。
	var textURL string // 声明一个字符串变量 textURL。
	err := json.Unmarshal(input, &textURL) // 解码 JSON 数据到 textURL。
	if err != nil { // 如果解码失败。
		return err // 返回错误。
	}
	url, err := parseURL(textURL) // 解析 URL 字符串。
	if err != nil { // 如果解析失败。
		return err // 返回错误。
	}
	u.Scheme = url.Scheme // 设置协议方案。
	u.Path = url.Path       // 设置路径。
	return nil             // 返回 nil 错误。
}

// Cmp compares x and y and returns:
//
//	-1 if x <  y
//	 0 if x == y
//	+1 if x >  y
//
// Cmp 比较 x 和 y 并返回：
//
//	-1 如果 x <  y
//	 0 如果 x == y
//	+1 如果 x >  y
func (u URL) Cmp(url URL) int { // 为 URL 类型定义 Cmp 方法，用于比较两个 URL。
	if u.Scheme == url.Scheme { // 如果协议方案相同。
		return strings.Compare(u.Path, url.Path) // 比较路径。
	}
	return strings.Compare(u.Scheme, url.Scheme) // 比较协议方案。
}
