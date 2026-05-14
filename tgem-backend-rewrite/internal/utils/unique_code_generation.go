package utils

import "fmt"

func UniqueCodeGeneration(codeIdentifier string, count int64, projectID uint) string {
	code := codeIdentifier + "-"
	projectIDStr := fmt.Sprint(projectID)
	projectIDLength := len(projectIDStr)
	amountOfZeroes := 2 - projectIDLength
	for amountOfZeroes != 0 {
		code += "0"
		amountOfZeroes--
	}
	code += projectIDStr + "-"

	countStr := fmt.Sprint(count)
	countlength := len(countStr)
	amountOfZeroes = 5 - countlength
	for amountOfZeroes != 0 {
		code += "0"
		amountOfZeroes--
	}

	code += countStr
	return code
}
