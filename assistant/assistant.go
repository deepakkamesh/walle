/* Package assistance initializes a new google assistant.
*
 */
package assistant

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"golang.org/x/oauth2"

	"github.com/golang/glog"

	"google.golang.org/api/option"
	"google.golang.org/api/transport"
	embedded "google.golang.org/genproto/googleapis/assistant/embedded/v1alpha1"
)

const (
	MAX_RUNTIME  = 240
	API_ENDPOINT = "embeddedassistant.googleapis.com:443"
)

type JSONToken struct {
	Installed struct {
		ClientID                string   `json:"client_id"`
		ProjectID               string   `json:"project_id"`
		AuthURI                 string   `json:"auth_uri"`
		TokenURI                string   `json:"token_uri"`
		AuthProviderX509CertURL string   `json:"auth_provider_x509_cert_url"`
		ClientSecret            string   `json:"client_secret"`
		RedirectUris            []string `json:"redirect_uris"`
	} `json:"installed"`
}

type GAssistant struct {
	audOut      chan bytes.Buffer
	audIn       chan bytes.Buffer
	oauthConfig *oauth2.Config
	oauthToken  *oauth2.Token
	secretsFile string
	scopes      []string
}

func New(out chan bytes.Buffer, in chan bytes.Buffer, secretsFile string, scope string) *GAssistant {

	return &GAssistant{
		audOut:      out,
		audIn:       in,
		secretsFile: secretsFile,
		scopes:      strings.Split(scope, ","),
	}
}

func (s *GAssistant) loadTokenSource() error {
	f, err := os.Open("oauthTokenCache")
	if err != nil {
		glog.Errorf("failed to load the token source  %v", err)
		return err
	}
	defer f.Close()
	var token oauth2.Token
	if err = json.NewDecoder(f).Decode(&token); err != nil {
		return err
	}
	s.oauthToken = &token
	return nil
}

// TODO: Need to get a single auth function for all google api authentication.
// TODO: Need to return a error.
func (s *GAssistant) Auth() error {
	f, err := os.Open(s.secretsFile)
	if err != nil {
		glog.Fatalf("Failed to open secrets file:%v", err)
	}
	defer f.Close()
	var token JSONToken
	if err = json.NewDecoder(f).Decode(&token); err != nil {
		glog.Fatalf("failed to decode json token: %v", err)
	}

	s.oauthConfig = &oauth2.Config{
		ClientID:     token.Installed.ClientID,
		ClientSecret: token.Installed.ClientSecret,
		Scopes:       s.scopes,
		RedirectURL:  "http://localhost:8080", // TODO: Verify if this can be replaced with token variable.
		Endpoint: oauth2.Endpoint{
			AuthURL:  token.Installed.AuthURI,
			TokenURL: token.Installed.TokenURI,
		},
	}

	// check if we have an oauth file on disk
	if true {
		err = s.loadTokenSource()
		if err == nil {
			fmt.Println("Launching the Google Assistant using cached credentials")
			return nil
		}
		fmt.Println("Failed to load the token source", err)
		fmt.Println("Continuing program without cached credentials")
	}
	return nil
}

func (s *GAssistant) ConverseWithAssistant() *bytes.Buffer {
	glog.V(2).Infof("Waiting for new conversation...")
	var convState []byte
	micStopCh := make(chan struct{})

	ctx, canceler := context.WithTimeout(context.Background(), MAX_RUNTIME*time.Second)
	tokenSource := s.oauthConfig.TokenSource(ctx, s.oauthToken)

	conn, err := transport.DialGRPC(ctx,
		option.WithTokenSource(tokenSource),
		option.WithEndpoint(API_ENDPOINT),
		option.WithScopes(s.scopes[0]),
	)
	if err != nil {
		glog.Fatalf("Failed to connect with rpc endpoint: %v", err)
	}

	// Clean up before finishing up.
	defer func() {
		glog.V(2).Infof("End of conversation. Cleaning up.")
		//	micStopCh <- struct{}{}
		conn.Close()
		ctx.Done()
		canceler()
	}()

	assistant := embedded.NewEmbeddedAssistantClient(conn)
	config := &embedded.ConverseRequest_Config{
		Config: &embedded.ConverseConfig{
			AudioInConfig: &embedded.AudioInConfig{
				Encoding:        embedded.AudioInConfig_LINEAR16,
				SampleRateHertz: 16000,
			},
			AudioOutConfig: &embedded.AudioOutConfig{
				Encoding:         embedded.AudioOutConfig_LINEAR16,
				SampleRateHertz:  16000,
				VolumePercentage: 70,
			},
		},
	}

	// TODO: add conversation state.
	if len(convState) > 0 {
		glog.V(2).Infof("continuing conversation")
		config.Config.ConverseState = &embedded.ConverseState{ConversationState: convState}
	}

	conversation, err := assistant.Converse(ctx)
	if err != nil {
		glog.Errorf("Failed to setup the conversation: %v", err)
		return nil
	}

	req := &embedded.ConverseRequest{
		ConverseRequest: config,
	}
	if err := conversation.Send(req); err != nil {
		glog.Errorf("Failed to send to Google Assistant: %v", err)
		return nil
	}

	// Get Audio from mic and send to Assistant.
	go func() {

		sendAudioFunc := func() {
			buff := <-s.audIn // get Audio packets from microphone channel (Audio In).

			req = &embedded.ConverseRequest{
				ConverseRequest: &embedded.ConverseRequest_AudioIn{
					AudioIn: buff.Bytes(),
				},
			}
			if err := conversation.Send(req); err != nil {
				glog.Errorf("Failed to send audio to Google Assistant: %v", err)
			}
		}

		for {
			select {
			// Close the send of conversation and return from goroutine.
			case <-micStopCh:
				glog.V(2).Infof("Turning off mic")
				conversation.CloseSend()
				return
			default:
				sendAudioFunc()
			}
		}
	}()

	var fullAudio bytes.Buffer
	// Process audio returned from assistant.
	for {
		resp, err := conversation.Recv()

		switch {
		case err == io.EOF:
			glog.V(2).Infof("Got EOF from Assistant API")
			return &fullAudio

		case err != nil:
			glog.Errorf("Failed to recieve a response from assistant: %v", err)
			continue
		}

		if err := resp.GetError(); err != nil {
			glog.Errorf("Received error from the assistant: %v", err)
		}

		result := resp.GetResult()
		if result != nil {
			glog.V(1).Infof("data %s- %s", result.SpokenResponseText, result.SpokenRequestText)
		}

		if resp.GetEventType() == embedded.ConverseResponse_END_OF_UTTERANCE {
			glog.V(1).Info("Google said you are done, are you?!")
			micStopCh <- struct{}{}
		}
		audioOut := resp.GetAudioOut()
		if audioOut != nil {
			glog.V(3).Infof("audio out from the assistant (%d bytes)\n", len(audioOut.AudioData))
			signal := bytes.NewBuffer(audioOut.AudioData)
			fullAudio.Write(audioOut.AudioData)
			s.audOut <- *signal // Send audio to AudioOut Channel.
		}
	}
}
