package fault

import(
 "time"
)

type Role int

const(
    RoleSlave Role= iota
    RoleMaster

)

type Config struct{
    StartupGrace  time.Duration
    MasterTimeout time.Duration
    PeerTimeout   time.Duration
    MotorTimeout  time.Duration
    Tick          time.Duration
}

