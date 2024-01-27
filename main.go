package main

import (
	"SleepingBarbersProblem/barbers"
	"SleepingBarbersProblem/constants"
	"SleepingBarbersProblem/customers"
	"context"
	"sync"
	"time"
)

var wg sync.WaitGroup

func main() {
	barberSeats := make([]int, constants.Barbers)
	waitingArea := make(chan int, constants.WaitingAreaSize)
	sleepingBarbers := make(chan chan int, constants.Barbers)

	defer close(sleepingBarbers)

	sleepingBarbersMutex, waitingAreaMutex := &sync.Mutex{}, &sync.Mutex{}

	ctx, cancel := context.WithTimeout(context.Background(), constants.ShopOpenDuration*time.Second)

	wg.Add(1)
	go barbers.StartBarbers(&wg, ctx, cancel, barberSeats, waitingArea, waitingAreaMutex, sleepingBarbers)

	wg.Add(1)
	go customers.StartCustomers(&wg, ctx, cancel, barberSeats, waitingArea, waitingAreaMutex, sleepingBarbers, sleepingBarbersMutex)

	<-ctx.Done()

	wg.Wait()
}
