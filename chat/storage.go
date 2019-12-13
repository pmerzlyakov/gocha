package chat

import (
	"encoding/json"
	"errors"
	"github.com/gomodule/redigo/redis"
	"log"
	"sort"
)

const (
	JoinChannel    = "join"
	LeaveChannel   = "leave"
	MessageChannel = "message"

	usersKey       = "users"
	messagesPrefix = "messages"
	roomsPrefix    = "rooms"
)

type StorageMessage struct {
	Channel string
	Data    []byte
}

type Storage struct {
	pool *redis.Pool
}

func NewStorage(addr string) *Storage {
	pool := &redis.Pool{
		Dial: func() (conn redis.Conn, err error) {
			return redis.Dial("tcp", addr)
		},
	}

	return &Storage{pool}
}

func (s *Storage) Publish(channel string, msg []byte) error {
	conn := s.pool.Get()
	defer conn.Close()

	_, err := conn.Do("PUBLISH", channel, msg)
	return err
}

func (s *Storage) Subscribe(ch chan<- *StorageMessage) error {
	conn := s.pool.Get()
	psc := redis.PubSubConn{Conn: conn}

	err := psc.Subscribe(JoinChannel, LeaveChannel, MessageChannel)
	if err != nil {
		return err
	}

	go func() {
		defer func() {
			psc.Unsubscribe()
			psc.Close()
		}()
		for {
			switch v := psc.Receive().(type) {
			case redis.Message:
				ch <- &StorageMessage{v.Channel, v.Data}
			case error:
				close(ch)
				return
			}
		}
	}()

	return nil
}

func (s *Storage) AddUser(u *User) error {
	conn := s.pool.Get()
	defer conn.Close()

	n, err := redis.Int(conn.Do("SADD", usersKey, u.Name))
	if err != nil {
		return err
	}
	if n == 0 {
		return errors.New("user with this name already logged in")
	}

	return nil
}

func (s *Storage) RemoveUser(u *User) error {
	conn := s.pool.Get()
	defer conn.Close()

	_, err := conn.Do("SREM", usersKey, u.Name)
	return err
}

func (s *Storage) ActiveUsers() []string {
	conn := s.pool.Get()
	defer conn.Close()

	names, err := redis.Strings(conn.Do("SMEMBERS", usersKey))
	if err != nil {
		log.Println(err)
		return nil
	}

	return names
}

func (s *Storage) SaveMessage(m *Message) error {
	conn := s.pool.Get()
	defer conn.Close()

	room := m.Room()
	v, _ := json.Marshal(m)

	conn.Send("MULTI")
	conn.Send("LPUSH", messagesKey(room), v)
	if m.Recipient != "" {
		conn.Send("SADD", roomsKey(m.Author), m.Recipient)
		conn.Send("SADD", roomsKey(m.Recipient), m.Author)
	}
	_, err := conn.Do("EXEC")
	if err != nil {
		return err
	}

	return nil
}

func (s *Storage) Rooms(u *User) []string {
	conn := s.pool.Get()
	defer conn.Close()

	rooms, err := redis.Strings(conn.Do("SMEMBERS", roomsKey(u.Name)))
	if err != nil {
		log.Println(err)
		return nil
	}
	rooms = append(rooms, "")
	sort.Strings(rooms)

	return rooms
}

func (s *Storage) MessageHistory(room string, size uint) []*Message {
	conn := s.pool.Get()
	defer conn.Close()

	messages := make([]*Message, 0, size)

	data, err := redis.ByteSlices(conn.Do("LRANGE", messagesKey(room), 0, size-1))
	if err != nil {
		return nil
	}

	for _, m := range data {
		msg := new(Message)
		if err := json.Unmarshal(m, msg); err == nil {
			messages = append(messages, msg)
		}
	}

	return messages
}

func roomsKey(name string) string {
	return roomsPrefix + ":" + name
}

func messagesKey(room string) string {
	if room == "" {
		return messagesPrefix
	}
	return messagesPrefix + ":" + room
}
