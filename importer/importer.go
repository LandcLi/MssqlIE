// importer/import.go
package importer

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/LandcLi/MssqlIE/config"
	"github.com/LandcLi/MssqlIE/utils"
)

// noopReadCloser 包装 io.Reader 为 io.ReadCloser（Close 无操作）
type noopReadCloser struct {
	io.Reader
}

func (n *noopReadCloser) Close() error { return nil }

// CSVToTable 从CSV文件导入数据到指定表
func CSVToTable(db *sql.DB, cfg config.ImportConfig) error {
	// 参数校验
	var err error
	if err := validateImportConfig(cfg); err != nil {
		return fmt.Errorf("配置校验失败: %w", err)
	}

	// 打开CSV输入：优先使用 Input 流，其次打开文件
	var readFile io.ReadCloser
	if cfg.Input != nil {
		readFile = &noopReadCloser{Reader: cfg.Input}
	} else {
		readFile, err = os.Open(cfg.CSVPath)
		if err != nil {
			return fmt.Errorf("打开CSV文件失败: %w", err)
		}
	}
	defer readFile.Close()

	// 应用字符集转换
	csvReader := csv.NewReader(utils.GetTransformersRead(readFile, cfg.FileCharset))
	csvReader.Comma = cfg.Delimiter
	if csvReader.Comma == 0 {
		csvReader.Comma = ',' // API 调用时未设置分隔符，默认逗号
	}

	// 读取列名
	var columnInfos []ColumnInfo
	// 如果没有标题行，尝试从数据库获取列名
	columnInfos, err = getTableColumns(db, cfg.Table)
	if err != nil {
		return fmt.Errorf("获取表列名失败: %w", err)
	}
	var headerRow []string
	var insertCols = make([]ColumnInfo, 0, len(columnInfos))
	if cfg.Header {
		headerRow, err = csvReader.Read()
		if err != nil {
			return fmt.Errorf("读取CSV列名失败: %w", err)
		}
		// 通过列名匹配，不需要列数完全一致
		for _, col := range headerRow {
			found := false
			for _, dbCol := range columnInfos {
				if strings.EqualFold(col, dbCol.Name) {
					insertCols = append(insertCols, dbCol)
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("CSV列 %s 与数据库列名不匹配", col)
			}
		}
		if len(insertCols) == 0 {
			return fmt.Errorf("CSV列名与数据库列名无匹配")
		}
	} else {
		headerRow = make([]string, len(columnInfos))
		for i, col := range columnInfos {
			headerRow[i] = col.Name
		}
		insertCols = columnInfos // 修复: 无表头时填充 insertCols
	}

	// 安全地转义列名
	safeCols := make([]string, len(headerRow))
	for i, col := range headerRow {
		safeCols[i] = utils.EscapeIdentifier(col)
	}

	// 构建插入SQL
	insertSQL, err := buildInsertSQL(cfg.Table, safeCols)
	if err != nil {
		return fmt.Errorf("构建插入SQL失败: %w", err)
	}

	// 如果需要，先清空表
	if cfg.Truncate {
		if err := truncateTable(db, cfg.Table); err != nil {
			return fmt.Errorf("清空表失败: %w", err)
		}
	}
	// 开始事务批量插入
	return batchInsert(BatchInsertConfig{
		DB:             db,
		InsertSQL:      insertSQL,
		Reader:         csvReader,
		Columns:        insertCols,
		TableName:      cfg.Table,
		BatchSize:      cfg.Batch,
		SkipErrors:     cfg.SkipErrors,
		SkipFirstRow:   false,
		IdentityInsert: cfg.IdentityInsert,
		BinaryFormat:   cfg.BinaryFormat,
		NullMarker:     cfg.NullMarker,
		FillDefaults:   cfg.FillDefaults,
	})
}

// validateImportConfig 校验导入配置
func validateImportConfig(cfg config.ImportConfig) error {
	if cfg.Table == "" {
		return fmt.Errorf("目标表名不能为空")
	}
	if cfg.CSVPath == "" && cfg.Input == nil {
		return fmt.Errorf("CSV文件路径或输入流必须设置一个")
	}
	if cfg.Batch <= 0 {
		return fmt.Errorf("批量大小必须大于0（建议500-2000）")
	}
	return nil
}

type ColumnInfo struct {
	Name     string
	DataType string
	Nullable bool
	CharLen  int // CHARACTER_MAXIMUM_LENGTH, -1 for MAX, 0 for unknown
}

// getTableColumns 从数据库获取表的列名
func getTableColumns(db *sql.DB, tableName string) ([]ColumnInfo, error) {
	// 解析限定名（schema.table）
	parts, err := utils.SplitQualifiedName(tableName)
	if err != nil {
		return nil, fmt.Errorf("解析表名失败: %w", err)
	}
	if len(parts) == 0 {
		return nil, fmt.Errorf("表名不能为空")
	}

	schema := "dbo"
	table := parts[len(parts)-1]
	if len(parts) >= 2 {
		schema = parts[len(parts)-2]
	}

	// 参数化查询，避免字符串拼接
	query := `
		/* mssql_ie tool query for check column*/
		SELECT COLUMN_NAME, DATA_TYPE, IS_NULLABLE, COALESCE(CHARACTER_MAXIMUM_LENGTH, 0)
		FROM INFORMATION_SCHEMA.COLUMNS 
		WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?
		ORDER BY ORDINAL_POSITION
	`

	rows, err := db.Query(query, schema, table)
	if err != nil {
		return nil, fmt.Errorf("查询表结构失败: %w", err)
	}
	defer rows.Close()

	var columns []ColumnInfo
	for rows.Next() {
		var col ColumnInfo
		var nullableStr string
		var charLen int
		if err := rows.Scan(&col.Name, &col.DataType, &nullableStr, &charLen); err != nil {
			return nil, err
		}
		col.Nullable = nullableStr == "YES"
		col.CharLen = charLen
		columns = append(columns, col)
	}

	if len(columns) == 0 {
		return nil, fmt.Errorf("表 %s 不存在或没有列", tableName)
	}

	return columns, nil
}

// truncateTable 清空表
func truncateTable(db *sql.DB, tableName string) error {
	escapedTable, err := utils.EscapeQualifiedName(tableName)
	if err != nil {
		return fmt.Errorf("转义表名失败: %w", err)
	}

	// 使用TRUNCATE TABLE
	_, err = db.Exec(fmt.Sprintf("TRUNCATE TABLE %s", escapedTable))
	return err
}

// buildInsertSQL 构建参数化插入SQL
func buildInsertSQL(table string, safeCols []string) (string, error) {
	if len(safeCols) == 0 {
		return "", fmt.Errorf("列名不能为空")
	}

	// 转义表名
	safeTable, err := utils.EscapeQualifiedName(table)
	if err != nil {
		return "", fmt.Errorf("转义表名失败: %w", err)
	}

	// 构建占位符
	placeholders := make([]string, len(safeCols))
	for i := range safeCols {
		placeholders[i] = "?"
	}

	return fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s)",
		safeTable,
		strings.Join(safeCols, ","),
		strings.Join(placeholders, ","),
	), nil
}

