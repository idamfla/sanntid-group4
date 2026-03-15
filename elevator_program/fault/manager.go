package fault

import (
	"fmt"
	"time"
)

type Manager struct{


    cfg Config
    role Role //TODO: do we need this?
    id int

    startedAt time.Time

    lastSeenMaster time.Time
    lastSeenPeer map[int]time.Time

    lastFloorEvent time.Time
    motorRunning bool

    online bool
    faulty bool

    OnBecomeMaster    func()
    OnMasterSuspected func(reason string)
    OnPeerDead        func(peerID int)
    OnGoOnline        func()
    OnGoOffline       func()
    OnMotorFault      func(reason string)
    OnNetworkFault    func(reason string)


}

func NewFaultManager(id int, cfg Config)*Manager{
    return&Manager{
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


func(fm*Manager) SeenMaster(){
    fm.lastSeenMaster= time.Now()
    if !fm.online{
        fm.online= true
        if fm.onGoOnline!= nil{
            fm.onGoOnline()
         }

 }
}

func(fm*Manager) SeenPeer(peerID int){
    fm.lastSeenPeer[peerID]= time.Now()
}

//TODO: FIX place
func(fm*Manager) FloorEvent(){
    fm.lastFloorEvent= time.Now()

}

func(fm*Manager) SetMotorRunning(running bool){
	fm.motorRunning= running
	if running {
		fm.lastFloorEvent = time.Now()
		fm.faulty = false
	}
}

func(fm*Manager) SetRoleMaster(){
    fm.role= RoleMaster
}

func(fm*Manager) SetRoleSlave(){
    fm.role= RoleSlave
 }


func (fm *Manager) checkMasterTimeout() {

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

        if fm.online {
            if fm.onMasterSuspected != nil {
                fm.onMasterSuspected("master timeout")
            }

            fm.online = false

           if fm.onGoOffline != nil {
                fm.onGoOffline()
            }

            if fm.onNetworkFault != nil {
                fm.onNetworkFault("master timeout")
            }
        }
    }
}




 func (fm *Manager) checkPeerTimeout() {
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

func (fm *Manager) checkMotorTimeout() {

    if !fm.motorRunning {
        return
    }
     if time.Since(fm.lastFloorEvent)> fm.cfg.MotorTimeout{

        if!fm.faulty{
            fm.faulty= true

            if fm.onMotorFault != nil{
                fm.onMotorFault("motor watchdog timeout")
            }
        }
    }
}

func (fm *Manager) Run() {
	ticker := time.NewTicker(fm.cfg.Tick)
	defer ticker.Stop()

	for range ticker.C {
		fm.checkMasterTimeout()
		fm.checkPeerTimeout()
		fm.checkMotorTimeout()
	}
}