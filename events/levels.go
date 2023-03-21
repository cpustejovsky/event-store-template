package events

import (
	"errors"
	pb "github.com/cpustejovsky/event-store/protos/levels"
	"google.golang.org/protobuf/proto"
)

type Levels struct {
	pb.Level
}

func (l *Levels) Aggregate(events [][]byte) ([]byte, error) {
	var lvl pb.Level
	var levelEvents []pb.Level
	m := make(map[pb.LevelType]struct{})
	//Unmarshal events; add LevelType to map
	for _, event := range events {
		err := proto.Unmarshal(event, &lvl)
		if err != nil {
			return nil, err
		}
		m[lvl.GetLevelType()] = struct{}{}
		levelEvents = append(levelEvents, lvl)
	}
	//Make check for consistent, non-empty level type
	_, milestone := m[pb.LevelType_Milestone]
	_, xp := m[pb.LevelType_XP]
	_, empty := m[pb.LevelType_Empty]
	if (milestone && xp) || empty {
		return nil, errors.New("either multiple leveling systems used or no leveling system provided")
	}

	//Aggregate events
	switch levelEvents[0].GetLevelType() {
	case pb.LevelType_XP:
		l.aggregateXP(levelEvents)
	case pb.LevelType_Milestone:
		l.aggregateMilestone(levelEvents)
	}
	//Apply static properties to aggregate
	first := levelEvents[0]
	l.Id = first.GetId()
	l.CharacterName = first.GetCharacterName()
	l.LevelType = first.GetLevelType()
	bin, err := proto.Marshal(l)
	if err != nil {
		return nil, err
	}
	return bin, nil
}

func (l *Levels) aggregateXP(events []pb.Level) {
	for _, e := range events {
		expSum := l.GetExperience() + e.GetExperience()
		l.Experience = &expSum
	}
	l.Levels = nil
}

func (l *Levels) aggregateMilestone(events []pb.Level) {
	for _, e := range events {
		lvlSum := l.GetLevels() + e.GetLevels()
		l.Levels = &lvlSum
	}
	l.Experience = nil
}
