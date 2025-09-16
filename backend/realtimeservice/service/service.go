package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"sortedstartup/realtimeservice/dao"
	pb "sortedstartup/realtimeservice/proto"
	"time"

	"github.com/pion/webrtc/v3"
	"github.com/pion/webrtc/v3/pkg/media/oggwriter"
)

type RealtimeService struct {
	dao dao.Dao
}

func NewRealtimeService(dao dao.Dao) *RealtimeService {
	return &RealtimeService{dao: dao}
}

func (s *RealtimeService) Init(config *dao.Config) {
	slog.Info("RealtimeService: Init", "config", config)
}

var OPENAI_API_KEY = os.Getenv("OPENAI_API_KEY")
var clientPC *webrtc.PeerConnection
var openaiPC *webrtc.PeerConnection
var clientOutboundTrack *webrtc.TrackLocalStaticRTP
var clientDataChannel *webrtc.DataChannel
var openaiDataChannel *webrtc.DataChannel
var openAIConnected = false

func (s *RealtimeService) Offer(offer *pb.OfferRequest) (string, error) {
	fmt.Println("🚀 Setting up WebRTC...")

	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{URLs: []string{"stun:stun.l.google.com:19302"}},
		},
	}

	var err error

	clientPC, err = webrtc.NewPeerConnection(config)
	if err != nil {
		return "", err
	}

	openaiPC, err = webrtc.NewPeerConnection(config)
	if err != nil {
		return "", err
	}

	// ICE forwarding
	clientPC.OnICECandidate(func(c *webrtc.ICECandidate) {
		if c != nil {
			openaiPC.AddICECandidate(c.ToJSON())
		}
	})

	openaiPC.OnICECandidate(func(c *webrtc.ICECandidate) {
		if c != nil {
			clientPC.AddICECandidate(c.ToJSON())
		}
	})

	// Connection status
	clientPC.OnConnectionStateChange(func(s webrtc.PeerConnectionState) {
		fmt.Printf("📡 Client: %s\n", s)
	})

	openaiPC.OnConnectionStateChange(func(s webrtc.PeerConnectionState) {
		fmt.Printf("🤖 OpenAI: %s\n", s)
	})

	// THIS IS THE KEY: Just like Node.js - add tracks when received from client
	// Audio from Client -> OpenAI (and save to file)
	clientPC.OnTrack(func(remoteTrack *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
		fmt.Println("🎤 Audio: Client -> OpenAI (relay + save)")
		fmt.Printf("Track details: kind=%s, id=%s\n", remoteTrack.Kind(), remoteTrack.ID())

		// Create local track for OpenAI
		localTrack, err := webrtc.NewTrackLocalStaticRTP(
			remoteTrack.Codec().RTPCodecCapability,
			remoteTrack.ID(),
			remoteTrack.StreamID(),
		)
		if err != nil {
			fmt.Printf("Error creating track: %v\n", err)
			return
		}

		if _, err := openaiPC.AddTrack(localTrack); err != nil {
			fmt.Printf("Error adding track to OpenAI PC: %v\n", err)
			return
		}

		// Save client audio to OGG (like we do for OpenAI)
		oggFile, err := oggwriter.New("client-output.ogg", 48000, 2)
		if err != nil {
			fmt.Println("oggwriter error:", err)
			return
		}

		go func() {
			defer oggFile.Close()
			for {
				pkt, _, readErr := remoteTrack.ReadRTP()
				if readErr != nil {
					fmt.Println("Client audio read end:", readErr)
					break
				}

				// 1. Save to OGG
				if err := oggFile.WriteRTP(pkt); err != nil {
					fmt.Println("ogg write err:", err)
					break
				}

				// 2. Relay to OpenAI
				if err := localTrack.WriteRTP(pkt); err != nil {
					fmt.Printf("forward to OpenAI err: %v\n", err)
				}
			}

		}()

		// Connect to OpenAI once we have audio
		if !openAIConnected {
			connectToOpenAIWithAudio()
		}
	})

	// Audio from OpenAI -> Client
	// NOTE: write into pre-created clientOutboundTrack (created below during answer creation)
	openaiPC.OnTrack(func(remoteTrack *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
		fmt.Println("🔊 Audio: OpenAI -> Client (relay + save)")

		oggFile, err := oggwriter.New("openai-output.ogg", 48000, 2)
		if err != nil {
			fmt.Println("oggwriter error:", err)
			return
		}

		go func() {
			defer oggFile.Close()

			for {
				pkt, _, readErr := remoteTrack.ReadRTP()
				if readErr != nil {
					fmt.Println("OpenAI audio read end:", readErr)
					break
				}

				// 1. Save to ogg
				if err := oggFile.WriteRTP(pkt); err != nil {
					fmt.Println("ogg write err:", err)
					break
				}

				// 2. Relay to browser
				if clientOutboundTrack != nil {
					if writeErr := clientOutboundTrack.WriteRTP(pkt); writeErr != nil {
						fmt.Printf("relay write err: %v\n", writeErr)
					}
				}
			}
		}()
	})

	// clientPC.OnDataChannel(func(dc *webrtc.DataChannel) {
	// 	clientDataChannel = dc
	// 	fmt.Println("📡 Client data channel")

	// 	dc.OnMessage(func(msg webrtc.DataChannelMessage) {
	// 		fmt.Println("openai data channel state", openaiDataChannel.ReadyState())
	// 		if openaiDataChannel != nil {
	// 			// forward raw bytes (may be string or binary)
	// 			fmt.Println("📥 forwarding", string(msg.Data))
	// 			if err := openaiDataChannel.Send(msg.Data); err != nil {
	// 				fmt.Printf("forward to openai datachannel err: %v\n", err)
	// 			}
	// 		}
	// 	})
	// })

	// openaiDataChannel, _ = openaiPC.CreateDataChannel("openai", nil)
	// openaiDataChannel.OnMessage(func(msg webrtc.DataChannelMessage) {
	// 	// Try parse JSON for logging; if not JSON, print raw
	// 	var message map[string]interface{}
	// 	if err := json.Unmarshal(msg.Data, &message); err != nil {
	// 		fmt.Printf("📥 OpenAI (raw): %s\n", string(msg.Data))
	// 	} else {
	// 		msgType, _ := message["type"].(string)
	// 		fmt.Printf("📥 OpenAI Event: %s\n", msgType)

	// 		// Extract and print usage tokens for response.done events
	// 		if msgType == "response.done" {

	// 			printResponseStatus(message)
	// 			dumpUsageToFile(message)

	// 		}

	// 		prettyJSON, err := json.MarshalIndent(message, "", "  ")
	// 		if err == nil {
	// 			fmt.Printf("📥 OpenAI Full Message:\n%s\n", string(prettyJSON))
	// 		}
	// 	}
	// 	// Forward to client (raw bytes)
	// 	if clientDataChannel != nil && clientDataChannel.ReadyState() == webrtc.DataChannelStateOpen {
	// 		if err := clientDataChannel.Send(msg.Data); err != nil {
	// 			fmt.Printf("forward to client datachannel err: %v\n", err)
	// 		}
	// 	}
	// })

	// openaiDataChannel.OnClose(func() {
	// 	fmt.Println("❌ OpenAI data channel closed")
	// })

	if err := clientPC.SetRemoteDescription(webrtc.SessionDescription{
		Type: webrtc.SDPTypeOffer,
		SDP:  offer.Offer,
	}); err != nil {
		return "", err
	}

	// --- Create placeholder outbound track BEFORE creating the answer ---
	// Typical browser codec is opus; create a track the browser will expect
	placeholder, err := webrtc.NewTrackLocalStaticRTP(
		webrtc.RTPCodecCapability{MimeType: "audio/opus", ClockRate: 48000, Channels: 2},
		"openai-audio", "pion-openai",
	)
	if err != nil {
		fmt.Printf("Error creating placeholder outbound track: %v\n", err)
	} else {
		clientOutboundTrack = placeholder
		if _, err := clientPC.AddTrack(clientOutboundTrack); err != nil {
			fmt.Printf("Error adding placeholder outbound track to clientPC: %v\n", err)
			// keep going; audio won't reach client unless added
		} else {
			fmt.Println("➕ Added placeholder outbound track to clientPC (pre-negotiation)")
		}
	}

	clientAnswer, err := clientPC.CreateAnswer(nil)
	if err != nil {
		return "", err
	}
	if err := clientPC.SetLocalDescription(clientAnswer); err != nil {
		return "", err
	}

	return clientAnswer.SDP, nil
}

