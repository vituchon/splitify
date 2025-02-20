package api

import (
	//"encoding/json"
	"fmt"
	"github.com/vituchon/splitify/model"
	"github.com/vituchon/splitify/repositories"
	"github.com/vituchon/splitify/util"
	"time"
)

var (
	groupsRepository               repositories.EntitiesRepository[*model.Group]
	participantsRepository         repositories.ParticipantsRepository
	movementsRepository            repositories.MovementsRepository
	participantMovementsRepository repositories.ParticipantMovementsRepository
)

func init() {
	groupsRepository = repositories.NewEntitiesMemoryStorage[*model.Group]()
	participantsRepository = repositories.NewParticipantsMemoryRepository()
	movementsRepository = repositories.NewMovementsMemoryRepository()
	participantMovementsRepository = repositories.NewParticipantMovementsMemoryRepository()
}

func CreateGroup(name string) (*model.Group, error) {
	group := &model.Group{
		Name: name,
	}
	return groupsRepository.Save(group)
}

func GetAllGroups() ([]*model.Group, error) {
	return groupsRepository.GetAll()
}

type Participant struct {
	GroupId int    `json:"GroupId"`
	Name    string `json:"name"`
}

func AddParticipant(participant Participant) (*model.Participant, error) {
	_, err := groupsRepository.GetById(participant.GroupId)
	if err != nil {
		return nil, err
	}
	p := &model.Participant{
		GroupId: participant.GroupId,
		Name:    participant.Name,
	}
	return participantsRepository.Save(p)
}

func GetParticipants(groupId int) ([]*model.Participant, error) {
	return participantsRepository.GetByGroupId(groupId)
}


type ParticipantMovement struct {
	ParticipantId int         `json:"participantId"`
	Amount        model.Price `json:"amount"`
}

type Movement struct {
	GroupId              int                   `json:"groupId"`
	Amount               model.Price           `json:"amount"`
	Concept              string                `json:"concept"`
	ParticipantMovements []ParticipantMovement `json:"participantMovement"`
}

func GetMovements(groupId int) ([]*model.Movement, error) {
	return movementsRepository.GetByGroupId(groupId)
}

func GetParticipantMovements(movementId int) ([]*model.ParticipantMovement, error) {
	return participantMovementsRepository.GetByMovementId(movementId)
}

func AddMovement(movement Movement) (*model.Movement, []*model.ParticipantMovement, error) {
	_, err := groupsRepository.GetById(movement.GroupId)
	if err != nil {
		return nil, nil, err
	}
	for _, participantMovement := range movement.ParticipantMovements {
		participant, err := participantsRepository.GetById(participantMovement.ParticipantId)
		if err != nil {
			return nil, nil, err
		}
		if participant.GroupId != movement.GroupId {
			return nil, nil, fmt.Errorf("Participant(id='%d') doesnt belong to movement's group(id='%d')", participant.Id, movement.GroupId)
		}
	}
	m := &model.Movement{
		GroupId:   movement.GroupId,
		Amount:    movement.Amount,
		CreatedAt: time.Now().Unix(),
		Concept:   movement.Concept,
	}
	m, err = movementsRepository.Save(m)
	if err != nil {
		return nil, nil, err
	}

	pms := make([]*model.ParticipantMovement, 0, len(movement.ParticipantMovements))
	for _, participantMovement := range movement.ParticipantMovements {
		pm := &model.ParticipantMovement{
			MovementId:    m.Id,
			ParticipantId: participantMovement.ParticipantId,
			Amount:        participantMovement.Amount,
		}
		pm, err = participantMovementsRepository.Save(pm)
		if err != nil {
			// TODO: must rollback   movementsRepository.Save(m), in memory repository shall have a transcation mechanishm
			return nil, nil, err
		}
		pms = append(pms, pm)
	}
	return m, pms, nil
}

func CalculateBalances(groupId int) (model.DebitCreditMap, model.ParticipantShareByParticipantId, error) {
	_, err := groupsRepository.GetById(groupId)
	if err != nil {
		return nil, nil, err
	}

	movements, err := movementsRepository.GetByGroupId(groupId)
	if err != nil {
		return nil, nil, err
	}

	acumulatedBalance := make(model.DebitCreditMap)
	acumulatedShare := make(model.ParticipantShareByParticipantId)
	for _, movement := range movements {
		participantMovementsPtr, err := participantMovementsRepository.GetByMovementId(movement.Id)
		if err != nil {
			return nil, nil, err
		}

		participantMovements := util.ToValues(participantMovementsPtr)
		model.EnsureMovementAmountMatchesParticipantAmounts(*movement, participantMovements)
		if err != nil {
			return nil, nil, err
		}
		participantShareByParticipantId := model.BuildParticipantsEqualShare(*movement, participantMovements)
		err = model.EnsureSharesSumToZero(participantShareByParticipantId)
		if err != nil {
			return nil, nil, err
		}
		acumulatedShare = model.SumParticipantShares(acumulatedShare, participantShareByParticipantId)

		balance := model.BuildDebitCreditMap(participantMovements, participantShareByParticipantId)
		acumulatedBalance = model.SumDebitCreditMaps(acumulatedBalance, balance)
	}
	return acumulatedBalance, acumulatedShare, nil
}


func CalculateBalance(groupId int, movementId int) (model.DebitCreditMap, model.ParticipantShareByParticipantId, error) {
	_, err := groupsRepository.GetById(groupId)
	if err != nil {
		return nil, nil, err
	}

	movement, err := movementsRepository.GetById(movementId)
	if err != nil {
		return nil, nil, err
	}

	participantMovementsPtr, err := participantMovementsRepository.GetByMovementId(movement.Id)
	if err != nil {
		return nil, nil, err
	}

	participantMovements := util.ToValues(participantMovementsPtr)

	err = model.EnsureMovementAmountMatchesParticipantAmounts(*movement, participantMovements)
	if err != nil {
		return nil, nil, err
	}

	shares := model.BuildParticipantsEqualShare(*movement, participantMovements)
	err = model.EnsureSharesSumToZero(shares)
	if err != nil {
		return nil, nil, err
	}
	/*a, _ := json.Marshal(*movement)
	b, _ := json.Marshal(participantMovements)
	c, _ := json.Marshal(shares)
	fmt.Println(string(a), "\n", string(b), "\nShares:", string(c))*/
	balance := model.BuildDebitCreditMap(participantMovements, shares)
	return balance, shares, nil
}

