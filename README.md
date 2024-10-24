# Distributed Systems: Global Snapshot Algorithm

This project implements an algorithm to solve the **Distributed Global Snapshot** problem as part of the Distributed Systems course during my second semester of 2024 in Software Engineering.

The algorithm simulates participants exchanging candies, where one participant holds a hidden piece of garlic wrapped like candy. The goal of the algorithm is to track who holds the garlic at any point in time using snapshots, capturing the system's state and channel conditions.

## Context

The problem is based on a distributed system where participants exchange messages concurrently. At certain points, a participant triggers a **snapshot** to capture the current state of the system and record it to a file. The system is designed to:

- Track the state of messages sent and received by each participant.
- Monitor which participant holds the garlic.
- Capture the status of each communication channel.
- Save results to files for each snapshot.

## Input Format

The program reads an input file, `acciones.txt`, which describes the actions participants take. The file format is as follows:

1. The first line contains `n`, `m`, `l`:
    - `n`: number of participants.
    - `m`: number of actions.
    - `l`: ID of the participant initially holding the garlic.

2. The following `m` lines describe actions in the format:

    ```txt
    participant_id:action:value
    ```

    - `SEND`: A participant sends a candy (or garlic if they hold it) to another participant.
    - `RECEIVE`: A participant receives a candy or snapshot marker from another participant.
    - `WAIT`: A participant waits for a given amount of time.
    - `SNAPSHOT`: A participant triggers a snapshot, sending markers to all other participants.

### Example Input

```txt
1 3 9 0 2 0:SEND:1 3 0:SNAPSHOT:0 4 0:RECEIVE:1 ...
```

## How to Run the Program

1. **Compilation**:
    First, build the project using Go:

    ```bash
    go build -o main main.go
    ```

2. **Execution**:
    To run the program, provide the path to the `acciones.txt` file as an argument:

    ```bash
    ./main path/to/acciones.txt
    ```

    The program will simulate the actions described in the input file and generate output files named `snapshot_n.txt`, where `n` corresponds to the snapshot number.

3. **Snapshot Output**:
    Each snapshot file contains the recorded state for each participant, including:
    - `stateatRecord`: Messages sent by each participant.
    - `messageatRecord`: Messages received by each participant.
    - `ajoatRecord`: Indicates if the participant has the garlic.
    - `channelatRecord`: Message count in input channels.
    - `ajoatMarker`: Indicates if the garlic is in the channel.

### Example Snapshot Output

For a snapshot with 3 participants:

```txt
0: 
stateatRecord: [0 3 1]
messageatRecord: [0 1 0]
ajoatRecord: false
channelatRecord: [0 0 0]
ajoatMarker: [false -1]

1:
stateatRecord: [0 0 0]
messageatRecord: [2 0 0]
ajoatRecord: true
channelatRecord: [0 0 0]
ajoatMarker: [false -1] ...
```

## Considerations

- The system ensures consistency by assuming enough **RECEIVE** actions are present to process both candies and snapshot markers.
- The input file `acciones.txt` must follow the required format for the algorithm to run correctly.

## Tools & Technologies

- **Go**: The task is implemented in Go, leveraging its concurrency model to handle multiple participants' actions simultaneously.
- **Distributed Systems Concepts**: The task applies principles such as message passing, state recording, and global snapshots to solve distributed system problems.
