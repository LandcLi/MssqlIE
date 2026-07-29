# Contributing to mssql_ie

感谢您考虑为 `mssql_ie` 贡献代码！

## 开发流程

1. Fork 本仓库
2. 创建功能分支: `git checkout -b feature/my-feature`
3. 提交更改: `git commit -am 'feat: add new feature'`
4. 推送分支: `git push origin feature/my-feature`
5. 创建 Pull Request

## 代码规范

- 运行 `make fmt` 格式化代码
- 运行 `make lint` 检查代码风格
- 运行 `make test` 确保测试通过
- 运行 `make vet` 静态分析

## 提交信息规范

我们使用 [Conventional Commits](https://www.conventionalcommits.org/):

- `feat:` 新功能
- `fix:` 修复
- `docs:` 文档
- `test:` 测试
- `refactor:` 重构
- `chore:` 构建/工具

## 测试

- 新功能应包含单元测试
- 运行 `make cover` 查看覆盖率
