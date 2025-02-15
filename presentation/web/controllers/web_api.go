package controllers

import (
	"fmt"
	"log"
	"net/http"

	model_api "github.com/vituchon/splitify/model/api"
)

func GetAllGroups(response http.ResponseWriter, request *http.Request) {
	groups, err := model_api.GetAllGroups()
	if err != nil {
		msg := fmt.Sprintf("error while retrieving groups : '%v'", err)
		log.Println(msg)
		http.Error(response, msg, http.StatusInternalServerError)
		return
	}
	WriteJsonResponse(response, http.StatusOK, groups)
}

func CreateGroup(response http.ResponseWriter, request *http.Request) {
	name, err := ParseSingleStringUrlQueryParam(request, "name")
	if err != nil {
		msg := fmt.Sprintf("error while creating group : '%v'", err)
		log.Println(msg)
		http.Error(response, msg, http.StatusBadRequest)
		return
	}

	createdGroup, err := model_api.CreateGroup(*name)
	if err != nil {
		msg := fmt.Sprintf("error while creating group : '%v'", err)
		log.Println(msg)
		http.Error(response, msg, http.StatusInternalServerError)
		return
	}
	WriteJsonResponse(response, http.StatusOK, createdGroup)
}

func GetGroupParticipants(response http.ResponseWriter, request *http.Request) {
	groupId, err := ParseRouteParamAsInt(request, "groupId")
	if err != nil {
		msg := fmt.Sprintf("error while retrieving group participants : '%v'", err)
		log.Println(msg)
		http.Error(response, msg, http.StatusBadRequest)
		return
	}
	participants, err := model_api.GetParticipants(groupId)
	if err != nil {
		msg := fmt.Sprintf("error while retrieving group participants: '%v'", err)
		log.Println(msg)
		http.Error(response, msg, http.StatusInternalServerError)
		return
	}
	WriteJsonResponse(response, http.StatusOK, participants)
}

func AddParcipantToGroup(response http.ResponseWriter, request *http.Request) {
	groupId, err := ParseRouteParamAsInt(request, "groupId")
	if err != nil {
		msg := fmt.Sprintf("error while retrieving group participants : '%v'", err)
		log.Println(msg)
		http.Error(response, msg, http.StatusBadRequest)
		return
	}
	name, err := ParseSingleStringUrlQueryParam(request, "name")
	if err != nil {
		msg := fmt.Sprintf("error while adding participant to group : '%v'", err)
		log.Println(msg)
		http.Error(response, msg, http.StatusBadRequest)
		return
	}

	participant := model_api.Participant{
		GroupId: groupId,
		Name:    *name,
	}

	createdParticipant, err := model_api.AddParticipant(participant)
	if err != nil {
		msg := fmt.Sprintf("error while adding participant to group : '%v'", err)
		log.Println(msg)
		http.Error(response, msg, http.StatusInternalServerError)
		return
	}
	WriteJsonResponse(response, http.StatusOK, createdParticipant)
}
