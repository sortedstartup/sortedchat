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

  // Handle audio context for autoplay policies
  useEffect(() => {
    const handleUserInteraction = () => {
      const audioContext = new (window.AudioContext || (window as any).webkitAudioContext)();
      if (audioContext.state === 'suspended') {
        audioContext.resume().then(() => {
          console.log('Audio context resumed after user interaction');
        });
      }
    };

    document.addEventListener('click', handleUserInteraction, { once: true });
    document.addEventListener('keydown', handleUserInteraction, { once: true });
    
    return () => {
      document.removeEventListener('click', handleUserInteraction);
      document.removeEventListener('keydown', handleUserInteraction);
    };
  }, []);

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

      // FIXED: Proper audio track handling
      pc.ontrack = (event) => {
        console.log("🔊 Audio track received:", {
          kind: event.track.kind,
          id: event.track.id,
          streams: event.streams.length,
          enabled: event.track.enabled,
          muted: event.track.muted
        });
        
        if (event.track.kind === 'audio') {
          // Create or get audio element
          if (!audioElementRef.current) {
            audioElementRef.current = document.createElement("audio");
            audioElementRef.current.autoplay = true;
            audioElementRef.current.controls = true; // For debugging
            audioElementRef.current.volume = 1.0;
            
            // Add to DOM temporarily for debugging
            audioElementRef.current.style.position = 'fixed';
            audioElementRef.current.style.bottom = '10px';
            audioElementRef.current.style.right = '10px';
            audioElementRef.current.style.zIndex = '9999';
            audioElementRef.current.style.width = '300px';
            document.body.appendChild(audioElementRef.current);
          }
          
          // Set audio source
          if (event.streams && event.streams[0]) {
            audioElementRef.current.srcObject = event.streams[0];
          } else {
            // Create stream from track if not provided
            const stream = new MediaStream([event.track]);
            audioElementRef.current.srcObject = stream;
          }
          
          // Ensure playback starts
          audioElementRef.current.play().then(() => {
            console.log("✅ Audio playback started successfully");
            setStatusMessage("🔊 Audio playing - Speaking...");
          }).catch(err => {
            console.error("❌ Audio play failed:", err);
            setStatusMessage("⚠️ Audio play failed - Check permissions");
            
            // Try to resume audio context if suspended
            const audioContext = new (window.AudioContext || (window as any).webkitAudioContext)();
            if (audioContext.state === 'suspended') {
              audioContext.resume().then(() => {
                console.log("Audio context resumed, retrying play...");
                audioElementRef.current?.play();
              });
            }
          });
          
          // Monitor track events
          event.track.addEventListener('ended', () => {
            console.log("Audio track ended");
          });
          
          event.track.addEventListener('mute', () => {
            console.log("Audio track muted");
          });
          
          event.track.addEventListener('unmute', () => {
            console.log("Audio track unmuted");
          });
        }
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
        try {
          let text;
          if (typeof event.data === "string") {
            text = event.data;
          } else if (event.data instanceof ArrayBuffer) {
            text = new TextDecoder().decode(event.data);
          } else if (event.data instanceof Blob) {
            text = await event.data.text();
          } else {
            console.warn("Unknown data type:", typeof event.data);
            return;
          }

          try {
            const msg = JSON.parse(text);
            console.log("📨", msg.type);
            if (msg.type === "session.created") {
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
    
    // Properly cleanup audio element
    if (audioElementRef.current) {
      audioElementRef.current.pause();
      audioElementRef.current.srcObject = null;
      
      // Remove from DOM if added
      if (audioElementRef.current.parentNode) {
        audioElementRef.current.parentNode.removeChild(audioElementRef.current);
      }
      
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
              <div className={`w-3 h-3 rounded-full ${
                connectionState === 'connected' ? 'bg-green-500 animate-pulse' :
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
            {connectionState === 'connected' && (
              <p className="text-blue-600 font-medium">Audio controls appear in bottom-right corner</p>
            )}
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}
