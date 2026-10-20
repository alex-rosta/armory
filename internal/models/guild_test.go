package models

import (
	"errors"
	"testing"
	"wowarmory/internal/api"
)

func TestNewGuildData(t *testing.T) {
	t.Run("wrong type does not panic", func(t *testing.T) {
		if _, err := NewGuildData(map[string]interface{}{}, "eu", "darkspear"); err == nil {
			t.Fatal("expected error")
		}
		if _, err := NewGuildData((*api.GuildResponse)(nil), "eu", "darkspear"); err == nil {
			t.Fatal("expected error for nil response")
		}
	})

	t.Run("missing guild is ErrGuildNotFound", func(t *testing.T) {
		_, err := NewGuildData(&api.GuildResponse{}, "eu", "darkspear")
		if !errors.Is(err, ErrGuildNotFound) {
			t.Fatalf("err = %v, want ErrGuildNotFound", err)
		}
	})

	t.Run("members are deduplicated and sorted", func(t *testing.T) {
		resp := &api.GuildResponse{}
		resp.GuildData.Guild.Name = "Test"
		resp.GuildData.Guild.Members.Total = 3
		resp.GuildData.Guild.Attendance.Data = []api.AttendanceEntry{{
			Players: []api.GuildMember{{Name: "Zed", Type: "Mage"}, {Name: "Amy", Type: "Priest"}, {Name: "Zed", Type: "Mage"}},
		}}

		got, err := NewGuildData(resp, "eu", "darkspear")
		if err != nil {
			t.Fatal(err)
		}
		if len(got.Members) != 2 || got.Members[0].Name != "Amy" || got.Members[1].Name != "Zed" {
			t.Fatalf("members = %+v", got.Members)
		}
		if got.MemberCount != 3 || got.Region != "eu" || got.Realm != "darkspear" {
			t.Fatalf("unexpected data: %+v", got)
		}
	})
}
