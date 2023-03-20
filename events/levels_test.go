package events

import (
	pb "github.com/cpustejovsky/event-store/protos/levels"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"testing"
)

func TestLevels_Aggregate(t *testing.T) {
	name := "cpustejovsky"
	t.Run("Aggregate Levels events with milestone system", func(t *testing.T) {
		var l int32 = 1
		levelEvents := []pb.Level{
			{
				CharacterName: name,
				LevelType:     pb.LevelType_Milestone,
				Experience:    nil,
				Levels:        &l,
			},
			{
				CharacterName: name,
				LevelType:     pb.LevelType_Milestone,
				Experience:    nil,
				Levels:        &l,
			},
			{
				CharacterName: name,
				LevelType:     pb.LevelType_Milestone,
				Experience:    nil,
				Levels:        &l,
			},
		}
		var bins [][]byte
		want := &pb.Level{}
		for _, e := range levelEvents {
			bin, err := proto.Marshal(&e)
			if err != nil {
				t.Fatal(err)
			}
			bins = append(bins, bin)
			sum := want.GetLevels() + e.GetLevels()
			want.Levels = &sum
			want.CharacterName = e.CharacterName
		}
		lvl := Levels{}
		got := &pb.Level{}
		gotbin, err := lvl.Aggregate(bins)
		require.Nil(t, err)
		err = proto.Unmarshal(gotbin, got)
		assert.Equal(t, want.CharacterName, got.CharacterName)
		assert.Equal(t, want.GetLevels(), got.GetLevels())
	})
	t.Run("Aggregate Levels events with XP system", func(t *testing.T) {
		var exp1 int32 = 50
		var exp2 int32 = 25
		var exp3 int32 = 75
		levelEvents := []pb.Level{
			{
				CharacterName: name,
				LevelType:     pb.LevelType_XP,
				Experience:    nil,
				Levels:        &exp1,
			},
			{
				CharacterName: name,
				LevelType:     pb.LevelType_XP,
				Experience:    nil,
				Levels:        &exp2,
			},
			{
				CharacterName: name,
				LevelType:     pb.LevelType_XP,
				Experience:    nil,
				Levels:        &exp3,
			},
		}
		var bins [][]byte
		want := &pb.Level{}
		for _, e := range levelEvents {
			bin, err := proto.Marshal(&e)
			if err != nil {
				t.Fatal(err)
			}
			bins = append(bins, bin)
			sum := want.GetExperience() + e.GetExperience()
			want.Experience = &sum
			want.CharacterName = e.CharacterName
		}
		lvl := Levels{}
		got := &pb.Level{}
		gotbin, err := lvl.Aggregate(bins)
		require.Nil(t, err)
		err = proto.Unmarshal(gotbin, got)
		assert.Equal(t, want.CharacterName, got.CharacterName)
		assert.Equal(t, want.GetExperience(), got.GetExperience())
	})
	t.Run("Aggregate Levels events with XP system", func(t *testing.T) {
		var l int32 = 1
		var exp1 int32 = 50
		var exp2 int32 = 75
		levelEvents := []pb.Level{
			{
				CharacterName: name,
				LevelType:     pb.LevelType_XP,
				Experience:    nil,
				Levels:        &exp1,
			},
			{
				CharacterName: name,
				LevelType:     pb.LevelType_Milestone,
				Experience:    nil,
				Levels:        &l,
			},
			{
				CharacterName: name,
				LevelType:     pb.LevelType_XP,
				Experience:    nil,
				Levels:        &exp2,
			},
		}
		var bins [][]byte
		want := &pb.Level{}
		for _, e := range levelEvents {
			bin, err := proto.Marshal(&e)
			if err != nil {
				t.Fatal(err)
			}
			bins = append(bins, bin)
			sum := want.GetExperience() + e.GetExperience()
			want.Experience = &sum
			want.CharacterName = e.CharacterName
		}
		lvl := Levels{}
		_, err := lvl.Aggregate(bins)
		require.Error(t, err)
	})
}
