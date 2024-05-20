package util

import "strings"

func ContainsInArray(strArray []string, targetStr string) bool {
	contains := false
	for _, str := range strArray {
		if strings.Contains(str, targetStr) {
			contains = true
			break
		}
	}
	return contains
}

func ContainsArray(str string, targetStr ...string) bool {
	contains := false
	for _, target := range targetStr {
		if strings.Contains(str, target) {
			contains = true
			break
		}
	}
	return contains
}
func EqualsAnyIgnoreCase(s string, strs ...string) bool {
	sLower := strings.ToLower(s)
	for _, str := range strs {
		if sLower == strings.ToLower(str) {
			return true
		}
	}
	return false
}
