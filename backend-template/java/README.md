# Java 后端模板

一个可独立启动的 Maven 多模块骨架，服务于后续管理后台项目。当前仅有启动配置、保守的安全边界及按功能预留的包，没有登录、用户、角色、权限、审计或 CRUD 业务代码。

每个派生项目只选择 Java 或 Go 其中一套后端；本目录仅构建 Java 实现，不要求 Go 后端或前端同时启动。

## 环境与版本

| 项目 | 基线 | 原因与管理方式 |
| --- | --- | --- |
| JDK | 21 | 使用已有 LTS 开发环境，源码与字节码目标由 `java.version` 固定为 21 |
| Spring Boot | 4.1.1 | 使用当前正式版本，统一管理 Spring、Web、日志、校验、安全和测试依赖 |
| Spring Cloud | 2025.1.3 BOM | 与 Spring Boot 4.1.x 按官方兼容关系配套；当前只导入 BOM，未启用具体 Cloud 组件 |
| MyBatis-Plus | 3.5.17 | 父 POM 导入 MyBatis-Plus BOM，使用 Boot 4 专用 starter 与同版本分页解析模块 |
| Flyway / PostgreSQL JDBC | Boot 管理版本 | 已引入迁移 starter、PostgreSQL 支持模块与 JDBC 驱动；数据库服务端版本待定 |
| Maven | 3.10.0 | 通过 Maven Wrapper 固定，避免依赖机器上的旧 Maven 版本 |
| 基础包名 | `com.example.template` | 模板占位命名；派生项目按自己的组织域名统一修改 |

