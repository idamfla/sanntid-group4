package main

import (
    "flag"
    "fmt"
	"elevator_program/elevator"
	"elevator_program/elevio"

)


// Remove this later if we see that the communication works
func testElevator() {
	var e elevator.Elevator
	// fmt.Println(e)

	id := 1
	numFloors := 4
	initFloor := 3 // NB! in the code the elevator floors are 0-index, on the controller it is not
	ip_address := "localhost"
	port := "15657"

	// "localhost:15657"
	elevio.Init(ip_address+":"+port, numFloors)

	e.InitElevator(id, numFloors, initFloor)
	e.RunElevatorProgram()
	/*
		TODO, bug - when cab to floor 2, then cab to floor 1, if floor 3 is pressed after reaching floor 2, elevator will go up to floor 3
	*/
	select {}
}




func runElevator(id int, ip string, port string, floors int, initFloor int) {

    addr:= fmt.Sprintf("%s:%s", *ip, *port)
    elevio.Init(addr, floors)

    var e elevator.Elevator
    e.InitElevator(id,floors, initFloor)
    e.RunElevatorProgram

    select {}

    }


func main() {

    id := flag.Int("id", 1, "elevator id") //Takes in id and port, makes it possible to run different elevators
    port := flag.String("port", "15657", "elevio port")//
    floors := flag.Int("floors", 4, "number of floors")
    initFloor := flag.int(initFloor, 3, "initial floor (0-index)")
    ip:= flag.String("ip", "localhost", "ip/host to elevio")

    flag.Parse()


    runElevator(*id, *ip, *port, *floors, *initFloor)



	//testElevator()
}
