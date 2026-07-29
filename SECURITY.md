# Security Policy

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| latest  | :white_check_mark: |

## Reporting a Vulnerability

请将安全漏洞报告发送至项目维护者。请勿在公开的 Issue 中披露安全漏洞。

我们承诺：
1. 收到报告后 48 小时内确认
2. 评估漏洞影响并提供修复时间表
3. 修复完成后发布安全更新

## 安全最佳实践

- 使用环境变量 `MSSQL_PASSWORD` 传递数据库密码，避免在命令行中明文输入
- 生产环境建议使用 SQL Server 的加密连接 (`--encrypt required` 或 `--encrypt strict`)
- 连接字符串不会记录在日志中
- 所有 SQL 查询使用参数化查询，防止注入攻击
- SQL 标识符使用安全转义函数处理
