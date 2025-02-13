package api

import (
	"reflect"
	"testing"

	"github.com/vituchon/splitify/model"
)

func TestNormalApiFlowFromGoodClient(t *testing.T) {
	group, err := CreateGroup("Group 1")
	if err != nil {
		t.Fatalf("Failed to create group: %v", err)
	}

	t.Run("TestCreateGroup", func(t *testing.T) {
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
	})

	participants := []string{"Vitu", "Chori", "Junior"}
	var participantModels []*model.Participant

	t.Run("TestAddParticipants", func(t *testing.T) {
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
	})

	participantId1 := participantModels[0].Id
	participantId2 := participantModels[1].Id
	participantId3 := participantModels[2].Id

	movements := []Movement{
		{
			GroupId: group.Id,
			Amount:  1000,
			Concept: "Almuerzo",
			ParticipantMovements: []ParticipantMovement{
				{ParticipantId: participantId1, Amount: 600},
				{ParticipantId: participantId2, Amount: 400},
			},
		},
		{
			GroupId: group.Id,
			Amount:  1500,
			Concept: "Merienda",
			ParticipantMovements: []ParticipantMovement{
				{ParticipantId: participantId1, Amount: 500},
				{ParticipantId: participantId2, Amount: 500},
				{ParticipantId: participantId3, Amount: 500},
			},
		},
		{
			GroupId: group.Id,
			Amount:  900,
			Concept: "Cena",
			ParticipantMovements: []ParticipantMovement{
				{ParticipantId: participantId1, Amount: 300},
				{ParticipantId: participantId2, Amount: 300},
				{ParticipantId: participantId3, Amount: 300},
			},
		},
	}
	var addedMovements []*model.Movement
	var addedParticipantMovements []*model.ParticipantMovement

	t.Run("TestAddMovements", func(t *testing.T) {
		for _, movement := range movements {
			m, pms, err := AddMovement(movement)
			if err != nil {
				t.Fatalf("Failed to add movement '%s': %v", movement.Concept, err)
			}
			addedMovements = append(addedMovements, m)
			addedParticipantMovements = append(addedParticipantMovements, pms...)
		}

		if len(addedMovements) != 3 {
			t.Fatalf("Expected 3 movements, got %d", len(addedMovements))
		}
		if len(addedParticipantMovements) != 8 {
			t.Fatalf("Expected 8 participant movements, got %d", len(addedParticipantMovements))
		}
	})

	t.Run("TestCalculateBalancesAndShares", func(t *testing.T) {
		generatedBalance, shares, err := CalculateBalances(group.Id)
		if err != nil {
			t.Fatalf("Failed to calculate balances: %v", err)
		}

		expectedBalance := model.DebitCreditMap{
			participantId2: {participantId1: 100},
		}
		if !reflect.DeepEqual(generatedBalance, expectedBalance) {
			t.Errorf("Balances mismatch. Expected: %v, got: %v", expectedBalance, generatedBalance)
		}
		expectedShares := model.ParticipantShareByParticipantId{
			participantId1: 100,
			participantId2: -100,
			participantId3: 0,
		}
		if !reflect.DeepEqual(shares, expectedShares) {
			t.Errorf("Shares mismatch. Expected: %v, got: %v", expectedShares, shares)
		}
	})
}

type MovementTest struct {
	Name                     string
	Movement                 Movement
	ExpectedMap              model.DebitCreditMap
	ExpectedAcumulatedMap    model.DebitCreditMap
	ExpectedShares           model.ParticipantShareByParticipantId
	ExpectedAcumulatedShares model.ParticipantShareByParticipantId
}

