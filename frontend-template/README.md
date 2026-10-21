# frontend-template

可复用的 Vue 管理端前端骨架，基于 create-vue 3.24.0 初始化。当前仅提供工程入口，不包含欢迎页、登录页、布局、菜单或业务页面；启动后显示空白内容，这是当前预期行为。

## 环境与安装

- Node.js：24.x，项目通过 `.node-version` 固定验证版本；`package.json` 限制支持的版本范围。
- pnpm：11.25.0，通过 `packageManager` 和 `engines` 固定；只提交 `pnpm-lock.yaml`。
- 各依赖的直接版本在 `package.json` 中固定，完整解析结果以锁文件为准。

安装 Node.js 后，在需要时安装指定版本的 pnpm：

```sh
npm install --global pnpm@11.25.0
```

在 `frontend-template/` 下执行：

```sh
pnpm install --frozen-lockfile
pnpm dev
```

开发地址为 `http://127.0.0.1:5173`，端口被占用时明确报错。无需后端即可启动空白骨架。

## 已接入内容

- Vue 3、TypeScript 严格检查、Vite 和 `@` 路径别名。
- Vue Router：保留 `/` 空白挂载路由；未添加实际页面、权限守卫或 404 页面。
- Pinia：注册全局状态管理入口，未创建示例或业务 store。
- Element Plus：接入中文配置与基础样式，组件采用显式导入；后续使用可视组件时同步引入对应样式，目前没有全量注册组件。
- CSS Variables、全局 SCSS 及 scoped SCSS。
- Axios：`src/api/http.ts` 提供统一实例及 10 秒超时；没有业务接口、认证注入、错误协议适配或自动刷新逻辑。
- ESLint、Prettier、vue-tsc、Vitest、Vue Test Utils 和 Playwright 配置。
- openapi-typescript 已安装；共同契约尚未建立，未生成虚构的 API 类型或请求 SDK。

## 配置

默认请求前缀为 `/api`，只作为 Axios 基础连接配置，不代表 API 路径已定稿或后端已实现。可参考 `.env.example` 在本地创建 `.env.local`：

```dotenv
VITE_API_BASE_URL=/api
```

该实例尚未被业务页面调用。开发环境未预设代理目标；接入选定后端时再配置 Vite 代理或跨域访问方式。每个派生项目只对接该项目选择的 Java 或 Go 后端。

`VITE_*` 会进入浏览器构建产物，只能放公开配置，不能放 Token、密码或密钥。Token 类型、存储、传递、刷新及撤销机制仍待确定；未使用 Cookie 或状态持久化插件保存认证凭据。

## 常用命令

| 命令                   | 用途                                                                       |
| ---------------------- | -------------------------------------------------------------------------- |
| `pnpm dev`             | 启动本地开发服务器                                                         |
| `pnpm type-check`      | 检查应用、工具配置、单元测试与端到端测试的类型                             |
| `pnpm build`           | 执行类型检查并生成 `dist/`                                                 |
| `pnpm preview`         | 在 `http://127.0.0.1:4173` 预览已有构建                                    |
| `pnpm lint`            | 只检查代码，不自动修改                                                     |
| `pnpm lint:fix`        | 修复可自动处理的代码规则问题                                               |
| `pnpm format:check`    | 检查格式                                                                   |
| `pnpm format`          | 格式化工程文件                                                             |
| `pnpm test:unit`       | 单次运行 Vitest；当前没有单元测试，显式允许空用例集                        |
| `pnpm test:unit:watch` | Vitest 监听模式，供后续添加用例时使用                                      |
| `pnpm test:e2e`        | 构建并启动隔离的预览服务，使用 Chromium 验证空白应用挂载、刷新及运行时错误 |

第一次执行端到端测试前安装浏览器：

```sh
pnpm exec playwright install chromium
pnpm test:e2e
```

端到端测试使用无头模式，不要求已有开发服务；它会占用 4173 端口，测试结束后停止自建预览服务。目前没有登录、权限或 CRUD 用例，不能将骨架冒烟验证视为第一期功能验收。

共同契约确定后，可按实际路径生成类型：

```sh
pnpm exec openapi-typescript <契约文件路径> -o src/api/generated/schema.d.ts
```

以上为待契约建立后的使用方式，目前没有可用的契约生成任务。

## 目录与扩展

```text
src/
├── api/http.ts       # Axios 基础实例
├── router/index.ts   # 静态路由入口
├── styles/main.scss # 全局样式与主题变量
├── App.vue          # Element Plus 配置与空白路由出口
└── main.ts          # Vue、Pinia、Router 注册
e2e/app.spec.ts      # 生产构建的应用挂载冒烟用例
```

构建产物位于 `dist/`。生产环境需提供 SPA history 回退，并将 API 请求独立路由到所选后端；`pnpm preview` 仅用于本地预览，尚未提供生产部署配置。

后续按 [技术栈规划](../docs/技术栈规划_2026-10-19.md)、[第一期功能规划](../docs/第一期功能规划_2026-10-19.md)与 [API 规划](../docs/API规划_2026-10-20.md)逐步添加能力。派生项目使用时记录来源提交及所选后端。
