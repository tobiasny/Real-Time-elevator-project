package synch

import (
	"fmt"
	"time"
	"sync"
	"os"
	"io/ioutil"
	"strconv"
	"../orders"
	"../network/bcast"
	"../Driver/elevio"
	"../network/localip"
)

var _mtx sync.Mutex

var statePort = 20100
var ackPort = 20200
var orderPort = 20300
var backupPort = 20302
var buttonClearPort = 20304

const _pollRate = 20 * time.Millisecond

const doorOpenCost int = 3000
const travelCost int = 2000

var b = []elevio.ButtonType{elevio.BT_HallUp, elevio.BT_HallDown, elevio.BT_Cab}
var backupFiles = []string{"backup1.txt","backup2.txt","backup3.txt"}

type OrderMsg struct {
	OrderButton elevio.ButtonEvent
	AuctionWinner string
}

var liveNodes = []string{}

var GlobalOrders = []byte{48,48,10,48,48,10,48,48,10,48,48}
/*
var GlobalOrders = [][]int{
	// Up  ,down
	{  0   ,0 }, // Floor 1
	{  0   ,0 }, // Floor 2
	{  0   ,0 }, // Floor 3
	{  0   ,0 }, // Floor 4
}
*/
type State struct {
	CurrentFloor int
	OrderFloor int
	CurrentDir int
	LastDir int
	IP string
	ID string
	PrivateOrders [][]int
}

type FloorOrderEvent struct{
	FloorOrders []int
	Floor int
	UpOrder int
	DownOrder int
}

var elevatorState State

var stateMap = make(map[string]State)
var cabBackup State

var clearOrderMsg FloorOrderEvent//{FloorOrders: [0,0,0], Floor: 0, UpOrder: 0, DownOrder: 0}

func SetClearOrderMsg(floorOrders []int, floor int, upOrder int, downOrder int){
	clearOrderMsg.FloorOrders = floorOrders
	clearOrderMsg.Floor = floor
	clearOrderMsg.UpOrder = upOrder
	clearOrderMsg.DownOrder = downOrder
}

func GetClearOrderMsg() FloorOrderEvent{
	return clearOrderMsg
}

func SetID(id string){
	elevatorState.ID = id
}

func GetID() string{
	return elevatorState.ID
}

func SizeOfLiveNodes() int{
	return len(liveNodes)
}

func Init(){
	_mtx = sync.Mutex{}
	thisIP,_ := localip.LocalIP()
	elevatorState.IP = thisIP
	elevatorState.CurrentFloor = orders.GetCurrentFloor()
	elevatorState.OrderFloor = orders.GetOrderFloor()
	elevatorState.CurrentDir = orders.GetCurrentDir()
	elevatorState.LastDir = orders.GetLastDir()
	elevatorState.PrivateOrders = orders.GetOrderTable()
	_mtx.Lock()
	stateMap[elevatorState.ID] = elevatorState
	_mtx.Unlock()
}

func UpdateElevatorState(){
	elevatorState.CurrentFloor = orders.GetCurrentFloor()
	elevatorState.OrderFloor = orders.GetOrderFloor()
	elevatorState.CurrentDir = orders.GetCurrentDir()
	elevatorState.LastDir = orders.GetLastDir()
	elevatorState.PrivateOrders = orders.GetOrderTable()
	_mtx.Lock()
	stateMap[elevatorState.ID] = elevatorState
	_mtx.Unlock()
}

func UpdateLiveNodes(nodes []string){
	liveNodes = nodes
}
/*
func UpdateGlobalOrders(floor int, button elevio.ButtonType){
	GlobalOrders[floor][int(button)] = 1
}

func ClearGlobalOrder(floor int){
	for button := 0; button < 2; button++{
		GlobalOrders[floor][button] = 0
	}
}
*/

func PrintAllStates(){
	for _,element := range liveNodes{
		PrintElevatorState(element)
	}
}

func PrintElevatorState(thisID string){
	fmt.Println("STATE FOR ID: ", stateMap[thisID].ID, " AT IP ", stateMap[thisID].IP)
	fmt.Println("Current floor: ",stateMap[thisID].CurrentFloor)
	fmt.Println("Current dir: ",stateMap[thisID].CurrentDir)
	fmt.Println("Order floor: ",stateMap[thisID].OrderFloor)
	fmt.Println("Last dir: ",stateMap[thisID].LastDir)
	fmt.Println("------------------------------------")
}

