// Copyright 2018 The go-ethereum Authors
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

// 版权所有 2018 The go-ethereum Authors
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

// AccountsByURL implements sort.Interface for []Account based on the URL field.
// AccountsByURL 基于 URL 字段为 []Account 实现 sort.Interface。
type AccountsByURL []Account // 定义 AccountsByURL 类型，它是一个 Account 的切片。

// Len 返回切片的长度。
func (a AccountsByURL) Len() int { return len(a) } // Len 方法返回切片 a 的长度。
// Swap 交换切片中两个元素的位置。
func (a AccountsByURL) Swap(i, j int) { a[i], a[j] = a[j], a[i] } // Swap 方法交换切片 a 中索引为 i 和 j 的两个元素。
// Less 比较两个元素的 URL，用于排序。
func (a AccountsByURL) Less(i, j int) bool { return a[i].URL.Cmp(a[j].URL) < 0 } // Less 方法比较切片 a 中索引为 i 和 j 的两个元素的 URL，如果第一个小于第二个，则返回 true。

// WalletsByURL implements sort.Interface for []Wallet based on the URL field.
// WalletsByURL 基于 URL 字段为 []Wallet 实现 sort.Interface。
type WalletsByURL []Wallet // 定义 WalletsByURL 类型，它是一个 Wallet 的切片。

// Len 返回切片的长度。
func (w WalletsByURL) Len() int { return len(w) } // Len 方法返回切片 w 的长度。
// Swap 交换切片中两个元素的位置。
func (w WalletsByURL) Swap(i, j int) { w[i], w[j] = w[j], w[i] } // Swap 方法交换切片 w 中索引为 i 和 j 的两个元素。
// Less 比较两个元素的 URL，用于排序。
func (w WalletsByURL) Less(i, j int) bool { return w[i].URL().Cmp(w[j].URL()) < 0 } // Less 方法比较切片 w 中索引为 i 和 j 的两个元素的 URL，如果第一个小于第二个，则返回 true。
