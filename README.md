# MssqlIE - SQL Server 数据导入导出工具

## 项目介绍

MssqlIE（MSSQL Import/Export）是一个功能强大的命令行工具，用于快速、高效地在 SQL Server 数据库和 CSV 文件之间进行数据导入和导出操作。

## 版本信息

当前版本：**v1.0.0**

> 版本号由构建工具通过 `-ldflags "-X main.version=..."` 注入，可在 `make build` 或 CI 发布时自动跟随 Git tag。

## 功能特性

### 📤 数据导出
- **表导出**：将整个表数据导出为 CSV 文件
- **SQL 查询导出**：执行自定义 SQL 查询并将结果导出为 CSV 文件
- **灵活配置**：支持自定义分隔符、包含/排除列标题
- **数据类型支持**：完整支持 SQL Server 各种数据类型，包括二进制数据
- **字符集转换**：支持 UTF-8、GBK、ISO-8859-1 等多种字符集
- **二进制格式**：支持二进制数据以十六进制（hex）、Base64 或原始格式导出
- **查询优化**：默认添加 WITH (NOLOCK) 提示以避免锁定
- **批量处理**：高效处理大量数据

### 📥 数据导入
- **CSV 导入**：将 CSV 文件数据导入到指定表
- **批量插入**：支持自定义批量大小，优化导入性能
- **自动匹配**：自动匹配 CSV 列和数据库表列
- **错误处理**：支持跳过错误行继续导入
- **字符集转换**：支持多种字符集的 CSV 文件
- **二进制格式**：支持多种二进制数据格式的导入
- **数据类型转换**：智能处理不同数据类型的转换

### 🔧 其他功能
- **数据库连接测试**：快速验证数据库连接配置
- **安全转义**：自动处理 SQL 标识符的安全转义
- **环境变量支持**：支持通过环境变量配置连接参数
- **友好提示**：详细的错误信息和操作提示

## 安装方法

### 前提条件
- Go 1.19 或更高版本
- SQL Server 2008 或更高版本

### 编译安装

```bash
# 克隆仓库
git clone https://github.com/LandcLi/MssqlIE.git
cd MssqlIE

# 编译
go build -o mssql-ie .

# 运行
./mssql-ie --help
```

### 直接安装（Go 1.19+）

```bash
go install github.com/LandcLi/MssqlIE@latest
```

### 直接使用

```bash
# 下载预编译二进制文件（如果有提供）
# 将可执行文件添加到系统 PATH 环境变量中

# 验证安装
mssql-ie --version
```

## 使用说明

### 基本语法

```bash
mssql-ie [全局参数] <命令> [命令参数]
```

### 全局参数

| 参数 | 别名 | 默认值 | 说明 | 环境变量 |
|------|------|--------|------|----------|
| --server | -S | localhost | SQL Server 地址 | MSSQL_SERVER, DB_SERVER |
| --port | -P | 1433 | SQL Server 端口 | MSSQL_PORT, DB_PORT |
| --user | -U | sa | 数据库用户名 | MSSQL_USER, DB_USER |
| --password | -W | 无 | 数据库密码 | MSSQL_PASSWORD, DB_PASSWORD |
| --db | -D | 无 | 数据库名 | MSSQL_DBNAME, DB_NAME |
| --encrypt | -E | off | 是否启用加密连接 | MSSQL_ENCRYPT |
| --charset | -C | utf8 | 字符集 | MSSQL_CHARSET |
| --timeout | -T | 30 | 连接超时时间(秒) | MSSQL_TIMEOUT |

### 命令

#### 1. 导出数据 (export)

```bash
mssql-ie [全局参数] export [命令参数]
```

**命令参数：**

| 参数 | 别名 | 默认值 | 说明 |
|------|------|--------|------|
| --csv | -o | 无 | CSV 输出文件路径（必填） |
| --table | -t | 无 | 要导出的表名（与 --sql 二选一） |
| --sql | -s | 无 | 自定义 SQL 查询（与 --table 二选一） |
| --header | - | true | 包含列标题 |
| --delimiter | - | , | CSV 分隔符 |
| --limit | -l | 0 | 限制导出记录数（0 表示无限制） |
| --binary-format | -bf | raw | 二进制数格式 {hex, base64, raw} |
| --file-charset | -fc | utf8 | 文件的字符集 {utf8, gbk, iso-8859-1} |
| --force | - | false | 输出文件已存在时直接覆盖（默认报错退出） |

#### 2. 导入数据 (import)

```bash
mssql-ie [全局参数] import [命令参数]
```

**命令参数：**

| 参数 | 别名 | 默认值 | 说明 |
|------|------|--------|------|
| --csv | -i | 无 | CSV 输入文件路径（必填） |
| --table | -t | 无 | 目标表名（必填） |
| --batch | -b | 1000 | 批量插入大小 |
| --header | - | true | CSV 文件包含列标题 |
| --delimiter | - | , | CSV 分隔符 |
| --truncate | - | false | 导入前清空表 |
| --skip-errors | - | false | 跳过错误行继续导入 |
| --binary-format | -bf | raw | 二进制数格式 {hex, base64, raw} |
| --file-charset | -fc | utf8 | 文件的字符集 {utf8, gbk, iso-8859-1} |
| --fill-defaults | - | false | 非空约束列为空时填充默认值(0/false/'')，默认报错退出 |

#### 3. 测试连接 (test)

```bash
mssql-ie [全局参数] test
```

## 使用示例

