package state

import (
	appUtils "bitbucket.org/proger4ever/draw-telegram-bot/utils/app"
	tuapi "github.com/PROger4ever/telegramapi"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"bitbucket.org/proger4ever/draw-telegram-bot/mongo/models/setting-state"
)

func Load() (state *tuapi.State, err error) {
	state = &tuapi.State{}
	err = mgo.ErrNotFound
	return state, err
}

func Save(state *tuapi.State) {
	ss := settingState.SettingState{
		Name:  "state",
		Value: settingState.NewStateSerializable(state),
	}
	ss.Init(settingState.NewCollectionDefault())

	_, err := ss.Upsert(bson.M{
		"name": "state",
	})
	appUtils.PanicIfExtended(err, "saving bot state")
}
