# Go 后端模板

单应用、单 Go module 的工程骨架。已实现配置校验、JSON 日志、GORM PostgreSQL 连接池、HTTP 生命周期、健康检查、显式 SQL 迁移命令及工程验证；没有登录、Token、RBAC、用户、角色、审计或 CRUD 业务功能。

## 工具链与依赖

| 项目 | 固定版本 / 状态 |
| --- | --- |
| Go | 1.27.3，见 `.go-version` 与 `go.mod`；支持 Go 自动工具链下载，不修改系统安装 |
| Gin | 1.12.0，使用其传递引入的 validator/v10 10.30.1；尚无业务 DTO 校验 |
| GORM | 1.31.2 |
| PostgreSQL 驱动 | gorm.io/driver/postgres 1.6.3，底层 pgx/v5 5.11.0 |
| goose | 3.29.0，Provider API + PostgreSQL 会话级 advisory lock |
| 漏洞检查 | golang.org/x/vuln 1.9.0，通过 `go tool govulncheck` 运行 |
| PostgreSQL | 本地验证 18.6，CI 使用 18 系列；这只是验证环境，Java、Go 共用的生产版本基线仍待确定 |

Gin 间接引入的 quic-go 显式提升到 0.59.1，修复扫描发现的 GO-2026-5676；本工程未启用 HTTP/3。实际依赖与校验信息以 `go.mod`、`go.sum` 为准，升级后重新执行检查。版本依据：[Go 下载](https://go.dev/dl/)、[Gin 发布](https://github.com/gin-gonic/gin/releases)、[GORM 连接](https://gorm.io/docs/connecting_to_the_database.html)、[goose](https://github.com/pressly/goose)、[漏洞说明](https://pkg.go.dev/vuln/GO-2026-5676)。

## 目录与边界

```text
go/
├── cmd/server/                 # 服务进程入口
├── cmd/manage/                 # 显式迁移命令
├── internal/app/               # 手动装配依赖、资源清理
├── internal/platform/
│   ├── buildinfo/              # 编译工具链与来源提交
│   ├── config/                 # 类型化配置与校验
│   ├── logging/                # slog JSON，UTC 毫秒时间
│   ├── postgres/               # 连接池与启动连通性检查
│   ├── httpserver/             # HTTP、中间件、健康检查、优雅停机
│   ├── migration/              # goose Provider
│   └── transaction/            # 显式事务回调
├── internal/security/          # 仅包说明，尚无密码或 Token 实现
├── internal/user/postgres/      # 管理用户持久化实体，尚无 Repository 或业务表
├── internal/member/postgres/    # 完整会员持久化实体，尚无会员业务或业务表
├── internal/auth,user,role,…/   # 业务根包仅包说明，不预建业务类型和空接口
├── migrations/sql/             # 嵌入式迁移资源，当前只有说明文件
├── tests/integration/          # PostgreSQL 隔离 schema 测试
├── scripts/check.ps1
└── scripts/check.sh
```

后续业务模块按实际需要增加 `service.go`、业务类型和窄 Repository 接口，再增加 `http/`、`postgres/` 适配包。HTTP DTO、业务类型与数据库实体分别维护；业务根包不导入 Gin 或 GORM，具体适配器由 app 装配。当前有 `internal/user/postgres/model.go` 的 `AdminUser` 和 `internal/member/postgres/model.go` 的 `MemberUser` 持久化实体；尚无 Service、Repository 或 HTTP 业务实现。

用户实体分别映射 `admin_user`、`member_user`，可空字段使用指针，密码哈希不参与 JSON 序列化，创建字段限制为插入时写入，更新时间由后续 Repository 显式维护。会员实体保留完整参考字段并补充 username；生日为 date，标签 ID 使用 GORM JSON 序列化映射 jsonb，积分和经验使用 int64。会员字段预留不代表已实现会员登录、积分、等级或标签管理。字段与后续约束见[业务规划](../../docs/业务规划_2026-10-21.md)。实体标签不创建数据库约束；建表、唯一性、外键及索引须通过 goose SQL 落实，禁止调用 AutoMigrate。当前没有迁移 SQL，也未创建用户表。

事务由业务 Service 决定边界。`transaction.Runner.Within` 提供绑定请求 context 的事务句柄，回调内把具体 Repository 绑定到该句柄，不能使用全局连接。当前只验证了基础事务能力，未实现跨业务模块的事务抽象或管理员并发保护。

## 配置与启动

服务仅读取进程环境变量，**不自动读取 `.env`**。完整变量与默认值见 [.env.example](.env.example)。实际 `.env` 与构建产物由本目录 `.gitignore` 排除，`docs/` 正常参与版本管理。

必填项为 `DB_NAME`、`DB_USERNAME`、`DB_PASSWORD`。默认数据库地址 `127.0.0.1:5432`、服务地址 `127.0.0.1:8081`。默认 `DB_SSLMODE=verify-full`；开发机可显式改为 `disable`，`APP_ENV=production` 要求 `verify-full`。受信证书可通过 `DB_SSLROOTCERT` 指定。默认仅监听本机，部署时显式指定监听地址。

Windows PowerShell，在本目录执行（替换密码占位值）：

```powershell
$env:DB_NAME = 'template_go'
$env:DB_USERNAME = 'template_go'
$env:DB_PASSWORD = 'replace-with-local-password'
$env:DB_SSLMODE = 'disable'
go run ./cmd/manage migrate up
go run ./cmd/server
```

Linux/macOS，在本目录执行：

```sh
export DB_NAME=template_go
export DB_USERNAME=template_go
export DB_PASSWORD='replace-with-local-password'
export DB_SSLMODE=disable
go run ./cmd/manage migrate up
go run ./cmd/server
```

先自行准备专用 PostgreSQL 数据库与账号，迁移命令不创建数据库或账号。不要使用上述占位值部署真实环境。Go Modules 与工具链下载需要网络；网络受限时按团队规范设置 GOPROXY，本工程不修改用户的全局 Go 配置。

配置缺失、参数非法、数据库连接失败或端口占用时，进程退出并返回非零状态码。错误只输出配置项名称或故障类别，不回显配置值、DSN 或 SQL。数据库池大小、连接寿命、HTTP 超时均由配置集中管理。

## HTTP 与运行行为

| 请求 | 当前行为 |
| --- | --- |
| `GET /health/live` | 200，`{"status":"up"}`，仅判断进程可响应 |
| `GET /health/ready` | 检查数据库；可用为 200，不可用或超时为 503，仅返回 up/down |
| 其他未注册请求 | 403，默认拒绝；没有实现业务认证、授权或共同 API 错误格式 |

以上是 Go 的运维端点，不替代 Java 的 `/actuator/health`，也不确定共同业务 API。未配置 CORS 时拒绝跨域 Origin；配置 `HTTP_CORS_ORIGINS` 后只允许精确来源，允许 Authorization、Content-Type 预检，暴露 X-Request-Id，不允许凭据 Cookie。预检通过不代表业务接口已开放。

`HTTP_TRUSTED_PROXIES` 默认为空，明确禁用 Gin 默认的全代理信任；有反向代理时填写实际 IP/CIDR。每次请求生成新的服务端请求标识；访问日志仅记录路由模板、方法、状态、耗时和请求标识，不记录原始路径、查询参数、请求体、请求头或客户端传入的标识。异常恢复不回显 panic 内容。

Ctrl+C / SIGTERM 触发优雅停机：停止接收新请求，在 `HTTP_SHUTDOWN_TIMEOUT` 内等待已有请求完成，再关闭数据库。超时则强制关闭连接并以非零状态退出。没有后台定时任务或自动迁移。

## 迁移与管理命令

```sh
go run ./cmd/manage migrate status
go run ./cmd/manage migrate up
go run ./cmd/server --version
go run ./cmd/manage --version
```

迁移资源编译进 manage 二进制，不依赖启动目录；新增 SQL 后重新构建。使用 `00001_description.sql` 等 goose 文件名与 SQL 注解，不修改已执行脚本。迁移采用数据库会话级 advisory lock，阻止并发迁移互相干扰。错误中不输出 SQL 内容。

当前没有业务 SQL：`up` / `status` 在连接校验后提示无迁移，不创建业务表或版本表。存在 SQL 后，goose 会维护自己的版本表；首次检查状态也可能初始化该元数据。管理命令只提供 `up` 与 `status`，未提供自动回滚或管理员初始化功能。

初始管理员创建依赖尚未定义的账号表和密码规则，留待认证业务实现，当前不会创建固定账号或密码。Java Flyway 与 Go goose 独立管理迁移，脚本复用边界仍待确定。

## 检查与构建

Windows PowerShell：

```powershell
./scripts/check.ps1
./scripts/check.ps1 -Integration -Security
./bin/server.exe --version
```

Linux/macOS：

```sh
sh scripts/check.sh
sh scripts/check.sh --integration --security
./bin/server --version
```

默认执行格式检查、模块校验、go vet、单元测试并构建 server/manage 到 `bin/`。集成检查必须通过 `TEST_DATABASE_DSN` 提供独立测试库的 PostgreSQL URL；密码中的特殊字符需 URL 编码。例：`postgres://test_user:replace-with-test-password@127.0.0.1:5432/template_go_test?sslmode=disable`。

集成测试为每个测试创建随机 schema，连接设置专属 search_path，结束时只删除该 schema；需要测试账号具有创建 schema 的权限。测试 SQL 仅位于测试代码中，不会成为应用迁移或业务实体。覆盖迁移重复执行、空迁移、失败迁移回滚、事务提交与回滚，以及真实数据库就绪探测。缺少测试库配置会失败，不会静默跳过。

漏洞检查需要访问 Go 漏洞数据库；本地通过 `-Security` / `--security` 启用。CI 始终执行集成和漏洞检查，工作流见 [Go CI](../../.github/workflows/go.yml)。本地运行并不等于 GitHub Actions 已在远程成功执行。

## 当前边界

多租户已提前[规划](../../docs/业务规划_2026-10-21.md)，当前两种用户实体均不含 tenant_id，没有租户上下文或查询隔离；共享库共享表、租户内账号和平台独立身份均为建议基线，尚未实施。

- 已提供工程与运维骨架及管理用户、会员持久化实体；不包含业务路由、DTO、Repository、业务表或业务迁移 SQL。
- 不透明 Token、Argon2id、PostgreSQL 会话、RBAC、登录限流与审计仍是已确定但尚未实现的方案。
- 共同 API 格式、前端 Token 存储、有效期、刷新、多端登录和管理员并发保护仍待落实。
- 生产部署拓扑、TLS 终止、资源限制和完整可观测平台仍待设计，当前没有生产部署模板。
