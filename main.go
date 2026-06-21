package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

type Payment struct {
	Description string `json:"descriotion"`
	USD         int64  `json:"usd"`
	FullName    string `json:"fullName"`
	Address     string `jsin:"address"`
	Time        time.Time
}

type HttpRespons struct {
	Money          int64
	PaymentHistory []Payment
}

var mtx = sync.Mutex{}
var money int64 = 1000
var paymentHistory = make([]Payment, 0)

func (p Payment) Println() {
	fmt.Println("Descripton: ", p.Description)
	fmt.Println("usd: ", p.USD)
	fmt.Println("FullName: ", p.FullName)
	fmt.Println("Address: ", p.Address)
}

func (p Payment) Validate() bool {

	if p.USD == 0 {
		return false
	}

	if p.Address == "" {
		return false
	}

	return true
}

func payHandler(w http.ResponseWriter, r *http.Request) {
	var payment Payment
	if err := json.NewDecoder(r.Body).Decode(&payment); err != nil {
		fmt.Println("ошибка при выполнении функции payHendler", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	payment.Time = time.Now()

	payment.Println()

	mtx.Lock()

	if money-payment.USD >= 0 {
		money -= payment.USD
	}

	paymentHistory = append(paymentHistory, payment)

	httpRespons := HttpRespons{
		Money:          money,
		PaymentHistory: paymentHistory,
	}

	b, err := json.MarshalIndent(httpRespons, "", "	")
	if err != nil {
		fmt.Println("err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if _, err := w.Write(b); err != nil {
		fmt.Println("err", err)

		return
	}

	mtx.Unlock()
}

func main() {
	http.HandleFunc("/pay", payHandler)

	if err := http.ListenAndServe(":9091", nil); err != nil {
		fmt.Println("Ошибка во время работы http сервера: ", err)
	}

}
