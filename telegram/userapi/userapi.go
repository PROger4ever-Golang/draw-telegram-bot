package userapi

import (
	"fmt"
	"time"

	"errors"
	"log"

	tuapi "github.com/PROger4ever/telegramapi"
	"github.com/PROger4ever/telegramapi/mtproto"
)

func UserTypesToUsers(userTypes *[]mtproto.TLUserType) *[]*mtproto.TLUser {
	users := []*mtproto.TLUser{}
	for _, userType := range *userTypes {
		user := userType.(*mtproto.TLUser)
		users = append(users, user)
	}
	return &users
}

type Tool struct {
	Conn    *tuapi.Conn
	State   *tuapi.State
	StateCh chan tuapi.State
	readyCh chan bool
}

func (tool *Tool) runProcessing(options *tuapi.Options) {
	tool.Conn = tuapi.New(*options, tool.State, tool)
	err := tool.Conn.Run()

	fmt.Printf("reconnecting runProcessing: %v", err)
	for err != nil {
		fmt.Printf("reconnecting runProcessing: %v", err)
		tool.State.PreferredDC = 0
		tool.Conn = tuapi.New(*options, tool.State, tool)
		err = tool.Conn.Run()
	}
}

func (tool *Tool) HandleConnectionReady() {
	tool.readyCh <- true
}

func (tool *Tool) HandleStateChanged(newState *tuapi.State) {
	tool.State = newState
	tool.StateCh <- *newState
}

func (tool *Tool) Run(state *tuapi.State, host string, port int, publicKey string, apiId int, apiHash string, verbose int) error {
	tool.State = state
	options := tuapi.Options{
		SeedAddr:  tuapi.Addr{IP: host, Port: port},
		PublicKey: publicKey,
		Verbose:   verbose,
		APIID:     apiId,
		APIHash:   apiHash,
	}

	tool.StateCh = make(chan tuapi.State, 5)
	tool.readyCh = make(chan bool, 5)

	go tool.runProcessing(&options)
	select {
	case <-tool.readyCh:
		return nil
	case <-time.After(30 * time.Second):
		return errors.New("timeout")
	}
}

func (tool *Tool) StartLogin(phoneNumber string) error {
	return tool.Conn.StartLogin(phoneNumber)
}

func (tool *Tool) CompleteLoginWithCode(phoneCode string) (*mtproto.TLUser, error) {
	auth, err := tool.Conn.CompleteLoginWithCode(phoneCode)
	if err != nil {
		return nil, err
	}

	if user, ok := auth.User.(*mtproto.TLUser); ok {
		return user, nil
	}
	return nil, errors.New("can't cast response to TLUser type")
}

func (tool *Tool) ChannelsGetParticipants(channelID int, accessHash uint64, filter mtproto.TLChannelParticipantsFilterType, offset int, limit int) (*mtproto.TLChannelsChannelParticipants, error) {
	r, err := tool.Conn.Send(&mtproto.TLChannelsGetParticipants{
		Channel: &mtproto.TLInputChannel{
			ChannelID:  channelID,
			AccessHash: accessHash,
		},
		Filter: filter,
		Offset: offset,
		Limit:  limit,
	})
	if err != nil {
		return nil, err
	}
	switch r2 := r.(type) {
	case *mtproto.TLChannelsChannelParticipants:
		if tool.Conn.Verbose >= 2 {
			log.Printf("Got channels.getParticipants response: %v", r2)
		}
		return r2, nil
	default:
		return nil, tool.Conn.HandleUnknownReply(r)
	}
}

func (tool *Tool) ContactsResolveUsername(username string) (*mtproto.TLContactsResolvedPeer, error) {
	r, err := tool.Conn.Send(&mtproto.TLContactsResolveUsername{
		Username: username,
	})
	if err != nil {
		return nil, err
	}
	switch r2 := r.(type) {
	case *mtproto.TLContactsResolvedPeer:
		if tool.Conn.Verbose >= 2 {
			log.Printf("Got contacts.ResolveUsername response: %v", r2)
		}
		return r2, nil
	default:
		return nil, tool.Conn.HandleUnknownReply(r)
	}
}

func (tool *Tool) MessagesGetDialogs() (*tuapi.ContactList, error) {
	contacts := tuapi.NewContactList()
	err := tool.Conn.LoadChats(contacts, 1000, &mtproto.TLInputPeerEmpty{})
	return contacts, err
}

func (tool *Tool) MessagesGetFullChat(chatId int) (*mtproto.TLMessagesChatFull, error) {
	r, err := tool.Conn.Send(&mtproto.TLMessagesGetFullChat{ChatID: chatId})
	if err != nil {
		return nil, err
	}
	switch r2 := r.(type) {
	case *mtproto.TLMessagesChatFull:
		if tool.Conn.Verbose >= 2 {
			log.Printf("Got messages.getFullChat response: %v", r2)
		}
		return r2, nil
	default:
		return nil, tool.Conn.HandleUnknownReply(r)
	}
}

// tool.Conn.Fail(err)
// tool.Conn.Shutdown()
