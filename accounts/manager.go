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
	"reflect" // 导入 reflect 包，用于运行时反射。
	"sort"    // 导入 sort 包，用于对数据进行排序。
	"sync"    // 导入 sync 包，提供基本的同步原语，如互斥锁。

	"github.com/ethereum/go-ethereum/common" // 导入 go-ethereum 的 common 包。
	"github.com/ethereum/go-ethereum/event"  // 导入 go-ethereum 的 event 包，用于处理事件。
)

// managerSubBufferSize determines how many incoming wallet events
// the manager will buffer in its channel.
// managerSubBufferSize 决定了管理器将在其通道中缓冲多少传入的钱包事件。
const managerSubBufferSize = 50 // 定义管理器订阅缓冲区的容量为 50。

// Config is a legacy struct which is not used
// Config 是一个未使用的旧结构体
type Config struct {
	InsecureUnlockAllowed bool // Unused legacy-parameter // 未使用的旧参数
}

// newBackendEvent lets the manager know it should
// track the given backend for wallet updates.
// newBackendEvent 让管理器知道它应该跟踪给定的后端以获取钱包更新。
type newBackendEvent struct {
	backend   Backend         // 后端接口。
	processed chan struct{} // Informs event emitter that backend has been integrated // 通知事件发射器后端已被集成
}

// Manager is an overarching account manager that can communicate with various
// backends for signing transactions.
// Manager 是一个首要的帐户管理器，可以与各种后端通信以签署交易。
type Manager struct {
	backends    map[reflect.Type][]Backend // Index of backends currently registered // 当前注册的后端索引
	updaters    []event.Subscription       // Wallet update subscriptions for all backends // 所有后端的钱包更新订阅
	updates     chan WalletEvent           // Subscription sink for backend wallet changes // 后端钱包更改的订阅接收器
	newBackends chan newBackendEvent       // Incoming backends to be tracked by the manager // 管理器要跟踪的传入后端
	wallets     []Wallet                   // Cache of all wallets from all registered backends // 来自所有注册后端的所有钱包的缓存

	feed event.Feed // Wallet feed notifying of arrivals/departures // 钱包到达/离开的通知源

	quit chan chan error // 退出信号通道。
	term chan struct{}   // Channel is closed upon termination of the update loop // 更新循环终止时关闭的通道
	lock sync.RWMutex   // 读写锁，用于保护对 Manager 内部状态的并发访问。
}

// NewManager creates a generic account manager to sign transaction via various
// supported backends.
// NewManager 创建一个通用帐户管理器，通过各种支持的后端签署交易。
func NewManager(config *Config, backends ...Backend) *Manager { // 定义 NewManager 函数，用于创建一个新的 Manager 实例。
	// Retrieve the initial list of wallets from the backends and sort by URL
	// 从后端检索钱包的初始列表并按 URL 排序
	var wallets []Wallet // 声明一个 Wallet 切片。
	for _, backend := range backends { // 遍历所有后端。
		wallets = merge(wallets, backend.Wallets()...) // 合并后端钱包到 wallets 切片中。
	}
	// Subscribe to wallet notifications from all backends
	// 订阅所有后端的钱包通知
	updates := make(chan WalletEvent, managerSubBufferSize) // 创建一个带缓冲的 WalletEvent 通道。

	subs := make([]event.Subscription, len(backends)) // 创建一个 Subscription 切片。
	for i, backend := range backends { // 遍历所有后端。
		subs[i] = backend.Subscribe(updates) // 订阅后端的钱包事件。
	}
	// Assemble the account manager and return
	// 组装帐户管理器并返回
	am := &Manager{ // 创建一个新的 Manager 实例。
		backends:    make(map[reflect.Type][]Backend), // 初始化 backends map。
		updaters:    subs,                             // 设置 updaters。
		updates:     updates,                          // 设置 updates 通道。
		newBackends: make(chan newBackendEvent),       // 初始化 newBackends 通道。
		wallets:     wallets,                          // 设置 wallets。
		quit:        make(chan chan error),            // 初始化 quit 通道。
		term:        make(chan struct{}),              // 初始化 term 通道。
	}
	for _, backend := range backends { // 遍历所有后端。
		kind := reflect.TypeOf(backend) // 获取后端的类型。
		am.backends[kind] = append(am.backends[kind], backend) // 将后端按类型添加到 backends map 中。
	}
	go am.update() // 启动 update goroutine。

	return am // 返回 Manager 实例。
}

// Close terminates the account manager's internal notification processes.
// Close 终止帐户管理器的内部通知进程。
func (am *Manager) Close() error { // 定义 Close 方法，用于关闭 Manager。
	errc := make(chan error) // 创建一个 error 通道。
	am.quit <- errc          // 发送退出信号。
	return <-errc            // 等待并返回错误。
}

