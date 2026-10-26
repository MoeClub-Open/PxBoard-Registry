# PxBoard 官方插件源

独立的静态插件索引、安装组合、平台产物与发布工具。需要同级 `../PxBoard-Plugins` 和 `../PxBoard`。

官方线上源：

```text
https://raw.githubusercontent.com/MoeClub-Open/PxBoard-Registry/main/index.json
```

源码与索引存放在此仓库，安装包放在该仓库的 GitHub Releases。主项目默认使用上述线上源，也可在模式安装页或插件中心注册其他源。

当前源包含 15 个官方插件，完整使用当前版本需要 PxBoard 核心 0.4.1 或更新版本。品牌、账户注册、兑换码和推广返佣均为独立插件；推广 `affiliate` 与第三方登录 `oauth` 不随使用模式或数据迁移插件自动安装，需要时在插件中心单独安装并配置。

自用组合安装品牌、节点与订阅；公益 / 分享组合增加账户注册、兑换码和流量审计；商业组合提供账户注册、套餐订单、支付、审计与用户服务。节点与订阅会自动安装访问账户依赖。商业站可另选兑换码和推广，公益站默认没有推广入口。切换组合保留原有插件与数据，停用或卸载插件由管理员单独操作。

```sh
go run ./cmd/publish
go run ./cmd/registry
```

在面板中注册 `http://127.0.0.1:8091/index.json`。公开源将 `index.json` 与 `packages/` 放到 HTTPS 静态托管即可；开发服务器默认只监听本机。安装组合定义在 `bundles.json`。

发布其他平台会保留同版本其他平台的产物记录：

```sh
go run ./cmd/publish -platform linux-amd64
go run ./cmd/publish -platform linux-arm64
```

可选发布签名：

```sh
go run ./cmd/publish -generate-key release.key
go run ./cmd/publish -platform linux-amd64 -key-file release.key
```

生成命令显示发布公钥，将该公钥填入面板源设置。私钥已被 Git 忽略，不放入静态服务目录，不随源文件发布。发布签名验证二进制，HTTPS 保护索引与传输；仍应信任插件作者及声明权限。

`packages/` 是生成产物，不跟踪到 Git。更新插件代码后应提升插件版本并重新发布；同版本安装不会覆盖已经安装的版本。

## 发布到 GitHub

创建一个发布标签，将安装包的下载目录写入索引：

```sh
go run ./cmd/publish -platform linux-amd64 -artifact-base https://github.com/MoeClub-Open/PxBoard-Registry/releases/download/plugins-2026.10.27
go run ./cmd/publish -platform linux-arm64 -artifact-base https://github.com/MoeClub-Open/PxBoard-Registry/releases/download/plugins-2026.10.27
go run ./cmd/publish -platform windows-amd64 -artifact-base https://github.com/MoeClub-Open/PxBoard-Registry/releases/download/plugins-2026.10.27
```

使用 GitHub CLI 创建 Release，上传 `index.json` 和对应安装包，发布 Release 后将索引提交到 `main`。也可手动运行仓库的 **Publish plugin registry** 工作流，填写发布标签、核心 SDK 与官方插件的具体版本；核心 SDK 默认使用 `v0.4.0`。工作流会完成构建、上传、发布和索引更新，需要仓库的内容写入权限。

每次公开发布使用新的标签，不覆盖已发布版本的安装包。只有二进制可以下载且索引已经提交后，才应把源交给面板使用。开发本地源时省略 `-artifact-base`，索引中的安装包路径将指向本地 `packages/`。
