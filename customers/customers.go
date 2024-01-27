package customers

import (
	"SleepingBarbersProblem/constants"
	"context"
	"fmt"
	"sync"
	"time"
)

var wg sync.WaitGroup

func StartCustomers(callerWG *sync.WaitGroup, ctx context.Context, cancel context.CancelFunc, barberSeats []int, waitingArea chan<- int, waitingAreaMutex *sync.Mutex, sleepingBarbers chan chan int, sleepingBarbersMutex *sync.Mutex) {
	defer callerWG.Done()
	defer cancel()

	for closed, customerID := false, 1; !closed; customerID++ {
		wg.Add(1)
		go cusomter(customerID, barberSeats, waitingArea, waitingAreaMutex, sleepingBarbers, sleepingBarbersMutex)
		time.Sleep(500 * time.Millisecond)

		select {
		case <-ctx.Done():
			closed = true
		default:
		}
	}

	fmt.Println("Shop is closed, no longer accepting new customers")

	close(waitingArea)

	wg.Wait()
}

func cusomter(customerID int, barberSeats []int, waitingArea chan<- int, waitingAreaMutex *sync.Mutex, sleepingBarbers chan chan int, sleepingBarbersMutex *sync.Mutex) {
	defer wg.Done()

	if !wakeBarber(customerID, sleepingBarbers, sleepingBarbersMutex) {
		if !gotoBarberSeats(customerID, barberSeats) {
			gotoWaitingArea(customerID, waitingArea, waitingAreaMutex)
		}
	}
}

func wakeBarber(customerID int, sleepingBarbers chan chan int, sleepingBarbersMutex *sync.Mutex) bool {
	defer sleepingBarbersMutex.Unlock()
	sleepingBarbersMutex.Lock()

	if len(sleepingBarbers) == 0 {
		return false
	}

	wakerChannel := <-sleepingBarbers
	wakerChannel <- customerID

	fmt.Println("Customer", customerID, "woke up a barber")

	return true
}

func gotoBarberSeats(customerID int, barberSeats []int) bool {
	availableBarberSeat := -1

	for seat, val := range barberSeats {
		if val == 0 {
			availableBarberSeat = seat
		}
	}

	if availableBarberSeat != -1 {
		barberSeats[availableBarberSeat] = customerID

		fmt.Println("Customer", customerID, "sat in seat of barber", availableBarberSeat+1)

		return true
	}

	fmt.Println("Customer", customerID, "checked barber seats, but it is full")

	return false
}

func gotoWaitingArea(customerID int, waitingArea chan<- int, waitingAreaMutex *sync.Mutex) {
	defer waitingAreaMutex.Unlock()
	waitingAreaMutex.Lock()

	fmt.Println("Customer", customerID, "is checking waiting area")

	if len(waitingArea) >= constants.WaitingAreaSize {
		fmt.Println("Customer", customerID, "is going back as waiting area is full")
		return
	}

	waitingArea <- customerID

	fmt.Println("Customer", customerID, "has entered waiting area")
}
