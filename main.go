package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/MarinX/keylogger"
	"github.com/atotto/clipboard"
	"github.com/sirupsen/logrus"
)

func get_latest_image() (string, error) {
	// Path to your screenshot or image file
	imagePath := "/home/x-ae-x/screenshots_flameshot/latest.png"

	// Read the image file
	imageData, err := ioutil.ReadFile(imagePath)
	if err != nil {
		return "", fmt.Errorf("failed to read image file: %v", err)
	}

	// Convert to base64
	base64Image := base64.StdEncoding.EncodeToString(imageData)
	return base64Image, nil
}



func ask_gpt(payload string) string {
	// Define the request payload
	requestData := map[string]any{
		"model":  "llama3:8b",
		"prompt": payload,
		"stream": false,
	}

	// Marshal the request data to JSON
	jsonData, err := json.Marshal(requestData)
	if err != nil {
		log.Fatalf("Failed to marshal request data: %v", err)
	}

	// Make the POST request
	resp, err := http.Post("http://localhost:11434/api/generate", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Failed to read response body: %v", err)
	}

	// Unmarshal the JSON response
	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		log.Fatalf("Failed to unmarshal response body: %v", err)
	}

	// Access the "response" field from the JSON response
	response, ok := result["response"].(string)
	if !ok {
		log.Fatalf("Response field not found or is not a string")
	}

	return response
}

func get_screenshot() (image.Image, error) {
	var path_to_screenshot_dir = ""

	file, err := os.Open(path_to_screenshot_dir)
	if err != nil {
		return nil, fmt.Errorf("failed to open image file: %v", err)
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %v", err)
	}

	return img, nil
}

func send_msg_with_a_screenshot(msg string) {

}


func send_req(msg string) {
    textFromClipboard, err := clipboard.ReadAll()
    if err != nil {
        fmt.Println(err)
        return
    }
	//! here add your own token
	var token = "radi"
        var requestData map[string]interface{}
        
        // Check if we need to include an image
        if strings.Contains(msg, "CLIP") {
            // Get and encode the image
            base64Image, err := get_latest_image()
            if err != nil {
                log.Printf("Failed to get image: %v", err)
                return
            }
            msg = strings.ReplaceAll( msg,"useClip","") 
            // Prepare request with both message and image
            requestData = map[string]interface{}{
                "message": textFromClipboard + msg,
                "image": base64Image,
				"token":token,
            }
        } else {
            // Text-only request
            requestData = map[string]interface{}{
                "message": textFromClipboard + msg,

				"token":token,
            }
        }

		

        // Marshal the request data to JSON
        jsonData, err := json.Marshal(requestData)
        if err != nil {
            log.Printf("Failed to marshal request data: %v", err)
            return
        }

        // Make the POST request to Flask server
        resp, err := http.Post("http://localhost:3003/chat", "application/json", bytes.NewBuffer(jsonData))
        if err != nil {
            log.Printf("Failed to send request to Flask server: %v", err)
            return
        }
        defer resp.Body.Close()

        // Read the response body
        body, err := ioutil.ReadAll(resp.Body)
        if err != nil {
            log.Printf("Failed to read response body: %v", err)
            return
        }

        // Unmarshal the JSON response
        var result map[string]interface{}
        err = json.Unmarshal(body, &result)
        if err != nil {
            log.Printf("Failed to unmarshal response body: %v", err)
            return
        }

        // Get the response from the result
        response, ok := result["response"].(string)
        if !ok {
            log.Printf("Response field not found or is not a string")
            return
        }

		fmt.Println("res from gpt", response)
        // Write the response to clipboard
        clipboard.WriteAll(response)
}

func main() {
	// find keyboard device, does not require a root permission
	keyboard := keylogger.FindKeyboardDevice()

	// check if we found a path to keyboard
	if len(keyboard) <= 0 {
		logrus.Error("No keyboard found...you will need to provide manual input path")
		return
	}

	logrus.Println("Found a keyboard at", keyboard)
	// init keylogger with keyboard
	k, err := keylogger.New(keyboard)
	if err != nil {
		logrus.Error(err)
		return
	}
	defer k.Close()

	// write to keyboard example:
	go func() {
		time.Sleep(5 * time.Second)
		// open text editor and focus on it, it should say "marin" and new line will be printed
		keys := []string{"m", "a", "r", "i", "n", "ENTER"}
		for _, key := range keys {
			// write once will simulate keyboard press/release, for long press or release, lookup at Write
			k.WriteOnce(key)
		}
	}()

	events := k.Read()
	var capture_strokes = false
	var buffer = ""
	// range of events
	for e := range events {
		switch e.Type {
		// EvKey is used to describe state changes of keyboards, buttons, or other key-like devices.
		// check the input_event.go for more events
		case keylogger.EvKey:

			// if the state of key is pressed
			if e.KeyPress() {
				logrus.Println("[event] press key ", e.KeyString())

				if e.KeyString() == "CAPS_LOCK" {
					if capture_strokes {
						send_req(buffer)
					}
					capture_strokes = !capture_strokes
					buffer = ""
				}

				if capture_strokes {
					var key = e.KeyString()
					if key == "SPACE" {
						buffer += " "
					} else if key == "CAPS_LOCK" {

					} else {
						if key == "BACKSPACE" {
							if len(buffer) > 0 {
								buffer = buffer[:len(buffer)-1]
							}
						}
						buffer += e.KeyString()
					}
				}

			}

			// if the state of key is released
			if e.KeyRelease() {
				logrus.Println("[event] release key ", e.KeyString())
			}

			break
		}
	}
}
