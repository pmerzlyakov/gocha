package chat

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Server struct {
	config  Config
	clients map[*Client]struct{}
	storage *Storage
	sub     chan *StorageMessage
}

func NewServer(cfg Config) (*Server, error) {
	return &Server{
		config:  cfg,
		clients: make(map[*Client]struct{}),
		storage: NewStorage(cfg.Redis),
		sub:     make(chan *StorageMessage),
	}, nil
}

func (s *Server) ListenAndServe() error {
	s.start()

	http.Handle("/", http.FileServer(http.Dir(s.config.WebRoot)))
	http.HandleFunc(s.config.Endpoint, s.connectHandler)

	return http.ListenAndServe(s.config.Address, nil)
}

func (s *Server) start() {
	err := s.storage.Subscribe(s.sub)
	if err != nil {
		return
	}

	go func() {
		for msg := range s.sub {
			switch msg.Channel {
			case JoinChannel:
				err := s.handleJoin(msg.Data)
				if err != nil {
					log.Println(err)
				}
			case LeaveChannel:
				err := s.handleLeave(msg.Data)
				if err != nil {
					log.Println(err)
				}
			case MessageChannel:
				err := s.handleMessage(msg.Data)
				if err != nil {
					log.Println(err)
				}
			}
		}
	}()
}

func (s *Server) connectHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	c, err := NewClient(conn, s)
	if err != nil {
		log.Println(err)
		return
	}

	s.AddClient(c)
	conn.SetCloseHandler(func(code int, text string) error {
		s.Logout(c.User())
		s.RemoveClient(c)
		return nil
	})
}

func (s *Server) AddClient(c *Client) {
	s.clients[c] = struct{}{}
}

func (s *Server) RemoveClient(c *Client) {
	delete(s.clients, c)
}

func (s *Server) handleJoin(username []byte) error {
	log.Printf("user `%s` logged in", username)

	v, err := json.Marshal(User{Name: string(username)})
	if err != nil {
		return err
	}

	s.broadcast(&WebSocketMessage{
		Type: "join",
		Data: v,
	})

	return nil
}

func (s *Server) handleLeave(username []byte) error {
	log.Printf("user `%s` logged out", username)

	v, err := json.Marshal(User{Name: string(username)})
	if err != nil {
		return err
	}

	s.broadcast(&WebSocketMessage{
		Type: "leave",
		Data: v,
	})

	return nil
}

func (s *Server) handleMessage(v []byte) error {
	msg := new(Message)
	err := json.Unmarshal(v, msg)
	if err != nil {
		return err
	}

	log.Println(msg)

	wsm := &WebSocketMessage{Type: "message", Data: v}

	if msg.Recipient == "" {
		s.broadcast(wsm)
	} else {
		for c := range s.clients {
			u := c.User()
			switch {
			case u == nil:
			case u.Name == msg.Author, u.Name == msg.Recipient:
				c.Send(wsm)
			}
		}
	}

	return nil
}

func (s *Server) broadcast(msg *WebSocketMessage) {
	for c := range s.clients {
		if c.User() != nil {
			c.Send(msg)
		}
	}
}

func (s *Server) Login(u *User) error {
	err := s.storage.AddUser(u)
	if err != nil {
		return err
	}

	err = s.storage.Publish(JoinChannel, []byte(u.Name))
	if err != nil {
		log.Println(err)
	}

	return nil
}

func (s *Server) Logout(u *User) error {
	if u == nil {
		return nil
	}

	err := s.storage.RemoveUser(u)
	if err != nil {
		return err
	}

	err = s.storage.Publish(LeaveChannel, []byte(u.Name))
	if err != nil {
		log.Println(err)
	}

	return nil
}

func (s *Server) SaveMessage(m *Message) error {
	err := s.storage.SaveMessage(m)
	if err != nil {
		return err
	}

	v, err := json.Marshal(m)
	if err != nil {
		return err
	}

	err = s.storage.Publish(MessageChannel, v)
	if err != nil {
		return err
	}

	return nil
}

func (s *Server) Storage() *Storage {
	return s.storage
}

func (s *Server) Config() Config {
	return s.config
}
