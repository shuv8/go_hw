package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	tgbotapi "github.com/skinass/telegram-bot-api/v5"
	"gitlab.vk-golang.com/vk-golang/lectures/04_net2/99_hw/taskbot/templates"
	"log"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"text/template"
)

var (
	BotToken   = ""
	WebhookURL = ""
)

var (
	BotTokenPtr   = flag.String("token", "", "token for telegram")
	WebhookURLPtr = flag.String("webhook", "", "webhook addr for telegram")
	ModePtr       = flag.String("mode", "webhook", "mode for bot (webhook or poll)")
)

var taskID = 1

type Task struct {
	ID           int
	Text         string
	Assignee     string
	AssigneeChat int64
	Owner        string
	OwnerChat    int64
}

type TGBot struct {
	botAPI    *tgbotapi.BotAPI
	commands  map[string]func(*TGBot, *tgbotapi.Message) error
	templates map[string]*template.Template
}

var tasks = make(map[int]*Task)

func (bot *TGBot) parseTemplates() error {
	bot.templates = make(map[string]*template.Template)
	tmpls := map[string]string{
		"tasks":        templates.TASKS,
		"myTasks":      templates.MYTASKS,
		"ownerTasks":   templates.OWNERTASKS,
		"newTask":      templates.NEWTASK,
		"assignTask":   templates.ASSIGNTASK,
		"unassignTask": templates.UNASSIGN,
		"resolveTask":  templates.RESOLVE,
	}
	for k, v := range tmpls {
		tmpl := template.New("Bot message")
		tmpl, err := tmpl.Parse(v)
		if err != nil {
			return err
		}
		bot.templates[k] = tmpl
	}
	return nil
}

func (bot *TGBot) sendMessage(chatID int64, messageText string) error {
	msg := tgbotapi.NewMessage(
		chatID,
		messageText,
	)
	_, err := bot.botAPI.Send(msg)
	if err != nil {
		return err
	}
	return nil
}

func (bot *TGBot) getTasks(message *tgbotapi.Message) error {
	if len(tasks) == 0 {
		err := bot.sendMessage(message.Chat.ID, "Нет задач")
		if err != nil {
			return err
		}
		return nil
	}
	keys := make([]int, 0, len(tasks))
	for k := range tasks {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})
	tasksSlice := make([]Task, 0, len(tasks))
	for _, id := range keys {
		tasksSlice = append(tasksSlice, *tasks[id])
	}
	var msg bytes.Buffer
	err := bot.templates["tasks"].Execute(&msg, struct {
		Tasks []Task
		User  string
	}{
		tasksSlice,
		message.Chat.UserName,
	})
	if err != nil {
		return err
	}
	err = bot.sendMessage(message.Chat.ID, strings.TrimRight(msg.String(), "\n\n"))
	if err != nil {
		return err
	}
	return nil
}

func (bot *TGBot) getMyTasks(message *tgbotapi.Message) error {
	if len(tasks) == 0 {
		err := bot.sendMessage(message.Chat.ID, "Нет задач")
		if err != nil {
			return err
		}
		return nil
	}
	keys := make([]int, 0, len(tasks))
	for k, v := range tasks {
		if v.Assignee == message.Chat.UserName {
			keys = append(keys, k)
		}
	}
	if len(keys) == 0 {
		err := bot.sendMessage(message.Chat.ID, "Нет задач")
		if err != nil {
			return err
		}
		return nil
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})
	tasksSlice := make([]Task, 0, len(tasks))
	for _, id := range keys {
		tasksSlice = append(tasksSlice, *tasks[id])
	}
	var msg bytes.Buffer
	err := bot.templates["myTasks"].Execute(&msg, struct {
		Tasks []Task
	}{
		tasksSlice,
	})
	if err != nil {
		return err
	}
	err = bot.sendMessage(message.Chat.ID, strings.TrimRight(msg.String(), "\n\n"))
	if err != nil {
		return err
	}
	return nil
}

func (bot *TGBot) getOwnerTasks(message *tgbotapi.Message) error {
	if len(tasks) == 0 {
		err := bot.sendMessage(message.Chat.ID, "Нет задач")
		if err != nil {
			return err
		}
		return nil
	}
	keys := make([]int, 0, len(tasks))
	for k, v := range tasks {
		if v.Owner == message.Chat.UserName {
			keys = append(keys, k)
		}
	}
	if len(keys) == 0 {
		err := bot.sendMessage(message.Chat.ID, "Нет задач")
		if err != nil {
			return err
		}
		return nil
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})
	tasksSlice := make([]Task, 0, len(tasks))
	for _, id := range keys {
		tasksSlice = append(tasksSlice, *tasks[id])
	}
	var msg bytes.Buffer
	err := bot.templates["ownerTasks"].Execute(&msg, struct {
		Tasks []Task
		User  string
	}{
		tasksSlice,
		message.Chat.UserName,
	})
	if err != nil {
		return err
	}
	err = bot.sendMessage(message.Chat.ID, strings.TrimRight(msg.String(), "\n\n"))
	if err != nil {
		return err
	}
	return nil
}

