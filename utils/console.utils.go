package utils

import "fmt"

func PrintRed(str string) {
	colorReset := "\033[0m"
	colorRed := "\033[31m"

	fmt.Print(string(colorRed), str, string(colorReset))
}

func PrintGreen(str string) {
	colorReset := "\033[0m"
	colorGreen := "\033[32m"

	fmt.Print(string(colorGreen), str, string(colorReset))
}

func PrintBlue(str string) {
	colorReset := "\033[0m"
	colorBlue := "\033[34m"

	fmt.Print(string(colorBlue), str, string(colorReset))
}

func PrintPurple(str string) {
	colorReset := "\033[0m"
	colorPurple := "\033[35m"

	fmt.Print(string(colorPurple), str, string(colorReset))
}
