package chat

import "testing"

func TestMessage_Room(t *testing.T) {
	cases := []struct {
		msg      Message
		wantRoom string
	}{
		{
			Message{Author: "A", Recipient: "B"},
			"A:B",
		},
		{
			Message{Author: "B", Recipient: "A"},
			"A:B",
		},
		{
			Message{Author: "A", Recipient: ""},
			"",
		},
	}

	for _, c := range cases {
		room := c.msg.Room()
		if room != c.wantRoom {
			t.Errorf("For message %+v room `%s` expect, but `%s` found", c.msg, c.wantRoom, room)
		}
	}
}
