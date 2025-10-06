package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"text/tabwriter"
	"time"

	"suah.dev/protect"
)

var verbose bool

// Push represents a message sent to the Pushover api
type Push struct {
	Message  string `json:"message"`
	Title    string `json:"title"`
	Priority int    `json:"priority"`
}

// PushResponse is a response from the Pushover api
type PushResponse struct {
	AppID      int       `json:"appid,omitempty"`
	Date       time.Time `json:"date,omitempty"`
	Error      string    `json:"error,omitempty"`
	ErrorCode  int       `json:"errorCode,omitempty"`
	ErrorDescr string    `json:"errorDescription",omitempty"`
}

func msg(msg interface{}) {
	if verbose {
		fmt.Println(msg)
	}
}

func main() {
	protect.Unveil("/etc/ssl", "r")
	protect.Pledge("stdio inet dns rpath")
	protect.UnveilBlock()

	var token string
	var err error
	var req *http.Request
	var client = *http.DefaultClient
	var pushURL = "https://notify.otter-alligator.ts.net/message"
	var title = flag.String("title", "", "title of message to send")
	var body = flag.String("body", "", "body of message to send")
	var priority = flag.Int("pri", 0, "priority of message")
	flag.BoolVar(&verbose, "v", false, "verbose")

	buf := new(bytes.Buffer)
	flag.Parse()

	if *title == "" || *body == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	token = os.Getenv("PUSHOVER_TOKEN")

	if token == "" {
		fmt.Println("please set PUSHOVER_TOKEN")
		os.Exit(1)
	}

	var push = &Push{
		Title:    *title,
		Message:  *body,
		Priority: *priority,
	}

	if err := json.NewEncoder(buf).Encode(push); err != nil {
		msg(err)
		os.Exit(1)
	}

	req, err = http.NewRequest("POST", pushURL, buf)
	if err != nil {
		msg(fmt.Sprintf("can't POST: %s\n", err))
		os.Exit(1)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Gotify-Key", token)

	res, err := client.Do(req)
	if err != nil {
		msg(fmt.Sprintf("can't make request: %s\n", err))
		os.Exit(1)
	}

	defer res.Body.Close()

	var resBody PushResponse
	if err = json.NewDecoder(res.Body).Decode(&resBody); err != nil {
		if err != nil {
			msg(err)
			os.Exit(1)
		}
	}
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 4, ' ', 0)
	defer w.Flush()

	if verbose {
		fmt.Fprintf(w, "Time:\t%s\n", resBody.Date)
		fmt.Fprintf(w, "AppID:\t%d\n", resBody.AppID)
		fmt.Fprintf(w, "Error:\t%s\n", resBody.Error)
		fmt.Fprintf(w, "Error Code:\t%d\n", resBody.ErrorCode)
		fmt.Fprintf(w, "Error Description:\t%s\n", resBody.ErrorDescr)
	}

}
