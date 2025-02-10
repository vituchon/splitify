package model

import (
	"errors"
	"sort"
)

type Movement struct {
	Id        int    `json:"id"`
	CreatedAt int64  `json:"createdAt"` // unix timestamp, in seconds since epoch
	Amount    Price  `json:"amount"`
	Concept   string `json:"concept"`
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

type ParticipantShareByParticipantId map[int]Price

type BalanceSheet interface {
	GetCredit(participantId int) (int, error)
	GetDebt(participantId int) (int, error)
}

type DebitCreditMap map[int]map[int]int

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

// generacion de deudas y créditos para cada participante en relación a los demás participantes
func BuildDebitCreditMap(participantMovements []ParticipantMovement, participantShareByParticipantId ParticipantShareByParticipantId) DebitCreditMap {
	debitCreditMap := make(DebitCreditMap)
	participantIds := getSortedParticipantIds(participantShareByParticipantId)
	for _, participantMovement := range participantMovements {
		participantShare := participantShareByParticipantId[participantMovement.ParticipantId]
		participantHasDebt := participantShare < 0
		if participantHasDebt {
			debitCreditMap[participantMovement.ParticipantId] = make(map[int]Price)
			// Dev notes: The order of processing must be taken into account to produce deterministic results ...
			//for id, share := range participantShareByParticipantId { // ... so relying on standard map iteration is not viable.
			for _, id := range participantIds {
				share := participantShareByParticipantId[id]
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
						participantShareByParticipantId[participantMovement.ParticipantId] = 0    // queda debiendo 0
						participantShareByParticipantId[id] = remainingShare                      // se le debe 100 al otro
						break
					} else { // 100 + -250 = -150, ejemplo considerando que participant tiene 250 de deuda y al otro se le debe 100
						debitCreditMap[participantMovement.ParticipantId][id] = share                       // le da lo que falta (100) al otro
						participantShare = remainingShare                                                   // sigue debiendo 150
						participantShareByParticipantId[participantMovement.ParticipantId] = remainingShare // sigue debiendo 150
						participantShareByParticipantId[id] = 0                                             // no se le debe más al otro
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
			target[i] = make(map[int]int)
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

func SumDebitCreditMaps(map1 DebitCreditMap, map2 DebitCreditMap) DebitCreditMap {
	result := make(DebitCreditMap)

	addDebitCreditMap(map1, result)
	addDebitCreditMap(map2, result)

	return result
}