func PrintAllAcceptedOrders(){
	for _,node := range liveNodes{
		PrintAcceptedOrders(node)
	}
}

func PrintAcceptedOrders(thisID string){
	myOrders := stateMap[thisID].PrivateOrders
	fmt.Println("Accepted orders for ID ",thisID)
	fmt.Println("Floor:     0       1       2       3")
	fmt.Println("------------------------------------")
	fmt.Println("Hall Up   ", myOrders[0][0],"     ",myOrders[1][0], "     ",myOrders[2][0], "     ",myOrders[3][0])
	fmt.Println("Hall Down ", myOrders[0][1],"     ",myOrders[1][1], "     ",myOrders[2][1], "     ",myOrders[3][1])
	fmt.Println("Cab       ", myOrders[0][2],"     ",myOrders[1][2], "     ",myOrders[2][2], "     ",myOrders[3][2])
}



func abs(number int) int{
	if number >= 0 {
		return number
	} else {
		return number*(-1)
	}
}




func simAbove(table [][]int, currentFloor int, orderFloor int) int{
	for floor := currentFloor+1; floor < 4; floor++{
		if table[floor][0] == 1 || table[floor][2] == 1 {
			return floor
		}
	}

	for floor := 3; floor >= 0; floor--{
		if table[floor][1] == 1 || table[floor][0] == 1 || table[floor][2] == 1  {
			return floor 
		}
	}

	return orderFloor
}

func simBelow(table [][]int, currentFloor int, orderFloor int) int{
	for floor := currentFloor-1; floor>= 0; floor--{
		if table[floor][1] == 1 || table[floor][2] == 1 {
			return floor
		}
	}

	for floor := 0; floor <= 3; floor++{
		if table[floor][1] == 1 || table[floor][0] == 1 || table[floor][2] == 1{
			return floor
		}
	}

	return orderFloor
}

func SimPriority(currentDir int, lastDir int, currentFloor int, table [][]int, orderFloor int) int{
	var floor int
	if currentDir == 1 {
		floor = simAbove(table, currentFloor,orderFloor)
		return floor
	} else if currentDir == -1{
		floor = simBelow(table, currentFloor,orderFloor)
		return floor
	} else if currentDir == 0 && lastDir == 1 && currentFloor != 3{
		floor = simAbove(table, currentFloor,orderFloor)
		return floor
	} else if currentDir == 0 && lastDir == -1 && currentFloor != 0{
		floor = simBelow(table, currentFloor,orderFloor)
		return floor
	} else if currentDir == 0 && lastDir == 1 && currentFloor == 3{
		floor = simBelow(table, currentFloor,orderFloor)
		return floor
	} else if currentDir == 0 && lastDir == -1 && currentFloor == 0{
		floor = simAbove(table, currentFloor,orderFloor)
		return floor
	}
	
	return orderFloor
}

func isEmpty(table [][]int) bool{
	for floor := 0; floor < 4; floor++ {
		for button := 0; button <3; button++{
			if  table[floor][button] == 1{
				return false
			}
		}
	}
	return true
}




func SimulateOrderExecution(order elevio.ButtonEvent, table [][]int, currentFloor int, currentDir int, lastDir int) []int{
	dummyTable := table
	dummyTable[order.Floor][int(order.Button)] = 1
	var orderQueue []int

	nextFloor := SimPriority(currentDir,lastDir,currentFloor,dummyTable,order.Floor)
	orderQueue = append(orderQueue,nextFloor)
	for button := 0; button < 3; button++{
		dummyTable[nextFloor][button] = 0
	}
	var dir int 
	if (currentFloor - nextFloor) > 0{
		dir = -1
	} else {
		dir = 1
	}
	currentDir = 0
	lastDir = dir

	for {
		if isEmpty(dummyTable) == true {
			break
		} else {
			nextFloor = SimPriority(currentDir,lastDir,currentFloor,dummyTable,nextFloor)
			orderQueue = append(orderQueue,nextFloor)
			if (currentFloor - nextFloor) > 0{
				dir = -1
			} else {
				dir = 1
			}
			currentDir = 0
			lastDir = dir
			for button := 0; button < 3; button++{
				dummyTable[nextFloor][button] = 0
			}

		}
	}
	//fmt.Println("OrderQueue: ",orderQueue)
	return orderQueue
}