版本依据见 [Spring Cloud 兼容表](https://spring.io/projects/spring-cloud#overview)、[Spring Boot 系统要求](https://docs.spring.io/spring-boot/4.1/system-requirements.html)和 [Maven 下载说明](https://maven.apache.org/download.cgi)。后续升级须同步校验 BOM、构建与安全配置。

请设置 `JAVA_HOME` 为 JDK 21，并确保可以访问 Maven Central。Wrapper 首次执行会下载 Maven，之后复用本地缓存，不修改系统 Maven 安装。

## 模块结构

```text
java/
├── pom.xml                        # 聚合、Java 基线、Boot 父 POM 与 Cloud BOM
├── .mvn/wrapper/                   # Maven Wrapper 配置
├── mvnw / mvnw.cmd                 # Linux/macOS 与 Windows 构建入口
├── template-common/               # 不依赖 Spring 的公共类型边界
├── template-framework/            # Web、校验、安全、持久层、日志及运维基础设施
└── template-app/                  # 唯一可运行应用，预留第一期功能包
```

模块依赖关系：`template-framework → template-common`；`template-app → template-framework、template-common`。仅 `template-app` 使用 Spring Boot 打包插件生成可执行 jar；其他模块是普通库，不能反向依赖启动模块。

该结构是一个应用的构建模块，不代表已经拆成多个服务。不预建网关、注册中心、配置中心或远程调用模块。后续只有在真实业务与复用边界明确后再增加 Maven 模块。

### 包职责

| 模块 | 包（均位于 `com.example.template` 下） | 当前内容与扩展方向 |
| --- | --- | --- |
| common | `common.error`、`common.page` | 仅 `package-info.java`，待契约确定后补充跨模块公共类型，不预设响应包装 |
| framework | `framework.security` | 默认拒绝策略配置，尚未实现 Token 认证 |
| framework | `framework.persistence` | PostgreSQL 分页拦截器配置；不包含业务 Mapper 或数据库实体 |
| framework | `framework.web`、`framework.validation`、`framework.logging` | 仅包说明，预留 HTTP、校验与技术日志基础设施 |
| app | 根包 `TemplateApplication` | 唯一启动类，扫描当前根包及各模块的子包 |
| app | `modules.auth`、`modules.user`、`modules.role`、`modules.permission` | 仅包说明，预留认证、个人中心、用户、角色及权限功能 |
| app | `modules.audit`、`modules.example` | 仅包说明，预留业务审计和可替换的 CRUD 示例 |

开发具体功能时，在对应功能包下按实际需要增加 `controller`、`service`、`repository`、`dto` 等子包；本次不创建无逻辑的 Controller、Service 或 Repository 类。技术日志与业务审计分别维护，权限不能只依赖前端。

## 已引入依赖

| 依赖 | 所在模块 | 用途 |
| --- | --- | --- |
| `spring-boot-starter-webmvc` | framework | Servlet MVC、嵌入式 Tomcat、JSON HTTP 基础能力 |
| `spring-boot-starter-validation` | framework | Jakarta Bean Validation 基础设施，未添加业务校验规则 |
| `spring-boot-starter-security` | framework | 默认拒绝的安全骨架，不代表登录或 Token 方案已实现 |
| `spring-boot-starter-actuator` | framework | 只暴露基础健康检查 |
| `mybatis-plus-spring-boot4-starter` | framework | MyBatis-Plus 自动配置，传递引入 JDBC starter 与 HikariCP |
| `mybatis-plus-jsqlparser` | framework | PostgreSQL 分页拦截器所需解析支持 |
| `spring-boot-starter-flyway` | framework | Boot 4 Flyway 自动配置与迁移支持 |
| `flyway-database-postgresql`、`postgresql` | framework，runtime | PostgreSQL 迁移支持模块及 JDBC 驱动，运行时传递到 app |
| `spring-boot-configuration-processor` | framework，optional | 为后续类型化配置生成元数据，不向使用方传播处理器依赖 |
| Boot 默认日志依赖 | 由 starter 传递引入 | SLF4J 与 Logback，当前使用默认控制台输出 |
| `spring-boot-starter-test`、`spring-security-test` | app，test | 应用集成测试与后续安全边界测试支持 |

MyBatis-Plus 版本由父 POM 的 `mybatis-plus.version` 与 BOM 管理；其他基础依赖由 Boot 父 POM 管理，Cloud 相关版本由 `spring-cloud-dependencies` BOM 管理，不逐个指定 Spring 子组件版本。

Java 已确定 PostgreSQL + MyBatis-Plus + Flyway，已提供依赖、数据源配置和分页拦截器；尚无业务表、迁移 SQL、Mapper 或持久化业务。PostgreSQL 服务端版本、Go 配套、Token 类型及存储仍待定。没有引入 JPA、Liquibase、Redis、JWT 或 OAuth Resource Server。

正常启动必须提供 `DB_URL`、`DB_USERNAME`、`DB_PASSWORD`，数据库应事先创建且可连接，账号须具备迁移所需权限。没有默认用户名或密码；不要将真实凭据写入仓库。连接或迁移失败会阻止启动。

Flyway 从 app 的 `src/main/resources/db/migration/` 加载版本化 SQL，如 `V1__initial_schema.sql`。当前只有目录说明，不创建业务表或默认管理员；Flyway 可能创建自身历史表。已执行脚本不得修改，使用新脚本演进；`schema.sql/data.sql` 初始化、自动 baseline 和 clean 均关闭，迁移校验开启。

业务 Mapper 随功能增加并标注 `@Mapper`，复杂查询可以使用 MyBatis XML；请求/响应 DTO 与实体分离，事务放在业务 Service 层。分页插件不替代分页大小、排序白名单与权限校验；逻辑删除、多租户及数据权限插件暂未开启。

## 构建与启动

在 `backend-template/java/` 下执行。

Windows PowerShell：

```powershell
.\mvnw.cmd clean verify
$env:DB_URL = 'jdbc:postgresql://127.0.0.1:5432/template'
$env:DB_USERNAME = '<数据库用户名>'
$env:DB_PASSWORD = '<数据库密码>'
java -jar .\template-app\target\template-app-0.1.0-SNAPSHOT.jar
```

Linux/macOS：

```sh
sh ./mvnw clean verify
export DB_URL='jdbc:postgresql://127.0.0.1:5432/template'
export DB_USERNAME='<数据库用户名>'
export DB_PASSWORD='<数据库密码>'
java -jar template-app/target/template-app-0.1.0-SNAPSHOT.jar
```

以上凭据为占位值，启动前替换为本地配置。默认监听 `127.0.0.1:8080`，可通过环境变量 `SERVER_ADDRESS`、`SERVER_PORT` 调整；容器或远程访问时显式配置监听地址。未提供生产环境部署配置。

健康检查：

```text
GET http://127.0.0.1:8080/actuator/health
```

正常响应为 HTTP 200 和仅包含基本状态的 JSON。没有首页或业务接口，其余路径被安全骨架拒绝，当前返回 403。

仅打包、不执行测试时可使用 `./mvnw.cmd -DskipTests package`（Windows）或 `sh ./mvnw -DskipTests package`（Linux/macOS）；交付验证应使用 `verify`。

## 当前安全与配置边界

- 仅放行 `GET /actuator/health`，不暴露健康详情、环境或 Bean 等管理信息，其余请求均拒绝。
- 关闭默认表单登录、HTTP Basic、默认退出入口和请求缓存，使用无状态 Session 策略；排除 Boot 默认内存用户自动配置，不生成示例账号或密码。
- 当前不接收任何业务写操作，也不使用浏览器自动附带的认证凭据；CSRF 过滤暂时关闭，避免空骨架生成会话。后续按最终凭据传递及浏览器交互方式重新评估，不能将该配置当作完整 Token 安全方案。
- 当前 403 是未接入认证时的封闭策略，不是已经实现 API 规划中的 401、错误码或 Problem Details 契约。Token、错误映射、权限生效及撤销规则须在实际功能开发时补充。
- 默认使用开发阶段易读的控制台日志；结构化 JSON、请求关联、脱敏、集中采集及业务审计尚未实现。
- 无真实密码、Token 或数据源密钥。环境配置后续按实际需求增加，不将生产凭据写入源码或示例。

## 验证范围与后续工作

`ApplicationSmokeTest` 使用随机端口启动真实 HTTP 服务，验证健康检查、未授权路径及写请求拒绝、伪造 Bearer 字符串不能获得访问、没有默认内存用户，以及相关响应不创建 Cookie 会话。

该测试在测试类内排除数据源与 Flyway 自动配置，因此 `clean verify` 的现有 HTTP/安全冒烟测试不需要外部数据库；正常应用启动没有这些排除项。测试通过不代表真实 PostgreSQL 连接、Flyway 迁移、Mapper CRUD 或事务回滚已验证。

2026-10-21 补齐持久层依赖后，在 Windows、JDK 21.0.10、Wrapper Maven 3.10.0 下通过 `clean verify`，5 项 HTTP/安全冒烟测试通过，打包成功。本机 Docker 服务不可用，未执行真实 PostgreSQL 迁移及持久化集成验证；这些验证随业务迁移和 Mapper 实现补充。

2026-10-20 已在 Windows、JDK 21.0.10、Wrapper Maven 3.10.0 下通过 `clean verify`，5 项测试通过；可执行 jar 独立启动及健康检查通过。POM XML、本地文档链接与构建产物忽略规则亦已检查，未验证其他操作系统或生产部署环境。

这些检查仅覆盖基础应用和当前安全骨架，不代表已经完成 Token 登录、RBAC、本人数据边界、管理员保护或业务验收。后续应先确定共同契约和数据方案，再在功能包内开发并增加对应测试。

完整规划见 [技术栈规划](../../docs/技术栈规划_2026-10-19.md)、[业务规划](../../docs/业务规划_2026-10-21.md)及 [API 规划](../../docs/API规划_2026-10-20.md)。派生项目使用时记录模板来源提交与所选后端。
