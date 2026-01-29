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

package accounts

import (
	"reflect"
	"sort"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/event"
)

// managerSubBufferSize determines how many incoming wallet events
// the manager will buffer in its channel.
// managerSubBufferSize 决定了管理器将在其通道中缓冲多少传入的钱包事件。
const managerSubBufferSize = 50

// Config is a legacy struct which is not used
// Config 是一个未使用的旧结构体
type Config struct {
	InsecureUnlockAllowed bool // Unused legacy-parameter // 未使用的旧参数
}

// newBackendEvent lets the manager know it should
// track the given backend for wallet updates.
// newBackendEvent 让管理器知道它应该跟踪给定的后端以获取钱包更新。
type newBackendEvent struct {
	backend   Backend
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

	quit chan chan error
	term chan struct{} // Channel is closed upon termination of the update loop // 更新循环终止时关闭的通道
	lock sync.RWMutex
}

// NewManager creates a generic account manager to sign transaction via various
// supported backends.
// NewManager 创建一个通用帐户管理器，通过各种支持的后端签署交易。
func NewManager(config *Config, backends ...Backend) *Manager {
	// Retrieve the initial list of wallets from the backends and sort by URL
	// 从后端检索钱包的初始列表并按 URL 排序
	var wallets []Wallet
	for _, backend := range backends {
		wallets = merge(wallets, backend.Wallets()...)
	}
	// Subscribe to wallet notifications from all backends
	// 订阅所有后端的钱包通知
	updates := make(chan WalletEvent, managerSubBufferSize)

	subs := make([]event.Subscription, len(backends))
	for i, backend := range backends {
		subs[i] = backend.Subscribe(updates)
	}
	// Assemble the account manager and return
	// 组装帐户管理器并返回
	am := &Manager{
		backends:    make(map[reflect.Type][]Backend),
		updaters:    subs,
		updates:     updates,
		newBackends: make(chan newBackendEvent),
		wallets:     wallets,
		quit:        make(chan chan error),
		term:        make(chan struct{}),
	}
	for _, backend := range backends {
		kind := reflect.TypeOf(backend)
		am.backends[kind] = append(am.backends[kind], backend)
	}
	go am.update()

	return am
}

// Close terminates the account manager's internal notification processes.
// Close 终止帐户管理器的内部通知进程。
func (am *Manager) Close() error {
	errc := make(chan error)
	am.quit <- errc
	return <-errc
}

// AddBackend starts the tracking of an additional backend for wallet updates.
// cmd/geth assumes once this func returns the backends have been already integrated.
// AddBackend 开始跟踪额外的后端以获取钱包更新。
// cmd/geth 假设一旦此函数返回，后端就已经被集成。
func (am *Manager) AddBackend(backend Backend) {
	done := make(chan struct{})
	am.newBackends <- newBackendEvent{backend, done}
	<-done
}

// update is the wallet event loop listening for notifications from the backends
// and updating the cache of wallets.
// update 是钱包事件循环，监听来自后端的通知并更新钱包缓存。
func (am *Manager) update() {
	// Close all subscriptions when the manager terminates
	// 当管理器终止时关闭所有订阅
	defer func() {
		am.lock.Lock()
		for _, sub := range am.updaters {
			sub.Unsubscribe()
		}
		am.updaters = nil
		am.lock.Unlock()
	}()

	// Loop until termination
	// 循环直到终止
	for {
		select {
		case event := <-am.updates:
			// Wallet event arrived, update local cache
			// 钱包事件到达，更新本地缓存
			am.lock.Lock()
			switch event.Kind {
			case WalletArrived:
				am.wallets = merge(am.wallets, event.Wallet)
			case WalletDropped:
				am.wallets = drop(am.wallets, event.Wallet)
			}
			am.lock.Unlock()

			// Notify any listeners of the event
			// 通知事件的任何监听器
			am.feed.Send(event)
		case event := <-am.newBackends:
			am.lock.Lock()
			// Update caches
			// 更新缓存
			backend := event.backend
			am.wallets = merge(am.wallets, backend.Wallets()...)
			am.updaters = append(am.updaters, backend.Subscribe(am.updates))
			kind := reflect.TypeOf(backend)
			am.backends[kind] = append(am.backends[kind], backend)
			am.lock.Unlock()
			close(event.processed)
		case errc := <-am.quit:
			// Close all owned wallets
			// 关闭所有拥有的钱包
			for _, w := range am.wallets {
				w.Close()
			}
			// Manager terminating, return
			// 管理器终止，返回
			errc <- nil
			// Signals event emitters the loop is not receiving values
			// to prevent them from getting stuck.
			// 信号事件发射器循环未接收值，以防止它们卡住。
			close(am.term)
			return
		}
	}
}

