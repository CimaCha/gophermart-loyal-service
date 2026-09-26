// Package luhn реализует алгоритм Луна для проверки номеров заказов.
package luhn

// Validate проверяет строку по алгоритму Луна.
// Возвращает false для пустой строки, строки из одних нулей и нецифровых символов.
func Validate(orderNumber string) bool {
	if orderNumber == "" {
		return false
	}

	sum := 0
	alternate := false

	for i := len(orderNumber) - 1; i >= 0; i-- {
		c := orderNumber[i]
		if c < '0' || c > '9' {
			return false
		}

		digit := int(c - '0')
		if alternate {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}
		sum += digit
		alternate = !alternate
	}

	return sum > 0 && sum%10 == 0
}
