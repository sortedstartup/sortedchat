import { useState, useEffect, useRef } from 'react';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Mic, MicOff, Phone, PhoneOff, Volume2 } from "lucide-react";
import { iceCandidate, offerRequest } from '@/store/realtime';

interface RealtimeAudioModalProps {
  isOpen: boolean;
  onClose: () => void;
}

type ConnectionState = 'disconnected' | 'connecting' | 'connected' | 'failed';

export function RealtimeAudioModal({ isOpen, onClose }: RealtimeAudioModalProps) {
  const [connectionState, setConnectionState] = useState<ConnectionState>('disconnected');
  const [isListening, setIsListening] = useState(false);
  const [isMuted, setIsMuted] = useState(false);
  const [statusMessage, setStatusMessage] = useState('Click Connect to start');

  const audioElementRef = useRef<HTMLAudioElement | null>(null);
  const dataChannelRef = useRef<RTCDataChannel | null>(null);
  const pcRef = useRef<RTCPeerConnection | null>(null);
  const streamRef = useRef<MediaStream | null>(null);

  const handleConnection = async () => {
    try {
      setConnectionState('connecting');
      setStatusMessage('Connecting...');

      const pc = new RTCPeerConnection();
      pcRef.current = pc;

      pc.onicecandidate = async (event) => {
        if (event.candidate) {
          iceCandidate(JSON.stringify({ candidate: event.candidate }));
        }
      };

      pc.onconnectionstatechange = () => {
        console.log("Connection state:", pc.connectionState);
        setStatusMessage(`Connection: ${pc.connectionState}`);

        if (pc.connectionState === 'connected') {
          setConnectionState('connected');
        } else if (pc.connectionState === 'failed' || pc.connectionState === 'disconnected') {
          setConnectionState('failed');
          setIsListening(false);
        }
      };

      pc.ontrack = (e) => {
        console.log("🔊 Audio received");
        if (!audioElementRef.current) {
          audioElementRef.current = document.createElement("audio");
          audioElementRef.current.autoplay = true;
        }
        audioElementRef.current.srcObject = e.streams[0];
      };

      // Add microphone
      const stream = await navigator.mediaDevices.getUserMedia({ audio: true });
      streamRef.current = stream;
      stream.getTracks().forEach(track => {
        pc.addTrack(track, stream);
      });

      // Data channel
      const dataChannel = pc.createDataChannel("oai-events");
      dataChannelRef.current = dataChannel;

      dataChannel.onopen = () => {
        console.log("✅ Data channel open");
        dataChannel.send(JSON.stringify({
          type: "session.update",
          session: {
            type: "realtime",
            modalities: ["audio"],
            voice: "alloy",
            turn_detection: { type: "server_vad" },
            instructions: "You are helpful. Answer in ENGLISH only."
          }
        }));
      };

      dataChannel.onmessage = async (event) => {
        console.log("event", event);

        try {
          let text;
          if (typeof event.data === "string") {
            console.log("event.data", event.data);
            text = event.data;
          } else if (event.data instanceof ArrayBuffer) {
            text = new TextDecoder().decode(event.data);
          } else if (event.data instanceof Blob) {
            text = await event.data.text();
          } else {
            console.warn("Unknown data type:", typeof event.data);
            return;
          }

          // Now try to parse the text as JSON
          try {
            const message = JSON.parse(text);

            if (message.type === "OpenAI:usage") {
              console.log("OpenAI:usage", message.data);
            }
            if (message.type === "OpenAI:input_details") {
              console.log("OpenAI:input_details", message.data);
            }
            if (message.type === "Gemini:output_details") {
              console.log("OpenAI:output_details", message.data);
            }

            

            console.log("📨", message.type);
            if (message.type === "session.created") {
              setStatusMessage("✅ Ready - Speak!");
              setConnectionState('connected');
              setIsListening(true);
            }
          } catch (parseErr) {
            console.log("Non-JSON data channel message:", text);
          }
        } catch (err) {
          console.error("dataChannel.onmessage error:", err);
        }
      };


      const offer = await pc.createOffer();
      await pc.setLocalDescription(offer);

      const response = await offerRequest(String(offer.sdp));

      const answer = {
        type: "answer",
        sdp: String(response)
      }


      await pc.setRemoteDescription(answer as unknown as RTCSessionDescription);

    } catch (error) {
      console.error("Failed:", error);
      setConnectionState('failed');
      setStatusMessage(`Failed: ${error}`);
    }
  };

  const handleDisconnect = () => {
    if (pcRef.current) {
      pcRef.current.close();
      pcRef.current = null;
    }

    if (streamRef.current) {
      streamRef.current.getTracks().forEach(track => track.stop());
      streamRef.current = null;
    }

    if (audioElementRef.current) {
      audioElementRef.current.srcObject = null;
      audioElementRef.current = null;
    }

    dataChannelRef.current = null;
    setConnectionState('disconnected');
    setIsListening(false);
    setIsMuted(false);
    setStatusMessage('Click Connect to start');
  };

  const toggleMute = () => {
    if (streamRef.current) {
      const audioTrack = streamRef.current.getAudioTracks()[0];
      if (audioTrack) {
        audioTrack.enabled = isMuted;
        setIsMuted(!isMuted);
      }
    }
  };

  const handleClose = () => {
    handleDisconnect();
    onClose();
  };

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      handleDisconnect();
    };
  }, []);

  const getStatusColor = () => {
    switch (connectionState) {
      case 'connected':
        return 'text-green-600';
      case 'connecting':
        return 'text-yellow-600';
      case 'failed':
        return 'text-red-600';
      default:
        return 'text-gray-600';
    }
  };

  const getStatusBgColor = () => {
    switch (connectionState) {
      case 'connected':
        return 'bg-green-50 border-green-200';
      case 'connecting':
        return 'bg-yellow-50 border-yellow-200';
      case 'failed':
        return 'bg-red-50 border-red-200';
      default:
        return 'bg-gray-50 border-gray-200';
    }
  };

  return (
    <Dialog open={isOpen} onOpenChange={handleClose}>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <DialogTitle className="flex items-center space-x-2">
            <Volume2 className="h-5 w-5" />
            <span>Realtime Audio</span>
          </DialogTitle>
        </DialogHeader>

        <div className="space-y-6">
          {/* Status Display */}
          <div className={`p-4 rounded-lg border ${getStatusBgColor()}`}>
            <div className="flex items-center space-x-2">
              <div className={`w-3 h-3 rounded-full ${connectionState === 'connected' ? 'bg-green-500 animate-pulse' :
                  connectionState === 'connecting' ? 'bg-yellow-500 animate-pulse' :
                    connectionState === 'failed' ? 'bg-red-500' : 'bg-gray-400'
                }`} />
              <span className={`text-sm font-medium ${getStatusColor()}`}>
                {statusMessage}
              </span>
            </div>
          </div>

          {/* Listening Indicator */}
          {isListening && connectionState === 'connected' && (
            <div className="flex items-center justify-center space-x-2">
              <div className="flex space-x-1">
                <div className="w-2 h-8 bg-blue-500 rounded animate-pulse" style={{ animationDelay: '0ms' }} />
                <div className="w-2 h-6 bg-blue-400 rounded animate-pulse" style={{ animationDelay: '100ms' }} />
                <div className="w-2 h-10 bg-blue-500 rounded animate-pulse" style={{ animationDelay: '200ms' }} />
                <div className="w-2 h-4 bg-blue-400 rounded animate-pulse" style={{ animationDelay: '300ms' }} />
                <div className="w-2 h-8 bg-blue-500 rounded animate-pulse" style={{ animationDelay: '400ms' }} />
              </div>
              <span className="text-sm text-blue-600 font-medium">Listening...</span>
            </div>
          )}

          {/* Controls */}
          <div className="flex justify-center space-x-4">
            {connectionState === 'disconnected' || connectionState === 'failed' ? (
              <Button
                onClick={handleConnection}
                className="bg-green-600 hover:bg-green-700 text-white px-6"
              // disabled={connectionState === 'connected'}
              >
                <Phone className="h-4 w-4 mr-2" />
                Connect
              </Button>
            ) : (
              <>
                <Button
                  onClick={toggleMute}
                  variant="outline"
                  size="sm"
                  className={isMuted ? 'bg-red-50 border-red-200 text-red-600' : ''}
                >
                  {isMuted ? <MicOff className="h-4 w-4" /> : <Mic className="h-4 w-4" />}
                </Button>

                <Button
                  onClick={handleDisconnect}
                  variant="destructive"
                  className="px-6"
                >
                  <PhoneOff className="h-4 w-4 mr-2" />
                  Disconnect
                </Button>
              </>
            )}
          </div>

          {/* Instructions */}
          <div className="text-xs text-gray-500 text-center space-y-1">
            <p>Click Connect to start voice conversation</p>
            <p>Speak naturally - the AI will respond with voice</p>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}