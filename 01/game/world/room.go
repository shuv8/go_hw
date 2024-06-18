package world

import "strings"

type Room struct {
	Name           string
	Storages       []*Storage
	NextRooms      []*Room
	DoorFromRoom   *Door
	Note           string
	Tasks          []Task
	LookAroundNote string
}

func (room *Room) NextRoomsList() (result string) {
	result = "можно пройти - "
	nextRoomsNames := make([]string, 0, len(room.NextRooms))
	for _, nextRoom := range room.NextRooms {
		nextRoomsNames = append(nextRoomsNames, nextRoom.Name)
	}
	result += strings.Join(nextRoomsNames, ", ")
	return
}

func (room *Room) TasksList() (result string) {
	result = "надо "
	var tasks []string
	for _, task := range room.Tasks {
		if !task.Done() {
			tasks = append(tasks, task.Text)
		}
	}
	result += strings.Join(tasks, " и ")
	return
}

func (item Item) OpenDoor(door *Door) string {
	if item != "ключи" {
		return cantUse
	}
	door.IsClosed = !door.IsClosed
	return door.States[door.IsClosed]
}

type Door struct {
	IsClosed bool
	States   map[bool]string
	Rooms    []*Room
}
