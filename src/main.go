package main

import (
	"./Driver/elevio"
	"./fsm"
	"./orders"
	"flag"
	"fmt"
	"time"
	"./network/localip"
	"./network/peers"
	"./synch"
)

const _pollRate = 20 * time.Millisecond
const heartBeatPeriod = 10 * time.Millisecond

var b = []elevio.ButtonType{elevio.BT_HallUp, elevio.BT_HallDown, elevio.BT_Cab}
var peerPort = 20124




func main() {

	// ****ASSIGNING ID TO CLIENT****

	var id string
	flag.StringVar(&id, "id", "", "id of this peer")
	flag.Parse()
	if id == "" {
		localIP, err := localip.LocalIP()
		if err != nil {
			fmt.Println(err)
			localIP = "DISCONNECTED"
		}
		id = fmt.Sprintf(localIP)
	}
	synch.SetID(id)

	// *****************************

	// ****INITIALIZATION****

	numFloors := 4
	elevio.Init("localhost:15657", numFloors)
	fsm.Init()
	synch.Init()
	synch.RetrieveHardBackup()
	fmt.Println("System initialized")

	//synch.InitGlobalOrders()
	//synch.RetrieveGlobalOrders()

	// *********************

	// ****CHANNELS****

	// Peer channels
	nodeUpdateCh := make(chan peers.PeerUpdate)
	nodeTxEnable := make(chan bool)

	// Synch channels
	synch_stateTx := make(chan synch.State)
	synch_stateRx := make(chan synch.State) 
	synch_orderTx := make(chan synch.OrderMsg)
	synch_orderRx := make(chan synch.OrderMsg)
	//synch_backupTx := make(chan synch.State)	<---- (1) UNCOMMENT TO ENABLE TRANSMIT BACKUP OVER NETWORK
	//synch_backupRx := make(chan synch.State)	<---- (1) UNCOMMENT TO ENABLE TRANSMIT BACKUP OVER NETWORK
	synch_ackRx := make(chan bool)
	synch_ackTx := make(chan bool)
	synch_heartBeatTick := make(chan bool)
	synch_hallEvent := make(chan synch.OrderMsg)
	//synch_sendBackup := make(chan bool)	<---- (1) UNCOMMENT TO ENABLE TRANSMIT BACKUP OVER NETWORK
	synch_floorOrdersTx := make(chan synch.FloorOrderEvent)
	synch_floorOrdersRx := make(chan synch.FloorOrderEvent)
	synch_transmitClear := make(chan synch.FloorOrderEvent)

	// Fsm channels
	fsm_stateChan := make(chan fsm.CurrentState)

	// Driver channels
	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	drv_stop := make(chan bool)

	// ***************

	// ****GOROUTINES****

	// Fsm Threads
	go fsm.StateHandler(fsm_stateChan)

	// Driver Threads
	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)

	// Peer Threads
	go peers.Transmitter(peerPort, id, nodeTxEnable)
	go peers.Receiver(peerPort, nodeUpdateCh)


	// Synch Threads
	go synch.TransmitState(synch_heartBeatTick, synch_stateTx, synch_ackRx)
	go synch.ReceiveState(synch_stateRx, synch_ackTx)
	go synch.TransmitOrderMsg(synch_hallEvent, synch_orderTx, synch_ackRx)
	go synch.ReceiveOrderMsg(synch_orderRx, synch_ackTx)
	//go synch.TransmitBackup(synch_sendBackup, synch_backupTx, synch_ackRx) <---- (1) UNCOMMENT TO ENABLE TRANSMIT BACKUP OVER NETWORK
	//go synch.ReceiveBackup(synch_backupRx, synch_ackTx)	<---- (1) UNCOMMENT TO ENABLE TRANSMIT BACKUP OVER NETWORK
	go synch.TransmitButtonClear(synch_floorOrdersTx, synch_transmitClear, synch_ackRx)
	go synch.ReceiveButtonClear(synch_floorOrdersRx, synch_ackTx)

	//******************

	// ****SYSTEM FLOW****

	for {
		select {
		case buttonEvent := <-drv_buttons:
			if buttonEvent.Button == b[0] || buttonEvent.Button == b[1] {
				if synch.SizeOfLiveNodes() > 1 {
					var message = synch.AuctionOrder(buttonEvent)
					synch_hallEvent <- message
				}
			} else if buttonEvent.Button == b[2] {
				elevio.SetButtonLamp(buttonEvent.Button, buttonEvent.Floor, true)
				orders.AcceptOrder(buttonEvent)
			}
			//orders.ClearOrder()
			//orders.SetPriority()

		case floorEvent := <-drv_floors:
			elevio.SetFloorIndicator(floorEvent)
			orders.SetCurrentFloor(floorEvent)
			//synch.CheckAndClearHallUp(floorEvent)

		case fsmState := <-fsm_stateChan:
			if fsmState == fsm.Idle {
				floor := orders.GetCurrentFloor()
				
				floorOrders := orders.GetOrderTable()[floor]
				synch.SetClearOrderMsg(floorOrders, floor, floorOrders[0], floorOrders[1])
				var floorOrderEvent = synch.GetClearOrderMsg()
				fmt.Printf("Clear msg initialized: %+v \n", floorOrderEvent)
				synch_transmitClear <- floorOrderEvent
				//orders.ClearOrder()
				orders.NewClearOrder()
				orders.SetPriority()

			}
		
		case nodeUpdate := <-nodeUpdateCh:
			synch.UpdateLiveNodes(nodeUpdate.Peers)
			fmt.Println("Live nodes: ", nodeUpdate.Peers)
			lostID := synch.IdentifyLostNode(nodeUpdate.Lost)
			fmt.Println("Dead nodes: ", lostID)
			if lostID != "" {
				synch.InheritOrders(lostID)
				//synch.CreateCabBackup(lostID)
			}
			//synch_sendBackup <- true 	<---- (1) UNCOMMENT TO ENABLE TRANSMIT BACKUP OVER NETWORK
			synch.SetGlobalLights()
			
			if synch.SizeOfLiveNodes() == 1{
				orders.ClearAllHalls()
			}

		case <-time.After(heartBeatPeriod):
			synch.UpdateElevatorState()
			synch.CreateHardBackup()
			synch_heartBeatTick <- true
			//synch.PrintAcceptedOrders(synch.GetID()) <---- FOR DEBUGGING
		}
	}

	//*****************
}
