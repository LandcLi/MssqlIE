package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	"github.com/LandcLi/MssqlIE/config"
	"github.com/LandcLi/MssqlIE/conn"
	"github.com/LandcLi/MssqlIE/exporter"
	"github.com/LandcLi/MssqlIE/importer"
	"github.com/urfave/cli/v2"
	"golang.org/x/term"
)

var (
	// version 等默认值仅用于本地 go run 开发；正式发布由 Makefile/CI 通过 -ldflags 注入
	version = "1.0.0"
	commit  = "unknown"
	date    = "unknown"
)

func main() {
	app := &cli.App{
		Name:     "mssql-ie",
		Version:  fmt.Sprintf("%s (commit: %s, date: %s)", version, commit, date),
		Usage:    "SQL Server 数据导入导出工具",
		Suggest:  true,
		HideHelp: false,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "server",
				Aliases: []string{"S"},
				Value:   "localhost",
				Usage:   "SQL Server地址",
				EnvVars: []string{"MSSQL_SERVER", "DB_SERVER"},
			},
			&cli.IntFlag{
				Name:    "port",
				Aliases: []string{"P"},
				Value:   1433,
				Usage:   "SQL Server端口",
				EnvVars: []string{"MSSQL_PORT", "DB_PORT"},
			},
			&cli.StringFlag{
				Name:    "user",
				Aliases: []string{"U"},
				Value:   "sa",
				Usage:   "数据库用户名",
				EnvVars: []string{"MSSQL_USER", "DB_USER"},
			},
			&cli.StringFlag{
				Name:    "password",
				Aliases: []string{"W"},
				Usage:   "数据库密码（优先使用环境变量 MSSQL_PASSWORD，避免明文出现在进程列表/日志）",
				EnvVars: []string{"MSSQL_PASSWORD", "DB_PASSWORD"},
			},
			&cli.StringFlag{
				Name:     "db",
				Aliases:  []string{"D"},
				Usage:    "数据库名",
				Required: true,
				EnvVars:  []string{"MSSQL_DBNAME", "DB_NAME"},
			},
			&cli.StringFlag{
				Name:    "encrypt",
				Aliases: []string{"E"},
				Value:   "off",
				Usage:   "是否启用加密连接",
				EnvVars: []string{"MSSQL_ENCRYPT"},
			},
			&cli.StringFlag{
				Name:    "charset",
				Aliases: []string{"C"},
				Usage:   "字符集 (例如: utf8, gbk)",
				Value:   "utf8",
				EnvVars: []string{"MSSQL_CHARSET"},
			},
			&cli.IntFlag{
				Name:    "timeout",
				Aliases: []string{"T"},
				Value:   30,
				Usage:   "连接超时时间(秒)",
				EnvVars: []string{"MSSQL_TIMEOUT"},
			},
		},
		Commands: []*cli.Command{
			{
				Name:    "export",
				Aliases: []string{"e"},
				Usage:   "导出数据到CSV文件",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "csv",
						Aliases:  []string{"o"},
						Usage:    "CSV输出文件路径",
						Required: true,
					},
					&cli.StringFlag{
						Name:    "table",
						Aliases: []string{"t"},
						Usage:   "要导出的表名 (与 --sql 二选一)",
					},
					&cli.StringFlag{
						Name:    "sql",
						Aliases: []string{"s"},
						Usage:   "自定义SQL查询 (与 --table 二选一)",
					},
					&cli.BoolFlag{
						Name:  "header",
						Usage: "包含列标题",
						Value: true,
					},
					&cli.StringFlag{
						Name:  "delimiter",
						Usage: "CSV分隔符",
						Value: ",",
					},
					&cli.IntFlag{
						Name:    "limit",
						Aliases: []string{"l"},
						Usage:   "限制导出记录数 (0表示无限制)",
						Value:   0,
					},
					&cli.StringFlag{
						Name:    "binary-format",
						Aliases: []string{"bf"},
						Usage:   "二进制数格式 {hex, base64, raw}",
						Value:   "raw",
					},
					&cli.StringFlag{
						Name:    "file-charset",
						Aliases: []string{"fc"},
						Usage:   "文件的字符集 {utf8,gbk,latin1}",
						Value:   "utf8",
					},
					&cli.StringFlag{
						Name:  "null-marker",
						Usage: "NULL 值在 CSV 中的标记字符串，为空时用空字段表示 NULL",
						Value: "",
					},
					&cli.BoolFlag{
						Name:  "force",
						Usage: "输出文件已存在时直接覆盖（默认报错退出）",
						Value: false,
					},
				},
				Before: validateExportFlags,
				Action: exportCommand,
			},
			{
				Name:    "import",
				Aliases: []string{"i"},
				Usage:   "从CSV文件导入数据",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "csv",
						Aliases:  []string{"i"},
						Usage:    "CSV输入文件路径",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "table",
						Aliases:  []string{"t"},
						Usage:    "目标表名",
						Required: true,
					},
					&cli.IntFlag{
						Name:    "batch",
						Aliases: []string{"b"},
						Usage:   "批量插入大小",
						Value:   1000,
					},
					&cli.BoolFlag{
						Name:  "header",
						Usage: "CSV文件包含列标题",
						Value: true,
					},
					&cli.StringFlag{
						Name:  "delimiter",
						Usage: "CSV分隔符",
						Value: ",",
					},
					&cli.BoolFlag{
						Name:  "truncate",
						Usage: "导入前清空表",
						Value: false,
					},
					&cli.BoolFlag{
						Name:  "skip-errors",
						Usage: "跳过错误行继续导入",
						Value: false,
					},
					&cli.BoolFlag{
						Name:  "identity-insert",
						Usage: "允许为自增列插入显式值 (SET IDENTITY_INSERT ON)",
						Value: false,
					},
					&cli.StringFlag{
						Name:    "binary-format",
						Aliases: []string{"bf"},
						Usage:   "二进制数格式 {hex, base64, raw}",
						Value:   "raw",
					},
					&cli.StringFlag{
						Name:    "file-charset",
						Aliases: []string{"fc"},
						Usage:   "文件的字符集 {utf8,gbk,latin1}",
						Value:   "utf8",
					},
					&cli.StringFlag{
						Name:  "null-marker",
						Usage: "CSV 中代表 NULL 的字符串，为空时空字段视为 NULL",
						Value: "",
					},
					&cli.BoolFlag{
						Name:  "fill-defaults",
						Usage: "非空约束列为空时填充默认值(0/false/'')，默认报错退出",
						Value: false,
					},
				},
				Before: validateImportFlags,
				Action: importCommand,
			},
			{
				Name:    "test",
				Aliases: []string{"t"},
				Usage:   "测试数据库连接",
				Action:  testConnection,
			},
		},
		Before: func(c *cli.Context) error {
			// 无子命令（仅显示帮助）时不做密码校验
			if !c.Args().Present() {
				return nil
			}
			if c.String("password") == "" {
				// 交互式输入（仅 TTY），避免密码出现在 shell 历史/进程列表
				if term.IsTerminal(int(os.Stdin.Fd())) {
					fmt.Fprint(os.Stderr, "请输入数据库密码: ")
					pw, err := term.ReadPassword(int(os.Stdin.Fd()))
					fmt.Fprintln(os.Stderr)
					if err != nil {
						return cli.Exit(fmt.Sprintf("错误: 读取密码失败: %v", err), 1)
					}
					if len(pw) == 0 {
						return cli.Exit("错误: 密码不能为空", 1)
					}
					return c.Set("password", string(pw))
				}
				return cli.Exit("错误: 密码不能为空，请通过 -password 参数、环境变量 MSSQL_PASSWORD 或交互式输入设置", 1)
			}
			return nil
		},
		Action: func(c *cli.Context) error {
			cli.ShowAppHelp(c)
			return nil
		},
		ExitErrHandler: func(c *cli.Context, err error) {
			if err != nil {
				fmt.Fprintf(os.Stderr, "[ERROR] %v\n", err)
			}
		},
	}

	if err := app.Run(os.Args); err != nil {
		os.Exit(1)
	}
}