// AddBackend starts the tracking of an additional backend for wallet updates.
// cmd/geth assumes once this func returns the backends have been already integrated.
// AddBackend 开始跟踪额外的后端以获取钱包更新。
// cmd/geth 假设一旦此函数返回，后端就已经被集成。
func (am *Manager) AddBackend(backend Backend) { // 定义 AddBackend 方法，用于添加后端。
	done := make(chan struct{}) // 创建一个 struct{} 通道。
	am.newBackends <- newBackendEvent{backend, done} // 发送 newBackendEvent 事件。
	<-done                   // 等待后端集成完成。
}

// update is the wallet event loop listening for notifications from the backends
// and updating the cache of wallets.
// update 是钱包事件循环，监听来自后端的通知并更新钱包缓存。
func (am *Manager) update() { // 定义 update 方法，用于处理钱包事件。
	// Close all subscriptions when the manager terminates
	// 当管理器终止时关闭所有订阅
	defer func() { // 定义一个延迟函数。
		am.lock.Lock() // 加锁。
		for _, sub := range am.updaters { // 遍历所有订阅。
			sub.Unsubscribe() // 取消订阅。
		}
		am.updaters = nil // 清空 updaters。
		am.lock.Unlock() // 解锁。
	}()

	// Loop until termination
	// 循环直到终止
	for { // 无限循环。
		select { // 使用 select 处理多个通道。
		case event := <-am.updates: // 接收钱包事件。
			// Wallet event arrived, update local cache
			// 钱包事件到达，更新本地缓存
			am.lock.Lock() // 加锁。
			switch event.Kind { // 根据事件类型进行处理。
			case WalletArrived: // 如果是钱包到达事件。
				am.wallets = merge(am.wallets, event.Wallet) // 合并钱包。
			case WalletDropped: // 如果是钱包掉线事件。
				am.wallets = drop(am.wallets, event.Wallet) // 移除钱包。
			}
			am.lock.Unlock() // 解锁。

			// Notify any listeners of the event
			// 通知事件的任何监听器
			am.feed.Send(event) // 发送事件通知。
		case event := <-am.newBackends: // 接收新后端事件。
			am.lock.Lock() // 加锁。
			// Update caches
			// 更新缓存
			backend := event.backend // 获取后端。
			am.wallets = merge(am.wallets, backend.Wallets()...) // 合并钱包。
			am.updaters = append(am.updaters, backend.Subscribe(am.updates)) // 添加订阅。
			kind := reflect.TypeOf(backend) // 获取后端类型。
			am.backends[kind] = append(am.backends[kind], backend) // 添加后端到 map。
			am.lock.Unlock() // 解锁。
			close(event.processed) // 关闭 processed 通道。
		case errc := <-am.quit: // 接收退出信号。
			// Close all owned wallets
			// 关闭所有拥有的钱包
			for _, w := range am.wallets { // 遍历所有钱包。
				w.Close() // 关闭钱包。
			}
			// Manager terminating, return
			// 管理器终止，返回
			errc <- nil // 发送 nil 错误。
			// Signals event emitters the loop is not receiving values
			// to prevent them from getting stuck.
			// 信号事件发射器循环未接收值，以防止它们卡住。
			close(am.term) // 关闭 term 通道。
			return         // 返回。
		}
	}
}

// Backends retrieves the backend(s) with the given type from the account manager.
// Backends 从帐户管理器检索具有给定类型的后端。
func (am *Manager) Backends(kind reflect.Type) []Backend { // 定义 Backends 方法，用于获取指定类型的后端。
	am.lock.RLock() // 加读锁。
	defer am.lock.RUnlock() // 延迟解锁。

	return am.backends[kind] // 返回指定类型的后端。
}

// Wallets returns all signer accounts registered under this account manager.
// Wallets 返回在此帐户管理器下注册的所有签名帐户。
func (am *Manager) Wallets() []Wallet { // 定义 Wallets 方法，用于获取所有钱包。
	am.lock.RLock() // 加读锁。
	defer am.lock.RUnlock() // 延迟解锁。

	return am.walletsNoLock() // 返回所有钱包。
}

// walletsNoLock returns all registered wallets. Callers must hold am.lock.
// walletsNoLock 返回所有注册的钱包。调用者必须持有 am.lock。
func (am *Manager) walletsNoLock() []Wallet { // 定义 walletsNoLock 方法，用于在持有锁的情况下获取所有钱包。
	cpy := make([]Wallet, len(am.wallets)) // 创建一个 Wallet 切片副本。
	copy(cpy, am.wallets)                  // 复制钱包数据。
	return cpy                             // 返回副本。
}

