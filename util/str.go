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
