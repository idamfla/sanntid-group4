package elevator

//nb
// run as go routine
func (e* Elevator) fault_loop() {
    for faultMsg := range <-e.faultMsg {
        switch faultMsg.FaultType {
        case FAULT_T_LostConn:
            // investigate
            // faultTolerence.killServer(e.srv)
            // e.NewServer()
            // e.StartServer()
            case FAULT_T_LostMaster:
          case FAULT_T_ElevatorFailed:
       case FAULT_T_BroadcastFailedToRespond:
       }
   }
}
             // manger
func lostconn(srv*Server){srv.Close}