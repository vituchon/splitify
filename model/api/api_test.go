package api

import (
	"reflect"
	"testing"

	"github.com/vituchon/splitify/model"
)

func TestNormalApiFlowFromGoodClient(t *testing.T) {
	// Crear un grupo y verificar que exista
	group, err := CreateGroup("Group 1")
	if err != nil {
		t.Fatalf("Failed to create group: %v", err)
	}
	if group.Name != "Group 1" {
		t.Fatalf("Expected group name to be 'Group 1', got '%s'", group.Name)
	}
	savedGroup, err := groupsRepository.GetAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(savedGroup) != 1 {
		t.Fatalf("Expected 1 group, got %d", len(savedGroup))
	}

	participants := []string{"Vitu", "Chori", "Junior"}
	var participantModels []*model.Participant
	for _, name := range participants {
		p, err := AddParticipant(Participant{GroupId: group.Id, Name: name})
		if err != nil {
			t.Fatalf("Failed to add participant '%s': %v", name, err)
		}
		participantModels = append(participantModels, p)
	}

	savedParticipants, err := participantsRepository.GetAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(savedParticipants) != 3 {
		t.Fatalf("Expected 3 participants, got %d", len(savedParticipants))
	}

	movements := []Movement{
		{
			GroupId: group.Id,
			Amount:  1000,
			Concept: "Almuerzo",
			ParticipantMovements: []ParticipantMovement{
				{ParticipantId: participantModels[0].Id, Amount: 600},
				{ParticipantId: participantModels[1].Id, Amount: 400},
			},
			/*acumulatedShare: model.ParticipantShareByParticipantId{
				1: 100,
				2: -100,
			},
			acumulatedBalance: model.DebitCreditMap{
				2: {1: 100},
			},*/
		},
		{
			GroupId: group.Id,
			Amount:  1500,
			Concept: "Merienda",
			ParticipantMovements: []ParticipantMovement{
				{ParticipantId: participantModels[0].Id, Amount: 500},
				{ParticipantId: participantModels[1].Id, Amount: 500},
				{ParticipantId: participantModels[2].Id, Amount: 500},
			},
			/*acumulatedShare: model.ParticipantShareByParticipantId{
				1: 100,
				2: -100,
			},
			acumulatedBalance: model.DebitCreditMap{
				2: {1: 100},
			},*/
		},
		{
			GroupId: group.Id,
			Amount:  900,
			Concept: "Cena",
			ParticipantMovements: []ParticipantMovement{
				{ParticipantId: participantModels[0].Id, Amount: 300},
				{ParticipantId: participantModels[1].Id, Amount: 300},
				{ParticipantId: participantModels[2].Id, Amount: 300},
			},
			/*acumulatedShare: model.ParticipantShareByParticipantId{
				1: 100,
				2: -100,
			},
			acumulatedBalance: model.DebitCreditMap{
				2: {1: 100},
			},*/
		},
	}

	var addedMovements []*model.Movement
	var addedParticipantMovements []*model.ParticipantMovement
	for _, movement := range movements {
		m, pms, err := AddMovement(movement)
		if err != nil {
			t.Fatalf("Failed to add movement '%s': %v", movement.Concept, err)
		}
		addedMovements = append(addedMovements, m)
		addedParticipantMovements = append(addedParticipantMovements, pms...)
	}

	// Verificar que los movimientos y participant movements sean los esperados
	if len(addedMovements) != 3 {
		t.Fatalf("Expected 3 movements, got %d", len(addedMovements))
	}
	if len(addedParticipantMovements) != 8 {
		t.Fatalf("Expected 8 participant movements, got %d", len(addedParticipantMovements))
	}

	generatedBalance, shares, err := CalculateBalances(group.Id)
	if err != nil {
		t.Fatalf("Failed to calculate balances: %v", err)
	}
	expectedBalance := model.DebitCreditMap{
		2: {1: 100},
	}
	if !reflect.DeepEqual(generatedBalance, expectedBalance) {
		t.Errorf("Balances mismatch. Expected: %v, got: %v", expectedBalance, generatedBalance)
	}

	expectedShares := model.ParticipantShareByParticipantId{
		1: 100,
		2: -100,
		3: 0,
	}
	if !reflect.DeepEqual(shares, expectedShares) {
		t.Errorf("Shares mismatch. Expected: %v, got: %v", expectedShares, shares)
	}
}
