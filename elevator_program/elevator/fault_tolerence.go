package elevator

/*
when not hearing from anybody else:
how to know if the elevator in question is the issue or the one that it is communicating with

how to make the elevator kill itself and then start anew


how to force another elevator to restart
*/

import(
 "fmt"
 "os"
 "os/exec"
 "time"
)

type Role int

const(
    RoleSlave Role= iota
    RoleMaster

)


type FaultConfig struct{
    StartupGrace  time.Duration
    MasterTimeout time.Duration
    PeerTimeout time.Duration
    MotorTimeout time.Duration
    Tick         time.Duration
}

type FaultManager struct{

    cfg FaultConfig
    role Role //TODO: do we need this?
    id int

    startedAt time.Time

    lastSeenMaster time.Time
    lastSeenPeer map[int]time.Time

    lastFloorEvent time.Time
    motorRunning bool

    online bool
    faulty bool

    onBecomeMaster func()
    onPeerDead     func(peerID int)
    onGoOnline     func()
    onGoOffline    func()
    onFaulty       func(reason string)



}

func NewFaultManager(id int, cfg FaultConfig)*FaultManager{
    return&FaultManager{
        cfg:            cfg,
        id:             id,
        role:           RoleSlave,
        online:         true,

        startedAt:      time.Now(),
        lastSeenPeer:   make(map[int]time.Time),
        lastFloorEvent: time.Now(),
        lastSeenMaster: time.Now(),

 }
}


func(fm*FaultManager) SeenMaster(){
    fm.lastSeenMaster= time.Now()
    if !fm.online{
        fm.online= true
        if fm.onGoOnline!= nil{
            fm.onGoOnline()
         }

 }
}

func(fm*FaultManager) SeenPeer(peerID int){
    fm.lastSeenPeer[peerID]= time.Now()
}

//TODO: FIX place
func(fm*FaultManager) FloorEvent(){
    fm.lastFloorEvent= time.Now()

}

func(fm*FaultManager) SetMotorRunning(running bool){
	fm.motorRunning= running
}

func(fm*FaultManager) SetRoleMaster(){
    fm.role= RoleMaster
}

func(fm*FaultManager) SetRoleSlave(){
    fm.role= RoleSlave
 }


func (fm *FaultManager) checkMasterTimeout() {

    if time.Since(fm.startedAt) < fm.cfg.StartupGrace {
            return
        }

    if fm.role != RoleSlave {
        return
    }

    if fm.lastSeenMaster.IsZero(){
        return
    }

    if time.Since(fm.lastSeenMaster) > fm.cfg.MasterTimeout{
        fmt.Println("Master timeout detected")

        if fm.online{
            fm.online= false

            if fm.onGoOffline!= nil{
                fm.onGoOffline()
            }
        }
    }
}




 func (fm *FaultManager) checkPeerTimeout() {
    if fm.role != RoleMaster{
        return
    }

    for peerID, ts:= range fm.lastSeenPeer{
        if (time.Since(ts) > fm.cfg.PeerTimeout){

            fmt.Println("Peer timeout:", peerID)
            delete(fm.lastSeenPeer, peerID)

            if fm.onPeerDead!= nil{
                fm.onPeerDead(peerID)
            }
        }
    }
}

func (fm *FaultManager) checkMotorTimeout() {

    if !fm.motorRunning {
        return
    }
     if time.Since(fm.lastFloorEvent)> fm.cfg.MotorTimeout{

        if!fm.faulty{
            fm.faulty= true

            if fm.onFaulty!= nil{
                fm.onFaulty("motor watchdog timeout")
            }
        }
    }
}


func restartSelf() {
    exe, err := os.Executable()
    if err != nil {
        os.Exit(1)
    }

    cmd := exec.Command(exe, os.Args[1:]...)
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr

    _ = cmd.Start()
    os.Exit(0)
}


func (fm *FaultManager) Run() {
    ticker := time.NewTicker(fm.cfg.Tick)
    defer ticker.Stop()

    for range ticker.C {
        fm.checkMasterTimeout()
        fm.checkPeerTimeout()
        fm.checkMotorTimeout()
    }
}