package model

import (
	"fmt"
)

type Group struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type Member struct {
	Id      int    `json:"id"`
	Name    string `json:"name"`
	GroupId int    `json:"groupId"`
}

type Movement struct {
	Id        int    `json:"id"`
	CreatedAt int64  `json:"createdAt"` // unix timestamp, in seconds since epoch
	Amount    Price  `json:"amount"`
	Concept   string `json:"concept"`
}

type MovementParticipant struct {
	Id         int   `json:"id"`
	MovementId int   `json:"movementId"`
	MemberId   int   `json:"memberId"`
	Amount     Price `json:"amount"`
}

type BalanceSheet interface {
}

type DebitCreditMap map[int]map[int]int

func BuildDebitCreditMap(movement Movement, participants []MovementParticipant) DebitCreditMap {
	// deberia verificar la invariante de que movement.Amount = SUM (participants[i].amount)

	// calculo de saldo total (balance) por participante
	// dev notes: si es una transferenca de un particpante a otro, esta estructur se arma disstinto!
	equalShare := movement.Amount / len(participants)
	participantShareByParticipantId := make(map[int]Price)
	for _, participant := range participants {
		participantShare := participant.Amount - equalShare
		participantShareByParticipantId[participant.MemberId] = participantShare
	}

	fmt.Printf("%+v", participantShareByParticipantId)
	// generacion de deudas y créditos para cada participante en relación a los demás participantes
	debitCreditMap := make(DebitCreditMap)
	for _, participant := range participants {
		participantShare := participantShareByParticipantId[participant.MemberId]
		if participantShare < 0 {
			debitCreditMap[participant.MemberId] = make(map[int]Price)
			for id, share := range participantShareByParticipantId {
				if id == participant.MemberId {
					continue
				}
				if share > 0 {
					remainingShare := share + participantShare // a: 150 + -50 = 100
					if remainingShare >= 0 {
						debitCreditMap[participant.MemberId][id] = -participantShare // le da todo (50) al otro
						participantShare = 0                                         // queda debiendo 0
						participantShareByParticipantId[participant.MemberId] = 0    // queda debiendo 0
						participantShareByParticipantId[id] = remainingShare         // se le debe 100 al otro
						break
					} else { // b : 100 + -250 = -150
						debitCreditMap[participant.MemberId][id] = share                       // le da lo que falta (100) al otro
						participantShare = remainingShare                                      // sigue debiendo 150
						participantShareByParticipantId[participant.MemberId] = remainingShare // sigue debiendo 150
						participantShareByParticipantId[id] = 0                                // no se le debe más al otro
					}
				}
			}
		}
	}
	return debitCreditMap
}
