package chat

import (
	"testing"
)

func TestServer_AddClient(t *testing.T) {
	client := new(Client)
	s := &Server{clients: make(map[*Client]struct{})}
	s.AddClient(client)
	if len(s.clients) == 0 {
		t.Error("client list is empty after adding in it")
	}
	_, ok := s.clients[client]
	if !ok {
		t.Error("can't access to added client")
	}
}

func TestServer_RemoveClient(t *testing.T) {
	client := new(Client)
	s := &Server{clients: map[*Client]struct{}{
		client: {},
	}}

	s.RemoveClient(client)
	if len(s.clients) > 0 {
		t.Error("client list's length don't change after removing from it")
	}
	_, ok := s.clients[client]
	if ok {
		t.Error("can access to removed client")
	}
}
