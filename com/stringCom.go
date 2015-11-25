package com

func Substring(str string, start int) string {
	return Substr(str, start, len(str)-start)
}

func Substr(str string, start, length int) string {
	temp := []rune(str)
	tempLength := len(temp)
	end := 0
	if start < 0 {
		start = tempLength - 1 + start
	}
	end = start + length

	if start > end {
		start, end = end, start
	}

	if start < 0 {
		start = 0
	}

	if end < 0 {
		end = 0
	}

	if start > tempLength {
		start = tempLength
	}

	if end > tempLength {
		end = tempLength
	}

	return string(temp[start:end])
}
