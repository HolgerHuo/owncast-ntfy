package main

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/k3a/html2text"
	"io"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
)

var (
	ntfyUrl       = flag.String("ntfy-url", "", "The ntfy url including the topic. e.g.: https://ntfy.sh/mytopic")
	ntfyBasicAuth = flag.String("ntfy-basic-auth", "", "Basic auth used for ntfy, e.g.: user:pass")
	allowInsecure = flag.Bool("allow-insecure", false, "Allow insecure connections to ntfy-url")
	port          = flag.Int("port", 8080, "The port to listen on")
	markdown      = flag.Bool("markdown", false, "Use Markdown message formatting")
)

var (
	bold = ""
)

var urlRe = regexp.MustCompile(`(https?://.*?)/([-a-zA-Z0-9()@:%_\+.~#?&=]+)$`)
var topic string
var serverUrl string

func main() {
	flag.Parse()
	var err error

	err = validateFlags()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if *allowInsecure {
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	matches := urlRe.FindStringSubmatch(*ntfyUrl)
	if len(matches) != 3 {
		fmt.Println("Error parsing ntfy-url")
		os.Exit(1)
	}
	serverUrl = matches[1]
	topic = matches[2]

	fmt.Println("ntfy-url:", *ntfyUrl)
	fmt.Println("topic:", topic)
	fmt.Println("serverUrl:", serverUrl)
	if *ntfyBasicAuth != "" {
		fmt.Println("basicAuth:", *ntfyBasicAuth)
	}

	if *markdown {
		bold = "**"
	}

	err = server()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func validateFlags() error {
	if *ntfyUrl == "" {
		return fmt.Errorf("ntfy-url is required")
	}
	if !strings.HasPrefix(*ntfyUrl, "http") {
		return fmt.Errorf("ntfy-url must start with http or https")
	}
	if !urlRe.MatchString(*ntfyUrl) {
		return fmt.Errorf("ntfy-url must follow the format https://ntfy.sh/<topic>. (you may use a custom ntfy server)")
	}
	return nil
}

// start a web server on port 8080 and output any json data to the console from a post request
func server() error {
	fmt.Println("Forwarding Owncast notifications to ntfy...", *ntfyUrl)
	http.HandleFunc("/", handleRequest)
	fmt.Println("Listening on port " + strconv.Itoa(*port) + "...")
	err := http.ListenAndServe(fmt.Sprintf(":%d", *port), nil)
	if err != nil {
		return err
	}
	return nil
}

func handleRequest(response http.ResponseWriter, request *http.Request) {
	if request.Method == "POST" {
		// Read the request body
		body, err := io.ReadAll(request.Body)
		if err != nil {
			http.Error(response, "Error reading request body", http.StatusBadRequest)
			return
		}

		// Parse the JSON payload
		var payload OwncastPayload
		err = json.Unmarshal(body, &payload)
		if err != nil {
			http.Error(response, "Error parsing JSON payload", http.StatusBadRequest)
			return
		}

		notificationPayload, err := prepareNotification(payload)
		if err != nil {
			http.Error(response, "Unknown Owncast Message", http.StatusNotImplemented)
		}
		err = sendNotification(notificationPayload)
		if err != nil {
			http.Error(response, "Error sending notification", http.StatusInternalServerError)
			return
		}

		// Send response
		response.WriteHeader(http.StatusOK)
		fmt.Fprint(response, "Payload received successfully\n")
	} else {
		http.Error(response, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
}

func prepareNotification(owncastPayload OwncastPayload) (NtfyNotification, error) {
	// Prepare the payload
	if owncastPayload.Type == "CHAT" {
		return NtfyNotification{
			Message: html2text.HTML2Text(owncastPayload.EventData.Body),
			Title:   owncastPayload.EventData.User.DisplayName + " said",
			Topic:   topic,
			Tags:    []string{"speech_balloon", "message"},
		}, nil
	} else if owncastPayload.Type == "NAME_CHANGE" {
		return NtfyNotification{
			Message: bold + owncastPayload.EventData.User.PreviousNames[len(owncastPayload.EventData.User.PreviousNames)-1] + bold + " changed its name to " + bold + owncastPayload.EventData.NewName + bold,
			Topic:   topic,
			Tags:    []string{"label", "name_change"},
		}, nil
	} else if owncastPayload.Type == "USER_JOINED" {
		return NtfyNotification{
			Message: bold + owncastPayload.EventData.User.DisplayName + bold + " joined stream",
			Topic:   topic,
			Tags:    []string{"sunglasses", "user_join"},
		}, nil
	} else if owncastPayload.Type == "USER_PARTED" {
		return NtfyNotification{
			Message: bold + owncastPayload.EventData.User.DisplayName + bold + " left stream",
			Topic:   topic,
			Tags:    []string{"dash", "user_leave"},
		}, nil
	} else if owncastPayload.Type == "STREAM_STARTED" {
		return NtfyNotification{
			Message: bold + owncastPayload.EventData.StreamTitle + bold + " started streaming",
			Topic:   topic,
			Tags:    []string{"green_circle", "stream_start"},
		}, nil
	} else if owncastPayload.Type == "STREAM_STOPPED" {
		return NtfyNotification{
			Message: bold + owncastPayload.EventData.StreamTitle + bold + " stopped streaming",
			Topic:   topic,
			Tags:    []string{"x", "stream_stop"},
		}, nil
	} else if owncastPayload.Type == "STREAM_TITLE_UPDATED" {
		return NtfyNotification{
			Message: "Stream title updated to " + bold + owncastPayload.EventData.StreamTitle + bold,
			Topic:   topic,
			Tags:    []string{"new", "stream_title_update"},
		}, nil
	}
	return NtfyNotification{}, errors.New("unknown owncast message")
}

func sendNotification(payload NtfyNotification) error {
	// Marshal the payload
	message, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	fmt.Println("Sending notification to ntfy...")
	fmt.Println(string(message))

	// Create a new request using http
	req, err := http.NewRequest("POST", serverUrl, bytes.NewBuffer(message))
	if err != nil {
		return err
	}

	// Set the content type to json
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Markdown", "yes")
	req.Header.Add("User-Agent", "owncast-ntfy/0.1.0")
	if *ntfyBasicAuth != "" {
		req.Header.Add("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(*ntfyBasicAuth)))
	}

	// Send the request
	defer req.Body.Close()
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	// Check the response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ntfy returned status code %d", resp.StatusCode)
	}

	fmt.Println("Notification sent to ntfy")

	return nil

}
