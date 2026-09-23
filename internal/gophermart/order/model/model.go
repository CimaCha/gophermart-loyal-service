package model

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

var (
	ErrInvalidOrderNumber = errors.New("invalid order number checksum")
)

type Order struct {
	OrderNum   int64
	UserID     uuid.UUID
	Status     string
	Accrual    *decimal.Decimal
	UploadedAt time.Time
}

func NewOrder(orderNum int64, uid uuid.UUID) *Order {
	return &Order{
		OrderNum:   orderNum,
		UserID:     uid,
		Status:     "NEW",
		Accrual:    nil,
		UploadedAt: time.Now(),
	}
}

func (o *Order) Validate() (*Order, error) {
	if o.OrderNum < 0 {
		return nil, errors.New("order id must be positive")
	}

	if !validLuhn(o.OrderNum) {
		return nil, ErrInvalidOrderNumber
	}

	return o, nil
}

// Идём справо на лево, забираем по одной цифре с конца и чередуем их обработку,
// умножая нечётные цифры в числе на 2. Если после умножения число получилось двузначное,
// то складываем две цифры в новое число (считаем цифровой корень).
// Суть валидации по Луна - проверить контрольную сумму, взвесив все цифры
// чтобы проверить, что в числе цифры не были перепутаны или если какие-то цифры потеряны.
func validLuhn(orderNumber int64) bool {

	if orderNumber <= 0 {
		return false
	}

	sum := 0
	alternate := false
	for orderNumber > 0 {
		// получаем mod - последнюю цифру в заказе
		mod := int(orderNumber % 10)
		if alternate {
			mod *= 2
			if mod > 9 {
				mod -= 9
			}
		}
		sum += mod
		// чередуем числа через раз
		alternate = !alternate
		// избавляемся от последней цифры в номере заказа
		orderNumber /= 10
	}

	// Проверяем что итоговая сумма кратна 10. Суть в том, что алгоритм Luna использует
	// контрольную цифру, добавляя её в конец номера. Эта контрольная цифра гарантирует, что
	// итоговая сумма будет всегда кратна 10. Если номер неверный, то итоговая сумма не будет кратна 10.
	return sum%10 == 0
}
