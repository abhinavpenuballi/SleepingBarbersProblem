package barbers

import (
	"SleepingBarbersProblem/constants"
	"context"
	"fmt"
	"sync"
	"time"
)

var wg sync.WaitGroup

func StartBarbers(callerWG *sync.WaitGroup, ctx context.Context, cancel context.CancelFunc, barberSeats []int, waitingArea <-chan int, waitingAreaMutex *sync.Mutex, sleepingBarbers chan<- chan int) {
	defer callerWG.Done()
	defer cancel()

	for barberID := 1; barberID <= constants.Barbers; barberID++ {
		wg.Add(1)
		go barber(ctx, barberID, barberSeats, waitingArea, waitingAreaMutex, sleepingBarbers)
	}

	wg.Wait()
}

func barber(ctx context.Context, barberID int, barberSeats []int, waitingArea <-chan int, waitingAreaMutex *sync.Mutex, sleepingBarbers chan<- chan int) {
	defer wg.Done()

	closed := false

	for !closed || stillHaveCustomers(barberID, barberSeats, waitingArea, waitingAreaMutex) {
		if !doHairCut(barberID, barberSeats) {
			if !checkWaitingArea(barberID, waitingArea, waitingAreaMutex, barberSeats) {
				sleep(ctx, barberID, sleepingBarbers, barberSeats)
			}
		}

		select {
		case <-ctx.Done():
			closed = true
		default:
		}
	}

	fmt.Println("Barber", barberID, "finished for the day")
}

func doHairCut(barberID int, barberSeats []int) bool {
	if barberSeats[barberID-1] > 0 {
		fmt.Println("Barber", barberID, "is doing hair cut for customer", barberSeats[barberID-1])
		time.Sleep(10 * time.Second)
		fmt.Println("Barber", barberID, "has done hair cut for customer", barberSeats[barberID-1])

		barberSeats[barberID-1] = 0
		time.Sleep(500 * time.Millisecond)

		return true
	}

	fmt.Println("Barber", barberID, "found no one in barber seat")

	return false
}

func checkWaitingArea(barberID int, waitingArea <-chan int, waitingAreaMutex *sync.Mutex, barberSeats []int) bool {
	defer waitingAreaMutex.Unlock()
	waitingAreaMutex.Lock()

	if len(waitingArea) == 0 {
		fmt.Println("Barber", barberID, "found no one in waiting area")
		return false
	}

	customerID := <-waitingArea
	barberSeats[barberID-1] = customerID
	fmt.Println("Barber", barberID, "picked customer", customerID, "from waiting area")

	return true
}

func sleep(ctx context.Context, barberID int, sleepingBarbers chan<- chan int, barberSeats []int) {
	fmt.Println("Barber", barberID, "is going to sleep")

	barberSeats[barberID-1] = -1

	wakerChannel := make(chan int, 1)
	sleepingBarbers <- wakerChannel

	defer close(wakerChannel)

	select {
	case customerID := <-wakerChannel:
		fmt.Println("Barber", barberID, "is awaken by customer", customerID)
		barberSeats[barberID-1] = customerID
	case <-ctx.Done():
		wakerChannel <- -1
		fmt.Println("Barber", barberID, "was awaken at the end of the day")
	}
}

func stillHaveCustomers(barberID int, barberSeats []int, waitingArea <-chan int, waitingAreaMutex *sync.Mutex) bool {
	if barberSeats[barberID-1] > 0 {
		return true
	}

	defer waitingAreaMutex.Unlock()
	waitingAreaMutex.Lock()

	return len(waitingArea) > 0
}
