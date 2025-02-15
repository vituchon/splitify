package repositories

import (
		"github.com/vituchon/splitify/model"

)

type ParticipantsRepository interface {
	EntitiesRepository[*model.Participant]
	GetByGroupId(groupId int) ([]*model.Participant, error)
}

type ParticipantsMemoryRepository struct {
	*EntitiesMemoryStorage[*model.Participant]
}

func NewParticipantsMemoryRepository() *ParticipantsMemoryRepository {
	return &ParticipantsMemoryRepository{
		EntitiesMemoryStorage: NewEntitiesMemoryStorage[*model.Participant](),
	}
}

func (repo *ParticipantsMemoryRepository) GetByGroupId(groupId int) ([]*model.Participant, error) {
	repo.mutex.Lock()
	defer repo.mutex.Unlock()

	var participants []*model.Participant
	for _, participant := range repo.entitiesById {
		if participant.GroupId == groupId { 
			participants = append(participants, participant)
		}
	}
	return participants, nil
}
