package agent

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/sensu/sensu-go/command"
	"github.com/sensu/sensu-go/types"
)

// TODO(greg): At some point, we're going to need max parallelism.
func (a *Agent) handleCheck(payload []byte) error {
	event := &types.Event{}
	err := json.Unmarshal(payload, event)
	if err != nil {
		return err
	}

	if event.Check == nil {
		return errors.New("no check found in event")
	}

	if event.Check.Command != "" {
		go func() {
			log.Println("executing check: ", event.Check.Name)
			ex := &command.Execution{}
			event.Check.Executed = time.Now().Unix()
			_, err := command.ExecuteCommand(context.Background(), ex)
			if err != nil {
				event.Check.Output = err.Error()
			} else {
				event.Check.Output = ex.Output
			}

			event.Check.Duration = ex.Duration
			msg, err := json.Marshal(event)
			if err != nil {
				log.Print("error marshaling check result: ", err.Error())
				return
			}

			a.sendMessage(types.EventType, msg)
		}()
	}

	return nil
}
