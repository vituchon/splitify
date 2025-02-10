package model

import (
	"fmt"
	"testing"
)

func TestCalculateDebitCreditMapForEqualShare(t *testing.T) {
	tests := []struct {
		name                 string
		movement             Movement
		participantMovements []ParticipantMovement
		expected             DebitCreditMap
	}{
		{
			name: "Movement fully covered by participant 1, resulting in participant 2 owing",
			movement: Movement{
				Id:        1,
				Amount:    1000,
				CreatedAt: 0,
				Concept:   "Test",
			},
			participantMovements: []ParticipantMovement{
				{Id: 1, ParticipantId: 1, MovementId: 1, Amount: 1000},
				{Id: 2, ParticipantId: 2, MovementId: 1, Amount: 0},
			},
			expected: DebitCreditMap{
				2: {1: 500},
			},
		},
		{
			name: "Movement fully covered by participant 2, resulting in participant 1 owing",
			movement: Movement{
				Id:        1,
				Amount:    1000,
				CreatedAt: 0,
				Concept:   "Test",
			},
			participantMovements: []ParticipantMovement{
				{Id: 1, ParticipantId: 1, MovementId: 1, Amount: 0},
				{Id: 2, ParticipantId: 2, MovementId: 1, Amount: 1000},
			},
			expected: DebitCreditMap{
				1: {2: 500},
			},
		},
		{
			name: "Equal movement split, no debts",
			movement: Movement{
				Id:        1,
				Amount:    1000,
				CreatedAt: 0,
				Concept:   "Test",
			},
			participantMovements: []ParticipantMovement{
				{Id: 1, ParticipantId: 1, MovementId: 1, Amount: 500},
				{Id: 2, ParticipantId: 2, MovementId: 1, Amount: 500},
			},
			expected: DebitCreditMap{},
		},
		{
			name: "Movement partially split, participant 2 owes participant 1",
			movement: Movement{
				Id:        1,
				Amount:    1000,
				CreatedAt: 0,
				Concept:   "Test",
			},
			participantMovements: []ParticipantMovement{
				{Id: 1, ParticipantId: 1, MovementId: 1, Amount: 800},
				{Id: 2, ParticipantId: 2, MovementId: 1, Amount: 200},
			},
			expected: DebitCreditMap{
				2: {1: 300},
			},
		},
		{
			name: "Movement fully covered by participant 1, participant 2,3 owes in equal shares",
			movement: Movement{
				Id:        1,
				Amount:    900,
				CreatedAt: 0,
				Concept:   "Test",
			},
			participantMovements: []ParticipantMovement{
				{Id: 1, ParticipantId: 1, MovementId: 1, Amount: 900},
				{Id: 2, ParticipantId: 2, MovementId: 1, Amount: 0},
				{Id: 3, ParticipantId: 3, MovementId: 1, Amount: 0},
			},
			expected: DebitCreditMap{
				2: {1: 300},
				3: {1: 300},
			},
		},
		{
			name: "Movement partially covered by participant 1 and 2, participant 2 and 3 owes in not equals shares",
			movement: Movement{
				Id:        1,
				Amount:    900,
				CreatedAt: 0,
				Concept:   "Test",
			},
			participantMovements: []ParticipantMovement{
				{Id: 1, ParticipantId: 1, MovementId: 1, Amount: 700},
				{Id: 2, ParticipantId: 2, MovementId: 1, Amount: 200},
				{Id: 3, ParticipantId: 3, MovementId: 1, Amount: 0},
			},
			expected: DebitCreditMap{
				2: {1: 100},
				3: {1: 300},
			},
		},
		{
			name: "Movement partially covered by participant 1 and 2, participant 3 and 4 owes in not equals shares",
			movement: Movement{
				Id:        1,
				Amount:    1000,
				CreatedAt: 0,
				Concept:   "Test",
			},
			participantMovements: []ParticipantMovement{
				{Id: 1, ParticipantId: 1, MovementId: 1, Amount: 400},
				{Id: 2, ParticipantId: 2, MovementId: 1, Amount: 400},
				{Id: 3, ParticipantId: 3, MovementId: 1, Amount: 0},
				{Id: 4, ParticipantId: 4, MovementId: 1, Amount: 200},
			},
			expected: DebitCreditMap{
				3: {1: 150, 2: 100},
				4: {2: 50},
			},
		},
		{
			name: "Movement partially covered by participant 1 and 2, participant 3 and 4 owes in not equals shares (different order)",
			movement: Movement{
				Id:        1,
				Amount:    1000,
				CreatedAt: 0,
				Concept:   "Test",
			},
			participantMovements: []ParticipantMovement{
				{Id: 1, ParticipantId: 1, MovementId: 1, Amount: 400},
				{Id: 2, ParticipantId: 2, MovementId: 1, Amount: 400},
				{Id: 3, ParticipantId: 3, MovementId: 1, Amount: 200},
				{Id: 4, ParticipantId: 4, MovementId: 1, Amount: 0},
			},
			expected: DebitCreditMap{
				3: {1: 50},
				4: {1: 100, 2: 150},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := EnsureMovementAmountMatchesParticipantAmounts(test.movement, test.participantMovements)
			if err != nil {
				t.Fatal(err.Error())
			}
			participantShareByParticipantId := BuildParticipantsEqualShare(test.movement, test.participantMovements)
			err = EnsureSharesSumToZero(participantShareByParticipantId)
			if err != nil {
				t.Fatal(err.Error())
			}
			generated := BuildDebitCreditMap(test.participantMovements, participantShareByParticipantId)
			if !areEquals(generated, test.expected) {
				t.Errorf("generated %v, expected %v", generated, test.expected)
			}
		})
	}
}

