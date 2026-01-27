# Go Ethereum (Geth)

Go Ethereum (通常称为 **Geth**) 是以太坊协议的官方 Golang 实现。它是以太坊网络中最流行的客户端。

> **中文说明**: 本文档是对官方 README 的中文注释版，旨在帮助开发者快速理解 Geth 的构建、运行和使用。

[![API Reference](
https://pkg.go.dev/badge/github.com/ethereum/go-ethereum
)](https://pkg.go.dev/github.com/ethereum/go-ethereum?tab=doc)
[![Go Report Card](https://goreportcard.com/badge/github.com/ethereum/go-ethereum)](https://goreportcard.com/report/github.com/ethereum/go-ethereum)
[![Travis](https://app.travis-ci.com/ethereum/go-ethereum.svg?branch=master)](https://app.travis-ci.com/github/ethereum/go-ethereum)
[![Discord](https://img.shields.io/badge/discord-join%20chat-blue.svg)](https://discord.gg/nthXNEv)
[![Twitter](https://img.shields.io/twitter/follow/go_ethereum)](https://x.com/go_ethereum)

我们为稳定版本和不稳定的 master 分支提供自动化构建。二进制归档文件发布在 https://geth.ethereum.org/downloads/。

## Building the source (源码构建)

> **注意**: 构建 Geth 需要安装 Go 语言环境（版本 1.23 或更高）以及 C 编译器（如 GCC）。

有关先决条件和详细的构建说明，请阅读 [安装说明](https://geth.ethereum.org/docs/getting-started/installing-geth)。

安装好依赖后，你可以运行以下命令来构建 `geth`：

```shell
make geth
```

或者，构建全套实用工具：

```shell
make all
```

## Executables (可执行文件)

Go-ethereum 项目在 `cmd` 目录下提供了多个实用工具。以下是它们的详细说明：

|  命令 (Command) | 描述 (Description)                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                        |
| :--------: | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **`geth`** | **以太坊主要 CLI 客户端**。它是进入以太坊网络（主网、测试网或私有网）的入口点。它可以作为全节点（默认）、归档节点（保留所有历史状态）或轻节点（实时检索数据）运行。它可以被其他进程用作进入以太坊网络的网关，通过 HTTP、WebSocket 和/或 IPC 传输层暴露 JSON RPC 端点。使用 `geth --help` 查看帮助，或访问 [CLI 页面](https://geth.ethereum.org/docs/fundamentals/command-line-options) 查看命令行选项。 |
|   `clef`   | **独立签名工具**。它可以作为 `geth` 的后端签名器使用，提供更安全的交易签名管理。                                                                                                                                                                                                                                                                                                                                                                                                                                                        |
|  `devp2p`  | **网络层调试工具**。用于与网络层上的节点进行交互，而无需运行完整的区块链。主要用于调试 P2P 协议。                                                                                                                                                                                                                                                                                                                                                                                                                                       |
|  `abigen`  | **源代码生成器**。用于将以太坊智能合约定义转换为易于使用的、编译时类型安全的 Go 包。它基于普通的 [以太坊合约 ABI](https://docs.soliditylang.org/en/develop/abi-spec.html) 运行，如果提供了合约字节码，还可以获得扩展功能。它也支持直接读取 Solidity 源文件，使开发更加流畅。详情请参阅我们的 [原生 DApps](https://geth.ethereum.org/docs/developers/dapp-developer/native-bindings) 页面。                                  |
|   `evm`    | **EVM (以太坊虚拟机) 开发工具**。它能够在可配置的环境和执行模式下运行字节码片段。它的目的是允许对 EVM 操作码进行隔离的、细粒度的调试（例如 `evm --code 60ff60ff --debug run`）。                                                                                                                                                                                                                                               |
| `rlpdump`  | **RLP 数据转储工具**。用于将二进制 RLP ([递归长度前缀](https://ethereum.org/en/developers/docs/data-structures-and-encoding/rlp)) 数据（以太坊协议在网络和共识层面使用的数据编码方式）转换为用户更友好的分层表示形式（例如 `rlpdump --hex CE0183FFFFFFC4C304050583616263`）。                                                                                                                                                                                |

## Running `geth` (运行 Geth)

在这里列出所有可能的命令行标志超出了范围（请查阅我们的 [CLI Wiki 页面](https://geth.ethereum.org/docs/fundamentals/command-line-options)），但我们列举了一些常用的参数组合，让你快速了解如何运行自己的 `geth` 实例。

### Hardware Requirements (硬件要求)

**最低配置:**

* CPU: 4核心以上
* 内存: 8GB RAM
* 存储: 1TB 可用空间（用于同步主网）
* 网络: 8 MBit/sec 下载速度

**推荐配置:**

* CPU: 快速 8核心以上
* 内存: 16GB+ RAM
* 存储: 高性能 SSD，至少 1TB 可用空间
* 网络: 25+ MBit/sec 下载速度

### Full node on the main Ethereum network (主网全节点)

目前为止最常见的场景是用户只想简单地与以太坊网络交互：创建账户、转移资金、部署和与合约交互。对于这种特定用例，用户并不关心多年的历史数据，所以我们可以快速同步到网络的当前状态。为此：

```shell
$ geth console
```

该命令将：
 * 以 **Snap Sync (快照同步)** 模式启动 `geth`（默认模式，可以通过 `--syncmode` 标志更改）。这将下载更多数据以换取避免处理以太坊网络的整个历史记录，因为处理历史记录非常消耗 CPU。
 * 启动内置的交互式 [JavaScript 控制台](https://geth.ethereum.org/docs/interacting-with-geth/javascript-console)（通过尾部的 `console` 子命令）。通过它你可以使用 [`web3` 方法](https://github.com/ChainSafe/web3.js/blob/0.20.7/DOCUMENTATION.md)（注意：`geth` 内置的 `web3` 版本很旧，可能与官方文档不完全同步）以及 `geth` 自己的 [管理 API](https://geth.ethereum.org/docs/interacting-with-geth/rpc) 进行交互。这个工具是可选的，如果你不加它，你以后总是可以通过 `geth attach` 附加到一个已经在运行的 `geth` 实例上。

### A Full node on the Holesky test network (Holesky 测试网全节点)

对于开发者来说，如果你想尝试创建以太坊合约，你几乎肯定想在没有任何真实资金风险的情况下进行，直到你掌握了整个系统。换句话说，你不想连接到主网，而是想加入 **测试 (Test)** 网络，它与主网完全等效，但只使用测试用的以太币。

```shell
$ geth --holesky console
```

`console` 子命令的含义同上，在测试网上同样有用。

指定 `--holesky` 标志会稍微重新配置你的 `geth` 实例：

 * 客户端将连接到 Holesky 测试网络，而不是连接到以太坊主网。Holesky 使用不同的 P2P 引导节点、不同的网络 ID 和创世状态。
 * `geth` 将把自己嵌套在一个 `holesky` 子文件夹中（例如 Linux 上的 `~/.ethereum/holesky`），而不是使用默认的数据目录（例如 Linux 上的 `~/.ethereum`）。注意，在 OSX 和 Linux 上，这也意味着附加到正在运行的测试网节点需要使用自定义端点，因为 `geth attach` 默认尝试附加到生产节点端点，例如 `geth attach <datadir>/holesky/geth.ipc`。Windows 用户不受此影响。

*注意：虽然一些内部保护措施可以防止交易在主网和测试网之间交叉，但你应该始终为测试和真实资金使用单独的账户。除非你手动移动账户，否则 `geth` 默认会正确地隔离这两个网络，并且不会让任何账户在它们之间可用。*

### Configuration (配置)

除了向 `geth` 二进制文件传递大量标志外，你还可以通过配置文件传递配置：

```shell
$ geth --config /path/to/your_config.toml
```

要了解该文件应该是什么样子，你可以使用 `dumpconfig` 子命令导出你现有的配置：

```shell
$ geth --your-favourite-flags dumpconfig
```

#### Docker quick start (Docker 快速入门)

在你的机器上启动并运行以太坊最快的方法之一是使用 Docker：

```shell
docker run -d --name ethereum-node -v /Users/alice/ethereum:/root \
           -p 8545:8545 -p 30303:30303 \
           ethereum/client-go
```

这将以 Snap-sync 模式启动 `geth`，并将 DB 内存限制设为 1GB。它还会在你的主目录中创建一个持久卷来保存你的区块链数据，并映射默认端口。还有一个 `alpine` 标签可用于该镜像的精简版本。

如果你想从其他容器和/或主机访问 RPC，不要忘记加上 `--http.addr 0.0.0.0`。默认情况下，`geth` 绑定到本地接口，RPC 端点无法从外部访问。

### Programmatically interfacing `geth` nodes (通过编程接口访问 Geth)

作为一个开发者，你迟早会想要通过自己的程序而不是通过控制台手动与 `geth` 和以太坊网络交互。为了帮助实现这一点，`geth` 内置了对基于 JSON-RPC 的 API 的支持（[标准 API](https://ethereum.org/en/developers/docs/apis/json-rpc/) 和 [`geth` 特定 API](https://geth.ethereum.org/docs/interacting-with-geth/rpc)）。这些 API 可以通过 HTTP、WebSocket 和 IPC（基于 UNIX 平台的 UNIX 套接字，Windows 上的命名管道）暴露。

IPC 接口默认启用，并暴露 `geth` 支持的所有 API，而 HTTP 和 WS 接口需要手动启用，并且出于安全原因只暴露一部分 API。这些可以按预期打开/关闭和配置。

HTTP JSON-RPC API 选项:

  * `--http` 启用 HTTP-RPC 服务器
  * `--http.addr` HTTP-RPC 服务器监听接口 (默认: `localhost`)
  * `--http.port` HTTP-RPC 服务器监听端口 (默认: `8545`)
  * `--http.api` 通过 HTTP-RPC 接口提供的 API (默认: `eth,net,web3`)
  * `--http.corsdomain` 允许跨域请求的域名逗号分隔列表 (浏览器强制执行)
  * `--ws` 启用 WS-RPC 服务器
  * `--ws.addr` WS-RPC 服务器监听接口 (默认: `localhost`)
  * `--ws.port` WS-RPC 服务器监听端口 (默认: `8546`)
  * `--ws.api` 通过 WS-RPC 接口提供的 API (默认: `eth,net,web3`)
  * `--ws.origins` 允许 WebSocket 请求的来源
  * `--ipcdisable` 禁用 IPC-RPC 服务器
  * `--ipcpath` 数据目录中 IPC 套接字/管道的文件名 (显式路径可转义)

你需要使用你自己的编程环境的能力（库、工具等）通过 HTTP、WS 或 IPC 连接到配置了上述标志的 `geth` 节点，并且你需要通过所有传输层使用 [JSON-RPC](https://www.jsonrpc.org/specification) 协议。你可以重用同一个连接进行多次请求！

**注意：在这样做之前，请务必理解打开基于 HTTP/WS 的传输层的安全隐患！互联网上的黑客正积极试图破坏暴露了 API 的以太坊节点！此外，所有浏览器标签页都可以访问本地运行的 Web 服务器，因此恶意网页可能会尝试破坏本地可用的 API！**

### Operating a private network (运行私有网络)

维护你自己的私有网络比较复杂，因为官方网络中理所当然的许多配置都需要手动设置。

遗憾的是，自从 [合并 (The Merge)](https://ethereum.org/en/roadmap/merge/) 以来，如果不设置相应的信标链 (Beacon Chain)，就不再可能轻松地设置 geth 节点网络。

根据你的使用情况，有三种不同的解决方案：

  * 如果你正在寻找一种简单的方法在 CI 中用 Go 测试智能合约，你可以使用 [模拟后端 (Simulated Backend)](https://geth.ethereum.org/docs/developers/dapp-developer/native-bindings#blockchain-simulator)。
  * 如果你想要一个方便的单节点环境进行测试，你可以使用我们的 [开发者模式 (Dev Mode)](https://geth.ethereum.org/docs/developers/dapp-developer/dev-mode)。
  * 如果你正在寻找多节点测试网络，你可以使用 [Kurtosis](https://geth.ethereum.org/docs/fundamentals/kurtosis) 轻松设置一个。

## Contribution (贡献)

感谢你考虑为源代码做出贡献！我们要欢迎来自互联网任何地方的贡献，哪怕是最小的修复我们也心存感激！

如果你想为 go-ethereum 做贡献，请 fork、修复、提交并发送 pull request，供维护者审查并合并到主代码库中。如果你希望提交更复杂的更改，请先在 [我们的 Discord 服务器](https://discord.gg/invite/nthXNEv) 上与核心开发人员联系，以确保这些更改符合项目的总体理念，并/或获得一些早期反馈，这可以使你的工作更加轻松，也能让我们的审查和合并过程快速简单。

请确保你的贡献遵守我们的编码指南：

 * 代码必须遵守官方 Go [格式化](https://golang.org/doc/effective_go.html#formatting) 指南（即使用 [gofmt](https://golang.org/cmd/gofmt/)）。
 * 代码必须按照官方 Go [注释](https://golang.org/doc/effective_go.html#commentary) 指南进行文档记录。
 * Pull requests 需要基于并针对 `master` 分支打开。
 * 提交消息应以它们修改的包为前缀。
   * 例如 "eth, rpc: make trace configs optional"

有关配置环境、管理项目依赖和测试过程的更多详细信息，请参阅 [开发者指南](https://geth.ethereum.org/docs/developers/geth-developer/dev-guide)。

### Contributing to geth.ethereum.org (为官网做贡献)

对于对 [go-ethereum 网站](https://geth.ethereum.org) 的贡献，请检出并针对 `website` 分支提出 pull request。
有关更多详细说明，请参阅 `website` 分支的 [README](https://github.com/ethereum/go-ethereum/tree/website#readme) 或网站的 [贡献页面](https://geth.ethereum.org/docs/developers/geth-developer/contributing)。

## License (许可证)

Go-ethereum 库（即 `cmd` 目录之外的所有代码）根据 [GNU Lesser General Public License v3.0](https://www.gnu.org/licenses/lgpl-3.0.en.html) 许可，该文件也包含在我们仓库的 `COPYING.LESSER` 文件中。

Go-ethereum 二进制文件（即 `cmd` 目录内的所有代码）根据 [GNU General Public License v3.0](https://www.gnu.org/licenses/gpl-3.0.en.html) 许可，该文件也包含在我们仓库的 `COPYING` 文件中。
