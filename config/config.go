// config/config.go
package config

// DBConfig 数据库连接配置
type DBConfig struct {
	Server   string
	Port     uint64
	User     string
	Password string
	DBName   string
	Encrypt  string
	Charset  string
	Timeout  uint64
}

// ExportConfig 导出配置
type ExportConfig struct {
	Table        string
	SQL          string
	CSVPath      string
	Header       bool
	Delimiter    rune
	Limit        int
	BinaryFormat string
	FileCharset  string
	NullMarker   string // NULL 值在 CSV 中的标记，空字符串表示用空字段
}

// ImportConfig 导入配置
type ImportConfig struct {
	Table          string
	CSVPath        string
	Batch          int
	Header         bool
	Delimiter      rune
	Truncate       bool
	SkipErrors     bool
	IdentityInsert bool
	BinaryFormat   string
	FileCharset    string
	NullMarker     string // CSV 中代表 NULL 的字符串，空字符串时空字段=null
}
