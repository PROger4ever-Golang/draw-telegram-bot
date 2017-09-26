package userapi

import (
	"fmt"
	"time"

	"errors"
	"log"

	tuapi "github.com/PROger4ever/telegramapi"
	"github.com/PROger4ever/telegramapi/mtproto"
)

type Tool struct {
	Conn    *tuapi.Conn
	State   *tuapi.State
	StateCh chan tuapi.State
	ErrCh   chan error
	readyCh chan bool
}

func (tool *Tool) HandleConnectionReady() {
	tool.readyCh <- true
}
func (tool *Tool) HandleStateChanged(newState *tuapi.State) {
	tool.State = newState
	tool.StateCh <- *newState
}

func (tool *Tool) Run(state *tuapi.State, host string, port int, publicKey string, apiId int, apiHash string, verbose int) error {
	var err error

	options := tuapi.Options{
		SeedAddr:  tuapi.Addr{IP: host, Port: port},
		PublicKey: publicKey,
		Verbose:   verbose,
		APIID:     apiId,
		APIHash:   apiHash,
	}

	tool.StateCh = make(chan tuapi.State, 5)
	tool.ErrCh = make(chan error, 5)
	tool.readyCh = make(chan bool, 5)
	tool.Conn = tuapi.New(options, state, tool)

	go tool.runProcessing()

	select {
	case <-tool.readyCh:
		return nil
	case err = <-tool.ErrCh:
		return err
	case <-time.After(1000 * time.Second):
		return fmt.Errorf("timeout")
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

func (tool *Tool) MessagesGetFullChat(chatId int) (*mtproto.TLMessagesChatFull, error) {
	r, err := tool.Conn.Send(&mtproto.TLMessagesGetFullChat{ChatID: chatId})
	if err != nil {
		return nil, err
	}
	switch r := r.(type) {
	case *mtproto.TLMessagesChatFull:
		if tool.Conn.Verbose >= 2 {
			log.Printf("Got messages.getFullChat response: %v", r)
		}
		return r, nil
	default:
		return nil, tool.Conn.HandleUnknownReply(r)
	}
}

func (tool *Tool) runProcessing() {
	err := tool.Conn.Run()
	if err != nil {
		tool.ErrCh <- err
	}
}

// tool.Conn.Fail(err)
// tool.Conn.Shutdown()