func TestCalculateDebitCreditMapForTransfer(t *testing.T) {
	tests := []struct {
		name             string
		transferMovement TransferMovement
		expected         DebitCreditMap
	}{
		{
			name: "Participant 1 transfer to participant 2, participant 2 (reciever) owes participant 1 (emiter)",
			transferMovement: TransferMovement{
				Movement: Movement{
					Id:        1,
					Amount:    1000,
					CreatedAt: 0,
					Concept:   "Test",
				},
				FromParticipantId: 1,
				ToParticipantId:   2,
			},
			expected: DebitCreditMap{
				2: {1: 1000},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			participantMovements := BuildParticipantsTransferMovements(test.transferMovement)
			err := EnsureMovementAmountMatchesParticipantAmounts(test.transferMovement.Movement, participantMovements)
			if err != nil {
				t.Fatal(err.Error())
			}
			participantShareByParticipantId := BuildParticipantsTransferShare(test.transferMovement)
			err = EnsureSharesSumToZero(participantShareByParticipantId)
			if err != nil {
				t.Fatal(err.Error())
			}
			generated := BuildDebitCreditMap(participantMovements, participantShareByParticipantId)
			if !areEquals(generated, test.expected) {
				t.Errorf("generated %v, expected %v", generated, test.expected)
			}
		})
	}
}

