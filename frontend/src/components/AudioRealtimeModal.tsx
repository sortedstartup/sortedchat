import { useState, useEffect, useRef } from 'react';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Phone, PhoneOff, Volume2, ChevronDown } from "lucide-react";
import { iceCandidate, offerRequest, $isConnected } from '@/store/realtime';
import { useStore } from '@nanostores/react';

interface RealtimeAudioModalProps {
  isOpen: boolean;
  onClose: () => void;
}

type Provider = 'openai' | 'gemini';

export function RealtimeAudioModal({ isOpen, onClose }: RealtimeAudioModalProps) {
  const isConnected = useStore($isConnected);
  const [provider, setProvider] = useState<Provider>('openai');
  const [showDropdown, setShowDropdown] = useState(false);
  const [statusMessage, setStatusMessage] = useState('Select provider and click Connect');
  const [isConnecting, setIsConnecting] = useState(false);
  const [inputTokens, setInputTokens] = useState(0);
  const [outputTokens, setOutputTokens] = useState(0);

  const pcRef = useRef<RTCPeerConnection | null>(null);
  const streamRef = useRef<MediaStream | null>(null);

  const handleConnection = async () => {
    try {
      setIsConnecting(true);
      setStatusMessage('Connecting...');

      const pc = new RTCPeerConnection();
      pcRef.current = pc;

      pc.onicecandidate = async (event) => {
        if (event.candidate) {
          await iceCandidate(JSON.stringify( event.candidate.toJSON() ));
        }
      };

      pc.onconnectionstatechange = () => {
        if (pc.connectionState === 'connected') {
          // $isConnected.set(true);
          setIsConnecting(false);
          setStatusMessage('');
        } else if (pc.connectionState === 'failed' || pc.connectionState === 'disconnected') {
          handleDisconnect();
          $isConnected.set(false);
          setIsConnecting(false);
          setStatusMessage('Connection failed');
        }
      };

      pc.ontrack = (e) => {
        const audio = document.createElement("audio");
        audio.autoplay = true;
        audio.srcObject = e.streams[0];
      };

      const stream = await navigator.mediaDevices.getUserMedia({ audio: true });
      streamRef.current = stream;
      stream.getTracks().forEach(track => pc.addTrack(track, stream));

      const dataChannel = pc.createDataChannel("oai-events");

      dataChannel.onopen = () => {
        dataChannel.send(JSON.stringify({
          type: "session.update",
          session: {
            type: "realtime",
            modalities: ["audio"],
            voice: provider === 'openai' ? "alloy" : "en-US-Neural2-A",
            turn_detection: { type: "server_vad" },
            instructions: "You are helpful. Answer in ENGLISH only."
          }
        }));
      };

      dataChannel.onmessage = async (event) => {
        try {
          let text = typeof event.data === "string" ? event.data :
                    event.data instanceof ArrayBuffer ? new TextDecoder().decode(event.data) :
                    event.data instanceof Blob ? await event.data.text() : "";

          const message = JSON.parse(text);
          
          if (message.type === "OpenAI:input_details") {
            console.log("OpenAI:input_details", message.data);
            if (message.data.audio_tokens) setInputTokens(prev => prev + message.data.audio_tokens);
          }
          if (message.type === "OpenAI:output_details") {
            console.log("OpenAI:output_details", message.data);
            if (message.data.audio_tokens) setOutputTokens(prev => prev + message.data.audio_tokens);
          }
          if (message.type === "Connection_closed") {
            handleDisconnect();
            $isConnected.set(false);
            setIsConnecting(false);
            setStatusMessage('Connection closed');
          }
        } catch (err) {
          console.error("Message handling error:", err);
        }
      };

      const offer = await pc.createOffer();
      await pc.setLocalDescription(offer);

      const response = await offerRequest(String(offer.sdp), provider);
      await pc.setRemoteDescription({ type: "answer", sdp: String(response) } as RTCSessionDescription);

    } catch (error) {
      setIsConnecting(false);
      setStatusMessage(`Connection failed. Verify your API keys.`);
      $isConnected.set(false);
    }
  };

  const handleDisconnect = () => {
    pcRef.current?.close();
    streamRef.current?.getTracks().forEach(track => track.stop());
    
    pcRef.current = null;
    streamRef.current = null;

    $isConnected.set(false);
    setIsConnecting(false);
    setStatusMessage('Select provider and click Connect');
    setInputTokens(0);
    setOutputTokens(0);
  };

  const handleClose = () => {
    handleDisconnect();
    onClose();
  };

  useEffect(() => {
    return handleDisconnect;
  }, []);

  return (
    <Dialog open={isOpen} onOpenChange={handleClose}>
      <DialogContent className="max-w-2xl w-full max-h-[90vh]">
        <DialogHeader className="pb-6">
          <DialogTitle className="flex items-center gap-3 text-lg">
            <div className="p-2 bg-primary/10 rounded-lg">
              <Volume2 className="h-5 w-5 text-primary" />
            </div>
            Realtime Audio Chat
          </DialogTitle>
        </DialogHeader>

        <div className="space-y-8">
          {/* Provider Selection */}
          <div className="space-y-2">
            <label className="text-sm font-medium text-foreground">AI Provider</label>
            <div className="relative">
              <button
                onClick={() => setShowDropdown(!showDropdown)}
                disabled={isConnected}
                className="w-full p-3 bg-card border border-border rounded-lg flex items-center justify-between hover:border-ring disabled:opacity-50 disabled:cursor-not-allowed"
              >
                <div className="flex items-center gap-2">
                  <div className={`w-2 h-2 rounded-full ${provider === 'openai' ? 'bg-green-500' : 'bg-blue-500'}`} />
                  <span className="font-medium capitalize text-foreground">{provider}</span>
                </div>
                <ChevronDown className="h-4 w-4 text-muted-foreground" />
              </button>
              
              {showDropdown && (
                <div className="absolute top-full left-0 right-0 mt-1 bg-popover border border-border rounded-lg shadow-lg z-10">
                  <button
                    onClick={() => { setProvider('openai'); setShowDropdown(false); }}
                    className="w-full p-3 hover:bg-accent text-popover-foreground flex items-center gap-2 border-b border-border"
                  >
                    <div className="w-2 h-2 rounded-full bg-green-500" />
                    <span>OpenAI</span>
                  </button>
                  <button
                    onClick={() => { setProvider('gemini'); setShowDropdown(false); }}
                    className="w-full p-3 hover:bg-accent text-popover-foreground flex items-center gap-2"
                  >
                    <div className="w-2 h-2 rounded-full bg-blue-500" />
                    <span>Gemini</span>
                  </button>
                </div>
              )}
            </div>
          </div>

          {/* Status */}
          {statusMessage && (
            <div className="p-4 rounded-xl border border-border bg-muted">
              <div className="flex items-center gap-3">
                <div className={`w-3 h-3 rounded-full ${isConnected ? 'bg-green-500 animate-pulse' : 'bg-muted-foreground'}`} />
                <span className={`font-medium ${isConnected ? 'text-green-600 dark:text-green-400' : 'text-muted-foreground'}`}>{statusMessage}</span>
              </div>
            </div>
          )}

          {/* Token Info */}
          {(inputTokens > 0 || outputTokens > 0) && (
            <div className="p-3 bg-primary/10 border border-primary/20 rounded-lg">
              <div className="flex justify-between text-sm text-primary font-mono">
                <span>Input: {inputTokens}</span>
                <span>Output: {outputTokens}</span>
              </div>
            </div>
          )}

          {/* Audio Visualizer */}
          {isConnected && (
            <div className="flex items-center justify-center py-6">
              <div className="flex items-end gap-1 mr-3">
                {[0, 100, 200, 300, 400].map((delay, i) => (
                  <div
                    key={i}
                    className="w-1 bg-primary rounded-full animate-pulse"
                    style={{ 
                      height: `${[32, 24, 40, 16, 32][i]}px`,
                      animationDelay: `${delay}ms` 
                    }}
                  />
                ))}
              </div>
              <span className="text-primary font-medium">Listening...</span>
            </div>
          )}

          {/* Controls */}
          <div className="flex justify-center gap-4 pt-4">
            {!isConnected ? (
              <Button
                onClick={handleConnection}
                disabled={isConnecting}
                className="bg-green-600 hover:bg-green-700 dark:bg-green-500 dark:hover:bg-green-600 px-8 py-3 text-white font-medium disabled:opacity-50 disabled:cursor-not-allowed"
              >
                <Phone className="h-4 w-4 mr-2" />
                {isConnecting ? 'Connecting...' : 'Connect'}
              </Button>
            ) : (
              <Button
                onClick={handleDisconnect}
                variant="destructive"
                className="px-8 py-3 font-medium"
              >
                <PhoneOff className="h-4 w-4 mr-2" />
                Disconnect
              </Button>
            )}
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}