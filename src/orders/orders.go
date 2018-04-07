package orders

import (
	"fmt"
	"time"
	"../Driver/elevio"
)

var orderTable = [][]int{
	// Up, Down, Cab
	{  0,  0,    0}, // Floor 1
	{  0,  0,    0}, // Floor 2
	{  0,  0,    0}, // Floor 3
	{  0,  0,    0},  // Floor 4
}

var numFloor = 4
var numButtons = 3
var currentFloor int
var orderFloor int
var currentDir int
var lastDir int
var onOrderFloor bool
const _pollRate = 20 * time.Millisecond
var button_string = []string{"Up","Down","Cab"}
var b = []elevio.ButtonType{elevio.BT_HallUp, elevio.BT_HallDown, elevio.BT_Cab}

func ClearOrder(){
	for i := 0; i<3; i++{
		orderTable[currentFloor][i] = 0
		elevio.SetButtonLamp(b[i],currentFloor,false)
	}
}

func NewClearOrder(){
	for i := 0; i < 3; i++{
		if orderTable[currentFloor][i] == 1{
			orderTable[currentFloor][i] = 0
			elevio.SetButtonLamp(b[i],currentFloor,false)
		}
	}
}

func ClearAllHalls(){
	var b elevio.ButtonType
	for floor := 0; floor < 4; floor++{
		for button := 0; button < 2; button++{
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
}

func SetOnOrderFloor(state bool){
	onOrderFloor = state
}

func GetOnOrderFloor() bool{
	return onOrderFloor
}

func SetOrderFloor(floor int){
	orderFloor = floor
}

func SetCurrentFloor(floor int){
	currentFloor = floor
}

func GetCurrentFloor() int{
	return currentFloor
}

func GetOrderFloor() int{
	return orderFloor
}

func UpdateOrderTable(floor int, button int){
	orderTable[floor][button] = 1
}

func GetOrderTable() [][]int {
	return orderTable
}

func SetOrderTable(table [][]int) {
	orderTable = table
}

func IsEmpty(table [][]int) bool{
	for i := 0; i < numFloor; i++ {
		for j := 0; j < numButtons; j++ {
			if table[i][j] == 1 {
				return false
			}
		}
	}
	return true
}

func PrintOrderTable(){
	for floor := 0; floor < 4; floor++{
		for button := 0; button < 3; button++{
			fmt.Println("Floor:",floor+1,"Button:",button_string[button],orderTable[floor][button])
		}
	}
	fmt.Printf("\n")
}

func PrintState(){
	fmt.Println("Current Floor:",currentFloor)
	fmt.Println("Order Floor:",orderFloor)
	fmt.Println("Current Dir:",currentDir)
	fmt.Println("Last Dir:",lastDir)
}

func PrintCurrentFloor(){
	fmt.Println("Current floor:",currentFloor)
}

func PrintOrderFloor(){
	fmt.Println("Order floor:",orderFloor)
}

func OrderExist() int {
	for floor := 0; floor < numFloor; floor++{
		for button := 0; button < numButtons; button++{
			if orderTable[floor][button] == 1 {
				return 1
			}
		}
	}
	return 0
}

func orderAbove(){
	for floor := currentFloor+1; floor < 4; floor++{
		if orderTable[floor][0] == 1 || orderTable[floor][2] == 1 {
			orderFloor = floor
			return 
		}
	}

	for floor := 3; floor >= 0; floor--{
		if orderTable[floor][1] == 1 || orderTable[floor][0] == 1 || orderTable[floor][2] == 1  {
			orderFloor = floor
			return
		}
	}
}

func orderBelow(){
	for floor := currentFloor-1; floor>= 0; floor--{
		if orderTable[floor][1] == 1 || orderTable[floor][2] == 1 {
			orderFloor = floor
			return
		}
	}

	for floor := 0; floor <= 3; floor++{
		if orderTable[floor][1] == 1 || orderTable[floor][0] == 1 || orderTable[floor][2] == 1{
			orderFloor = floor
			return 
		}
	}
}

func orderHere(){
	if (orderTable[currentFloor][0] == 1 || orderTable[currentFloor][1] == 1 || orderTable[currentFloor][2] == 1) && currentDir == 0{
		orderFloor = currentFloor
		return
	}
}


func SetPriority(){
	if OrderExist() == 1{
		
		if currentDir == 1 {
			orderAbove()
		} else if currentDir == -1{
			orderBelow()
		} else if currentDir == 0 && lastDir == 1 && currentFloor != 3{
			orderAbove()
		} else if currentDir == 0 && lastDir == -1 && currentFloor != 0{
			orderBelow()
		} else if currentDir == 0 && lastDir == 1 && currentFloor == 3{
			orderBelow()
		} else if currentDir == 0 && lastDir == -1 && currentFloor == 0{
			orderAbove()
		}
		//orderHere()
	} else {
		orderFloor = -1
	}
	
}


func SetLastDir(dir int) {
	lastDir = dir
}

func GetLastDir() int{
	return lastDir
}

func SetCurrentDir(dir int) {
	currentDir = dir
}

func GetCurrentDir() int{
	return currentDir
}

func AcceptOrder(buttonEvent elevio.ButtonEvent){
	UpdateOrderTable(buttonEvent.Floor, int(buttonEvent.Button))
	SetPriority()
	if buttonEvent.Floor == currentFloor && currentDir == 0 {
		orderFloor = buttonEvent.Floor
	}
}

