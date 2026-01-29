// Copyright 2022 The go-ethereum Authors
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

// 版权所有 2022 The go-ethereum Authors
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

import "fmt"

// ResolveNameConflict returns the next available name for a given thing.
// This helper can be used for lots of purposes:
//
//   - In solidity function overloading is supported, this function can fix
//     the name conflicts of overloaded functions.
//   - In golang binding generation, the parameter(in function, event, error,
//     and struct definition) name will be converted to camelcase style which
//     may eventually lead to name conflicts.
//
// Name conflicts are mostly resolved by adding number suffix. e.g. if the abi contains
// Methods "send" and "send1", ResolveNameConflict would return "send2" for input "send".
// ResolveNameConflict 为给定事物返回下一个可用的名称。
// 这个辅助函数可以用于多种目的：
//
//   - 在 Solidity 中，函数重载是支持的，此函数可以修复
//     重载函数的名称冲突。
//   - 在 Go 语言绑定生成中，参数（在函数、事件、错误
//     和结构体定义中）的名称将被转换为驼峰式风格，这
//     最终可能导致名称冲突。
//
// 名称冲突主要通过添加数字后缀来解决。例如，如果 abi 包含
// 方法 "send" 和 "send1"，则 ResolveNameConflict 对于输入 "send" 将返回 "send2"。
//
// rawName: 原始名称
// used: 一个函数，用于检查某个名称是否已被使用
// return: 返回一个不冲突的名称
func ResolveNameConflict(rawName string, used func(string) bool) string {
	// 初始名称为原始名称
	name := rawName
	// 检查初始名称是否已被使用
	ok := used(name)
	// 如果名称已被使用，则开始循环添加数字后缀
	for idx := 0; ok; idx++ {
		// 格式化新名称，例如 send0, send1, ...
		name = fmt.Sprintf("%s%d", rawName, idx)
		// 检查新名称是否已被使用
		ok = used(name)
	}
	// 返回最终确定的不冲突的名称
	return name
}
