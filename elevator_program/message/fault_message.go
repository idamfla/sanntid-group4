package fault

type FaultType int

const (
    FAULT_T_LostConn FaultType = iota // Who lost? // send with ip
    FAULT_T_LostMaster // Dont know master, who is master? master not answering
    FAULT_T_ElevatorFailed // commit took too long
    FAULT_T_BroadcastFailedToRespond // someone did not answer
    // FAULT_T_BroadcastErr
    FAULT_T_TaskRunningErr // task completion took to long, suspect motor stop?
  )

type FaultMessage struct {
    ID string // probably the same as addr
    FaultType FaultType
    IsMaster bool
    //Do you know master?
    SelfAddr string // your addr
    Peers []string // who did you talk too
}