// 构建数据库配置
func buildDBConfig(c *cli.Context) config.DBConfig {
	return config.DBConfig{
		Server:   c.String("server"),
		Port:     uint64(c.Int("port")),
		User:     c.String("user"),
		Password: c.String("password"),
		DBName:   c.String("db"),
		Encrypt:  c.String("encrypt"),
		Charset:  c.String("charset"),
		Timeout:  uint64(c.Int("timeout")),
	}
}

// 连接数据库
func connectDB(c *cli.Context) (*sql.DB, error) {
	dbCfg := buildDBConfig(c)
	return conn.Connect(dbCfg)
}

// 导出命令
func exportCommand(c *cli.Context) error {
	db, err := connectDB(c)
	if err != nil {
		return fmt.Errorf("数据库连接失败: %w", err)
	}
	defer db.Close()

	// 测试连接
	if err := db.PingContext(context.Background()); err != nil {
		return fmt.Errorf("数据库连接测试失败: %w", err)
	}

	// 解析分隔符
	delimiter := ','
	if delim := c.String("delimiter"); len(delim) > 0 {
		delimiter = []rune(delim)[0]
	}

	cfg := config.ExportConfig{
		Table:        c.String("table"),
		SQL:          c.String("sql"),
		CSVPath:      c.String("csv"),
		Header:       c.Bool("header"),
		Delimiter:    delimiter,
		Limit:        c.Int("limit"),
		BinaryFormat: c.String("binary-format"),
		FileCharset:  c.String("file-charset"),
		NullMarker:   c.String("null-marker"),
	}

	if cfg.Table != "" {
		if err := exporter.TableToCSV(db, cfg); err != nil {
			return fmt.Errorf("导出表失败: %w", err)
		}
	} else {
		if err := exporter.SQLToCSV(db, cfg); err != nil {
			return fmt.Errorf("导出SQL结果失败: %w", err)
		}
	}

	fmt.Fprintf(os.Stderr, "[OK] 导出成功: 数据已保存到 %s\n", cfg.CSVPath)
	return nil
}

