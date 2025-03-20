package format

import "strings"

// ModelStringCorrect 校正modelstring格式
func ModelStringCorrect(modaltype string) string {
	switch modaltype {
	case "ipv4":
		return "IP"
	default:
		return strings.ToUpper(modaltype)
	}
}
