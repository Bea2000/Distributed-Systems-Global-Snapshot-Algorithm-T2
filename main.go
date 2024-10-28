package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type AjoAtMarker struct {
	boolValue bool
	intValue  int
}

type LastMarker struct {
	snapshotId    int
	participantId int
	count         int
}

type Participant struct {
	id              int
	ajo             bool
	stateAtRecord   []int
	messageAtRecord []int
	ajoAtRecord     bool
	channelAtRecord []int
	ajoAtMarker     AjoAtMarker
	lastSnapshotId  int
	lastReceived    []int
	lastSend        []int
	lastMarker      []LastMarker
	lastMessageId   int
}

type Action string

const (
	Send     Action = "SEND"
	Receive  Action = "RECEIVE"
	Wait     Action = "WAIT"
	Snapshot Action = "SNAPSHOT"
	Marker   Action = "MARKER"
)

type Message struct {
	from       int
	action     Action
	to         int
	ajo        bool
	snapshotId int
}

var messageChannels = make(map[int]map[int]chan Message)

func createChannels(participants []Participant) {
	for i := range participants {
		messageChannels[i] = make(map[int]chan Message)
		for j := range participants {
			if i != j {
				messageChannels[i][j] = make(chan Message, 50)
			}
		}
	}
}

func validateAction(a Action) error {
	switch a {
	case Send, Receive, Wait, Snapshot:
		return nil
	default:
		return errors.New("acción no válida")
	}
}

func NewMessage(from int, action Action, to int) (*Message, error) {
	if err := validateAction(action); err != nil {
		return nil, err
	}
	return &Message{from: from, action: action, to: to, ajo: false, snapshotId: 0}, nil
}

func NewParticipant(id int, n_participants int) Participant {
	return Participant{
		id:              id,
		stateAtRecord:   make([]int, n_participants),
		messageAtRecord: make([]int, n_participants),
		ajoAtRecord:     false,
		channelAtRecord: make([]int, n_participants),
		ajoAtMarker:     AjoAtMarker{},
		lastSnapshotId:  0,
		lastReceived:    make([]int, n_participants),
		lastSend:        make([]int, n_participants),
		lastMarker:      []LastMarker{},
		ajo:             false,
		lastMessageId:   0,
	}
}

