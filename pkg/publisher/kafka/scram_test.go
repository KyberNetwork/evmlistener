package kafka

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/xdg-go/scram"
)

// TestXDGSCRAMClient_SHA512Handshake drives XDGSCRAMClient through a full
// SCRAM-SHA-512 exchange against a real scram.Server conversation, so a
// regression that breaks message framing (e.g. wiring the wrong
// HashGeneratorFcn, or mishandling Begin/Step/Done) fails here instead of
// only showing up as an opaque auth rejection from a live Kafka broker.
func TestXDGSCRAMClient_SHA512Handshake(t *testing.T) {
	const username, password = "evmlistener", "s3cret"

	kf := scram.KeyFactors{Salt: "c2FsdA==", Iters: 4096}
	setupClient, err := scram.SHA512.NewClient(username, password, "")
	require.NoError(t, err)
	storedCreds, err := setupClient.GetStoredCredentialsWithError(kf)
	require.NoError(t, err)

	server, err := scram.SHA512.NewServer(func(user string) (scram.StoredCredentials, error) {
		require.Equal(t, username, user)
		return storedCreds, nil
	})
	require.NoError(t, err)
	serverConv := server.NewConversation()

	client := &XDGSCRAMClient{HashGeneratorFcn: scram.SHA512}
	require.NoError(t, client.Begin(username, password, ""))

	challenge := ""
	for !client.Done() {
		msg, err := client.Step(challenge)
		require.NoError(t, err)
		if client.Done() {
			// Final step only validates the server's signature locally;
			// there is nothing left to send.
			break
		}

		challenge, err = serverConv.Step(msg)
		require.NoError(t, err)
	}

	require.True(t, client.Valid())
	require.True(t, serverConv.Valid())
}
