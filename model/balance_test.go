package model

import (
	"fmt"
	"testing"
)

func TestCalculateMap(t *testing.T) {
	tests := []struct {
		name         string
		movement     Movement
		participants []MovementParticipant
		expected     map[int]map[int]int
	}{
		{
			name: "Movement fully covered by participant 1, resulting in participant 2 owing",
			movement: Movement{
				Id:        1,
				Amount:    1000,
				CreatedAt: 0,
				Concept:   "Test",
			},
			participants: []MovementParticipant{
				{Id: 1, MemberId: 1, MovementId: 1, Amount: 1000},
				{Id: 2, MemberId: 2, MovementId: 1, Amount: 0},
			},
			expected: map[int]map[int]int{
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
			participants: []MovementParticipant{
				{Id: 1, MemberId: 1, MovementId: 1, Amount: 0},
				{Id: 2, MemberId: 2, MovementId: 1, Amount: 1000},
			},
			expected: map[int]map[int]int{
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
			participants: []MovementParticipant{
				{Id: 1, MemberId: 1, MovementId: 1, Amount: 500},
				{Id: 2, MemberId: 2, MovementId: 1, Amount: 500},
			},
			expected: map[int]map[int]int{},
		},
		{
			name: "Movement partially split, participant 2 owes participant 1",
			movement: Movement{
				Id:        1,
				Amount:    1000,
				CreatedAt: 0,
				Concept:   "Test",
			},
			participants: []MovementParticipant{
				{Id: 1, MemberId: 1, MovementId: 1, Amount: 800},
				{Id: 2, MemberId: 2, MovementId: 1, Amount: 200},
			},
			expected: map[int]map[int]int{
				2: {1: 300},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			generated := BuildDebitCreditMap(test.movement, test.participants)
			if !areEquals(generated, test.expected) {
				t.Errorf("generated %v, expected %v", generated, test.expected)
			}
		})
	}
}

func areEquals(left, right map[int]map[int]int) bool {
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
