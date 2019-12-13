package chat

import (
	"testing"
)

func TestClient_SetUser(t *testing.T) {
	c := &Client{}
	u := &User{"user"}
	c.SetUser(u)
	if c.user == nil {
		t.Errorf("nil user found")
	}
}
