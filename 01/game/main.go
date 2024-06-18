package main

import (
	"bufio"
	"fmt"
	"gitlab.vk-golang.com/vk-golang/lectures/01_intro/99_hw/game/world"
	"os"
	"slices"
	"strings"
)

// Глобальные переменные с экземплярами Игрока и мира
var player world.Player
var gameWorld map[string]*world.Room

func main() {
	initGame()
	in := bufio.NewScanner(os.Stdin)
	for in.Scan() {
		command := in.Text()
		fmt.Println(handleCommand(command))
	}
}

func initGame() {
	kitchen := world.Room{
		Name:      "кухня",
		NextRooms: make([]*world.Room, 0, 1),
		Storages: []*world.Storage{
			{
				NameInCase: "на столе",
				Items:      []world.Item{"чай"},
			},
		},
		Note:           "кухня, ничего интересного",
		LookAroundNote: "ты находишься на кухне",
		Tasks: []world.Task{
			{Text: "собрать рюкзак", Done: func() bool {
				inventory := player.Inventory
				return slices.Contains(inventory, world.Item("ключи")) &&
					slices.Contains(inventory, world.Item("конспекты"))
			}},
			{Text: "идти в универ", Done: func() bool {
				return false
			}},
		},
	}
	hall := world.Room{
		Name:      "коридор",
		NextRooms: make([]*world.Room, 0, 3),
		Note:      "ничего интересного",
	}
	bedroom := world.Room{
		Name:      "комната",
		NextRooms: make([]*world.Room, 0, 1),
		Storages: []*world.Storage{
			{
				NameInCase: "на столе",
				Items:      []world.Item{"ключи", "конспекты"},
			},
			{
				NameInCase: "на стуле",
				Items:      []world.Item{"рюкзак"},
			},
		},
		Note: "ты в своей комнате",
	}
	outside := world.Room{
		Name:      "улица",
		NextRooms: make([]*world.Room, 0, 1),
		Note:      "на улице весна",
	}
	home := world.Room{
		Name: "домой",
	}

	hallToOutsideDoor := world.Door{
		IsClosed: true,
		States:   map[bool]string{true: "дверь закрыта", false: "дверь открыта"},
		Rooms:    []*world.Room{&hall, &outside},
	}

	kitchen.NextRooms = append(kitchen.NextRooms, &hall)
	bedroom.NextRooms = append(bedroom.NextRooms, &hall)
	hall.NextRooms = append(hall.NextRooms, &kitchen, &bedroom, &outside)
	outside.NextRooms = []*world.Room{&home}

	hall.DoorFromRoom = &hallToOutsideDoor
	outside.DoorFromRoom = &hallToOutsideDoor

	gameWorld = map[string]*world.Room{
		"кухня":   &kitchen,
		"коридор": &hall,
		"комната": &bedroom,
		"улица":   &outside,
	}

	player = world.Player{
		CurrentRoom: &kitchen,
		Inventory:   make([]world.Item, 0, 5),
	}
}

func handleCommand(command string) string {
	commandSplit := strings.Split(command, " ")
	var res string
	switch commandSplit[0] {
	case "осмотреться":
		res = player.LookAround()
	case "идти":
		res = player.GoToRoom(gameWorld[commandSplit[1]])
	case "взять":
		res = player.TakeItem(world.Item(commandSplit[1]))
	case "надеть":
		res = player.WearItem(world.Item(commandSplit[1]))
	case "применить":
		res = player.UseItem(world.Item(commandSplit[1]), commandSplit[2])
	default:
		res = "неизвестная команда"
	}
	return res
}
