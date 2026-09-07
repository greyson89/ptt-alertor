package pushsum

import (
	"os"
	"reflect"
	"sort"
	"testing"

	"github.com/watain666/ptt-alertor/connections"
)

func TestMain(m *testing.M) {
	os.Setenv("SQLITE_PATH", ":memory:")
	os.Exit(m.Run())
}

func resetDiffTable(t *testing.T) {
	t.Helper()
	if _, err := connections.DB().Exec("DELETE FROM pushsum_diff_ids"); err != nil {
		t.Fatal(err)
	}
}

func sorted(ids []int) []int {
	sort.Ints(ids)
	return ids
}

func TestDiffList(t *testing.T) {
	resetDiffTable(t)

	// First call ever for this account/board/kind only establishes the
	// baseline; nothing should be reported even though every id is "new".
	if got := DiffList("alice", "gossiping", "up", 1, 2, 3); len(got) != 0 {
		t.Errorf("first DiffList() = %v, want empty (baseline only)", got)
	}

	// Same ids again: still known, nothing new to report.
	if got := DiffList("alice", "gossiping", "up", 1, 2, 3); len(got) != 0 {
		t.Errorf("repeat DiffList() = %v, want empty", got)
	}

	// A genuinely new id shows up alongside old ones: only the new one is reported.
	got := DiffList("alice", "gossiping", "up", 1, 2, 3, 4)
	if !reflect.DeepEqual(sorted(got), []int{4}) {
		t.Errorf("DiffList() with new id = %v, want [4]", got)
	}

	// It's now known too, so it won't be reported again.
	if got := DiffList("alice", "gossiping", "up", 1, 2, 3, 4); len(got) != 0 {
		t.Errorf("DiffList() after learning 4 = %v, want empty", got)
	}
}

func TestDiffList_EmptyIDs(t *testing.T) {
	resetDiffTable(t)
	if got := DiffList("alice", "gossiping", "up"); len(got) != 0 {
		t.Errorf("DiffList() with no ids = %v, want empty", got)
	}
}

func TestReplaceBenchKeys(t *testing.T) {
	resetDiffTable(t)

	// Establish a base with id 1.
	DiffList("alice", "gossiping", "up", 1)

	// Rotate base -> bench. The article is still suppressed via bench.
	if err := ReplaceBenchKeys(); err != nil {
		t.Fatalf("ReplaceBenchKeys() error = %v", err)
	}

	// Right after the rotation there is no base yet, so even a brand new id
	// only seeds the baseline and nothing is reported.
	if got := DiffList("alice", "gossiping", "up", 1, 2); len(got) != 0 {
		t.Errorf("DiffList() right after rotation = %v, want empty", got)
	}

	// Id 1 stays suppressed (still remembered via bench), id 2 was already
	// learned into the new base by the previous call, so nothing new here...
	if got := DiffList("alice", "gossiping", "up", 1, 2); len(got) != 0 {
		t.Errorf("DiffList() = %v, want empty", got)
	}

	// ...but a truly new id is reported.
	got := DiffList("alice", "gossiping", "up", 1, 2, 3)
	if !reflect.DeepEqual(sorted(got), []int{3}) {
		t.Errorf("DiffList() = %v, want [3]", got)
	}
}

func TestDelDiffList(t *testing.T) {
	resetDiffTable(t)
	DiffList("alice", "gossiping", "up", 1)
	DiffList("alice", "gossiping", "down", 1)

	if err := DelDiffList("alice", "gossiping", "up"); err != nil {
		t.Fatalf("DelDiffList() error = %v", err)
	}

	// "up" baseline is gone, so this call re-establishes it silently again.
	if got := DiffList("alice", "gossiping", "up", 1); len(got) != 0 {
		t.Errorf("DiffList() after DelDiffList = %v, want empty", got)
	}
	// "down" baseline is untouched, so id 1 is already known there.
	if got := DiffList("alice", "gossiping", "down", 1); len(got) != 0 {
		t.Errorf("DiffList() for untouched kind = %v, want empty", got)
	}
}

func TestBoardsAndSubscribers(t *testing.T) {
	db := connections.DB()
	db.Exec("DELETE FROM pushsum_boards")
	db.Exec("DELETE FROM pushsum_subscribers")

	if Exist("gossiping") {
		t.Fatal("Exist() = true before Add()")
	}
	if err := Add("gossiping"); err != nil {
		t.Fatalf("Add() error = %v", err)
	}
	if !Exist("gossiping") {
		t.Fatal("Exist() = false after Add()")
	}
	if got := List(); !reflect.DeepEqual(got, []string{"gossiping"}) {
		t.Errorf("List() = %v, want [gossiping]", got)
	}

	if err := AddSubscriber("gossiping", "alice"); err != nil {
		t.Fatalf("AddSubscriber() error = %v", err)
	}
	if got := ListSubscribers("gossiping"); !reflect.DeepEqual(got, []string{"alice"}) {
		t.Errorf("ListSubscribers() = %v, want [alice]", got)
	}

	if err := RemoveSubscriber("gossiping", "alice"); err != nil {
		t.Fatalf("RemoveSubscriber() error = %v", err)
	}
	if got := ListSubscribers("gossiping"); len(got) != 0 {
		t.Errorf("ListSubscribers() after remove = %v, want empty", got)
	}

	if err := Remove("gossiping"); err != nil {
		t.Fatalf("Remove() error = %v", err)
	}
	if Exist("gossiping") {
		t.Fatal("Exist() = true after Remove()")
	}
}