// Wallet retrieves the wallet associated with a particular URL.
// Wallet 检索与特定 URL 关联的钱包。
func (am *Manager) Wallet(url string) (Wallet, error) { // 定义 Wallet 方法，用于根据 URL 获取钱包。
	am.lock.RLock() // 加读锁。
	defer am.lock.RUnlock() // 延迟解锁。

	parsed, err := parseURL(url) // 解析 URL。
	if err != nil {               // 如果解析失败。
		return nil, err // 返回错误。
	}
	for _, wallet := range am.walletsNoLock() { // 遍历所有钱包。
		if wallet.URL() == parsed { // 如果 URL 匹配。
			return wallet, nil // 返回钱包。
		}
	}
	return nil, ErrUnknownWallet // 返回未知钱包错误。
}

// Accounts returns all account addresses of all wallets within the account manager
// Accounts 返回帐户管理器内所有钱包的所有帐户地址
func (am *Manager) Accounts() []common.Address { // 定义 Accounts 方法，用于获取所有账户地址。
	am.lock.RLock() // 加读锁。
	defer am.lock.RUnlock() // 延迟解锁。

	addresses := make([]common.Address, 0) // return [] instead of nil if empty // 如果为空，返回 [] 而不是 nil
	for _, wallet := range am.wallets { // 遍历所有钱包。
		for _, account := range wallet.Accounts() { // 遍历钱包中的所有账户。
			addresses = append(addresses, account.Address) // 添加账户地址到切片。
		}
	}
	return addresses // 返回所有账户地址。
}

// Find attempts to locate the wallet corresponding to a specific account. Since
// accounts can be dynamically added to and removed from wallets, this method has
// a linear runtime in the number of wallets.
// Find 尝试定位对应于特定帐户的钱包。
// 由于帐户可以动态添加到钱包中或从钱包中移除，此方法在钱包数量上具有线性运行时间。
func (am *Manager) Find(account Account) (Wallet, error) { // 定义 Find 方法，用于查找特定账户对应的钱包。
	am.lock.RLock() // 加读锁。
	defer am.lock.RUnlock() // 延迟解锁。

	for _, wallet := range am.wallets { // 遍历所有钱包。
		if wallet.Contains(account) { // 如果钱包包含该账户。
			return wallet, nil // 返回钱包。
		}
	}
	return nil, ErrUnknownAccount // 返回未知账户错误。
}

// Subscribe creates an async subscription to receive notifications when the
// manager detects the arrival or departure of a wallet from any of its backends.
// Subscribe 创建异步订阅，以便在管理器检测到任何后端的钱包到达或离开时接收通知。
func (am *Manager) Subscribe(sink chan<- WalletEvent) event.Subscription { // 定义 Subscribe 方法，用于订阅钱包事件。
	return am.feed.Subscribe(sink) // 订阅 feed。
}

// merge is a sorted analogue of append for wallets, where the ordering of the
// origin list is preserved by inserting new wallets at the correct position.
//
// The original slice is assumed to be already sorted by URL.
// merge 是钱包 append 的排序模拟，其中通过在正确位置插入新钱包来保留原始列表的顺序。
//
// 假定原始切片已按 URL 排序。
func merge(slice []Wallet, wallets ...Wallet) []Wallet { // 定义 merge 函数，用于合并钱包切片。
	for _, wallet := range wallets { // 遍历所有要合并的钱包。
		n := sort.Search(len(slice), func(i int) bool { return slice[i].URL().Cmp(wallet.URL()) >= 0 }) // 查找插入位置。
		if n == len(slice) { // 如果插入位置在末尾。
			slice = append(slice, wallet) // 直接追加。
			continue                      // 继续下一个。
		}
		slice = append(slice[:n], append([]Wallet{wallet}, slice[n:]...)...) // 插入钱包。
	}
	return slice // 返回合并后的切片。
}

// drop is the counterpart of merge, which looks up wallets from within the sorted
// cache and removes the ones specified.
// drop 是 merge 的对应部分，它从排序的缓存中查找钱包并删除指定的钱包。
func drop(slice []Wallet, wallets ...Wallet) []Wallet { // 定义 drop 函数，用于从切片中移除钱包。
	for _, wallet := range wallets { // 遍历所有要移除的钱包。
		n := sort.Search(len(slice), func(i int) bool { return slice[i].URL().Cmp(wallet.URL()) >= 0 }) // 查找钱包位置。
		if n == len(slice) { // 如果找不到。
			// Wallet not found, may happen during startup
			// 找不到钱包，可能在启动期间发生
			continue // 继续下一个。
		}
		slice = append(slice[:n], slice[n+1:]...) // 移除钱包。
	}
	return slice // 返回处理后的切片。
}