func readFile(file_path string) ([]Participant, []Message) {
	file, err := os.Open(file_path)
	if err != nil {
		fmt.Println("error", err)
		return nil, nil
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Scan()
	firstLine := scanner.Text()
	parts := strings.Split(firstLine, ",")
	n_participants, _ := strconv.Atoi(parts[0])
	m_actions, _ := strconv.Atoi(parts[1])
	ajo_participant, _ := strconv.Atoi(parts[2])

	var participants = make([]Participant, n_participants)
	for i := 0; i < n_participants; i++ {
		participants[i] = NewParticipant(i, n_participants)
		if i == ajo_participant {
			participants[i].ajo = true
		}
	}

	var messages = make([]Message, m_actions)
	for i := 0; i < m_actions; i++ {
		scanner.Scan()
		line := scanner.Text()
		parts := strings.Split(line, ":")
		from, _ := strconv.Atoi(parts[0])
		action := Action(parts[1])
		to, _ := strconv.Atoi(parts[2])
		message, _ := NewMessage(from, action, to)
		messages[i] = *message
	}

	createChannels(participants)

	return participants, messages
}

func (participant *Participant) sendMessage(to int) {
	var ajo bool
	if participant.ajo {
		ajo = true
	} else {
		ajo = false
	}
	messageChannels[participant.id][to] <- Message{from: participant.id, action: Send, to: to, ajo: ajo, snapshotId: 0}
	fmt.Println(participant.id, "SENT MESSAGE to", to)
	participant.lastSend[to] = participant.lastSend[to] + 1
}

func markerExists(participant *Participant, marker LastMarker) int {
	for i, participant_marker := range participant.lastMarker {
		if participant_marker.participantId == marker.participantId && participant_marker.snapshotId == marker.snapshotId {
			return i
		}
	}
	return -1
}

func (participant *Participant) receiveMessage(from int, participants []Participant) {
	message := <-messageChannels[from][participant.id]
	if message.action == Marker {
		fmt.Println(participant.id, "RECEIVED MARKER from", message.from)
		marker := LastMarker{message.snapshotId, message.from, 0}
		marker_index := markerExists(participant, marker)
		if marker_index == -1 {
			participant.lastMarker = append(participant.lastMarker, marker)
			participant.stateAtRecord = participant.lastSend
			participant.messageAtRecord = participant.lastReceived
			participant.ajoAtRecord = participant.ajo
			for i := range participants {
				if i != participant.id {
					messageChannels[participant.id][i] <- Message{from: message.from, action: Marker, to: i, ajo: false, snapshotId: message.snapshotId}
					fmt.Println(participant.id, "SENT MARKER to", i)
				}
			}
		} else {
			participant.lastMarker[marker_index].count = participant.lastMarker[marker_index].count + 1
			if message.ajo {
				participant.ajo = true
				participant.ajoAtMarker = AjoAtMarker{boolValue: true, intValue: participant.id}
			}
			if marker.participantId == participant.id {
				if participant.lastMarker[marker_index].count == len(participants)-1 {
					registerSnapshot(*participant, participants, marker.snapshotId)
				}
			}
		}
		return
	}
	participant.lastReceived[from] = participant.lastReceived[from] + 1
	if message.ajo {
		participant.ajo = true
	}
	if len(participant.lastMarker) > 0 {
		participant.channelAtRecord[from] = participant.channelAtRecord[from] + 1
	}
	fmt.Println(participant.id, "RECEIVED MESSAGE from", from)
}

func (participant *Participant) wait(seconds int) {
	fmt.Println(participant.id, "WAIT for", seconds, "seconds")
	time.Sleep(time.Duration(seconds) * time.Second)
}

func (participant *Participant) snapshot(participants []Participant) {
	participant.stateAtRecord = participant.lastSend
	participant.messageAtRecord = participant.lastReceived
	marker := LastMarker{participant.lastSnapshotId, participant.id, 0}
	participant.lastMarker = append(participant.lastMarker, marker)

	for i := range participants {
		if i != participant.id {
			messageChannels[participant.id][i] <- Message{from: participant.id, action: Marker, to: i, ajo: false, snapshotId: participant.lastSnapshotId}
			fmt.Println(participant.id, "SEND SNAPSHOT to", i)
		}
	}
	participant.lastSnapshotId = participant.lastSnapshotId + 1
}

func (participant *Participant) processMessages(messages []Message, participants []Participant, wg *sync.WaitGroup) {
	defer wg.Done()
	for i := participant.lastMessageId; i < len(messages); i++ {
		message := messages[i]
		if message.from == participant.id {
			participant.lastMessageId = i + 1
			switch message.action {
			case Send:
				participant.sendMessage(message.to)
			case Receive:
				participant.receiveMessage(message.to, participants)
			case Wait:
				participant.wait(message.to)
			case Snapshot:
				participant.snapshot(participants)
			}
		}
	}
}

func registerSnapshot(participant Participant, participants []Participant, snapshotId int) {
	fmt.Println("REGISTERED SNAPSHOT", snapshotId)
	filename := fmt.Sprintf("snapshot_%d.txt", snapshotId)
	file, err := os.Create(filename)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()
	for i := range participants {
		file.WriteString(fmt.Sprintf("%d:\n", i))
		file.WriteString(fmt.Sprintf("stateatRecord: %v\n", participants[i].stateAtRecord))
		file.WriteString(fmt.Sprintf("messageatRecord: %v\n", participants[i].messageAtRecord))
		file.WriteString(fmt.Sprintf("ajoAtRecord: %v\n", participants[i].ajoAtRecord))
		file.WriteString(fmt.Sprintf("channelAtRecord: %v\n", participants[i].channelAtRecord))
		file.WriteString(fmt.Sprintf("ajoAtMarker: %v\n", participants[i].ajoAtMarker))
		removeMarkers(participants[i], snapshotId, participant.id)
	}
}

func removeMarkers(participant Participant, snapshotId int, participantId int) Participant {
	var newMarkers []LastMarker
	for _, marker := range participant.lastMarker {
		if marker.participantId == participantId && marker.snapshotId == snapshotId {
			continue
		}
		newMarkers = append(newMarkers, marker)
	}
	participant.lastMarker = newMarkers

	participant.stateAtRecord = []int{}
	participant.messageAtRecord = []int{}
	participant.channelAtRecord = []int{}
	participant.ajoAtMarker = AjoAtMarker{}
	participant.ajoAtRecord = false

	return participant
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Uso: ./main path_input")
		return
	}

	pathInput := os.Args[1]

	participants, messages := readFile(pathInput)

	var wg sync.WaitGroup

	for i := range participants {
		wg.Add(1)
		go func(participant *Participant) {
			participant.processMessages(messages, participants, &wg)
		}(&participants[i])
	}

	wg.Wait()
}
