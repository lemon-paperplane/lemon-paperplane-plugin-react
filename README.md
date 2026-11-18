# Lemon Paperplane React 插件

依赖 nodejs 插件，在 Node 模块上识别 React：

- facet `react`（Next.js 工程另有 `nextjs`），关键信息如「React 19.1 · Next.js 15.1」「React 17.0 · Create React App 4.0」「React 19.1 · React Router 7.6」。
- Next / CRA / React Router / Vite 的开发脚本作为默认脚本偏好。
- react-scripts < 5、Next < 12 在 Node 17+ 上预先提示 OpenSSL 旧版兼容。
- 诊断：CRA 端口被占用（非交互时不会换端口）、Next.js 要求的 Node 版本、SWC 二进制缺失。

```sh
./build.sh
go test ./...
```
