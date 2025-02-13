package repositories

import (
	"sync"
)

type Identificable interface {
	GetId() int
	SetId(id int)
}

type EntitiesMemoryStorage[E Identificable] struct {
	entitiesById map[int]E
	idSequence  int
	mutex       sync.Mutex
}

func NewEntitiesMemoryStorage[E Identificable]() *EntitiesMemoryStorage[E] {
	return &EntitiesMemoryStorage[E]{entitiesById: make(map[int]E), idSequence: 0}
}

func (repo *EntitiesMemoryStorage[E]) GetAll() ([]E, error) {
	repo.mutex.Lock()
	defer repo.mutex.Unlock()
	entities := make([]E, 0, len(repo.entitiesById))
	for _, entity := range repo.entitiesById {
		entities = append(entities, entity)
	}
	return entities, nil
}

func (repo *EntitiesMemoryStorage[E]) GetById(id int) (E, error) {
	repo.mutex.Lock()
	defer repo.mutex.Unlock()
	entity, exists := repo.entitiesById[id]
	if !exists {
		var zeroValue E 
		return zeroValue, EntityNotExistsErr
	}
	return entity, nil
}

func (repo *EntitiesMemoryStorage[E]) Save(entity E) (E, error) {
	repo.mutex.Lock()
	defer repo.mutex.Unlock()
	nextId := repo.idSequence + 1
	entity.SetId(nextId)
	repo.entitiesById[nextId] = entity
	repo.idSequence++
	return entity, nil
}

func (repo *EntitiesMemoryStorage[E]) Update(entity E) (E, error) {
	repo.mutex.Lock()
	defer repo.mutex.Unlock()
	entity, exists := repo.entitiesById[entity.GetId()] 
	if !exists {
		var zeroValue E 
		return zeroValue, EntityNotExistsErr
	}
	repo.entitiesById[entity.GetId()] = entity
	return entity, nil
}

func (repo *EntitiesMemoryStorage[E]) Delete(id int) error {
	repo.mutex.Lock()
	defer repo.mutex.Unlock()
	delete(repo.entitiesById, id)
	return nil
}
