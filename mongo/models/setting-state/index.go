package settingState

import (
	"bitbucket.org/proger4ever/draw-telegram-bot/mongo"
	"bitbucket.org/proger4ever/draw-telegram-bot/mongo/models/base-model"
	tuapi "github.com/PROger4ever/telegramapi"
	"github.com/PROger4ever/telegramapi/mtproto"
	"gopkg.in/mgo.v2/bson"
)

type SettingState struct {
	*baseModel.BaseModel
	ID    bson.ObjectId `bson:"_id,omitempty"`
	Name  string
	Value *StateSerializable
}

func (m *SettingState) Init(connection *mongo.Connection) *SettingState {
	m.BaseModel = baseModel.New(connection, "mazimotaBot", "settings")
	return m
}

func New(connection *mongo.Connection) *SettingState {
	m := SettingState{}
	m.Init(connection)
	return &m
}

func (ss *SettingState) DecodeValue() *tuapi.State {
	ser := ss.Value
	st := &tuapi.State{
		PreferredDC: ser.PreferredDC,

		LoginState:    ser.LoginState,
		PhoneNumber:   ser.PhoneNumber,
		PhoneCodeHash: ser.PhoneCodeHash,

		UserID:    ser.UserID,
		FirstName: ser.FirstName,
		LastName:  ser.LastName,
		Username:  ser.Username,
	}

	st.DCs = make(map[int]*tuapi.DCState, len(st.DCs))
	for _, dc := range ser.DCs {
		st.DCs[dc.ID] = &tuapi.DCState{
			ID:          dc.ID,
			PrimaryAddr: dc.PrimaryAddr,
			FramerState: dc.FramerState,
			Auth: mtproto.AuthResult{
				Key:        dc.Auth.Key,
				KeyID:      (uint64(dc.Auth.KeyIDHigh) << 32) + uint64(dc.Auth.KeyIDLow),
				ServerSalt: dc.Auth.ServerSalt,
				TimeOffset: dc.Auth.TimeOffset,
				SessionID:  dc.Auth.SessionID,
			},
		}
	}
	return st
}

type StateSerializable struct {
	PreferredDC int

	DCs []*DCState

	LoginState    tuapi.LoginState
	PhoneNumber   string
	PhoneCodeHash string

	UserID    int
	FirstName string
	LastName  string
	Username  string
}

type DCState struct {
	ID int

	PrimaryAddr tuapi.Addr

	Auth        AuthResult
	FramerState mtproto.FramerState
}

type AuthResult struct {
	Key        []byte
	KeyIDHigh  uint32
	KeyIDLow   uint32
	ServerSalt [8]byte
	TimeOffset int
	SessionID  [8]byte
}

func NewStateSerializable(st *tuapi.State) *StateSerializable {
	ser := &StateSerializable{
		PreferredDC: st.PreferredDC,

		LoginState:    st.LoginState,
		PhoneNumber:   st.PhoneNumber,
		PhoneCodeHash: st.PhoneCodeHash,

		UserID:    st.UserID,
		FirstName: st.FirstName,
		LastName:  st.LastName,
		Username:  st.Username,
	}

	ser.DCs = make([]*DCState, len(st.DCs))
	i := 0
	for _, dc := range st.DCs {
		ser.DCs[i] = &DCState{
			ID:          dc.ID,
			PrimaryAddr: dc.PrimaryAddr,
			FramerState: dc.FramerState,
			Auth: AuthResult{
				Key:        dc.Auth.Key,
				KeyIDHigh:  uint32(dc.Auth.KeyID >> 32),
				KeyIDLow:   uint32(dc.Auth.KeyID),
				ServerSalt: dc.Auth.ServerSalt,
				TimeOffset: dc.Auth.TimeOffset,
				SessionID:  dc.Auth.SessionID,
			},
		}
		i++
	}
	return ser
}
