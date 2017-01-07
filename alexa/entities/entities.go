package entities

import (
	"encoding/json"
)

type Application struct {
	ApplicationID string `json:"applicationId"`
}

type User struct {
	UserID      string `json:"userId"`
	AccessToken string `json:"accessToken"`
}

type Session struct {
	New         bool                       `json:"new"`
	SessionID   string                     `json:"sessionId"`
	Application Application                `json:"application"`
	Attributes  map[string]json.RawMessage `json:"attributes"`
	User        User                       `json:"user"`
}

type System struct {
	Application Application                `json:"application"`
	User        User                       `json:"user"`
	Device      map[string]json.RawMessage `json:"device"`
}

type Context struct {
	New         bool                   `json:"new"`
	SessionID   string                 `json:"sessionId"`
	Application Application            `json:"application"`
	Attributes  map[string]interface{} `json:"attributes"`
	User        User                   `json:"user"`
}

type BaseRequest struct {
	Type      string `json:"type"`
	RequestID string `json:"requestId"`
	Timestamp string `json:"timestamp"`
	Locale    string `json:"locale"`
}

type LaunchRequest struct {
	BaseRequest
}

type IntentSlot struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type IntentRequest struct {
	BaseRequest
	Intent struct {
		Name  string                `json:"name"`
		Slots map[string]IntentSlot `json:"slots"`
	} `json:"intent"`
}

type SessionEndedRequest struct {
	BaseRequest
	Error struct {
		Type    string `json:"string"`
		Message string `json:"message"`
	} `json:"error"`
}

type Request struct {
	Version string          `json:"version"`
	Session Session         `json:"session"`
	Context Context         `json:"context"`
	Request json.RawMessage `json:"request"`
}

type OutputSpeech struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
	SSML string `json:"ssml,omitempty"`
}

type Card struct {
	Type    string `json:"type"`
	Title   string `json:"title,omitempty"`
	Content string `json:"content,omitempty"`
	Text    string `json:"text,omitempty"`
	Image   struct {
		SmallImageURL string `json:"smallImageUrl"`
		LargeImageURL string `json:"largeImageUrl"`
	} `json:"image"`
}

type Directive struct {
	Type string `json:"type"`
}

type Response struct {
	Version           string            `json:"version,omitempty"`
	SessionAttributes map[string]string `json:"sessionAttributes,omitempty"`
	Response          struct {
		OutputSpeech OutputSpeech `json:"outputSpeech,omitempty"`
		Card         Card         `json:"card,omitempty"`
		Reprompt     struct {
			OutputSpeech OutputSpeech `json:"outputSpeech,omitempty"`
		} `json:"reprompt,omitempty"`
		Directives []Directive `json:"directives,omitempty"`
	} `json:"response,omitempty"`
}
