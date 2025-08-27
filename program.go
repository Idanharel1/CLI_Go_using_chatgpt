package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/websocket"
)

// the function that will be activated in the function calling of multiplication of 2 numbers
func getMulti(num1 float64, num2 float64) float64 {
	fmt.Println("The multiplication function has been activated!")
	return num1 * num2
}
func main() {
	//the tool's (getMulti) setup
	tools := []map[string]interface{}{
		map[string]interface{}{
			"type":        "function",
			"name":        "getMulti",
			"description": "Get two float numbers and returns the multiplication of them",
			"parameters": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"num1": map[string]interface{}{
						"type":        "number",
						"description": "the first number to be multiplied",
					},
					"num2": map[string]interface{}{
						"type":        "number",
						"description": "the second number to be multiplied",
					},
				},
				"required":             []string{"num1", "num2"},
				"additionalProperties": false,
			},
		},
	}
	readWrite := false
	//api request setup
	KEY := os.Getenv("API_KEY")
	// fmt.Println(KEY)
	tokenRequest := map[string]interface{}{
		"model":        "gpt-4o-realtime-preview",
		"modalities":   []string{"text"},
		"instructions": "You are a friendly assistant. if you get any ask for multiplication of two numbers ALWAYS USE THE getMulti function",
		"tools":        tools,
	}
	requestJson, _ := json.Marshal(tokenRequest)
	request, _ := http.NewRequest("POST", "https://api.openai.com/v1/realtime/sessions", bytes.NewBuffer(requestJson))
	request.Header.Set("Authorization", "Bearer "+KEY)
	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		fmt.Println("error in generating a token " + err.Error())
	} else {
		// fmt.Println("response has came back:")
		resBody, err := io.ReadAll(response.Body)
		if err != nil {
			println("error in converting the resposne")
		} else {
			var tokenResponse map[string]interface{}
			json.Unmarshal(resBody, &tokenResponse)
			// fmt.Println(tokenResponse["client_secret"].(map[string]interface{})["value"].(string))
		}

	}
	defer response.Body.Close()
	//creating the specific chat
	url := "wss://api.openai.com/v1/realtime?model=gpt-4o-mini-realtime-preview-2024-12-17"

	headers := http.Header{
		"Authorization": []string{"Bearer " + KEY},
		"OpenAI-Beta":   []string{"realtime=v1"},
	}
	conn, _, err := websocket.DefaultDialer.Dial(url, headers)
	if err != nil {
		fmt.Println("error with the creation of the websocket")
		fmt.Println(err)
	} else {
		// fmt.Println("conn")
	}

	//update the conversation to know the tool getMulti
	updateGetMulti := map[string]interface{}{
		"type": "session.update",
		"session": map[string]interface{}{
			"tools":       tools,
			"tool_choice": "auto",
		},
	}
	updateEv, _ := json.Marshal(updateGetMulti)
	conn.WriteMessage(websocket.TextMessage, updateEv)

	//a function that handles the using of the function calling, will be activated as the chat needs an output from the function in order to continue answering
	handleFunction := func(num1 float64, num2 float64, call_id string) {
		funcRes := map[string]interface{}{
			"type": "conversation.item.create",
			"item": map[string]interface{}{
				"type":    "function_call_output",
				"call_id": call_id,
				"output":  fmt.Sprintf("%g", getMulti(num1, num2)),
			},
		}
		funcResEv, _ := json.Marshal(funcRes)
		conn.WriteMessage(websocket.TextMessage, funcResEv)
		respEvent := map[string]interface{}{
			"type": "response.create",
			"response": map[string]interface{}{
				"modalities":  []string{"text"},
				"tool_choice": "none",
			},
		}
		resev, err := json.Marshal(respEvent)
		if err != nil {
			fmt.Println("err in marshaling")
			return
		}
		conn.WriteMessage(websocket.TextMessage, resev)
	}

	reader := func() {
		for { //infinite busy wait loop
			time.Sleep(time.Millisecond * 500)

			if !readWrite || true {
				_, message, err := conn.NextReader()
				if err != nil {
					fmt.Println("loop stoppped " + err.Error())
					return
				} else {
					var msg4 map[string]interface{}
					msg, _ := io.ReadAll(message)
					msg3 := string(msg)
					json.Unmarshal([]byte(msg3), &msg4)
					switch msg4["type"] {
					case "response.text.delta": //stream of answer
						if msg4["delta"] != nil {
							fmt.Print(msg4["delta"])
						}
					case "response.text.done": //finishes answering - enable writing new message
						fmt.Println()
						readWrite = true

					case "response.function_call_arguments.done": //finishes generating the arguments for the getMulti function
						var data map[string]float64
						json.Unmarshal([]byte(msg4["arguments"].(string)), &data)
						num1 := data["num1"]
						num2 := data["num2"]
						handleFunction(num1, num2, msg4["call_id"].(string))
					}

				}

			}

		}
	}
	go reader() //goroutine that runs in the infinite loop concurrently

	writer := func() {
		scanner := bufio.NewReader(os.Stdin)
		for {
			if readWrite { //indicates finishes answering the previous message
				fmt.Print("User: ")
				msg, _ := scanner.ReadString('\n')
				for msg == "" {
					msg, _ = scanner.ReadString('\n')
				}
				if msg == "close" { //a keyword for finishing conversation
					fmt.Println("Converstaion Ended")
					defer conn.Close()
					break
				}
				time.Sleep(time.Second * 3)
				msgEvent := map[string]interface{}{
					"type": "conversation.item.create",
					"item": map[string]interface{}{
						"type": "message",
						"role": "user",
						"content": [](map[string]interface{}){
							map[string]interface{}{
								"type": "input_text",
								"text": msg,
							},
						},
					},
				}
				msgev, _ := json.Marshal(msgEvent)
				conn.WriteMessage(websocket.TextMessage, msgev)
				respEvent := map[string]interface{}{
					"type": "response.create",
					"response": map[string]interface{}{
						"modalities": []string{"text"},
					},
				}
				resev, err := json.Marshal(respEvent)
				if err != nil {
					fmt.Println("err in marshaling")
					return
				}
				conn.WriteMessage(websocket.TextMessage, resev)
				fmt.Print("Chat: ")
				readWrite = false
			}
		}

	}
	readWrite = true
	go writer()

	time.Sleep(time.Minute * 10) //time limitation for session (avoiding unlimited session)

	defer conn.Close()

}