func TestCalculateSumDebitCreditMaps(t *testing.T) {
	tests := []struct {
		name                  string
		movement              Movement
		participantMovements  []ParticipantMovement
		expectedStepMap       DebitCreditMap
		expectedAcumulatedMap DebitCreditMap
	}{
		{
			name: "Movement fully covered by participant 1, resulting in participant 2 owing",
			movement: Movement{
				Id:        1,
				Amount:    1000,
				CreatedAt: 0,
				Concept:   "Test",
			},
			participantMovements: []ParticipantMovement{
				{Id: 1, ParticipantId: 1, MovementId: 1, Amount: 1000},
				{Id: 2, ParticipantId: 2, MovementId: 1, Amount: 0},
			},
			expectedStepMap: DebitCreditMap{
				2: {1: 500},
			},
			expectedAcumulatedMap: DebitCreditMap{
				2: {1: 500},
			},
		},
		{
			name: "Movement fully covered by participant 2, resulting in participant 1 owing",
			movement: Movement{
				Id:        1,
				Amount:    1000,
				CreatedAt: 0,
				Concept:   "Test",
			},
			participantMovements: []ParticipantMovement{
				{Id: 1, ParticipantId: 1, MovementId: 1, Amount: 0},
				{Id: 2, ParticipantId: 2, MovementId: 1, Amount: 1000},
			},
			expectedStepMap: DebitCreditMap{
				1: {2: 500},
			},
			expectedAcumulatedMap: DebitCreditMap{
				1: {2: 500},
				2: {1: 500},
			},
		},
		{
			name: "Equal movement split, no debts",
			movement: Movement{
				Id:        1,
				Amount:    1000,
				CreatedAt: 0,
				Concept:   "Test",
			},
			participantMovements: []ParticipantMovement{
				{Id: 1, ParticipantId: 1, MovementId: 1, Amount: 500},
				{Id: 2, ParticipantId: 2, MovementId: 1, Amount: 500},
			},
			expectedStepMap: DebitCreditMap{},
			expectedAcumulatedMap: DebitCreditMap{
				1: {2: 500},
				2: {1: 500},
			},
		},
		{
			name: "Movement partially split, participant 2 owes participant 1",
			movement: Movement{
				Id:        1,
				Amount:    1000,
				CreatedAt: 0,
				Concept:   "Test",
			},
			participantMovements: []ParticipantMovement{
				{Id: 1, ParticipantId: 1, MovementId: 1, Amount: 800},
				{Id: 2, ParticipantId: 2, MovementId: 1, Amount: 200},
			},
			expectedStepMap: DebitCreditMap{
				2: {1: 300},
			},
			expectedAcumulatedMap: DebitCreditMap{
				1: {2: 500},
				2: {1: 800},
			},
		},
		{
			name: "Movement fully covered by participant 1, participant 2,3 owes in equal shares",
			movement: Movement{
				Id:        1,
				Amount:    900,
				CreatedAt: 0,
				Concept:   "Test",
			},
			participantMovements: []ParticipantMovement{
				{Id: 1, ParticipantId: 1, MovementId: 1, Amount: 900},
				{Id: 2, ParticipantId: 2, MovementId: 1, Amount: 0},
				{Id: 3, ParticipantId: 3, MovementId: 1, Amount: 0},
			},
			expectedStepMap: DebitCreditMap{
				2: {1: 300},
				3: {1: 300},
			},
			expectedAcumulatedMap: DebitCreditMap{
				1: {2: 500},
				2: {1: 1100},
				3: {1: 300},
			},
		},
		{
			name: "Movement partially covered by participant 1 and 2, participant 2 and 3 owes in not equals shares",
			movement: Movement{
				Id:        1,
				Amount:    900,
				CreatedAt: 0,
				Concept:   "Test",
			},
			participantMovements: []ParticipantMovement{
				{Id: 1, ParticipantId: 1, MovementId: 1, Amount: 700},
				{Id: 2, ParticipantId: 2, MovementId: 1, Amount: 200},
				{Id: 3, ParticipantId: 3, MovementId: 1, Amount: 0},
			},
			expectedStepMap: DebitCreditMap{
				2: {1: 100},
				3: {1: 300},
			},
			expectedAcumulatedMap: DebitCreditMap{
				1: {2: 500},
				2: {1: 1200},
				3: {1: 600},
			},
		},
		{
			name: "Movement partially covered by participant 1 and 2, participant 3 and 4 owes in not equals shares",
			movement: Movement{
				Id:        1,
				Amount:    1000,
				CreatedAt: 0,
				Concept:   "Test",
			},
			participantMovements: []ParticipantMovement{
				{Id: 1, ParticipantId: 1, MovementId: 1, Amount: 400},
				{Id: 2, ParticipantId: 2, MovementId: 1, Amount: 400},
				{Id: 3, ParticipantId: 3, MovementId: 1, Amount: 0},
				{Id: 4, ParticipantId: 4, MovementId: 1, Amount: 200},
			},
			expectedStepMap: DebitCreditMap{
				3: {1: 150, 2: 100},
				4: {2: 50},
			},
			expectedAcumulatedMap: DebitCreditMap{
				1: {2: 500},
				2: {1: 1200},
				3: {1: 750, 2: 100},
				4: {2: 50},
			},
		},
		{
			name: "Movement partially covered by participant 1 and 2, participant 3 and 4 owes in not equals shares (different order)",
			movement: Movement{
				Id:        1,
				Amount:    1000,
				CreatedAt: 0,
				Concept:   "Test",
			},
			participantMovements: []ParticipantMovement{
				{Id: 1, ParticipantId: 1, MovementId: 1, Amount: 400},
				{Id: 2, ParticipantId: 2, MovementId: 1, Amount: 400},
				{Id: 3, ParticipantId: 3, MovementId: 1, Amount: 200},
				{Id: 4, ParticipantId: 4, MovementId: 1, Amount: 0},
			},
			expectedStepMap: DebitCreditMap{
				3: {1: 50},
				4: {1: 100, 2: 150},
			},
			expectedAcumulatedMap: DebitCreditMap{
				1: {2: 500},
				2: {1: 1200},
				3: {1: 800, 2: 100},
				4: {1: 100, 2: 200},
			},
		},
	}

	acumulatedMap := make(DebitCreditMap)
	for _, test := range tests {
		err := EnsureMovementAmountMatchesParticipantAmounts(test.movement, test.participantMovements)
		if err != nil {
			t.Fatal(err.Error())
		}
		participantShareByParticipantId := BuildParticipantsEqualShare(test.movement, test.participantMovements)
		err = EnsureSharesSumToZero(participantShareByParticipantId)
		if err != nil {
			t.Fatal(err.Error())
		}
		generated := BuildDebitCreditMap(test.participantMovements, participantShareByParticipantId)
		if !areEquals(generated, test.expectedStepMap) {
			t.Errorf("generated %v, expected %v", generated, test.expectedStepMap)
		}

		acumulatedMap = SumDebitCreditMaps(generated, acumulatedMap)
		if !areEquals(acumulatedMap, test.expectedAcumulatedMap) {
			t.Errorf("%s acumulated generated %v, acumulated expected %v", test.name, acumulatedMap, test.expectedAcumulatedMap)
		}
	}
	transferMovement := TransferMovement{
		Movement: Movement{
			Id:        1,
			Amount:    1000,
			CreatedAt: 0,
			Concept:   "Test",
		},
		FromParticipantId: 2,
		ToParticipantId:   1,
	}
	participantMovements := BuildParticipantsTransferMovements(transferMovement)
	err := EnsureMovementAmountMatchesParticipantAmounts(transferMovement.Movement, participantMovements)
	if err != nil {
		t.Fatal(err.Error())
	}
	participantShareByParticipantId := BuildParticipantsTransferShare(transferMovement)
	err = EnsureSharesSumToZero(participantShareByParticipantId)
	if err != nil {
		t.Fatal(err.Error())
	}
	generated := BuildDebitCreditMap(participantMovements, participantShareByParticipantId)
}

func areEquals(left, right DebitCreditMap) bool {
	if len(left) != len(right) {
		fmt.Printf("Length mismatch: left=%d, right=%d\n", len(left), len(right))
		return false
	}

	for key, leftInnerMap := range left {
		rightInnerMap, exists := right[key]
		if !exists {
			fmt.Printf("Key %d found in left but not in right\n", key)
			return false
		}

		if len(leftInnerMap) != len(rightInnerMap) {
			fmt.Printf("Inner map length mismatch for key %d: left=%d, right=%d\n", key, len(leftInnerMap), len(rightInnerMap))
			return false
		}

		for innerKey, leftValue := range leftInnerMap {
			rightValue, exists := rightInnerMap[innerKey]
			if !exists {
				fmt.Printf("Inner key %d for outer key %d found in left but not in right\n", innerKey, key)
				return false
			}
			if leftValue != rightValue {
				fmt.Printf("Value mismatch at key %d -> %d: left=%d, right=%d\n", key, innerKey, leftValue, rightValue)
				return false
			}
		}
	}

	return true
}
