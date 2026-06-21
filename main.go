package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
)

type Payment struct {
	Description string `json:"descriotion"`
	USD         int64  `json:"usd"`
	FullName    string `json:"fullName"`
	Address     string `jsin:"address"`
}

var mtx = sync.Mutex{}
var money int64 = 1000
var paymentHistory = make([]Payment, 0)

func payHandler(w http.ResponseWriter, r *http.Request) {

	httpBody, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Println("ошибка при выполнении функции payHendler", err)
		return
	}

	var payment Payment
	json.Unmarshal(httpBody, &payment)

	mtx.Lock()
	defer mtx.Unlock()

	if money-payment.USD >= 0 {
		money -= payment.USD
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