func DetermineCost(elevator State, order elevio.ButtonEvent) int{
	cost := 0
	orderQueue := SimulateOrderExecution(order,elevator.PrivateOrders,elevator.CurrentFloor,elevator.CurrentDir,elevator.LastDir)
	cost += abs(elevator.CurrentFloor - order.Floor)*travelCost
	if orderQueue[0] == order.Floor{	
		return cost
	}
	cost += doorOpenCost
	for index := 1; index < len(orderQueue); index++ {
		cost += abs(orderQueue[index]-orderQueue[index-1])
		if orderQueue[index] == order.Floor {
			break
		}
		cost += doorOpenCost
	}

	//fmt.Println("Cost of ID ", elevator.ID, " is: ", cost)
	return cost
}

/*
func DetermineCost(elevator State, order elevio.ButtonEvent) int {
	cost := abs(order.Floor - elevator.CurrentFloor)*abs(order.Floor - elevator.CurrentFloor)
	fmt.Println("Cost of ID ", elevator.ID, " is: ", cost)
	return cost
}
*/


func AuctionOrder(order elevio.ButtonEvent) OrderMsg{
	var currentHolder State
	var bidder State
	iter := 1

	for _,node := range liveNodes{
		if iter == 1{
			currentHolder = stateMap[node]
		} else {
			bidder = stateMap[node]
			if DetermineCost(bidder,order) < DetermineCost(currentHolder,order) {
				currentHolder = bidder
			}
		}
		iter++
	}


	fmt.Println("THE WINNER OF THE AUCTION: ", currentHolder.ID)
	var message = OrderMsg{OrderButton: order, AuctionWinner: currentHolder.ID}
	return message


}


func TransmitState(heartBeatTick chan bool, stateTx chan State, ackRx chan bool){
	go bcast.Transmitter(statePort,stateTx)
	go bcast.Receiver(ackPort, ackRx)
	for {
		select{
		case <- heartBeatTick:
			stateTx <- elevatorState
			//fmt.Println("State sent")
		case <- ackRx:
			//fmt.Println("Ack received")
		}
	}
}

func ReceiveState(stateRx chan State, ackTx chan bool){
	go bcast.Transmitter(ackPort, ackTx)
	go bcast.Receiver(statePort, stateRx)
	for {
		select{
		case state:= <-stateRx:
			//fmt.Println("State received")
			ackTx <- true
			_mtx.Lock()
			stateMap[state.ID] = state
			_mtx.Unlock()
		}
	}
}

func TransmitOrderMsg(hallEvent chan OrderMsg, orderTx chan OrderMsg, ackRx chan bool){
	go bcast.Transmitter(orderPort, orderTx)
	go bcast.Receiver(ackPort, ackRx)
	for {
		select {
		case message := <- hallEvent:
			for i := 0; i < 50; i++{
				orderTx <- message
				time.Sleep(2 * time.Millisecond)
			}
		case <- ackRx:

		}
	}

}

func ReceiveOrderMsg(orderRx chan OrderMsg, ackTx chan bool){
	go bcast.Transmitter(ackPort,ackTx)
	go bcast.Receiver(orderPort, orderRx)

	for {
		select {
		case message := <-orderRx:
			elevio.SetButtonLamp(message.OrderButton.Button, message.OrderButton.Floor, true)
			//RetrieveGlobalOrders()
			//WriteToGlobalOrders(message.OrderButton.Floor, int(message.OrderButton.Button),true)
			if message.AuctionWinner == elevatorState.ID{
				// This node should accept order
				orders.AcceptOrder(message.OrderButton)
			}
		}
	}
}

func TransmitBackup(nodeUpdate chan bool, backupTx chan State, ackRx chan bool){
	go bcast.Transmitter(backupPort,backupTx)
	go bcast.Receiver(ackPort, ackRx)
	for {
		select{
		case <- nodeUpdate:
			for i := 0; i < 50; i++{
				backupTx <- cabBackup
				time.Sleep(2 * time.Millisecond)
			}
			//fmt.Println("State sent")
		case <- ackRx:
			//fmt.Println("Ack received")
		}
	}
}

