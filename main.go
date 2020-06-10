package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"golang.org/x/sys/unix"
)

// Push represents a message sent to the Pushover api
type Push struct {
	Token     string    `json:"token"`
	User      string    `json:"user"`
	Message   string    `json:"message"`
	Device    string    `json:"device"`
	Title     string    `json:"title"`
	URL       string    `json:"url"`
	URLTitle  string    `json:"url_title"`
	Priority  int       `json:"priority"`
	Sound     string    `json:"sound"`
	Timestamp time.Time `json:"timestamp"`
}

// PushResponse is a response from the Pushover api
type PushResponse struct {
	Status  int      `json:"status,omitempty"`
	Request string   `json:"request,omitempty"`
	User    string   `json:"user,omitempty"`
	Errors  []string `json:"errors,omitempty"`
}

func msg(msg interface{}, quiet *bool) {
	if !*quiet {
		fmt.Println(msg)
	}
}

func main() {
	unix.Unveil("", "")
	unix.PledgePromises("stdio inet dns rpath")

	var err error
	var req *http.Request
	var client = *http.DefaultClient
	var pushURL = "https://api.pushover.net/1/messages.json"
	var title = flag.String("title", "", "title of message to send")
	var body = flag.String("body", "", "body of message to send")
	var url = flag.String("url", "", "url to send")
	var priority = flag.Int("pri", 0, "priority of message")
	var sound = flag.String("sound", "pushover", "sound")
	var verbose = flag.Bool("v", false, "verbose")

	buf := new(bytes.Buffer)
	flag.Parse()

	if *title == "" || *body == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	var push = &Push{
		Token:     os.Getenv("PUSHOVER_TOKEN"),
		User:      os.Getenv("PUSHOVER_USER"),
		Timestamp: time.Now(),
		Title:     *title,
		Message:   *body,
		URL:       *url,
		Priority:  *priority,
		Sound:     *sound,
	}

	if err := json.NewEncoder(buf).Encode(push); err != nil {
		msg(err, verbose)
		os.Exit(1)
	}

	req, err = http.NewRequest("POST", pushURL, buf)
	if err != nil {
		msg(fmt.Sprintf("can't POST: %s\n", err), verbose)
		os.Exit(1)
	}

	req.Header.Set("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		msg(fmt.Sprintf("can't make request: %s\n", err), verbose)
		os.Exit(1)
	}

	defer res.Body.Close()

	var resBody PushResponse
	if err = json.NewDecoder(res.Body).Decode(&resBody); err != nil {
		if err != nil {
			msg(err, verbose)
			os.Exit(1)
		}
	}
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 4, ' ', 0)
	defer w.Flush()
	if len(resBody.Errors) > 0 {
		if *verbose {
			fmt.Fprintf(w, "Errors:\t%s\n", strings.Join(resBody.Errors, ", "))
			if resBody.User != "" {
				fmt.Fprintf(w, "User:\t%s\n", resBody.User)
			}
		}
		os.Exit(1)
	} else {
		if *verbose {
			fmt.Fprintf(w, "Request:\t%s\n", resBody.Request)
			fmt.Fprintf(w, "Status:\t%d\n", resBody.Status)
		}
	}
}
