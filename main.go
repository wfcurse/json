package main

import (
	"encoding/json" // Пакет для работы с JSON: читать JSON и превращать Go-структуры в JSON
	"fmt"           // Пакет для вывода текста в консоль
	"net/http"      // Пакет для создания HTTP-сервера
	"sync"          // Пакет для Mutex, чтобы безопасно работать с общими переменными
	"time"          // Пакет для работы со временем
)

// Payment описывает один платёж
type Payment struct {
	Description string    `json:"description"` // json:"description" означает: брать поле description из JSON
	USD         int64     `json:"usd"`         // Сумма платежа
	FullName    string    `json:"fullName"`    // Имя плательщика
	Address     string    `json:"address"`     // Адрес плательщика
	Time        time.Time `json:"time"`        // Время платежа
}

// HttpResponse описывает ответ, который сервер отправит клиенту
type HttpResponse struct {
	Money          int64     `json:"money"`          // Текущий остаток денег
	PaymentHistory []Payment `json:"paymentHistory"` // История платежей
}

// mtx нужен, чтобы два запроса одновременно не меняли money и paymentHistory
var mtx = sync.Mutex{}

// money — общий баланс
var money int64 = 1000

// paymentHistory — общий список всех платежей
var paymentHistory = make([]Payment, 0)

// Println — метод структуры Payment
// Он просто выводит данные платежа в консоль
func (p Payment) Println() {
	fmt.Println("Description:", p.Description)
	fmt.Println("USD:", p.USD)
	fmt.Println("FullName:", p.FullName)
	fmt.Println("Address:", p.Address)
}

// payHandler вызывается каждый раз, когда приходит запрос на /pay
func payHandler(w http.ResponseWriter, r *http.Request) {
	var payment Payment // Здесь будет храниться платёж из тела запроса

	// r.Body — это тело HTTP-запроса.
	// Например клиент отправляет:
	// {
	//   "description": "coffee",
	//   "usd": 100,
	//   "fullName": "Ivan",
	//   "address": "Moscow"
	// }

	// json.NewDecoder(r.Body) создаёт JSON-декодер.
	// Декодер — это объект, который умеет читать JSON из потока данных.
	//
	// Decode(&payment) читает JSON из r.Body
	// и записывает данные в переменную payment.
	//
	// &payment — это адрес переменной payment.
	// Он нужен, потому что Decode должен изменить саму переменную.
	// Если передать payment без &, Decode получил бы копию и не смог бы заполнить структуру.
	if err := json.NewDecoder(r.Body).Decode(&payment); err != nil {
		fmt.Println("ошибка чтения JSON:", err)

		// Если JSON неправильный, отправляем клиенту ошибку
		http.Error(w, "Некорректный JSON", http.StatusBadRequest)
		return
	}

	// Важно:
	// r.Body можно прочитать только один раз.
	// После Decode тело запроса уже прочитано.
	// Поэтому после Decode обычно НЕ нужно делать io.ReadAll(r.Body),
	// потому что там уже будет пусто.

	// Добавляем к платежу текущее время сервера
	payment.Time = time.Now()

	// Выводим платёж в консоль сервера
	payment.Println()

	// Блокируем общие данные перед изменением
	mtx.Lock()

	// defer выполнится в конце функции.
	// Это гарантирует, что mutex разблокируется даже если ниже будет return.
	defer mtx.Unlock()

	// Проверяем, хватает ли денег
	if money-payment.USD >= 0 {
		// Если хватает, уменьшаем баланс
		money -= payment.USD
	}

	// Добавляем платёж в историю
	paymentHistory = append(paymentHistory, payment)

	// Создаём структуру ответа клиенту
	response := HttpResponse{
		Money:          money,
		PaymentHistory: paymentHistory,
	}

	// json.MarshalIndent превращает Go-структуру response в красивый JSON.
	//
	// response — что превращаем в JSON
	// ""       — префикс для строк, обычно пустой
	// "\t"     — отступы табуляцией
	b, err := json.MarshalIndent(response, "", "\t")
	if err != nil {
		fmt.Println("ошибка создания JSON:", err)
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}

	// Говорим клиенту, что ответ будет в формате JSON
	w.Header().Set("Content-Type", "application/json")

	// Отправляем JSON клиенту
	if _, err := w.Write(b); err != nil {
		fmt.Println("ошибка отправки ответа:", err)
		return
	}
}

func main() {
	// Регистрируем обработчик.
	// Когда клиент отправит запрос на /pay,
	// Go вызовет функцию payHandler.
	http.HandleFunc("/pay", payHandler)

	// Запускаем HTTP-сервер на порту 9091
	if err := http.ListenAndServe(":9091", nil); err != nil {
		fmt.Println("Ошибка во время работы http сервера:", err)
	}
}