func connectToOpenAIWithAudio() {
	openAIConnected = true
	fmt.Println("🔗 Now connecting to OpenAI with audio...")

	// Create offer AFTER we have audio tracks (like Node.js)
	openaiOffer, err := openaiPC.CreateOffer(nil)
	if err != nil {
		fmt.Printf("Error creating OpenAI offer: %v\n", err)
		return
	}

	openaiPC.SetLocalDescription(openaiOffer)
	fmt.Printf("OpenAI Offer created with %d characters\n", len(openaiOffer.SDP))

	// Get token
	token, err := getOpenAIToken()
	if err != nil {
		fmt.Printf("Token error: %v\n", err)
		return
	}

	// Connect to OpenAI
	if err := connectToOpenAI(openaiOffer, token); err != nil {
		fmt.Printf("OpenAI error: %v\n", err)
		return
	}

	fmt.Println("✅ Connected to OpenAI!")
}

func getOpenAIToken() (string, error) {
	config := map[string]interface{}{
		"session": map[string]interface{}{
			"type":  "realtime",
			"model": "gpt-4o-mini-realtime-preview",
			"audio": map[string]interface{}{
				"output": map[string]interface{}{"voice": "alloy"},
				"input":  map[string]interface{}{"turn_detection": map[string]interface{}{"type": "server_vad"}},
			},
			"instructions": "You are helpful. Answer briefly in ENGLISH only.",
		},
	}

	data, _ := json.Marshal(config)
	req, _ := http.NewRequest("POST", "https://api.openai.com/v1/realtime/client_secrets", bytes.NewBuffer(data))
	req.Header.Set("Authorization", "Bearer "+OPENAI_API_KEY)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	if value, ok := result["value"].(string); ok {
		return value, nil
	}
	return "", fmt.Errorf("no token received")
}

