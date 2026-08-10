package kafka

import (
	"github.com/xdg-go/scram"
)

// XDGSCRAMClient adapts xdg-go/scram to sarama's SCRAMClient interface.
// Step and Done are satisfied via the embedded *scram.ClientConversation.
type XDGSCRAMClient struct {
	*scram.Client
	*scram.ClientConversation
	scram.HashGeneratorFcn
}

func (x *XDGSCRAMClient) Begin(userName, password, authzID string) error {
	client, err := x.NewClient(userName, password, authzID)
	if err != nil {
		return err
	}
	x.Client = client
	x.ClientConversation = x.NewConversation()
	return nil
}
