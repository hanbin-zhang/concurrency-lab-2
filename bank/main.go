package main

import (
	"container/list"
	"flag"
	"fmt"
	"github.com/ChrisGora/semaphore"
	"math/rand"
	"strconv"
	"time"
)

var debug *bool

// An executor is a type of a worker goroutine that handles the incoming transactions.
func executor(bank *bank, executorId int, ProcessedTransaction <-chan transaction, done chan<- bool,
	semaphores map[string]semaphore.Semaphore) {
	for {
		t := <-ProcessedTransaction

		from := bank.getAccountName(t.from)
		to := bank.getAccountName(t.to)

		fmt.Println("Executor\t", executorId, "attempting transaction from", from, "to", to)
		e := bank.addInProgress(t, executorId) // Removing this line will break visualisations.

		bank.lockAccount(t.from, strconv.Itoa(executorId))
		fmt.Println("Executor\t", executorId, "locked account", from)
		/*bank.lockAccount(t.to, strconv.Itoa(executorId))
		fmt.Println("Executor\t", executorId, "locked account", to)*/

		bank.execute(t, executorId)

		bank.unlockAccount(t.from, strconv.Itoa(executorId))
		fmt.Println("Executor\t", executorId, "unlocked account", from)
		/*bank.unlockAccount(t.to, strconv.Itoa(executorId))
		fmt.Println("Executor\t", executorId, "unlocked account", to)*/

		bank.removeCompleted(e, executorId) // Removing this line will break visualisations.
		semaphores[from].Post()
		semaphores[to].Post()
		done <- true
	}
}

func toChar(i int) rune {
	return rune('A' + i)
}

func manager(bank *bank, transactionQueue <-chan transaction, processedTransaction chan transaction,
	Semaphores map[string]semaphore.Semaphore) {
	ts := make([]transaction, 1000)
	for i := 0; i < 1000; i++ {
		t := <-transactionQueue
		ts[i] = t
	}

	i := 0
LOOP:
	for {
		if len(ts) != 0 {
			if i >= len(ts) {
				i = 0
			}
			t := ts[i]

			from := bank.getAccountName(t.from)
			to := bank.getAccountName(t.to)
			if Semaphores[from].GetValue() == 0 || Semaphores[to].GetValue() == 3 {
				i = i + 1
				continue
			} else {

				Semaphores[from].Wait()
				Semaphores[to].Wait()

				ts = append(ts[:i], ts[i+1:]...)
				processedTransaction <- t
				i = i + 1
			}
		} else {
			break LOOP
		}
	}
}

// main creates a bank and executors that will be handling the incoming transactions.
func main() {
	rand.Seed(time.Now().UTC().UnixNano())
	debug = flag.Bool("debug", false, "generate DOT graphs of the state of the bank")
	flag.Parse()

	bankSize := 6 // Must be even for correct visualisation.
	transactions := 1000

	accounts := make([]*account, bankSize)
	for i := range accounts {
		accounts[i] = &account{name: string(toChar(i)), balance: 1000}
	}

	bank := bank{
		accounts:               accounts,
		transactionsInProgress: list.New(),
		gen:                    newGenerator(),
	}

	startSum := bank.sum()

	transactionQueue := make(chan transaction, transactions)
	expectedMoneyTransferred := 0
	for i := 0; i < transactions; i++ {
		t := bank.getTransaction()
		expectedMoneyTransferred += t.amount
		transactionQueue <- t
	}

	done := make(chan bool)
	processedTransaction := make(chan transaction)
	semaphores := make(map[string]semaphore.Semaphore)
	/*	accountsOccupied := make([]bool, 6)
		for i:=0;i<6;i++ {
			accountsOccupied[i] = false
		}*/
	for i := 0; i < 6; i++ {
		semaphores[bank.getAccountName(i)] = semaphore.Init(1, 1)
	}
	go manager(&bank, transactionQueue, processedTransaction, semaphores)
	for i := 0; i < bankSize; i++ {

		go executor(&bank, i, processedTransaction, done, semaphores)
	}

	for total := 0; total < transactions; total++ {
		fmt.Println("Completed transactions\t", total)
		<-done
	}

	fmt.Println()
	fmt.Println("Expected transferred", expectedMoneyTransferred)
	fmt.Println("Actual transferred", bank.moneyTransferred)
	fmt.Println("Expected sum", startSum)
	fmt.Println("Actual sum", bank.sum())
	if bank.sum() != startSum {
		panic("sum of the account balances does not much the starting sum")
	} else if len(transactionQueue) > 0 {
		panic("not all transactions have been executed")
	} else if bank.moneyTransferred != expectedMoneyTransferred {
		panic("incorrect amount of money was transferred")
	} else {
		fmt.Println("The bank works!")
	}
}