func connectToOpenAI(offer webrtc.SessionDescription, token string) error {
	req, _ := http.NewRequest("POST",
		"https://api.openai.com/v1/realtime/calls?model=gpt-4o-mini-realtime-preview",
		bytes.NewBufferString(offer.SDP))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/sdp")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	responseBody := buf.String()

	fmt.Printf("Response Status: %d\n", resp.StatusCode)
	fmt.Printf("Response Body Length: %d chars\n", len(responseBody))

	// Accept both 200 and 201
	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		return fmt.Errorf("status %d: %s", resp.StatusCode, responseBody)
	}

	// Success! Set the SDP answer from OpenAI
	fmt.Println("✅ Got SDP answer from OpenAI!")
	return openaiPC.SetRemoteDescription(webrtc.SessionDescription{
		Type: webrtc.SDPTypeAnswer,
		SDP:  responseBody,
	})
}

func printResponseStatus(message map[string]interface{}) {
	if response, ok := message["response"].(map[string]interface{}); ok {
		if status, ok := response["status"].(string); ok {
			fmt.Printf("🔄 Response Status: %s\n", status)
		}

		// Also print status_details if available
		if statusDetails := response["status_details"]; statusDetails != nil {
			fmt.Printf("📋 Status Details: %v\n", statusDetails)
		}
	}
}

func dumpUsageToFile(message map[string]interface{}) {
	// Open file for appending (create if not exists)
	file, err := os.OpenFile("token_usage.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Error opening token file: %v\n", err)
		return
	}
	defer file.Close()

	// Navigate to response.usage and write detailed breakdown to file
	if response, ok := message["response"].(map[string]interface{}); ok {
		if usage, ok := response["usage"].(map[string]interface{}); ok {
			// Write timestamp
			timestamp := time.Now().Format("2006-01-02 15:04:05")
			fmt.Fprintf(file, "[%s] ", timestamp)

			// Main token counts
			if totalTokens, ok := usage["total_tokens"].(float64); ok {
				fmt.Fprintf(file, "Total:%.0f ", totalTokens)
			}
			if inputTokens, ok := usage["input_tokens"].(float64); ok {
				fmt.Fprintf(file, "Input:%.0f ", inputTokens)
			}
			if outputTokens, ok := usage["output_tokens"].(float64); ok {
				fmt.Fprintf(file, "Output:%.0f ", outputTokens)
			}

			// Input token details
			if inputDetails, ok := usage["input_token_details"].(map[string]interface{}); ok {
				fmt.Fprintf(file, "| InputDetails: ")
				if textTokens, ok := inputDetails["text_tokens"].(float64); ok {
					fmt.Fprintf(file, "Text:%.0f ", textTokens)
				}
				if audioTokens, ok := inputDetails["audio_tokens"].(float64); ok {
					fmt.Fprintf(file, "Audio:%.0f ", audioTokens)
				}
				if cachedTokens, ok := inputDetails["cached_tokens"].(float64); ok {
					fmt.Fprintf(file, "Cached:%.0f ", cachedTokens)
				}
			}

			// Output token details
			if outputDetails, ok := usage["output_token_details"].(map[string]interface{}); ok {
				fmt.Fprintf(file, "| OutputDetails: ")
				if textTokens, ok := outputDetails["text_tokens"].(float64); ok {
					fmt.Fprintf(file, "Text:%.0f ", textTokens)
				}
				if audioTokens, ok := outputDetails["audio_tokens"].(float64); ok {
					fmt.Fprintf(file, "Audio:%.0f ", audioTokens)
				}
			}

			fmt.Fprintf(file, "\n")
		}
	}
}

func (s *RealtimeService) IceCandidate(candidate string) (string, error) {
	if clientPC == nil {
		return "", fmt.Errorf("client peer connection not initialized")
	}

	// Check if remote description is set
	if clientPC.RemoteDescription() == nil {
		return "", fmt.Errorf("remote description not set, cannot add ICE candidate")
	}

	if err := clientPC.AddICECandidate(webrtc.ICECandidateInit{
		Candidate: candidate,
	}); err != nil {
		fmt.Println("Error adding ICE candidate", err)
		return "", err
	}
	fmt.Println("Connecte")
	return "connected", nil
}