// BatchInsertConfig 批量插入配置
type BatchInsertConfig struct {
	DB             *sql.DB
	InsertSQL      string
	Reader         *csv.Reader
	Columns        []ColumnInfo
	TableName      string
	BatchSize      int
	SkipErrors     bool
	SkipFirstRow   bool
	IdentityInsert bool
	BinaryFormat   string
	NullMarker     string
	FillDefaults   bool
}

// batchInsert 批量插入数据
func batchInsert(cfg BatchInsertConfig) error {
	db := cfg.DB
	insertSQL := cfg.InsertSQL
	reader := cfg.Reader
	safeCols := cfg.Columns
	tableName := cfg.TableName
	batchSize := cfg.BatchSize
	skipErrors := cfg.SkipErrors
	skipFirstRow := cfg.SkipFirstRow
	identityInsert := cfg.IdentityInsert
	binaryFormat := cfg.BinaryFormat
	nullMarker := cfg.NullMarker
	fillDefaults := cfg.FillDefaults
	// 开始事务
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("开启事务失败: %w", err)
	}

	// 异常回滚处理
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p) // 重新抛出panic
		}
	}()

	// 如果需要允许自增列插入，在事务内启用
	if identityInsert {
		safeTable, err := utils.EscapeQualifiedName(tableName)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("转义表名失败: %w", err)
		}
		if _, err := tx.Exec(fmt.Sprintf("SET IDENTITY_INSERT %s ON", safeTable)); err != nil {
			tx.Rollback()
			return fmt.Errorf("启用 IDENTITY_INSERT 失败: %w", err)
		}
		fmt.Fprintln(os.Stderr, "[INFO] 已启用 IDENTITY_INSERT")
	}

	// 预处理插入语句
	stmt, err := tx.Prepare(insertSQL)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("预处理插入语句失败: %w", err)
	}
	defer stmt.Close()

	batchCount := 0
	totalCount := 0
	rowNum := 0
	errorRows := []int{}

	// 循环读取CSV行
	for {
		row, err := reader.Read()
		rowNum++

		// 跳过标题行
		if skipFirstRow && rowNum == 1 {
			continue
		}

		if err != nil {
			if err == io.EOF {
				break
			}
			if skipErrors {
				errorRows = append(errorRows, rowNum)
				continue
			}
			tx.Rollback()
			return fmt.Errorf("读取CSV行失败(行%d): %w", rowNum, err)
		}

		// 列数校验
		if len(row) != len(safeCols) {
			if skipErrors {
				errorRows = append(errorRows, rowNum)
				continue
			}
			tx.Rollback()
			return fmt.Errorf("行%d数据列数不匹配（期望%d列，实际%d列）", rowNum, len(safeCols), len(row))
		}

		// 准备参数
		args := make([]interface{}, len(row))
		for i, v := range row {
			if nullMarker != "" {
				// 自定义 NULL 标记模式：仅 nullMarker 代表 NULL，空串保持为空串
				if v == nullMarker {
					setArgToNull(&args[i], safeCols[i])
				} else {
					args[i], err = convertValue(v, safeCols[i], binaryFormat)
				}
		} else {
			// 传统模式：空字段 = NULL
			if v == "" {
				if safeCols[i].Nullable {
					setArgToNull(&args[i], safeCols[i])
				} else if fillDefaults {
					args[i] = getDefaultValue(safeCols[i].DataType)
				} else {
					tx.Rollback()
					return fmt.Errorf("行%d列%d(%s)为空，但该列不允许NULL。请填充数据或使用 --fill-defaults 显式填充默认值",
						rowNum, i+1, safeCols[i].Name)
				}
			} else {
				args[i], err = convertValue(v, safeCols[i], binaryFormat)
			}
		}
			if err != nil {
				if skipErrors {
					errorRows = append(errorRows, rowNum)
					err = nil // reset err for next iteration
					continue
				}
				tx.Rollback()
				return fmt.Errorf("转换值失败(行%d,列%d): %w", rowNum, i+1, err)
			}
		}

		// 执行插入
		if _, err := stmt.Exec(args...); err != nil {
			if skipErrors {
				errorRows = append(errorRows, rowNum)
				continue
			}
			tx.Rollback()
			return fmt.Errorf("插入行失败(行%d): %w", rowNum, err)
		}

		batchCount++
		totalCount++

		// 达到批量大小提交事务
		if batchCount >= batchSize {
			if err := tx.Commit(); err != nil {
				return fmt.Errorf("提交批量事务失败(累计%d行): %w", totalCount, err)
			}

			// 开始新事务
			tx, err = db.Begin()
			if err != nil {
				return fmt.Errorf("重新开启事务失败: %w", err)
			}

			// 每个新事务都需要重新启用 IDENTITY_INSERT（连接可能不同）
			if identityInsert {
				safeTable, _ := utils.EscapeQualifiedName(tableName)
				if _, err := tx.Exec(fmt.Sprintf("SET IDENTITY_INSERT %s ON", safeTable)); err != nil {
					tx.Rollback()
					return fmt.Errorf("启用 IDENTITY_INSERT 失败: %w", err)
				}
			}

			// 重新预处理语句
			stmt, err = tx.Prepare(insertSQL)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("重新预处理语句失败: %w", err)
			}
			batchCount = 0

			fmt.Fprintf(os.Stderr, "已导入 %d 行...\n", totalCount)
		}
	}

	// 提交剩余数据
	if batchCount > 0 {
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("提交剩余数据失败: %w", err)
		}
	}

	// 输出结果
	fmt.Fprintf(os.Stderr, "[OK] CSV导入完成，共插入 %d 行数据\n", totalCount)
	if len(errorRows) > 0 {
		fmt.Fprintf(os.Stderr, "[WARN] 跳过 %d 行错误数据: %v\n", len(errorRows), errorRows)
	}

	return nil
}

