package main

import (
	"fmt"
	"strings"
	"github.com/mattermost/mattermost/server/public/model"
)

const (
	ActionList = "list"
	ActionAdd = "add"
	ActionDelete = "delete"
)


type Bot struct {
	name string
}

type BotError struct {
	msg string
}

func (err BotError) Error() string {
	return err.msg
}

func (bot *Bot) String() string {
	result := fmt.Sprintf("Bot name: %s\n", bot.name)

	return result
}

func (bot *Bot) handleActions(post *model.Post) (result string, err error) {
	message := post.Message
					
	words := strings.Fields(message)
	wordsLen := len(words)
	if  wordsLen < 2 {
		err = BotError{fmt.Sprintf("Error: Too few arguments %v - %v\n", wordsLen, words)}
		return result, err
	}

	actionType := words[0]
	actionPayload := strings.Join(words[1:], " ")

	switch actionType {
	case ActionList:
		// list all
		result, err = bot.IPList()
	case ActionAdd:
		// add <ip>
		if wordsLen < 2 {
			err = BotError{fmt.Sprintf("Error: Too few arguments %v - %v\n", wordsLen, words)}
			return result, err
		}
		result, err = bot.IPAdd(actionPayload)
	case ActionDelete:
		// delete <entry id>
		if wordsLen < 2 {
			err = BotError{fmt.Sprintf("Error: Too few arguments %v - %v\n", wordsLen, words)}
			return result, err
		}
		result, err = bot.IPDelete(actionPayload)
	}


	return result, err
}

func (bot *Bot) IPList() (string, error){
	var response string
	items, err := getIPDenyData()
	if err != nil {
		return "", err
	}

	formattedItems := formatIPDenyData(items)
	if formattedItems == "" {
		response = "No IPs added in IPDeny"
	} else {
		response = formattedItems
	}

	return response, nil
}

func (bot *Bot) IPAdd(payload string) (string, error){
	var response string = "Done"
	
	err := addIP(payload)
	if err != nil {
		return "", err
	}

	return response, nil
}

func (bot *Bot) IPDelete(entry_id string) (string, error){
	var response string = "Done"

	err := deleteIP(entry_id)
	if err != nil {
		return "", err
	}

	return response, nil
}
