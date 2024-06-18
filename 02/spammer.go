package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
)

func RunPipeline(cmds ...cmd) {
	channels := make([]chan interface{}, 0, 2)
	channels = append(channels, make(chan interface{}))
	wg := &sync.WaitGroup{}

	for _, command := range cmds {
		channels = append(channels, make(chan interface{}))
		wg.Add(1)
		go func(command cmd, wg *sync.WaitGroup, in, out chan interface{}) {
			defer wg.Done()
			defer close(out)
			command(in, out)
		}(command, wg, channels[len(channels)-2], channels[len(channels)-1])
	}
	wg.Wait()
}

func SelectUsers(in, out chan interface{}) {
	// 	in - string
	// 	out - User
	wg := &sync.WaitGroup{}
	mu := &sync.Mutex{}
	usedEmails := map[string]struct{}{}

	for userEmail := range in {
		if email, okCast := userEmail.(string); okCast {
			wg.Add(1)
			go func(email string, out chan interface{}, usedEmails map[string]struct{}, wg *sync.WaitGroup, mu *sync.Mutex) {
				defer wg.Done()
				user := GetUser(email)
				mu.Lock()
				if _, ok := usedEmails[user.Email]; !ok {
					out <- user
					usedEmails[user.Email] = struct{}{}
				}
				mu.Unlock()
			}(email, out, usedEmails, wg, mu)
		}
	}
	wg.Wait()
}

func SelectMessages(in, out chan interface{}) {
	// 	in - User
	// 	out - MsgID
	wg := &sync.WaitGroup{}
	batch := make([]User, 0, GetMessagesMaxUsersBatch)

	for user := range in {
		if userCasted, ok := user.(User); ok {
			batch = append(batch, userCasted)
		}
		for len(batch) < GetMessagesMaxUsersBatch {
			secondUser, ok := <-in
			if !ok {
				break
			}
			if secondUserCasted, okCast := secondUser.(User); okCast {
				batch = append(batch, secondUserCasted)
			}
		}
		wg.Add(1)
		go func(batch []User, wg *sync.WaitGroup) {
			defer wg.Done()
			messages, err := GetMessages(batch...)
			if err != nil {
				fmt.Println(err)
			} else {
				for _, msg := range messages {
					out <- msg
				}
			}
		}(batch, wg)
		batch = nil
	}
	wg.Wait()
}

func CheckSpam(in, out chan interface{}) {
	// in - MsgID
	// out - MsgData
	wg := &sync.WaitGroup{}
	quotaCh := make(chan struct{}, HasSpamMaxAsyncRequests)
	defer close(quotaCh)

	for msg := range in {
		if msgID, ok := msg.(MsgID); ok {
			wg.Add(1)
			go func(msgID MsgID, wg *sync.WaitGroup, quotaCh chan struct{}) {
				defer wg.Done()
				quotaCh <- struct{}{}
				defer func() { <-quotaCh }()

				msgSpam, err := HasSpam(msgID)
				if err != nil {
					fmt.Println(err)
				} else {
					msgData := MsgData{
						ID:      msgID,
						HasSpam: msgSpam,
					}
					out <- msgData
				}
			}(msgID, wg, quotaCh)
		}
	}
	wg.Wait()
}

func CombineResults(in, out chan interface{}) {
	// in - MsgData
	// out - string
	messages := make([]MsgData, 0, 10)

	for msg := range in {
		if msgData, ok := msg.(MsgData); ok {
			messages = append(messages, msgData)
		}
	}
	sort.SliceStable(messages, func(i, j int) bool {
		if messages[i].HasSpam != messages[j].HasSpam {
			return messages[i].HasSpam
		}
		return messages[i].ID < messages[j].ID
	})
	for _, sortedMessage := range messages {
		out <- strings.Join([]string{strconv.FormatBool(sortedMessage.HasSpam), strconv.FormatUint(uint64(sortedMessage.ID), 10)}, " ")
	}
}
