package world

func deleteItem(slice []Item, index int) []Item {
	return append(slice[:index], slice[index+1:]...)
}

type Task struct {
	Text string
	Done func() bool
}

type Item string

type RoomItems map[string]string

type Storage struct {
	NameInCase string
	Items      []Item
}
