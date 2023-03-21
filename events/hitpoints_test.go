package events

import (
	"fmt"
	"github.com/cpustejovsky/event-store/protos/hitpoints"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"testing"
)

func TestHitPoints_Aggregate(t *testing.T) {
	name := "cpustejovsky"
	id := uuid.NewString()
	hitPointEvents := []hitpoints.PlayerCharacterHitPoints{{
		Id:                 id,
		CharacterName:      name,
		CharacterHitPoints: 8,
		Note:               "Init",
	}, {
		Id:                 id,
		CharacterName:      name,
		CharacterHitPoints: -2,
		Note:               "Slashing damage from goblin",
	}, {
		Id:                 id,
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
		want.CharacterHitPoints += e.GetCharacterHitPoints()
		want.Id = e.GetId()
		want.CharacterName = e.GetCharacterName()
		want.Note += fmt.Sprintf("hit point change of %d with note '%s'", e.GetCharacterHitPoints(), e.GetNote())
	}
	hp := HitPoints{}
	got := &hitpoints.PlayerCharacterHitPoints{}
	gotbin, err := hp.Aggregate(bins)
	require.Nil(t, err)
	err = proto.Unmarshal(gotbin, got)
	assert.Equal(t, want.GetId(), got.GetId())
	assert.Equal(t, want.GetCharacterName(), got.GetCharacterName())
	assert.Equal(t, want.GetCharacterHitPoints(), got.GetCharacterHitPoints())
	assert.Equal(t, want.GetNote(), got.GetNote())
}