func (bot *TGBot) addTask(message *tgbotapi.Message) error {
	newTask := Task{
		ID:        taskID,
		Text:      strings.Join(strings.Fields(message.Text)[1:], " "),
		Owner:     message.Chat.UserName,
		OwnerChat: message.Chat.ID,
	}
	tasks[taskID] = &newTask
	taskID++
	var msg bytes.Buffer
	err := bot.templates["newTask"].Execute(&msg, newTask)
	if err != nil {
		return err
	}
	err = bot.sendMessage(message.Chat.ID, msg.String())
	if err != nil {
		return err
	}
	return nil
}

func (bot *TGBot) assignTask(message *tgbotapi.Message) error {
	if len(strings.Split(message.Text, "_")) < 2 {
		err := bot.sendMessage(message.Chat.ID, "Неверное использование команды!\n/help")
		if err != nil {
			return err
		}
		return nil
	}
	assignTaskID, err := strconv.Atoi(strings.Split(message.Text, "_")[1])
	if err != nil {
		return err
	}
	task, ok := tasks[assignTaskID]
	if !ok {
		err = bot.sendMessage(message.Chat.ID, "Несуществующая задача!")
		if err != nil {
			return err
		}
		return nil
	}
	oldAssignee := ""
	if task.Assignee != "" && task.Assignee != message.Chat.UserName {
		oldAssignee = task.Assignee
	}
	task.Assignee = message.Chat.UserName
	var msg bytes.Buffer
	err = bot.templates["assignTask"].Execute(&msg, struct {
		Text     string
		Assignee string
		User     string
	}{
		task.Text,
		task.Assignee,
		message.Chat.UserName,
	})
	if err != nil {
		return err
	}
	err = bot.sendMessage(message.Chat.ID, msg.String())
	if err != nil {
		return err
	}
	if task.AssigneeChat != 0 && task.AssigneeChat != message.Chat.ID {
		var msgToOld bytes.Buffer
		err = bot.templates["assignTask"].Execute(&msgToOld, struct {
			Text     string
			Assignee string
			User     string
		}{
			task.Text,
			task.Assignee,
			oldAssignee,
		})
		if err != nil {
			return err
		}
		err = bot.sendMessage(task.AssigneeChat, msgToOld.String())
		if err != nil {
			return err
		}
	} else if task.Owner != message.Chat.UserName {
		var msgToOwner bytes.Buffer
		err = bot.templates["assignTask"].Execute(&msgToOwner, struct {
			Text     string
			Assignee string
			User     string
		}{
			task.Text,
			task.Assignee,
			task.Owner,
		})
		if err != nil {
			return err
		}
		err = bot.sendMessage(task.OwnerChat, msgToOwner.String())
		if err != nil {
			return err
		}
	}
	task.AssigneeChat = message.Chat.ID
	return nil
}

func (bot *TGBot) unassignTask(message *tgbotapi.Message) error {
	if len(strings.Split(message.Text, "_")) < 2 {
		err := bot.sendMessage(message.Chat.ID, "Неверное использование команды!\n/help")
		if err != nil {
			return err
		}
		return nil
	}
	unassignTaskID, err := strconv.Atoi(strings.Split(message.Text, "_")[1])
	if err != nil {
		return err
	}
	task, ok := tasks[unassignTaskID]
	if !ok {
		err = bot.sendMessage(message.Chat.ID, "Несуществующая задача!")
		if err != nil {
			return err
		}
		return nil
	}
	if task.Assignee != message.Chat.UserName {
		err = bot.sendMessage(message.Chat.ID, "Задача не на вас")
		if err != nil {
			return err
		}
		return nil
	}
	task.Assignee = ""
	task.AssigneeChat = 0
	err = bot.sendMessage(message.Chat.ID, "Принято")
	if err != nil {
		return err
	}
	var msg bytes.Buffer
	err = bot.templates["unassignTask"].Execute(&msg, struct {
		Text string
	}{
		task.Text,
	})
	if err != nil {
		return err
	}
	err = bot.sendMessage(task.OwnerChat, msg.String())
	if err != nil {
		return err
	}
	return nil
}

