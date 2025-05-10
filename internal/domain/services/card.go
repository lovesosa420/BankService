package services

import (
	mathRand "math/rand"
	"strconv"
	"strings"
	"time"
)

func GenerateCardNumber() string {
	mas := make([]int, 16)
	for i := 0; i < 15; i++ {
		mas[i] = mathRand.Intn(10)
	}
	sum := 0
	for i := 0; i < 15; i++ {
		if i%2 == 0 {
			num := mas[i] * 2
			if num > 9 {
				num -= 9
			}
			sum += num
		} else {
			sum += mas[i]
		}
	}
	mas[15] = (10 - (sum % 10)) % 10
	var number strings.Builder
	for _, elem := range mas {
		number.WriteString(strconv.Itoa(elem))
	}
	return number.String()
}

func GenerateCVV() string {
	var cvv strings.Builder
	for i := 0; i < 3; i++ {
		num := mathRand.Intn(10)
		cvv.WriteString(strconv.Itoa(num))
	}
	return cvv.String()
}

func GenerateDate() string {
	now := time.Now()
	futureDate := time.Date(now.Year()+3, now.Month(), 1, 0, 0, 0, 0, time.UTC)
	return futureDate.Format("2006-01-02")
}
