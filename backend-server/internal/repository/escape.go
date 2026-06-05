package repository

import "strings"

// escapeLike 转义 SQL LIKE 通配符（%、_、\），防止用户输入意外匹配
func escapeLike(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `%`, `\%`)
	s = strings.ReplaceAll(s, `_`, `\_`)
	return s
}
