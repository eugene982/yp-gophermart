package utils

import (
	"fmt"
	"strconv"

	"github.com/eugene982/yp-gophermart/internal/model"
)

type HasherFunc func(model.LoginReqest) string

func (f HasherFunc) Hash(r model.LoginReqest) string {
	return f(r)
}

// Проверка корректности номера заказа
func OrderNumberToInt(order string) (int64, error) {

	number, err := strconv.ParseInt(order, 10, 64)
	if err != nil {
		return 0, err
	}

	// Valid check number is valid or not based
	// on Luhn algorithm
	var luhn int64
	num := number / 10
	for i := 0; num > 0; i++ {
		cur := num % 10

		if i%2 == 0 { // even
			cur = cur * 2
			if cur > 9 {
				cur = cur%10 + cur/10
			}
		}
		luhn += cur
		num = num / 10
	}

	if (number%10+luhn%10)%10 != 0 {
		return 0, fmt.Errorf("invalid check number %s", order)
	}
	return number, nil
}
