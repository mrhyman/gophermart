package util

import "strconv"

func ValidateLuhn(number string) bool {
	if len(number) == 0 {
		return false
	}

	// Проверяем, что все символы - цифры
	for _, r := range number {
		if r < '0' || r > '9' {
			return false
		}
	}

	sum := 0
	isSecond := false

	// Проходим с конца
	for i := len(number) - 1; i >= 0; i-- {
		digit, _ := strconv.Atoi(string(number[i]))

		if isSecond {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}

		sum += digit
		isSecond = !isSecond
	}

	return sum%10 == 0
}