func TestNormalApiFlowFromGoodClientSpepByStep(t *testing.T) {
	group, err := CreateGroup("Group 1")
	if err != nil {
		t.Fatalf("Failed to create group: %v", err)
	}

	participants := []string{"Vitu", "Chori", "Junior"}
	var participantModels []*model.Participant

	t.Run("TestAddParticipants", func(t *testing.T) {
		for _, name := range participants {
			p, err := AddParticipant(Participant{GroupId: group.Id, Name: name})
			if err != nil {
				t.Fatalf("Failed to add participant '%s': %v", name, err)
			}
			participantModels = append(participantModels, p)
		}
	})

	participantId1 := participantModels[0].Id
	participantId2 := participantModels[1].Id
	participantId3 := participantModels[2].Id

	tests := []MovementTest{
		{
			Name: "Almuerzo",
			Movement: Movement{
				GroupId: group.Id,
				Amount:  1000,
				Concept: "Almuerzo",
				ParticipantMovements: []ParticipantMovement{
					{ParticipantId: participantId1, Amount: 600},
					{ParticipantId: participantId2, Amount: 400},
				},
			},
			ExpectedMap: model.DebitCreditMap{
				participantId2: {participantId1: 100},
			},
			ExpectedAcumulatedMap: model.DebitCreditMap{
				participantId2: {participantId1: 100},
			},
			ExpectedShares: model.ParticipantShareByParticipantId{
				participantId1: 100,
				participantId2: -100,
			},
			ExpectedAcumulatedShares: model.ParticipantShareByParticipantId{
				participantId1: 100,
				participantId2: -100,
			},
		},
		{
			Name: "Merienda",
			Movement: Movement{
				GroupId: group.Id,
				Amount:  1500,
				Concept: "Merienda",
				ParticipantMovements: []ParticipantMovement{
					{ParticipantId: participantId1, Amount: 500},
					{ParticipantId: participantId2, Amount: 500},
					{ParticipantId: participantId3, Amount: 500},
				},
			},
			ExpectedMap: model.DebitCreditMap{},
			ExpectedAcumulatedMap: model.DebitCreditMap{
				participantId2: {participantId1: 100},
			},
			ExpectedShares: model.ParticipantShareByParticipantId{
				participantId1: 0,
				participantId2: 0,
				participantId3: 0,
			},
			ExpectedAcumulatedShares: model.ParticipantShareByParticipantId{
				participantId1: 100,
				participantId2: -100,
				participantId3: 0,
			},
		},
		{
			Name: "Cena",
			Movement: Movement{
				GroupId: group.Id,
				Amount:  900,
				Concept: "Cena",
				ParticipantMovements: []ParticipantMovement{
					{ParticipantId: participantId1, Amount: 800},
					{ParticipantId: participantId2, Amount: 0},
					{ParticipantId: participantId3, Amount: 100},
				},
			},
			ExpectedMap: model.DebitCreditMap{
				participantId2: {participantId1: 300},
				participantId3: {participantId1: 200},
			},
			ExpectedAcumulatedMap: model.DebitCreditMap{
				participantId2: {participantId1: 400},
				participantId3: {participantId1: 200},
			},
			ExpectedShares: model.ParticipantShareByParticipantId{
				participantId1: 500,
				participantId2: -300,
				participantId3: -200,
			},
			ExpectedAcumulatedShares: model.ParticipantShareByParticipantId{
				participantId1: 600,
				participantId2: -400,
				participantId3: -200,
			},
		},
	}

	acumulatedMap := make(model.DebitCreditMap)
	acumulatedShares := make(model.ParticipantShareByParticipantId)

	t.Run("TestAddMovementsAndVerifyStepByStep", func(t *testing.T) {
		for _, test := range tests {
			t.Log(test.Name)
			m, _, err := AddMovement(test.Movement)
			if err != nil {
				t.Fatalf("Failed to add movement '%s': %v", test.Name, err)
			}
			generatedMap, generatedShares, err := CalculateBalance(group.Id, m.Id)
			if err != nil {
				t.Fatalf("Failed to calculate balances: %v", err)
			}

			if !reflect.DeepEqual(generatedMap, test.ExpectedMap) {
				t.Errorf("Balances mismatch. Expected: %v, got: %v", test.ExpectedMap, generatedMap)
			}
			if !reflect.DeepEqual(generatedShares, test.ExpectedShares) {
				t.Errorf("Shares mismatch. Expected: %v, got: %v", test.ExpectedShares, generatedShares)
			}

			acumulatedMap = model.SumDebitCreditMaps(acumulatedMap, generatedMap)
			if !reflect.DeepEqual(acumulatedMap, test.ExpectedAcumulatedMap) {
				t.Errorf("Acumulated balances mismatch. Expected: %v, got: %v", test.ExpectedAcumulatedMap, acumulatedMap)
			}

			acumulatedShares = model.SumParticipantShares(acumulatedShares, generatedShares)
			if !reflect.DeepEqual(acumulatedShares, test.ExpectedAcumulatedShares) {
				t.Errorf("Acumulated shares mismatch. Expected: %v, got: %v", test.ExpectedAcumulatedShares, acumulatedShares)
			}
		}
	})
}
