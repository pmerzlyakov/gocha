package chat

import (
	"fmt"
	"strings"
)

type Message struct {
	Author    string
	Recipient string
	Body      string
	Time      int64
}

func (m *Message) String() string {
	if m.Recipient == "" {
		return fmt.Sprintf("user `%s` send message `%s` to the public room", m.Author, m.Body)
	}

	return fmt.Sprintf("user `%s` send message `%s` to user `%s`", m.Author, m.Body, m.Recipient)
}

func (m *Message) Room() string {
	if m.Recipient == "" {
		return ""
	}
	if strings.Compare(m.Author, m.Recipient) < 0 {
		return fmt.Sprintf("%s:%s", m.Author, m.Recipient)
	}
	return fmt.Sprintf("%s:%s", m.Recipient, m.Author)
}