func (bot *TGBot) resolveTask(message *tgbotapi.Message) error {
	if len(strings.Split(message.Text, "_")) < 2 {
		err := bot.sendMessage(message.Chat.ID, "Неверное использование команды!\n/help")
		if err != nil {
			return err
		}
		return nil
	}
	resolveTaskID, err := strconv.Atoi(strings.Split(message.Text, "_")[1])
	if err != nil {
		return err
	}
	task, ok := tasks[resolveTaskID]
	if !ok {
		err = bot.sendMessage(message.Chat.ID, "Несуществующая задача!")
		if err != nil {
			return err
		}
		return nil
	}
	if task.Assignee != message.Chat.UserName {
		err = bot.sendMessage(message.Chat.ID, "Задача не на вас")
		if err != nil {
			return err
		}
		return nil
	}
	var msg bytes.Buffer
	err = bot.templates["resolveTask"].Execute(&msg, struct {
		Text     string
		Receiver string
		User     string
	}{
		task.Text,
		message.Chat.UserName,
		message.Chat.UserName,
	})
	if err != nil {
		return err
	}
	err = bot.sendMessage(message.Chat.ID, msg.String())
	if err != nil {
		return err
	}
	if task.Owner != message.Chat.UserName {
		var msgToOwner bytes.Buffer
		err = bot.templates["resolveTask"].Execute(&msgToOwner, struct {
			Text     string
			Receiver string
			User     string
		}{
			task.Text,
			task.Owner,
			message.Chat.UserName,
		})
		if err != nil {
			return err
		}
		err = bot.sendMessage(task.OwnerChat, msgToOwner.String())
		if err != nil {
			return err
		}
	}
	delete(tasks, resolveTaskID)
	return nil
}

func (bot *TGBot) help(message *tgbotapi.Message) error {
	msg := tgbotapi.NewMessage(
		message.Chat.ID,
		templates.HELP,
	)
	_, err := bot.botAPI.Send(msg)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func startTaskBot(_ context.Context) error {
	flag.Parse()
	if BotToken == "" {
		BotToken = *BotTokenPtr
	}
	if WebhookURL == "" {
		WebhookURL = *WebhookURLPtr
	}

	bot, err := tgbotapi.NewBotAPI(BotToken)
	fmt.Println(BotToken, WebhookURL, *ModePtr)
	if err != nil {
		log.Printf("NewBotAPI failed: %s", err)
		return err
	}

	bot.Debug = true
	fmt.Printf("Authorized on account %s\n", bot.Self.UserName)

	var updates tgbotapi.UpdatesChannel
	switch *ModePtr {
	case "webhook":
		wh, whErr := tgbotapi.NewWebhook(WebhookURL)
		if whErr != nil {
			log.Printf("NewWebhook failed: %s", err)
			return err
		}

		_, err = bot.Request(wh)
		if err != nil {
			log.Printf("SetWebhook failed: %s", err)
			return err
		}

		updates = bot.ListenForWebhook("/")

		port := os.Getenv("PORT")
		if port == "" {
			port = "8081"
		}
		go func() {
			log.Println("http err:", http.ListenAndServe(":"+port, nil))
		}()
		fmt.Println("start listen :" + port)
	case "poll":
		u := tgbotapi.NewUpdate(0)
		u.Timeout = 60
		updates = bot.GetUpdatesChan(u)
	}

	tgBot := &TGBot{
		botAPI: bot,
		commands: map[string]func(*TGBot, *tgbotapi.Message) error{
			"/tasks":    (*TGBot).getTasks,
			"/new":      (*TGBot).addTask,
			"/assign":   (*TGBot).assignTask,
			"/unassign": (*TGBot).unassignTask,
			"/resolve":  (*TGBot).resolveTask,
			"/my":       (*TGBot).getMyTasks,
			"/owner":    (*TGBot).getOwnerTasks,
			"/help":     (*TGBot).help,
		},
	}
	err = tgBot.parseTemplates()
	if err != nil {
		return err
	}
	regex := regexp.MustCompile(`[\s_]`)
	for update := range updates {
		log.Printf("upd: %#v\n", update)
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}

		cmd, ok := tgBot.commands[regex.Split(update.Message.Text, 2)[0]]

		if !ok {
			err = tgBot.sendMessage(update.Message.Chat.ID, "Неизвестная команда!")
			if err != nil {
				return err
			}
			continue
		}

		err = cmd(tgBot, update.Message)
		if err != nil {
			log.Println(err)
			err = tgBot.sendMessage(update.Message.Chat.ID, "Непредвиденная ошибка при выполнении команды!")
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func main() {
	err := startTaskBot(context.Background())
	if err != nil {
		panic(err)
	}
}
