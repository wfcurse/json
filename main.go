package main

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

type Payment struct {
	Description string
	USD         int64
}

var mtx = sync.Mutex{}
var money int64 = 1000
var paymentHistory = make([]Payment, 0)

func readAmount(w http.ResponseWriter, r *http.Request) (int64, []string, error) {

	body, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Println("Ошибка чтения тела запроса")
		return 0, nil, err
	}

	bodyString := strings.TrimSpace(string(body))

	bodyStringParts := strings.SplitN(bodyString, ",", 2)

	if len(bodyStringParts) != 2 {
		http.Error(w, "Ошибка запроса", http.StatusBadRequest)
		return 0, nil, err
	}

	usd, err := strconv.ParseInt(bodyStringParts[0], 10, 64)
	if err != nil {
		http.Error(w, "Ошибка перевода string в int", http.StatusBadRequest)
		fmt.Println("Ошибка перевода string в int")
		return 0, nil, nil
	}

	return usd, bodyStringParts, nil

}

func payHandler(w http.ResponseWriter, r *http.Request) {
	usd, bodyStringParts, err := readAmount(w, r)

	if err != nil {
		fmt.Println("ошибка при выполнении функции payHendler")
	}
	payment := Payment{
		Description: bodyStringParts[1],
		USD:         usd,
	}

	defer mtx.Unlock()
	mtx.Lock()

	if money-usd >= 0 {
		money -= usd
	}

	paymentHistory = append(paymentHistory, payment)

	fmt.Println("money: ", money)
	fmt.Println("paymentHistory", paymentHistory)

}

func main() {
	http.HandleFunc("/pay", payHandler)

	if err := http.ListenAndServe(":9091", nil); err != nil {
		fmt.Println("Ошибка во время работы http сервера: ", err)
	}

}