func ReceiveBackup(backupRx chan State, ackTx chan bool){
	go bcast.Transmitter(ackPort, ackTx)
	go bcast.Receiver(backupPort, backupRx)
	for {
		select{
		case backup:= <-backupRx:
			ackTx <- true
			if  backup.ID == elevatorState.ID{
				RetreiveCabOrders(backup)
			}
		}
	}
}

func TransmitButtonClear(floorOrdersTx chan FloorOrderEvent,transmitClear chan FloorOrderEvent, ackRx chan bool){
	go bcast.Transmitter(buttonClearPort,floorOrdersTx)
	go bcast.Receiver(ackPort, ackRx)
	for {
		select{
		case clearMsg := <- transmitClear:
			//fmt.Printf("Clear msg sent: %+v \n", clearMsg)
			//fmt.Println("Clear msg sent: ",clearMsg.UpOrder,clearMsg.DownOrder," at floor ", clearMsg.Floor)
			for i := 0; i < 50; i++{
				floorOrdersTx <- clearMsg
				time.Sleep(2 * time.Millisecond)
			}
		case <- ackRx:
			//fmt.Println("Ack received")
		}
	}
}

func ReceiveButtonClear(floorOrdersRx chan FloorOrderEvent, ackTx chan bool){
	go bcast.Transmitter(ackPort, ackTx)
	go bcast.Receiver(buttonClearPort, floorOrdersRx)
	for {
		select{
		case floorOrderEvent:= <-floorOrdersRx:
			//fmt.Println("Clear msg received")
			ackTx <- true
			//fmt.Printf("Clear msg received: %+v \n", floorOrderEvent)
			//fmt.Println("Floor orders: ",floorOrderEvent.UpOrder, floorOrderEvent.DownOrder," at floor ", floorOrderEvent.Floor)
			if floorOrderEvent.UpOrder == 1 {
				elevio.SetButtonLamp(b[0],floorOrderEvent.Floor,false)
				//RetrieveGlobalOrders()
				//WriteToGlobalOrders(floorOrderEvent.Floor,0,false)
			}

			if floorOrderEvent.DownOrder == 1 {
				elevio.SetButtonLamp(b[1],floorOrderEvent.Floor,false)
				//RetrieveGlobalOrders()
				//WriteToGlobalOrders(floorOrderEvent.Floor,1,false)
			}

			/*
			for button := 0; button < 2; button++ {

				if floorOrderEvent.FloorOrders[button] == 1 {
					//fmt.Println("Shutting of light")
					elevio.SetButtonLamp(b[button],floorOrderEvent.Floor,false)
				}
			}
			*/
		}
	}
}

func IdentifyLostNode(deadNodes []string) string{
	if len(deadNodes) >= 1{
		return deadNodes[0]
	} else {
		return ""
	}
}

// Inherit hall-orders
func InheritOrders(lostID string){ 
	lostOrders := stateMap[lostID].PrivateOrders
	//descendant := liveNodes[0] // The first node in liveNodes inherits orders
	if /*descendant == elevatorState.ID &&*/len(lostOrders) == 4 {
		for floor := 0; floor < 4; floor++ {
			for button := 0; button < 2; button++ {
				if lostOrders[floor][button] == 1 {
					var buttonEvent = elevio.ButtonEvent{Floor: floor, Button: b[button]}
					message := AuctionOrder(buttonEvent)
					if message.AuctionWinner == elevatorState.ID {
						elevio.SetButtonLamp(b[button], floor, true)
						orders.AcceptOrder(buttonEvent)

					}
				}
			}
		}
	}
}

func RetreiveCabOrders(backup State){
	backupOrders := backup.PrivateOrders
	for floor := 3; floor >= 0; floor--{
		if backupOrders[floor][2] == 1{
			orders.UpdateOrderTable(floor,2)
			orders.SetPriority()
			elevio.SetButtonLamp(b[2],floor,true)
		}
	}
}

func CreateCabBackup(lostID string){
	cabBackup = stateMap[lostID]
}

func GetBackupID() string{
	return cabBackup.ID
}