func convertValue(value string, col ColumnInfo, binaryFormat string) (interface{}, error) {
	if value == "" {
		return getDefaultValue(col.DataType), nil
	}
	switch strings.ToLower(col.DataType) {
	case "bit":
		value = strings.ToLower(value)
		if value == "true" || value == "1" || value == "y" || value == "yes" || value == "t" {
			return true, nil
		}
		if value == "false" || value == "0" || value == "n" || value == "no" || value == "f" {
			return false, nil
		}
		return nil, fmt.Errorf("无效的位值: %s", value)
	case "binary", "varbinary", "image":
		return convertBinary(value, binaryFormat)
	case "geometry", "geography":
		return convertGeo(value, binaryFormat)
	case "hierarchyid":
		return convertHierarchyID(value, binaryFormat)
	case "uniqueidentifier":
		return convertGUID(value, binaryFormat)
	default: // 其他类型保持字符串
		return value, nil
	}
}
func convertGUID(value, binaryFormat string) (interface{}, error) {
	if utils.IsValidGUID(value) {
		return value, nil
	}
	return convertBinary(value, binaryFormat)
}
func convertHierarchyID(value, binaryFormat string) ([]byte, error) {
	return convertBinary(value, binaryFormat)
}
func convertGeo(value, binaryFormat string) (interface{}, error) {
	if isValidWKT(value) {
		return value, nil
	}
	return convertBinary(value, binaryFormat)
}
func isValidWKT(wkt string) bool {
	return (strings.HasPrefix(strings.ToUpper(wkt), "POINT(") ||
		strings.HasPrefix(strings.ToUpper(wkt), "LINESTRING(") ||
		strings.HasPrefix(strings.ToUpper(wkt), "POLYGON(") ||
		strings.HasPrefix(strings.ToUpper(wkt), "MULTIPOINT(") ||
		strings.HasPrefix(strings.ToUpper(wkt), "MULTILINESTRING(") ||
		strings.HasPrefix(strings.ToUpper(wkt), "MULTIPOLYGON(") ||
		strings.HasPrefix(strings.ToUpper(wkt), "GEOMETRYCOLLECTION(")) && (strings.HasSuffix(strings.ToUpper(wkt), ")") ||
		strings.HasSuffix(strings.ToUpper(wkt), "ZM)") ||
		strings.HasSuffix(strings.ToUpper(wkt), "M)"))
}
func convertBinary(value string, binaryFormat string) ([]byte, error) {
	switch strings.ToLower(binaryFormat) {
	case "hex":
		return utils.HexToBytes(value)
	case "base64":
		return utils.Base64ToBytes(value)
	default: // raw
		return []byte(value), nil
	}
}
// setArgToNull 将参数设为 NULL（二进制类型用带类型 nil 以避免 go-mssqldb 误用 nvarchar）
func setArgToNull(arg *interface{}, col ColumnInfo) {
	if isBinaryType(col.DataType) {
		var b []byte = nil
		*arg = b
	} else {
		*arg = nil
	}
}

func isBinaryType(dataType string) bool {
	switch strings.ToLower(dataType) {
	case "binary", "varbinary", "image":
		return true
	}
	return false
}

func getDefaultValue(dataType string) interface{} {
	switch strings.ToLower(dataType) {
	case "int", "smallint", "tinyint", "bigint", "numeric", "decimal", "real", "float":
		return 0
	case "bit":
		return false
	case "binary", "varbinary", "image":
		return []byte{}
	case "geometry", "geography":
		return []byte{}
	case "hierarchyid":
		return []byte{}
	default: // 其他类型返回空字符串
		return ""
	}
}
