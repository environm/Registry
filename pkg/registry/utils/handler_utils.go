package utils

import "strings"

// GetQueryParamCaseInsensitive 获取不区分大小写的查询参数
func GetQueryParamCaseInsensitive(params map[string][]string, paramName string) string {
	for key, values := range params {
		if strings.EqualFold(key, paramName) && len(values) > 0 {
			return values[0]
		}
	}
	return ""
}
