package chat

import (
	"encoding/json"
	"errors"
	"github.com/gorilla/websocket"
	"log"
	"strings"
	"time"
)

type WebSocketMessage struct {
	Type string
	Data json.RawMessage
}

type LoginResponse struct {
	Username string
	Users    []string
	Rooms    []string
	Messages []*Message
}

type HistoryResponse struct {
	Messages []*Message
}

type ErrorResponse struct {
	Message string
}

type Client struct {
	conn   *websocket.Conn
	server *Server
	user   *User
	send   chan *WebSocketMessage
}

func NewClient(conn *websocket.Conn, s *Server) (*Client, error) {
	if conn == nil {
		return nil, errors.New("connection is nil")
	}

	if s == nil {
		return nil, errors.New("server is nil")
	}

	c := &Client{
		conn:   conn,
		server: s,
		send:   make(chan *WebSocketMessage),
	}

	go c.read()
	go c.write()

	return c, nil
}

func (c *Client) SetUser(u *User) {
	c.user = u
}

func (c *Client) User() *User {
	return c.user
}

func (c *Client) Conn() *websocket.Conn {
	return c.conn
}

func (c *Client) Send(m *WebSocketMessage) {
	c.send <- m
}

func (c *Client) read() {
	defer close(c.send)
	for {
		msg := new(WebSocketMessage)
		if err := c.conn.ReadJSON(msg); err != nil {
			return
		}

		switch msg.Type {
		case "login":
			u := new(User)
			_ = json.Unmarshal(msg.Data, u)

			err := c.server.Login(u)
			if err != nil {
				v, _ := json.Marshal(ErrorResponse{Message: err.Error()})
				c.Send(&WebSocketMessage{
					Type: "error",
					Data: v,
				})
			} else {
				c.SetUser(u)

				v, _ := json.Marshal(LoginResponse{
					Username: u.Name,
					Users:    c.server.Storage().ActiveUsers(),
					Rooms:    c.server.Storage().Rooms(u),
					Messages: c.server.Storage().MessageHistory("", c.server.Config().HistorySize),
				})

				c.Send(&WebSocketMessage{
					Type: "login",
					Data: v,
				})
			}
		case "history":
			r := new(struct {
				User string
				Room string
			})
			_ = json.Unmarshal(msg.Data, r)
			room := ""
			if r.Room != "" {
				if strings.Compare(r.User, r.Room) < 0 {
					room = r.User + ":" + r.Room
				} else {
					room = r.Room + ":" + r.User
				}
			}
			v, _ := json.Marshal(HistoryResponse{
				Messages: c.server.Storage().MessageHistory(room, c.server.Config().HistorySize),
			})

			c.Send(&WebSocketMessage{
				Type: "history",
				Data: v,
			})
		case "message":
			err := c.handleMessage(msg)
			if err != nil {
				log.Println(err)
			}
		}
	}
}

func (c *Client) write() {
	for msg := range c.send {
		if err := c.conn.WriteJSON(msg); err != nil {
			log.Println(err)
		}
	}
}

func (c *Client) handleMessage(m *WebSocketMessage) error {
	msg := new(Message)
	err := json.Unmarshal(m.Data, msg)
	if err != nil {
		return err
	}

	msg.Time = time.Now().Unix()

	err = c.server.SaveMessage(msg)
	if err != nil {
		return err
	}

	return nil
}
