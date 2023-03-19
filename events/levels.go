package events

import (
	"errors"
	pb "github.com/cpustejovsky/event-store/protos/levels"
	"google.golang.org/protobuf/proto"
)

type Level struct {
	pb.Level
}

type InconsistentLevelTypeError struct {
}

func (l *Level) Aggregate(events [][]byte) ([]byte, error) {
	var lvl pb.Level
	m := make(map[pb.LevelType]struct{})
	for _, event := range events {
		err := proto.Unmarshal(event, &lvl)
		if err != nil {
			return nil, err
		}
		m[lvl.GetLevelType()] = struct{}{}
		lvlSum := l.GetLevels() + lvl.GetLevels()
		l.Levels = &lvlSum
		expSum := l.GetExperience() + lvl.GetExperience()
		l.Experience = &expSum
		l.CharacterName = lvl.CharacterName
		l.LevelType = lvl.GetLevelType()
	}
	_, milestone := m[pb.LevelType_Milestone]
	_, xp := m[pb.LevelType_XP]
	_, empty := m[pb.LevelType_Empty]
	if (milestone && xp) || empty {
		return nil, errors.New("either multiple leveling systems used or no leveling system provided")
	}
	bin, err := proto.Marshal(l)
	if err != nil {
		return nil, err
	}
	return bin, nil
}
