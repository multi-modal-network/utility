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

// ExtractSupportModal 从pipeconf中提取模态
func ExtractSupportModal(s string) (string, error) {
	return strings.Split(s, ".")[2], nil
}

// ExtractGroup 从deviceID中提取group
func ExtractGroup(s string) (int32, error) {
	if s == "device:domain2:p1" || s == "device:domain4:p4" || s == "device:domain6:p6" ||
		s == "device:satellite1" || s == "device:satellite2" || s == "device:satellite3" {
		return 0, nil
	}
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
	if s == "device:satellite1" || s == "device:satellite2" || s == "device:satellite3" {
		return 3, nil
	}
	if s == "device:domain2:p1" {
		return 2, nil
	}
	if s == "device:domain4:p4" {
		return 4, nil
	}
	if s == "device:domain6:p6" {
		return 6, nil
	}
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
func ExtractSwitchID(s string) (int32, error) {
	if s == "device:satellite1" {
		return 3100, nil
	}
	if s == "device:satellite2" {
		return 3200, nil
	}
	if s == "device:satellite3" {
		return 3300, nil
	}
	if s == "device:domain2:p1" {
		return 2000, nil
	}
	if s == "device:domain4:p4" {
		return 4000, nil
	}
	if s == "device:domain6:p6" {
		return 6000, nil
	}
	lastIndex := strings.LastIndex(s, ":")
	if lastIndex == -1 {
		// 如果没有找到冒号，返回空字符串和false
		return 0, errors.New("extractSwitchID failed")
	}
	// 获取最后一个冒号之后的所有字符
	switchID, err := strconv.ParseInt(s[lastIndex+2:], 10, 32)
	if err != nil {
		return 0, errors.New("extractSwitchID failed")
	}
	return int32(switchID), nil
}