// 导入命令
func importCommand(c *cli.Context) error {
	db, err := connectDB(c)
	if err != nil {
		return fmt.Errorf("数据库连接失败: %w", err)
	}
	defer db.Close()

	// 测试连接
	if err := db.PingContext(context.Background()); err != nil {
		return fmt.Errorf("数据库连接测试失败: %w", err)
	}

	// 解析分隔符
	delimiter := ','
	if delim := c.String("delimiter"); len(delim) > 0 {
		delimiter = []rune(delim)[0]
	}

	cfg := config.ImportConfig{
		Table:          c.String("table"),
		CSVPath:        c.String("csv"),
		Batch:          c.Int("batch"),
		Header:         c.Bool("header"),
		Delimiter:      delimiter,
		Truncate:       c.Bool("truncate"),
		SkipErrors:     c.Bool("skip-errors"),
		IdentityInsert: c.Bool("identity-insert"),
		BinaryFormat:   c.String("binary-format"),
		FileCharset:    c.String("file-charset"),
		NullMarker:     c.String("null-marker"),
		FillDefaults:   c.Bool("fill-defaults"),
	}

	if err := importer.CSVToTable(db, cfg); err != nil {
		return fmt.Errorf("导入失败: %w", err)
	}

	fmt.Fprintf(os.Stderr, "[OK] 导入成功: 数据已导入到表 %s\n", cfg.Table)
	return nil
}

// 测试连接命令
func testConnection(c *cli.Context) error {
	db, err := connectDB(c)
	if err != nil {
		return fmt.Errorf("数据库连接失败: %w", err)
	}
	defer db.Close()

	if err := db.PingContext(context.Background()); err != nil {
		return fmt.Errorf("数据库连接测试失败: %w", err)
	}

	// 获取数据库信息
	var version, dbName string
	err = db.QueryRowContext(context.Background(), "SELECT @@VERSION, DB_NAME()").Scan(&version, &dbName)
	if err != nil {
		return fmt.Errorf("查询数据库信息失败: %w", err)
	}

	fmt.Println("[OK] 数据库连接测试成功!")
	fmt.Printf("   数据库: %s\n", dbName)
	fmt.Printf("   服务器: %s:%d\n", c.String("server"), c.Int("port"))
	fmt.Printf("   版本: %s\n", version)

	return nil
}

// 导出参数验证
func validateExportFlags(c *cli.Context) error {
	table := c.String("table")
	sql := c.String("sql")
	csv := c.String("csv")

	if csv == "" {
		return cli.Exit("错误: 必须指定 --csv 参数", 1)
	}

	if (table == "" && sql == "") || (table != "" && sql != "") {
		return cli.Exit("错误: 必须且只能指定 --table 或 --sql 参数之一", 1)
	}

	// 输出文件已存在时，需显式 --force 才覆盖（避免交互确认在 CI/管道中挂起）
	if _, err := os.Stat(csv); err == nil && !c.Bool("force") {
		return cli.Exit(fmt.Sprintf("错误: 输出文件已存在: %s。如需覆盖请添加 --force 参数", csv), 1)
	}

	return nil
}

// 导入参数验证
func validateImportFlags(c *cli.Context) error {
	table := c.String("table")
	csv := c.String("csv")
	batch := c.Int("batch")

	if table == "" {
		return cli.Exit("错误: 必须指定 --table 参数", 1)
	}

	if csv == "" {
		return cli.Exit("错误: 必须指定 --csv 参数（或使用 API 传入 Input 流）", 1)
	}

	if batch <= 0 {
		return cli.Exit("错误: --batch 参数必须大于0", 1)
	}

	// 检查文件是否存在
	if _, err := os.Stat(csv); os.IsNotExist(err) {
		return cli.Exit(fmt.Sprintf("错误: CSV文件不存在: %s", csv), 1)
	}

	return nil
}
