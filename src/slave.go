package main

import (
	"fmt"
	//"flag"
	"time"
	"./Driver/elevio"
	"./fsm"
	"../src/orders"
	"./network/bcast"
)

const _pollRate = 20 * time.Millisecond

func main() {

	/*
	// Our id can be anything. Here we pass it on the command line, using
	//  `go run main.go -id=our_id`
	var id string
	flag.StringVar(&id, "id", "", "id of this peer")
	flag.Parse()
	*/
	fmt.Println("System initialized")
	numFloors := 4

	elevio.Init("localhost:15658", numFloors)
	fsm.Init()
	
	// Network channels
	//tableTx := make(chan [][]int)
	tableRx := make(chan [][]int)

	// Fsm channels
	fsm_idle := make(chan bool)
	fsm_driveup := make(chan bool)
	fsm_drivedown := make(chan bool)

	// Driver channels
	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	drv_stop := make(chan bool)

	// Network Threads
	//go bcast.Transmitter(20006,tableTx)
	go bcast.Receiver(20006,tableRx)
	
	go func(){
		for {
			time.Sleep(1 * time.Second)
			//orderTable := orders.GetOrderTable()
		}
	}()
	
	// Fsm Threads
	go fsm.StateHandler(fsm_idle,fsm_driveup,fsm_drivedown)
	go fsm.StateMachine(fsm_idle,fsm_driveup,fsm_drivedown)
	
	// Driver Threads
	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)


	for {
		select {
		case a := <-drv_buttons:
			elevio.SetButtonLamp(a.Button, a.Floor, true)
			orders.UpdateOrderTable(a.Floor,int(a.Button))
			orders.PrintState()
			orders.SetPriority()
			orders.PrintOrderFloor()
			orders.PrintOrderTable()

		case a := <-drv_floors:
			elevio.SetFloorIndicator(a)
			orders.SetCurrentFloor(a)
			orders.PrintCurrentFloor()
		
		case a := <- tableRx:
			orders.SetOrderTable(a)
			orders.PrintOrderTable()
			//elevio.SetButtonLamp(a.Button, a.Floor, true)
			orders.SetPriority()
		}
	}
}
