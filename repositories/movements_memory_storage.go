package repositories

import (
		"github.com/vituchon/splitify/model"

)

type MovementsRepository interface {
	EntitiesRepository[*model.Movement]
	GetByGroupId(groupId int) ([]*model.Movement, error)
}

type MovementsMemoryRepository struct {
	*EntitiesMemoryStorage[*model.Movement]
}

func NewMovementsMemoryRepository() *MovementsMemoryRepository {
	return &MovementsMemoryRepository{
		EntitiesMemoryStorage: NewEntitiesMemoryStorage[*model.Movement](),
	}
}

func (repo *MovementsMemoryRepository) GetByGroupId(groupId int) ([]*model.Movement, error) {
	repo.mutex.Lock()
	defer repo.mutex.Unlock()

	var movements []*model.Movement
	for _, movement := range repo.entitiesById {
		if movement.GroupId == groupId { 
			movements = append(movements, movement)
		}
	}
	return movements, nil
}
