package settingState

import (
	"reflect"

	"github.com/PROger4ever/draw-telegram-bot/error"
	"github.com/PROger4ever/draw-telegram-bot/mongo"
	tuapi "github.com/PROger4ever/telegramapi"
	"github.com/PROger4ever/telegramapi/mtproto"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type SettingState struct {
	*mongo.BaseModel
	SettingStateCollection *SettingStateCollection
	Name                   string
	Value                  *StateSerializable
}

func (m *SettingState) Init(collection *SettingStateCollection) *SettingState {
	m.SettingStateCollection = collection
	m.BaseModel = mongo.NewModel(collection, m)
	return m
}

func (m *SettingState) GetBaseModel() *mongo.BaseModel {
	return m.BaseModel
}

func (m *SettingState) SetBaseModel(bm *mongo.BaseModel) {
	m.BaseModel = bm
}

func (m *SettingState) ClearModel() {
	m.Name = ""
	m.Value = nil
}

func (m *SettingState) GetContent() bson.M {
	return bson.M{
		"name":  m.Name,
		"value": m.Value,
	}
}

func (m *SettingState) SetContent(theMap bson.M) {
	if nameI, okM := theMap["name"]; okM {
		if name, okC := nameI.(string); okC {
			m.Name = name
		}
	}

	if valueI, okM := theMap["value"]; okM {
		if value, okC := valueI.(*StateSerializable); okC { //TODO: check casting: pointer or value?
			m.Value = value
		}
	}
}

type SettingStateCollection struct {
	*mongo.BaseCollection
}

func (c *SettingStateCollection) Init(connection *mongo.Connection) {
	c.BaseCollection = mongo.NewCollection(connection /*c,*/, "mazimotaBot", "settings", reflect.TypeOf(&SettingState{}))
}

func (c *SettingStateCollection) GetIndexes() []mgo.Index {
	return []mgo.Index{}
}

func (c *SettingStateCollection) EnsureIndexes() *eepkg.ExtendedError {
	return c.BaseCollection.EnsureIndexes(c.GetIndexes())
}

func (c *SettingStateCollection) GetBaseCollection() *mongo.BaseCollection {
	return c.BaseCollection
}

func (c *SettingStateCollection) Upsert(query bson.M, value *SettingState) (info *mgo.ChangeInfo, err error) {
	return c.BaseCollection.UpsertInterface(query, value)
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

func New(collection *SettingStateCollection) *SettingState {
	m := SettingState{}
	return m.Init(collection)
}

func NewCollection(connection *mongo.Connection) (c *SettingStateCollection) {
	c = &SettingStateCollection{}
	c.Init(connection)
	return
}

func NewCollectionDefault() (c *SettingStateCollection) {
	c = &SettingStateCollection{}
	c.Init(mongo.DefaultConnection)
	return
}
