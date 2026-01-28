# Elevator Program

- [Elevator Program](#elevator-program)
  - [Run Elevator Simulation](#run-elevator-simulation)
    - [Default buttons](#default-buttons)
  - [Run the elevator at the lab](#run-the-elevator-at-the-lab)


## Run Elevator Simulation

1. Start by opening the terminal and navigate to `../elevator_program`
2. Then run

```bash
chmod +x SimElevatorServer
```

3. Lastly run

```bash
./SimElevatorServer
```

and see the simulator appare.

Now, to run the program just open a new terminal window, navigate to `../elevator_program` and run
```bash
go run main.go
```

### Default buttons
Default keyboard controls

* Up: `qwertyui`
* Down: `sdfghjkl`
* Cab: `zxcvbnm`,.
* Stop: `p`
* Obstruction: `-`
* Motor manual override: Down: `7`, Stop: `8`, Up: `9`
* Move elevator back in bounds (away from the end stop switches): `0`

## Run the elevator at the lab

To make the elevator run at the lab:

1. Check if the everything is set up correctly
   - Turn _on_ the PC
   - Make sure everything is up to date
   - Toggle to `pc` and `obstruction` on the elevator panel
2. On the PC, open the terminal and go to `/elevator_program`
3. Run
   ```bash
   chmod +x elevatorserver
   ./elevatorserver
   ```
4. Open a new elevator, make sure you are in the correct folder (`/elevator_program`), and run
   ```bash
   go run main.go
   ```
5. Now the elevator should run
