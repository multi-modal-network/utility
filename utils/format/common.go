package format

import "strings"

// ModelStringCorrect 校正modelstring格式
func ModelStringCorrect(modaltype string) string {
	switch modaltype {
	case "ipv4":
		return "ip"
	default:
		return strings.ToLower(modaltype)
	}
}