func CreateHardBackup(){
	ourID, _ := strconv.ParseInt(elevatorState.ID,10,0)
	inFile := backupFiles[ourID-1]
	file, err := os.Create(inFile)
    if err != nil {
        return
    }
    defer file.Close()

    myOrders := elevatorState.PrivateOrders
    var UTFArray = []int{0,0,0,0}
    for floor := 0; floor <= 3; floor++{
    	if myOrders[floor][2] == 1{
    		UTFArray[floor] = 49
    	} else {
    		UTFArray[floor] = 48    	}
    }

    stringToFile := string(UTFArray[0]) + string(UTFArray[1]) + string(UTFArray[2]) + string(UTFArray[3])
    //fmt.Println("UTF array: ", UTFArray)
    //fmt.Println("Writing following string to file: ", stringToFile)

    file.WriteString(stringToFile)
}

func RetrieveHardBackup(){
	ourID, _ := strconv.ParseInt(elevatorState.ID,10,0)
	inFile := backupFiles[ourID-1]
	fmt.Println("Accessing file: ",inFile)
	data, err := ioutil.ReadFile(inFile)
    if err != nil {
        return
    }

    // Data is on format []int -- for example data = [48 48 49 49]

    // Data contains 0s or 1s if there exists an order or not
    // The txt-file contains characters of UTF-8
    // 0s = 48 ----------- 1s = 49

    fmt.Println("Retrieved backup:",data)

	for floor := 0; floor <= 3; floor++{
		if data[floor] == 49{
			orders.UpdateOrderTable(floor,2)
			orders.SetPriority()
			elevio.SetButtonLamp(b[2],floor,true)
			fmt.Println("Order in floor ", floor)
		}
	}
	
}

func WriteToGlobalOrders(floor int, button int, set bool){
	//RetrieveGlobalOrders()
	_mtx = sync.Mutex{}
	_mtx.Lock()
	file, err := os.Create("globalorders.txt")
	if err != nil{
		return
	}
	defer file.Close()
	
	if set == true{
		GlobalOrders[2*floor + button] = 49
	} else {
		GlobalOrders[2*floor + button] = 48
	}

	stringToFile := ""
	for i := 0; i < len(GlobalOrders); i++{
		stringToFile = stringToFile + string(GlobalOrders[i])
	}

	fmt.Println("stringToFile: ", stringToFile)

	file.WriteString(stringToFile)
	fmt.Println("Write func using mtx")
	_mtx.Unlock()
	
}

func InitGlobalOrders(){
	_mtx = sync.Mutex{}
	_mtx.Lock()
	file, err := os.Create("globalorders.txt")
	if err != nil{
		return
	}
	defer file.Close()
	fmt.Println("Init func using mtx")
	file.WriteString("00000000")
	_mtx.Unlock()
}

func RetrieveGlobalOrders(){
	_mtx = sync.Mutex{}
	_mtx.Lock()
	data, err := ioutil.ReadFile("globalorders.txt")
	if err != nil{
		return
	}
	fmt.Println("Global data from file: ",data)

	GlobalOrders = data
	fmt.Println("Read func using mtx")
	_mtx.Unlock()
}

/*

func SetGlobalLights(){
	//RetrieveGlobalOrders()
	floor := 0
	iter := 0

	for index,element := range GlobalOrders{
		if element == 48  && index%2 == 0{
			elevio.SetButtonLamp(b[0],floor,false)
		}
		if element == 48 && index%2 == 1 {
			elevio.SetButtonLamp(b[1],floor,false)
		}
		if element == 49 && index%2 == 0 {
			elevio.SetButtonLamp(b[0],floor,true)
		} 
		if element == 49 && index%2 == 1{
			elevio.SetButtonLamp(b[1],floor,true)
		}
		iter++ 
		if iter>1 {
			iter = 0
			floor++
		}

	}
}
*/

func SetGlobalLights(){
	for _,node := range liveNodes{
		privateOrderList := stateMap[node].PrivateOrders
		if len(privateOrderList) != 4 {
			continue
		}
		//fmt.Println("Private orders for node ", node, " : ", privateOrderList)
		for floor := 0; floor < 4; floor++ {
			for button := 0; button < 2; button++{
				if privateOrderList[floor][button] == 1{
					elevio.SetButtonLamp(b[button],floor,true)
				} else {
					//elevio.SetButtonLamp(b[button],floor,false)
				}
			}
		}
	}
}