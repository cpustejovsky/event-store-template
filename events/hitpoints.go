package events

import (
	"fmt"
	"github.com/cpustejovsky/event-store/protos/hitpoints"
	"google.golang.org/protobuf/proto"
)

type HitPoints struct {
	hitpoints.PlayerCharacterHitPoints
}

func (h *HitPoints) Aggregate(events [][]byte) ([]byte, error) {
	h.Note = "Aggregated Notes: "
	var hp hitpoints.PlayerCharacterHitPoints
	for _, event := range events {
		err := proto.Unmarshal(event, &hp)
		if err != nil {
			return nil, err
		}
		h.CharacterName = hp.CharacterName
		h.CharacterHitPoints += hp.CharacterHitPoints
		h.Note += fmt.Sprintf("hit point change of %d with note '%s'", hp.GetCharacterHitPoints(), hp.GetNote())
	}
	bin, err := proto.Marshal(h)
	if err != nil {
		return nil, err
	}
	return bin, nil
}
