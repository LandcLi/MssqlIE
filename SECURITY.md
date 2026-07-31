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

- 使用环境变量 `MSSQL_PASSWORD` 传递数据库密码，避免在命令行中明文输入（明文密码会出现在 shell 历史、`ps` 进程列表和 CI 日志中）
- 未设置密码且终端为交互模式时，工具会隐藏回显地提示输入；非交互环境（如 CI）必须使用环境变量
- 生产环境建议使用 SQL Server 的加密连接 (`--encrypt required` 或 `--encrypt strict`)
- 连接字符串不会记录在日志中
- 所有 SQL 查询使用参数化查询，防止注入攻击（表结构查询、批量插入均参数化）
- SQL 标识符使用安全转义函数处理
- `--sql` 自定义查询原样执行，仅可用于可信来源，工具只对表名/列名做转义
