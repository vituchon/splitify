package model

import (
	"errors"
	"sort"
)

type Movement struct {
	Id        int    `json:"id"`
	GroupId   int    `json:"groupId"`
	CreatedAt int64  `json:"createdAt"` // unix timestamp, in seconds since epoch
	Amount    Price  `json:"amount"`
	Concept   string `json:"concept"`
}

func (movement Movement) GetId() int {
	return movement.Id
}

func (movement *Movement) SetId(id int) {
	movement.Id = id
}

type TransferMovement struct {
	Movement
	FromParticipantId int `json:"fromParticipantId"`
	ToParticipantId   int `json:"toParticipantId"`
}

type ParticipantMovement struct {
	Id            int   `json:"id"`
	MovementId    int   `json:"movementId"`
	ParticipantId int   `json:"participantId"`
	Amount        Price `json:"amount"`
}

func (participantMovement ParticipantMovement) GetId() int {
	return participantMovement.Id
}

func (participantMovement *ParticipantMovement) SetId(id int) {
	participantMovement.Id = id
}

type ParticipantShareByParticipantId map[int]Price

type BalanceSheet interface {
	GetCredit(participantId int) (int, error)
	GetDebt(participantId int) (int, error)
}

type DebitCreditMap map[int]map[int]Price

func BuildParticipantsEqualShare(movement Movement, participantMovements []ParticipantMovement) ParticipantShareByParticipantId {
	equalShare := movement.Amount / len(participantMovements)
	participantShareByParticipantId := make(map[int]Price)
	for _, participantMovement := range participantMovements {
		participantShare := participantMovement.Amount - equalShare
		participantShareByParticipantId[participantMovement.ParticipantId] = participantShare
	}
	return participantShareByParticipantId
}

func BuildParticipantsTransferShare(movement TransferMovement) ParticipantShareByParticipantId {
	participantShareByParticipantId := make(map[int]Price)
	participantShareByParticipantId[movement.FromParticipantId] = movement.Amount // el que da queda acreditando
	participantShareByParticipantId[movement.ToParticipantId] = -movement.Amount  // el que recibe queda adeudando
	return participantShareByParticipantId
}

func BuildParticipantsTransferMovements(movement TransferMovement) []ParticipantMovement {
	return []ParticipantMovement{
		{ParticipantId: movement.FromParticipantId, MovementId: movement.Id, Amount: movement.Amount}, // el que da pone todo el monto (amount)
		{ParticipantId: movement.ToParticipantId, MovementId: movement.Id, Amount: 0},                 // el que recibe no pone (0)
	}
}

var ErrMovementAmountMismatch error = errors.New("The movement amount must match the sum of all participants' amounts.")

// invariante de que movement.Amount = SUM (participantMovements[i].amount)
func EnsureMovementAmountMatchesParticipantAmounts(movement Movement, participantMovements []ParticipantMovement) error {
	totalAmount := 0
	for _, participantMovement := range participantMovements {
		totalAmount += participantMovement.Amount
	}
	if movement.Amount != totalAmount {
		return ErrMovementAmountMismatch
	} else {
		return nil
	}
}

var ErrSharesDoNotSumToZero error = errors.New("The sum of shares must equal zero")

// invariante de que 0 = SUM (participantShareByParticipantId[i].amount)
func EnsureSharesSumToZero(participantShareByParticipantId ParticipantShareByParticipantId) error {
	totalAmount := 0
	for _, share := range participantShareByParticipantId {
		totalAmount += share
	}
	if 0 != totalAmount {
		return ErrSharesDoNotSumToZero
	} else {
		return nil
	}
}

func deepCopyParticipantShareByParticipantId(original map[int]Price) map[int]Price {
	_copy := make(map[int]Price)
	for key, value := range original {
		_copy[key] = value
	}
	return _copy
}

// generacion de deudas y créditos para cada participante en relación a los demás participantes
func BuildDebitCreditMap(participantMovements []ParticipantMovement, shares ParticipantShareByParticipantId) DebitCreditMap {
	debitCreditMap := make(DebitCreditMap)
	sharesCopy := deepCopyParticipantShareByParticipantId(shares)
	shares = sharesCopy // using a copy in order to leave untouch the "shares" argument
	participantIds := getSortedParticipantIds(shares)
	for _, participantMovement := range participantMovements {
		participantShare := shares[participantMovement.ParticipantId]
		participantHasDebt := participantShare < 0
		if participantHasDebt {
			debitCreditMap[participantMovement.ParticipantId] = make(map[int]Price)
			// Dev notes: The order of processing must be taken into account to produce deterministic results ...
			//for id, share := range shares { // ... so relying on standard map iteration is not viable.
			for _, id := range participantIds {
				share := shares[id]
				if id == participantMovement.ParticipantId {
					continue
				}
				participantHasCredit := share > 0
				if participantHasCredit {
					remainingShare := share + participantShare
					if remainingShare >= 0 {
						// 150 + -50 = 100, ejemplo considerando que participant tiene 50 de deuda y hay otro que se le debe 150
						debitCreditMap[participantMovement.ParticipantId][id] = -participantShare // le da todo (50) al otro
						participantShare = 0                                                      // queda debiendo 0
						shares[participantMovement.ParticipantId] = 0                             // queda debiendo 0
						shares[id] = remainingShare                                               // se le debe 100 al otro
						break
					} else { // 100 + -250 = -150, ejemplo considerando que participant tiene 250 de deuda y al otro se le debe 100
						debitCreditMap[participantMovement.ParticipantId][id] = share // le da lo que falta (100) al otro
						participantShare = remainingShare                             // sigue debiendo 150
						shares[participantMovement.ParticipantId] = remainingShare    // sigue debiendo 150
						shares[id] = 0                                                // no se le debe más al otro
					}
				}
			}
		}
	}
	return debitCreditMap
}

func getSortedParticipantIds(m ParticipantShareByParticipantId) []int {
	keys := make([]int, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	return keys
}

func addDebitCreditMap(source DebitCreditMap, target DebitCreditMap) {
	for i, innerMap := range source {
		_, exists := target[i]
		if !exists {
			target[i] = make(map[int]Price)
		}
		for j, value := range innerMap {
			_, exists := target[i][j]
			if !exists {
				target[i][j] = 0
			}
			target[i][j] += value
		}
	}
}

func SumDebitCreditMaps(left DebitCreditMap, right DebitCreditMap) DebitCreditMap {
	result := make(DebitCreditMap)

	addDebitCreditMap(left, result)
	addDebitCreditMap(right, result)

	return result
}

func addParticipantShare(source ParticipantShareByParticipantId, target ParticipantShareByParticipantId) {
	for id, value := range source {
		_, exists := target[id]
		if !exists {
			target[id] = 0
		}
		target[id] += value
	}
}

func SumParticipantShares(left ParticipantShareByParticipantId, right ParticipantShareByParticipantId) ParticipantShareByParticipantId {
	result := make(ParticipantShareByParticipantId)

	addParticipantShare(left, result)
	addParticipantShare(right, result)

	return result
}