// Backends retrieves the backend(s) with the given type from the account manager.
// Backends 从帐户管理器检索具有给定类型的后端。
func (am *Manager) Backends(kind reflect.Type) []Backend {
	am.lock.RLock()
	defer am.lock.RUnlock()

	return am.backends[kind]
}

// Wallets returns all signer accounts registered under this account manager.
// Wallets 返回在此帐户管理器下注册的所有签名帐户。
func (am *Manager) Wallets() []Wallet {
	am.lock.RLock()
	defer am.lock.RUnlock()

	return am.walletsNoLock()
}

// walletsNoLock returns all registered wallets. Callers must hold am.lock.
// walletsNoLock 返回所有注册的钱包。调用者必须持有 am.lock。
func (am *Manager) walletsNoLock() []Wallet {
	cpy := make([]Wallet, len(am.wallets))
	copy(cpy, am.wallets)
	return cpy
}

// Wallet retrieves the wallet associated with a particular URL.
// Wallet 检索与特定 URL 关联的钱包。
func (am *Manager) Wallet(url string) (Wallet, error) {
	am.lock.RLock()
	defer am.lock.RUnlock()

	parsed, err := parseURL(url)
	if err != nil {
		return nil, err
	}
	for _, wallet := range am.walletsNoLock() {
		if wallet.URL() == parsed {
			return wallet, nil
		}
	}
	return nil, ErrUnknownWallet
}

// Accounts returns all account addresses of all wallets within the account manager
// Accounts 返回帐户管理器内所有钱包的所有帐户地址
func (am *Manager) Accounts() []common.Address {
	am.lock.RLock()
	defer am.lock.RUnlock()

	addresses := make([]common.Address, 0) // return [] instead of nil if empty // 如果为空，返回 [] 而不是 nil
	for _, wallet := range am.wallets {
		for _, account := range wallet.Accounts() {
			addresses = append(addresses, account.Address)
		}
	}
	return addresses
}

// Find attempts to locate the wallet corresponding to a specific account. Since
// accounts can be dynamically added to and removed from wallets, this method has
// a linear runtime in the number of wallets.
// Find 尝试定位对应于特定帐户的钱包。
// 由于帐户可以动态添加到钱包中或从钱包中移除，此方法在钱包数量上具有线性运行时间。
func (am *Manager) Find(account Account) (Wallet, error) {
	am.lock.RLock()
	defer am.lock.RUnlock()

	for _, wallet := range am.wallets {
		if wallet.Contains(account) {
			return wallet, nil
		}
	}
	return nil, ErrUnknownAccount
}

// Subscribe creates an async subscription to receive notifications when the
// manager detects the arrival or departure of a wallet from any of its backends.
// Subscribe 创建异步订阅，以便在管理器检测到任何后端的钱包到达或离开时接收通知。
func (am *Manager) Subscribe(sink chan<- WalletEvent) event.Subscription {
	return am.feed.Subscribe(sink)
}

// merge is a sorted analogue of append for wallets, where the ordering of the
// origin list is preserved by inserting new wallets at the correct position.
//
// The original slice is assumed to be already sorted by URL.
// merge 是钱包 append 的排序模拟，其中通过在正确位置插入新钱包来保留原始列表的顺序。
//
// 假定原始切片已按 URL 排序。
func merge(slice []Wallet, wallets ...Wallet) []Wallet {
	for _, wallet := range wallets {
		n := sort.Search(len(slice), func(i int) bool { return slice[i].URL().Cmp(wallet.URL()) >= 0 })
		if n == len(slice) {
			slice = append(slice, wallet)
			continue
		}
		slice = append(slice[:n], append([]Wallet{wallet}, slice[n:]...)...)
	}
	return slice
}

// drop is the counterpart of merge, which looks up wallets from within the sorted
// cache and removes the ones specified.
// drop 是 merge 的对应部分，它从排序的缓存中查找钱包并删除指定的钱包。
func drop(slice []Wallet, wallets ...Wallet) []Wallet {
	for _, wallet := range wallets {
		n := sort.Search(len(slice), func(i int) bool { return slice[i].URL().Cmp(wallet.URL()) >= 0 })
		if n == len(slice) {
			// Wallet not found, may happen during startup
			// 找不到钱包，可能在启动期间发生
			continue
		}
		slice = append(slice[:n], slice[n+1:]...)
	}
	return slice
}
