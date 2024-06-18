package world

import (
	"slices"
	"strings"
)

const cantUse = "не к чему применить"

type Player struct {
	CurrentRoom *Room
	Inventory   []Item
	HasBackpack bool
}

func (player *Player) GoToRoom(room *Room) (result string) {
	if slices.Contains(player.CurrentRoom.NextRooms, room) {
		if door := player.CurrentRoom.DoorFromRoom; door == nil || !slices.Contains(door.Rooms, room) || !door.IsClosed {
			player.CurrentRoom = room
			result = player.CurrentRoom.Note + ". " + player.CurrentRoom.NextRoomsList()
		} else {
			result = "дверь закрыта"
		}
	} else {
		result = strings.Join([]string{"нет пути в", room.Name}, " ")
	}
	return
}

func (player *Player) LookAround() (result string) {
	res := player.CurrentRoom.LookAroundNote
	if res != "" {
		res += ", "
	}
	emptyRoom := true
	storages := player.CurrentRoom.Storages
	var storageList []string
	for _, storage := range storages {
		if len(storage.Items) > 0 {
			emptyRoom = false
			var items []string
			for _, item := range storage.Items {
				items = append(items, string(item))
			}
			itemsList := strings.Join(items, ", ")
			storageList = append(storageList, strings.Join([]string{storage.NameInCase, itemsList}, ": "))
		}
	}
	res += strings.Join(storageList, ", ")
	if emptyRoom {
		res += "пустая комната"
	}
	if len(player.CurrentRoom.Tasks) > 0 {
		res = strings.Join([]string{res, player.CurrentRoom.TasksList()}, ", ")
	}
	res = strings.Join([]string{res, player.CurrentRoom.NextRoomsList()}, ". ")
	return res
}

func (player *Player) TakeItem(item Item) (result string) {
	storages := player.CurrentRoom.Storages
	for _, storage := range storages {
		for i, storageItem := range storage.Items {
			if storageItem == item {
				if !player.HasBackpack {
					return "некуда класть"
				}
				result = strings.Join([]string{"предмет добавлен в инвентарь", string(item)}, ": ")
				player.Inventory = append(player.Inventory, storageItem)
				storage.Items = deleteItem(storage.Items, i)
				return
			}
		}
	}
	return "нет такого"
}

func (player *Player) WearItem(item Item) (result string) {
	storages := player.CurrentRoom.Storages
	for _, storage := range storages {
		for i, storageItem := range storage.Items {
			if storageItem == item {
				result = strings.Join([]string{"вы надели", string(item)}, ": ")
				player.HasBackpack = true
				storage.Items = deleteItem(storage.Items, i)
				return
			}
		}
	}
	return "нет такого"
}

func (player *Player) UseItem(item Item, target string) (result string) {
	if !slices.Contains(player.Inventory, item) {
		result = "нет предмета в инвентаре - " + string(item)
		return
	}
	// В switch по типу предмета можно добавить другие цели для применения,
	// если в игре добавятся используемые предметы кроме ключа
	switch target {
	case "дверь":
		if player.CurrentRoom.DoorFromRoom != nil {
			result = item.OpenDoor(player.CurrentRoom.DoorFromRoom)
		} else {
			result = cantUse
		}
	default:
		result = cantUse
	}
	return
}