### 连接测试

```bash
# 使用命令行参数
mssql-ie -S localhost -P 1433 -U sa -W your_password -D your_database test

# 使用环境变量
export MSSQL_SERVER=localhost
export MSSQL_PORT=1433
export MSSQL_USER=sa
export MSSQL_PASSWORD=your_password
export MSSQL_DBNAME=your_database
mssql-ie test
```

### 导出表数据

```bash
# 导出整个表到 CSV
mssql-ie -S localhost -P 1433 -U sa -W your_password -D your_database export -t your_table -o output.csv

# 导出表的前 1000 行
mssql-ie -S localhost -P 1433 -U sa -W your_password -D your_database export -t your_table -o output.csv -l 1000

# 使用自定义分隔符（制表符）
mssql-ie -S localhost -P 1433 -U sa -W your_password -D your_database export -t your_table -o output.csv --delimiter "\t"

# 使用 GBK 字符集
mssql-ie -S localhost -P 1433 -U sa -W your_password -D your_database export -t your_table -o output.csv -fc gbk

# 二进制数据以十六进制格式导出
mssql-ie -S localhost -P 1433 -U sa -W your_password -D your_database export -t your_table -o output.csv -bf hex
```

### 导出 SQL 查询结果

```bash
# 执行自定义 SQL 查询并导出结果
mssql-ie -S localhost -P 1433 -U sa -W your_password -D your_database export -s "SELECT id, name FROM your_table WHERE status = 1" -o output.csv

# 使用带参数的复杂查询
mssql-ie -S localhost -P 1433 -U sa -W your_password -D your_database export -s "SELECT * FROM orders WHERE order_date BETWEEN '2024-01-01' AND '2024-12-31' ORDER BY order_date" -o orders_2024.csv
```

### 导入数据

```bash
# 从 CSV 导入数据到表
mssql-ie -S localhost -P 1433 -U sa -W your_password -D your_database import -t your_table -i input.csv

# 导入前清空表
mssql-ie -S localhost -P 1433 -U sa -W your_password -D your_database import -t your_table -i input.csv --truncate

# 使用更大的批量大小
mssql-ie -S localhost -P 1433 -U sa -W your_password -D your_database import -t your_table -i input.csv -b 2000

# 跳过错误行
mssql-ie -S localhost -P 1433 -U sa -W your_password -D your_database import -t your_table -i input.csv --skip-errors

# 导入 GBK 编码的 CSV 文件
mssql-ie -S localhost -P 1433 -U sa -W your_password -D your_database import -t your_table -i input.csv -fc gbk

# 从包含二进制数据（Base64 格式）的 CSV 导入
mssql-ie -S localhost -P 1433 -U sa -W your_password -D your_database import -t your_table -i input.csv -bf base64
```

## 支持的数据类型

### 导出时支持的类型
- 字符类型：char, varchar, text, nchar, nvarchar, ntext
- 数值类型：tinyint, smallint, int, bigint, decimal, numeric, float, real, money, smallmoney
- 日期时间：date, time, datetime, datetime2, datetimeoffset, smalldatetime
- 二进制类型：binary, varbinary, image
- 特殊类型：bit, uniqueidentifier, xml, geometry, geography, hierarchyid

### 导入时支持的类型
- 完整支持所有 SQL Server 数据类型
- 自动处理 NULL 值和默认值
- 智能类型转换

## 安全注意事项

1. **密码安全**：
   - 优先使用环境变量 `MSSQL_PASSWORD`（推荐，CI 中必须使用环境变量）
   - 未提供密码且终端为交互模式时，工具会隐藏回显地提示输入密码
   - 避免使用 `-W` 明文传参：密码会出现在 shell 历史、进程列表（`ps`）和 CI 日志中
2. **自定义 SQL 警示**：`--sql` 参数由用户自写并**原样执行**，工具仅对表名/列名做安全转义，请勿传入不可信来源的 SQL
3. **导入空字段**：非空约束列遇空字段会**报错退出**（防止静默写入错误默认值）；如确实需要填充默认值，请显式使用 `--fill-defaults`
4. **二进制原始模式**：`-bf raw` 将原始字节写入 CSV，非 UTF-8 数据可能损坏，建议使用 `hex` 或 `base64`
5. **数据安全**：在生产环境中使用时，确保适当的权限控制
6. **网络安全**：在不安全的网络环境中，建议启用加密连接（--encrypt 选项）

## 性能优化

1. **批量大小**：导入时根据数据库性能调整批量大小（建议 500-2000）
2. **索引管理**：大规模导入前可考虑暂时禁用索引，导入完成后重新创建
3. **事务控制**：工具已实现高效的事务管理，无需额外配置
4. **查询优化**：导出时自定义 SQL 查询可包含 WHERE 条件减少数据量

## 限制

1. 当前版本不支持 XML 数据的特殊处理
2. 复杂的层次结构数据（如 hierarchyid）在 CSV 中可能不易阅读
3. 超大文件导入时建议分批处理

## 许可证

本项目采用 [MIT 许可证](LICENSE) 开源。

## 贡献

欢迎提交 Issue 和 Pull Request 来帮助改进这个项目！

## 联系方式

如有问题或建议，请通过以下方式联系：

- 项目地址：[GitHub Repository](https://github.com/LandcLi/MssqlIE)
- 邮箱：206131925@qq.com

---

**MssqlIE v1.0.0** - 让 SQL Server 数据导入导出变得简单高效！
