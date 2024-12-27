package calc

import (
	"errors"
	"strconv"
	"strings"
)

func extractStringBetweenSubStr(s, substr1, substr2 string) (string, bool) {
	start := strings.Index(s, substr1)
	if start == -1 {
		return "", false
	}
	start += len(substr1)
	end := strings.Index(s, substr2)
	if end == -1 {
		return "", false
	}
	return s[start:end], true
}

// ExtractGroup 从deviceID中提取group
func ExtractGroup(s string) (int32, error) {
	groupStr, ok := extractStringBetweenSubStr(s, "group", ":level")
	if !ok {
		return 0, errors.New("extractGroup failed")
	}
	group, err := strconv.ParseInt(groupStr, 10, 32)
	if err != nil {
		return 0, errors.New("extractGroup failed")
	}
	return int32(group), nil
}

// ExtractDomain 从deviceID中提取domain
func ExtractDomain(s string) (int32, error) {
	domainStr, ok := extractStringBetweenSubStr(s, "domain", ":group")
	if !ok {
		return 0, errors.New("extractDomain failed")
	}
	domain, err := strconv.ParseInt(domainStr, 10, 32)
	if err != nil {
		return 0, errors.New("extractDomain failed")
	}
	return int32(domain), nil
}

// ExtractSwitchID 从deviceID中提取switchID
func ExtractSwitchID(s string) (string, error) {
	lastIndex := strings.LastIndex(s, ":")
	if lastIndex == -1 {
		// 如果没有找到冒号，返回空字符串和false
		return "", errors.New("ExtractSwitchID failed")
	}
	// 获取最后一个冒号之后的所有字符
	return s[lastIndex+1:], nil
}
