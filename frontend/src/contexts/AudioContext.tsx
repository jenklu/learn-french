import React, { createContext, useContext, useState, useRef, useCallback, ReactNode } from 'react';

interface AudioContextType {
  audioUrl: string | null;
  isPlaying: boolean;
  playAudio: (url: string) => Promise<void>;
  pauseAudio: () => void;
  stopAudio: () => void;
  cacheAudio: (url: string, blob: Blob) => Promise<void>;
  getCachedAudio: (url: string) => Promise<string | null>;
}

const AudioContext = createContext<AudioContextType>({
  audioUrl: null,
  isPlaying: false,
  playAudio: async () => {},
  pauseAudio: () => {},
  stopAudio: () => {},
  cacheAudio: async () => {},
  getCachedAudio: async () => null,
});

// Simple in-memory cache (in production, use IndexedDB)
const audioCache: Map<string, string> = new Map();

export const AudioProvider: React.FC<{ children: ReactNode }> = ({ children }) => {
  const [audioUrl, setAudioUrl] = useState<string | null>(null);
  const [isPlaying, setIsPlaying] = useState<boolean>(false);
  const audioRef = useRef<HTMLAudioElement | null>(null);

  const playAudio = useCallback(async (url: string) => {
    // Check cache first
    const cachedUrl = await getCachedAudio(url);
    if (cachedUrl) {
      url = cachedUrl;
    }

    setAudioUrl(url);
    setIsPlaying(true);

    // In a real implementation, we'd use the HTML5 audio element
    // For now, we'll just set the URL
    if (audioRef.current) {
      audioRef.current.src = url;
      audioRef.current.play().catch(err => {
        console.error('Failed to play audio:', err);
        setIsPlaying(false);
      });
    }
  }, []);

  const pauseAudio = useCallback(() => {
    if (audioRef.current) {
      audioRef.current.pause();
    }
    setIsPlaying(false);
  }, []);

  const stopAudio = useCallback(() => {
    if (audioRef.current) {
      audioRef.current.pause();
      audioRef.current.currentTime = 0;
    }
    setAudioUrl(null);
    setIsPlaying(false);
  }, []);

  const cacheAudio = useCallback(async (url: string, blob: Blob) => {
    try {
      const objectUrl = URL.createObjectURL(blob);
      audioCache.set(url, objectUrl);
    } catch (err) {
      console.error('Failed to cache audio:', err);
    }
  }, []);

  const getCachedAudio = useCallback(async (url: string) => {
    return audioCache.get(url) || null;
  }, []);

  const value: AudioContextType = {
    audioUrl,
    isPlaying,
    playAudio,
    pauseAudio,
    stopAudio,
    cacheAudio,
    getCachedAudio,
  };

  return (
    <AudioContext.Provider value={value}>
      <audio ref={audioRef} onEnded={() => setIsPlaying(false)} />
      {children}
    </AudioContext.Provider>
  );
};

export const useAudio = (): AudioContextType => {
  return useContext(AudioContext);
};
