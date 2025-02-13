package repositories

import (
	"errors"
)

var EntityNotExistsErr error = errors.New("Entity doesn't exists")
var DuplicatedEntityErr error = errors.New("Duplicated Entity")
var InvalidEntityStateErr error = errors.New("Entity state is invalid")

type EntitiesRepository[E Identificable] interface {
	GetAll() ([]E, error)
	GetById(id int) (E, error)
	Save(entity E) (E, error)
	Update(entity E) (E, error)
	Delete(id int) error
}
