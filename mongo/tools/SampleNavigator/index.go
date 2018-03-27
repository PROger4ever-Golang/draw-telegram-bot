package sampleNavigator

import (
	"github.com/PROger4ever/draw-telegram-bot/mongo"
	"errors"
	"math"

	"gopkg.in/mgo.v2/bson"
)

var ErrNotEnough = errors.New("not enough docs in db")

type SampleNavigator struct {
	Collection *mongo.BaseCollection

	MatchM   bson.M
	SampleM  bson.M
	Pipeline []bson.M

	Buf    []bson.M
	BufLen int
	Index  int

	ShownIds  []bson.ObjectId
	ShownIdsM bson.M
}

func (sn *SampleNavigator) Load() (err error) {
	sn.Buf = []bson.M{}
	sn.Index = 0

	pipe := sn.Collection.PipeInterface(sn.Pipeline)
	err = pipe.All(&sn.Buf)
	if err != nil {
		return err
	}
	if len(sn.Buf) == 0 {
		return ErrNotEnough
	}
	return
}

func (sn *SampleNavigator) Next(model mongo.Model) (err error) {
	if len(sn.Buf) == 0 || sn.Index == len(sn.Buf) {
		err = sn.Load()
	}
	if err != nil {
		return
	}
	theMap := sn.Buf[sn.Index]
	bm := model.GetBaseModel()

	bm.SetContent(theMap)
	sn.ShownIds = append(sn.ShownIds, bm.ID)
	sn.ShownIdsM["$nin"] = sn.ShownIds
	sn.Index++
	return
}

func New(collection *mongo.BaseCollection, matchM bson.M, length int) *SampleNavigator {
	bufLen := int(math.Ceil(float64(length)*0.2)) + length

	sampleM := bson.M{
		"size": bufLen,
	}

	//if match has $and already
	var and []bson.M
	if a, ok := matchM["$and"]; ok {
		and = a.([]bson.M)
	} else {
		and = []bson.M{}
		matchM["$and"] = and
	}
	shownIdsM := bson.M{ //append our $nin to skip ids which are already shown
		"$nin": []bson.ObjectId{},
	}
	idM := bson.M{
		"_id": shownIdsM,
	}
	matchM["$and"] = append(and, idM)

	pipeline := []bson.M{{
		"$match": matchM,
	}, {
		"$sample": sampleM,
	}}

	sn := SampleNavigator{
		Collection: collection,
		BufLen:     bufLen,
		MatchM:     matchM,
		SampleM:    sampleM,

		Pipeline:  pipeline,
		ShownIds:  make([]bson.ObjectId, 0, bufLen),
		ShownIdsM: shownIdsM,
	}
	return &sn
}
