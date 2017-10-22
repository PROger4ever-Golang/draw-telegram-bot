package state

import (
	tuapi "github.com/PROger4ever/telegramapi"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"bitbucket.org/proger4ever/drawtelegrambot/common"
	"bitbucket.org/proger4ever/drawtelegrambot/mongo"
	"bitbucket.org/proger4ever/drawtelegrambot/mongo/models/setting-state"
)

func Load(conn *mongo.Connection) (state *tuapi.State, err error) {
	state = &tuapi.State{}
	err = mgo.ErrNotFound
	// settingState := settingState.SettingState{}
	// err = settingState.New(conn).FindOne(&settingState)
	return state, err
}

func Save(conn *mongo.Connection, state *tuapi.State) {
	err := settingState.New(conn).Upsert(bson.M{
		"name": "state",
	}, &settingState.SettingState{
		Name:  "state",
		Value: settingState.NewStateSerializable(state),
	})
	common.PanicIfError(err, "saving bot state")
}
