package fsm

import (
	"fmt"
	"../Driver/elevio"
	"../orders"
	"time"
)

const _pollRate = 20 * time.Millisecond

var numFloor = 4
var numButtons = 3

type CurrentState int

const (
	Idle CurrentState 	= 0
	DriveUp 			= 1
	DriveDown 			= 2
)

var currentState CurrentState

func GetCurrentState() CurrentState{
	return currentState
}


func Init(){
	fmt.Println("fsm init")
	var b elevio.ButtonType
	for floor := 0; floor < numFloor; floor++{
		for button := 0; button < numButtons; button++{
			if button == 0{
				b = elevio.BT_HallUp
			} else if button == 1{
				b = elevio.BT_HallDown
			} else {
				b = elevio.BT_Cab
			}
			elevio.SetButtonLamp(b, floor, false)
		}
	}
	elevio.SetDoorOpenLamp(false)

	orders.SetCurrentDir(0)
	orders.SetLastDir(-1)
	orders.SetOrderFloor(-1)

	elevio.SetMotorDirection(elevio.MD_Down)
	for elevio.GetFloor() == -1{
		time.Sleep(_pollRate)
	}
	orders.SetOnOrderFloor(true)
	elevio.SetMotorDirection(elevio.MD_Stop)
	orders.SetCurrentFloor(elevio.GetFloor())
}


func StateHandler(stateChan chan CurrentState){
	for {
		time.Sleep(_pollRate)
		orderFloor := orders.GetOrderFloor()
		currentFloor := orders.GetCurrentFloor()
		if orders.GetOnOrderFloor() == true || orderFloor == -1 || orderFloor == currentFloor /*&& orders.GetCurrentDir() == 0*/{
			if orders.GetCurrentDir() == 0 {
				currentState = Idle
				stateChan <- currentState
				idle_state(stateChan)
			}
		} else {
			d := orderFloor - currentFloor
			if d > 0{
				currentState = DriveUp
				stateChan <- currentState
				driveUp_state()
			} else if d < 0{
				currentState = DriveDown
				stateChan <- currentState
				driveDown_state()
			} 
		}

	}
}


func idle_state(stateChan chan CurrentState){
	fmt.Println("Enter IDLE")
	if orders.OrderExist() == 1 {
		fmt.Println("Order exists")
	} else{
		fmt.Println("No orders")
	}
	elevio.SetDoorOpenLamp(true)
	orders.SetCurrentDir(0)
	timer := time.After(3 * time.Second)
	for {
		select{
		case <- timer:
			elevio.SetDoorOpenLamp(false)
			for {
				if orders.GetOrderFloor() != -1 {
					orders.SetOnOrderFloor(false)
					return
				} else {
					time.Sleep(_pollRate)
				}
			}
		default:
			if orders.GetCurrentFloor() == orders.GetOrderFloor() {
				fmt.Println("Order in this floor. Deleting it")
				timer = time.After(3 * time.Second)
				stateChan <- Idle
			}
		}
	}
}

func driveUp_state(){
	// Routine for starting motor
	elevio.SetMotorDirection(elevio.MD_Up)
	orders.SetCurrentDir(1) // UP
	for {
		time.Sleep(_pollRate)
		if orders.GetCurrentFloor() == orders.GetOrderFloor(){
			break
		}
	}
	// Routine for stopping motor
	elevio.SetMotorDirection(elevio.MD_Stop)
	orders.SetOnOrderFloor(true)
	orders.SetLastDir(1)
	orders.SetCurrentDir(0)
	
}

func driveDown_state(){
	// Routine for starting motor
	elevio.SetMotorDirection(elevio.MD_Down)
	orders.SetCurrentDir(-1) // DOWN
	for {
		time.Sleep(_pollRate)
		if orders.GetCurrentFloor() == orders.GetOrderFloor(){
			break
		}
	}
	// Routine for stopping motor
	elevio.SetMotorDirection(elevio.MD_Stop)
	orders.SetOnOrderFloor(true)
	orders.SetCurrentDir(0) // NO DIRECTION
	orders.SetLastDir(-1)

}