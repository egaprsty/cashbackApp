package main

import (
	"html/template"
	"log"
	"net/http"
	"strconv"
	"sync"
)

type User struct {
	ID       int
	Username string
	Email    string
	Balance  float64
}

type Transaction struct {
	UserID int
	Amount float64
}

var (
	users        = make(map[int]User)
	transactions []Transaction
	mutex        = &sync.Mutex{}
)

func main() {
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/addUser", addUserHandler)
	http.HandleFunc("/addTransaction", addTransactionHandler)
	http.HandleFunc("/cashback", cashbackHandler)

	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	log.Println("Server running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/index.html"))
	tmpl.Execute(w, nil)
}

func addUserHandler(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	email := r.FormValue("email")

	newUser := User{
		ID:       len(users) + 1,
		Username: username,
		Email:    email,
		Balance:  0,
	}

	mutex.Lock()
	users[newUser.ID] = newUser
	mutex.Unlock()

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func addTransactionHandler(w http.ResponseWriter, r *http.Request) {
	userID, _ := strconv.Atoi(r.FormValue("userID"))
	amount, _ := strconv.ParseFloat(r.FormValue("amount"), 64)

	newTransaction := Transaction{
		UserID: userID,
		Amount: amount,
	}

	mutex.Lock()
	transactions = append(transactions, newTransaction)
	user, exists := users[userID]
	if exists {
		user.Balance += amount
		users[userID] = user
	}
	mutex.Unlock()

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func cashbackHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/cashback.html"))

	userID, _ := strconv.Atoi(r.FormValue("userID"))
	user, exists := users[userID]
	if !exists {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	const cashbackThreshold = 100
	const cashbackRate = 0.05

	totalCashback := 0.0
	for _, transaction := range transactions {
		if transaction.UserID == userID && transaction.Amount > cashbackThreshold {
			cashbackAmount := transaction.Amount * cashbackRate
			totalCashback += cashbackAmount
		}
	}

	user.Balance += totalCashback
	users[userID] = user

	data := struct {
		User       User
		Cashback   float64
	}{
		User:     user,
		Cashback: totalCashback,
	}

	tmpl.Execute(w, data)
}
