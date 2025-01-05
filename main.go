package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/fiatjaf/eventstore/sqlite3"
	"github.com/fiatjaf/khatru"
	"github.com/nbd-wtf/go-nostr"
)

type TreeNode struct {
	Parent   string
	Level    int
}

var tribe = make(map[string]TreeNode)

func rejectNonMember(ctx context.Context, event *nostr.Event) (reject bool, msg string) {
	_, ok := tribe[event.PubKey]
	if ok {
		return false, ""
	}
	return true, "Not a member"
}

func updateTribe(ctx context.Context, stamp *nostr.Event) error {
	if stamp.Kind != 77 {
		return nil
	}
	ptag := stamp.Tags.GetFirst([]string{"p", ""})
	stamped := ptag.Value()
	_, ok := tribe[stamped]
	if !ok {
		parentnode, ok2 := tribe[stamp.PubKey]
		if !ok2 {
			panic(errors.New("some strange error"))
		}
		tribe[stamped] = TreeNode{Level: parentnode.Level + 1, Parent: stamp.PubKey}
		fmt.Println("> has", len(tribe), "members")
	}
	return nil
}

func mapkeys(m map[string]bool) []string {
	r := []string{}
	for k, _ := range m {
		r = append(r, k)
	}
	return r
}

func populateTribe(leader string, tribeid string, db sqlite3.SQLite3Backend) {
	tribe[leader] = TreeNode{Level: 0, Parent: ""}
	level := 1
	stampers := map[string]bool{leader: true}
	for level < 21 && len(stampers) > 0 {
		stampers_list := mapkeys(stampers)
		ch, err := db.QueryEvents(context.TODO(),
					nostr.Filter{ Authors: stampers_list, Tags: nostr.TagMap{"c": []string{tribeid}}, Kinds: []int{77} })
		if err != nil {
			panic(err)
		}
		stampers = make(map[string]bool)
		for stamp := range ch {
			ptag := stamp.Tags.GetFirst([]string{"p", ""})
			stamped := ptag.Value()

			// refuse to add member at a lower level if already added
			_, ok := tribe[stamped]
			if !ok {
				tribe[stamped] = TreeNode{Level: level, Parent: stamp.PubKey}
				stampers[stamped] = true
			}
		}
		fmt.Println("level:", level, "stamped:", len(stampers), "members:", len(tribe))
		level = level + 1
	}
	fmt.Println("Tribe has", len(tribe), "members")
}

func main() {
	// create the relay instance
	relay := khatru.NewRelay()
	leader := os.Args[1]
	tribeid := os.Args[2]

	// set up some basic properties (will be returned on the NIP-11 endpoint)
	relay.Info.Name = "tribe relay"
	relay.Info.PubKey = leader
	relay.Info.Description = "yet another pyramid scheme home of "+tribeid
	relay.Info.Icon = "https://nostr.hu/kobuki.jpg"

	db := sqlite3.SQLite3Backend{DatabaseURL: "./khatru-tribe.sqlite"}
	if err := db.Init(); err != nil {
		panic(err)
	}
	populateTribe(leader, tribeid, db)

	relay.StoreEvent = append(relay.StoreEvent, updateTribe, db.SaveEvent)
	relay.QueryEvents = append(relay.QueryEvents, db.QueryEvents)
	relay.CountEvents = append(relay.CountEvents, db.CountEvents)
	relay.DeleteEvent = append(relay.DeleteEvent, db.DeleteEvent)
	relay.RejectEvent = append(relay.RejectEvent, rejectNonMember)

	// start the server
	fmt.Println("running on :3334")
	http.ListenAndServe(":3334", relay)
}
