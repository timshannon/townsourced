// Townsourced
// Copyright 2015 Tim Shannon. All rights reserved.

package app

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path"
	"strings"
	"text/template"

	"git.townsourced.com/townsourced/townsourced/data"
)

func init() {
	messages = make(map[string]*MessageTemplate)
}

// MessageTemplate is a templated message with a subject and body
// used for notifications and emails
type MessageTemplate struct {
	subject *template.Template
	body    *template.Template
}

type message struct {
	subject  string
	body     string
	bodyPath string
}

var messages msgMap

type msgMap map[string]*MessageTemplate

func addMessageType(msgType string, msg message) {
	_, ok := messages[msgType]
	if ok {
		panic(fmt.Sprintf("A message of type '%s' already exists!", msgType))
	}

	m := &MessageTemplate{}

	m.subject = template.Must(template.New(msgType + "-subject").Funcs(m.funcMap()).Parse(strings.TrimSpace(msg.subject)))
	body := msg.body

	if msg.bodyPath != "" {
		b, err := ioutil.ReadFile(msg.bodyPath)
		if err != nil {
			panic(fmt.Sprintf("Error loading template from file %s: %s", msg.bodyPath, err))
		}
		body = string(b)
	}

	m.body = template.Must(template.New(msgType + "-body").Funcs(m.funcMap()).Parse(strings.TrimSpace(body)))
	messages[msgType] = m
}

func (m msgMap) use(msgType string) *MessageTemplate {
	msg, ok := m[msgType]
	if !ok {
		panic(fmt.Sprintf("No message type '%s' found!", msgType))
	}
	return msg
}

// Execute executes the given message with the passed in data
func (m *MessageTemplate) Execute(data interface{}) (subject, body string, err error) {
	sub := bytes.NewBuffer([]byte{})
	bdy := bytes.NewBuffer([]byte{})

	err = m.subject.Execute(sub, data)
	if err != nil {
		return "", "", err
	}

	err = m.body.Execute(bdy, data)
	if err != nil {
		return "", "", err
	}

	subject = sub.String()
	body = bdy.String()
	err = nil
	return

}

func (m *MessageTemplate) funcMap() template.FuncMap {
	return template.FuncMap{
		"FromUUID": data.FromUUID,
		"url": func(paths ...string) string {
			return baseURL + "/" + path.Join(paths...)
		},
	}
}
