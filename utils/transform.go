package utils

import (
	"io"
	"strings"

	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

// encoderMap 字符集名称到 Encoder 的映射表
var encoderMap = map[string]transform.Transformer{
	"gbk":            simplifiedchinese.GBK.NewEncoder(),
	"iso-8859-1":    charmap.ISO8859_1.NewEncoder(),
	"iso-8859-2":    charmap.ISO8859_2.NewEncoder(),
	"iso-8859-3":    charmap.ISO8859_3.NewEncoder(),
	"iso-8859-4":    charmap.ISO8859_4.NewEncoder(),
	"iso-8859-9":    charmap.ISO8859_9.NewEncoder(),
	"iso-8859-10":   charmap.ISO8859_10.NewEncoder(),
	"iso-8859-13":   charmap.ISO8859_13.NewEncoder(),
	"iso-8859-14":   charmap.ISO8859_14.NewEncoder(),
	"iso-8859-15":   charmap.ISO8859_15.NewEncoder(),
	"iso-8859-16":   charmap.ISO8859_16.NewEncoder(),
	"cp1252":        charmap.Windows1252.NewEncoder(),
	"windows-1252":  charmap.Windows1252.NewEncoder(),
}

// decoderMap 字符集名称到 Decoder 的映射表
var decoderMap = map[string]transform.Transformer{
	"gbk":            simplifiedchinese.GBK.NewDecoder(),
	"iso-8859-1":    charmap.ISO8859_1.NewDecoder(),
	"iso-8859-2":    charmap.ISO8859_2.NewDecoder(),
	"iso-8859-3":    charmap.ISO8859_3.NewDecoder(),
	"iso-8859-4":    charmap.ISO8859_4.NewDecoder(),
	"iso-8859-9":    charmap.ISO8859_9.NewDecoder(),
	"iso-8859-10":   charmap.ISO8859_10.NewDecoder(),
	"iso-8859-13":   charmap.ISO8859_13.NewDecoder(),
	"iso-8859-14":   charmap.ISO8859_14.NewDecoder(),
	"iso-8859-15":   charmap.ISO8859_15.NewDecoder(),
	"iso-8859-16":   charmap.ISO8859_16.NewDecoder(),
	"cp1252":        charmap.Windows1252.NewDecoder(),
	"windows-1252":  charmap.Windows1252.NewDecoder(),
}

// GetTransformersWrite 根据 format 返回对应的编码转换 Writer
func GetTransformersWrite(writer io.Writer, format string) io.Writer {
	format = strings.ToLower(format)
	if t, ok := encoderMap[format]; ok {
		return transform.NewWriter(writer, t)
	}
	return writer // raw / utf8
}

// GetTransformersRead 根据 format 返回对应的编码转换 Reader
func GetTransformersRead(reader io.Reader, format string) io.Reader {
	format = strings.ToLower(format)
	if t, ok := decoderMap[format]; ok {
		return transform.NewReader(reader, t)
	}
	return reader // raw / utf8
}
