package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

type gcalImpl struct {
	srv    *calendar.Service
	events []*calendar.Event
}

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func OpenGCal(credentialsFile string) (CalendarInterface, error) {
	ctx := context.Background()
	b, err := os.ReadFile(credentialsFile)
	if err != nil {
		return nil, err
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, calendar.CalendarReadonlyScope)
	if err != nil {
		return nil, err
	}
	client := getClient(config)

	srv, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, err
	}

	//call := srv.Events.List("primary").ShowDeleted(false)
	//events, err := call.Do()
	////if err != nil {
	//return nil, err
	//}

	call := srv.Events.List("primary").ShowDeleted(false)
	//call := service.Events.List(calendarID).SingleEvents(true)
	events, err := call.Do()
	if err != nil {
		return nil, err
	}
	var allEvents []*calendar.Event
	allEvents = append(allEvents, events.Items...)
	for events.NextPageToken != "" {
		nextPageToken := events.NextPageToken
		call := call.PageToken(nextPageToken)
		events, err = call.Do()
		if err != nil {
			return nil, err
		}
		allEvents = append(allEvents, events.Items...)
	}

	return &gcalImpl{srv: srv, events: allEvents}, nil
}

func eventTime(ev *calendar.EventDateTime) time.Time {
	if ev == nil {
		return time.Time{}
	}
	if ev.DateTime != "" {
		t, err := time.Parse(time.RFC3339, ev.DateTime)
		if err == nil {
			return t
		}
	} else if ev.Date != "" {
		t, err := time.Parse(time.DateOnly, ev.Date)
		if err == nil {
			return t
		}
	}
	return time.Time{}
}

func (gc *gcalImpl) Years() []int {
	years := map[int]struct{}{}
	for _, e := range gc.events {
		t := eventTime(e.Start)
		if t.IsZero() {
			continue
		}
		years[t.Year()] = struct{}{}
	}

	retval := make([]int, 0, len(years))
	for y := range years {
		retval = append(retval, y)
	}
	return retval
}

func (gc *gcalImpl) Months(year int) []time.Month {
	months := map[time.Month]struct{}{}
	for _, e := range gc.events {
		t := eventTime(e.Start)
		if t.IsZero() {
			continue
		}
		if t.Year() != year {
			continue
		}
		months[t.Month()] = struct{}{}
	}
	retval := make([]time.Month, 0, len(months))
	for m := range months {
		retval = append(retval, m)
	}
	return retval
}

func (gc *gcalImpl) Days(year int, month time.Month) []int {
	days := map[int]struct{}{}
	for _, e := range gc.events {
		t := eventTime(e.Start)
		if t.IsZero() {
			continue
		}
		if t.Year() != year || t.Month() != month {
			continue
		}
		days[t.Day()] = struct{}{}
	}
	retval := make([]int, 0, len(days))
	for d := range days {
		retval = append(retval, d)
	}
	return retval
}

func (gc *gcalImpl) Entries(day time.Time) []CalendarEntry {
	var retval []CalendarEntry
	for _, e := range gc.events {
		t := eventTime(e.Start)
		if t.IsZero() {
			continue
		}
		if t.Year() != day.Year() || t.Month() != day.Month() || t.Day() != day.Day() {
			continue
		}
		retval = append(retval, CalendarEntry{
			Start:       t,
			End:         eventTime(e.End),
			Summary:     e.Summary,
			Description: e.Description,
		})
	}
	return retval
}
