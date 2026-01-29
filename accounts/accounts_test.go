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

package accounts

import (
	"bytes"
	"testing"

	"github.com/ethereum/go-ethereum/common/hexutil"
)

// TestTextHash tests the TextHash function.
// TestTextHash 测试 TextHash 函数。
func TestTextHash(t *testing.T) {
	t.Parallel()
	// Calculate the hash of the text "Hello Joe"
	// 计算文本 "Hello Joe" 的哈希
	hash := TextHash([]byte("Hello Joe"))
	// Define the expected hash result
	// 定义预期的哈希结果
	want := hexutil.MustDecode("0xa080337ae51c4e064c189e113edd0ba391df9206e2f49db658bb32cf2911730b")
	// Verify if the calculated hash matches the expected hash
	// 验证计算出的哈希是否与预期哈希匹配
	if !bytes.Equal(hash, want) {
		t.Fatalf("wrong hash: %x", hash)
	}
}
