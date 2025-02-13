package repositories

import (
		"github.com/vituchon/splitify/model"

)


type ParticipantMovementsRepository interface {
	EntitiesRepository[*model.ParticipantMovement]
	GetByMovementId(movementId int) ([]*model.ParticipantMovement, error)
}

type ParticipantMovementsMemoryRepository struct {
	*EntitiesMemoryStorage[*model.ParticipantMovement]
}

func NewParticipantMovementsMemoryRepository() *ParticipantMovementsMemoryRepository {
	return &ParticipantMovementsMemoryRepository{
		EntitiesMemoryStorage: NewEntitiesMemoryStorage[*model.ParticipantMovement](),
	}
}

func (repo *ParticipantMovementsMemoryRepository) GetByMovementId(movementId int) ([]*model.ParticipantMovement, error) {
	repo.mutex.Lock()
	defer repo.mutex.Unlock()

	var participantMovements []*model.ParticipantMovement
	for _, participantMovement := range repo.entitiesById {
		if participantMovement.MovementId == movementId { 
			participantMovements = append(participantMovements, participantMovement)
		}
	}
	return participantMovements, nil
}
