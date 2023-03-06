package events

import (
	"fmt"
	"github.com/cpustejovsky/event-store/protos/hitpoints"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"testing"
)

func TestHitPoints_Aggregate(t *testing.T) {
	name := "cpustejovsky"
	hitPointEvents := []hitpoints.PlayerCharacterHitPoints{{
		CharacterName:      name,
		CharacterHitPoints: 8,
		Note:               "Init",
	}, {
		CharacterName:      name,
		CharacterHitPoints: -2,
		Note:               "Slashing damage from goblin",
	}, {
		CharacterName:      name,
		CharacterHitPoints: -3,
		Note:               "bludgeoning damage from bugbear",
	}}
	var bins [][]byte
	want := &hitpoints.PlayerCharacterHitPoints{}
	want.Note = "Aggregated Notes: "
	for _, e := range hitPointEvents {
		bin, err := proto.Marshal(&e)
		if err != nil {
			t.Fatal(err)
		}
		bins = append(bins, bin)
		want.CharacterHitPoints += e.CharacterHitPoints
		want.CharacterName = e.CharacterName
		want.Note += fmt.Sprintf("hit point change of %d with note '%s'", e.GetCharacterHitPoints(), e.GetNote())
	}
	hp := HitPoints{}
	got := &hitpoints.PlayerCharacterHitPoints{}
	gotbin, err := hp.Aggregate(bins)
	require.Nil(t, err)
	err = proto.Unmarshal(gotbin, got)
	assert.Equal(t, want.CharacterName, got.CharacterName)
	assert.Equal(t, want.CharacterHitPoints, got.CharacterHitPoints)
	assert.Equal(t, want.Note, got.Note)
}
