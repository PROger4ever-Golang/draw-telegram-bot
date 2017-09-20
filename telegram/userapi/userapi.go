package userapi

import (
	"fmt"
	"log"
	"time"

	"github.com/andreyvit/telegramapi"
)

type Tool struct {
	TG      *telegramapi.Conn
	readyCh chan bool
	ErrCh   chan error
}

func (tool *Tool) HandleConnectionReady() {
	tool.readyCh <- true
}
func (tool *Tool) HandleStateChanged(newState *telegramapi.State) {
	log.Printf("HandleStateChanged" /*, pretty.Sprint(newState)*/)
}

func (tool *Tool) Run(host string, port int, publicKey string, apiId int, apiHash string, verbose int) error {
	var err error

	options := telegramapi.Options{
		SeedAddr:  telegramapi.Addr{host, port},
		PublicKey: publicKey,
		Verbose:   verbose,
		APIID:     apiId,
		APIHash:   apiHash,
	}

	tool.ErrCh = make(chan error, 5)
	tool.readyCh = make(chan bool, 5)
	state := new(telegramapi.State)
	tool.TG = telegramapi.New(options, state, tool)

	go tool.runProcessing()

	select {
	case <-tool.readyCh:
		return nil
	case err = <-tool.ErrCh:
		fmt.Println("1111")
		return err
	case <-time.After(10 * time.Second):
		return fmt.Errorf("timeout")
	}
}

func (tool *Tool) runProcessing() {
	err := tool.TG.Run()
	if err != nil {
		fmt.Println("2222")
		tool.ErrCh <- err
	}
}

// tool.TG.Fail(err)
// tool.TG.Shutdown()
