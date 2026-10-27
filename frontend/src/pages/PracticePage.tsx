import React, { useState, useEffect, useCallback } from 'react';
import { useAuth } from '../contexts/AuthContext';
import { useAudio } from '../contexts/AudioContext';

interface Exercise {
  id: number;
  level: string;
  question: string;
  options: string[];
  correct_answer: number;
  audio_url: string;
  text: string;
}

interface ExerciseResult {
  session_id: number;
  exercise_id: number;
  exercise_index: number;
  selected_option: number;
  is_correct: boolean;
  score: number;
  duration: number;
}

const PracticePage: React.FC = () => {
  const { user } = useAuth();
  const { playAudio, pauseAudio, stopAudio } = useAudio();
  const [exercises, setExercises] = useState<Exercise[]>([]);
  const [currentExerciseIndex, setCurrentExerciseIndex] = useState(0);
  const [selectedOption, setSelectedOption] = useState<number | null>(null);
  const [isCorrect, setIsCorrect] = useState<boolean | null>(null);
  const [showFeedback, setShowFeedback] = useState(false);
  const [isPlaying, setIsPlaying] = useState(false);
  const [score, setScore] = useState(0);
  const [sessionId, setSessionId] = useState<number | null>(null);
  const [startTime, setStartTime] = useState<number | null>(null);
  const [completed, setCompleted] = useState(false);
  const [level, setLevel] = useState<'A2' | 'B1'>('A2');
  const [loading, setLoading] = useState(true);

  const API_BASE = import.meta.env.VITE_API_BASE_URL || '/api';

  const fetchExercises = useCallback(async () => {
    try {
      setLoading(true);
      const response = await fetch(`${API_BASE}/exercises?level=${level}&count=5`, {
        credentials: 'include',
      });
      
      if (response.ok) {
        const data = await response.json();
        setExercises(data);
        setCurrentExerciseIndex(0);
        setSelectedOption(null);
        setIsCorrect(null);
        setShowFeedback(false);
        setScore(0);
        setStartTime(Date.now());
      }
    } catch (err) {
      console.error('Failed to fetch exercises:', err);
    } finally {
      setLoading(false);
    }
  }, [level]);

  useEffect(() => {
    if (user) {
      fetchExercises();
    }
  }, [user, fetchExercises]);

  const handleOptionSelect = (optionIndex: number) => {
    setSelectedOption(optionIndex);
  };

  const handleSubmit = async () => {
    if (selectedOption === null) return;

    const currentExercise = exercises[currentExerciseIndex];
    const isCorrectAnswer = selectedOption === currentExercise.correct_answer;
    const scoreIncrement = isCorrectAnswer ? 20 : 0;

    setIsCorrect(isCorrectAnswer);
    setShowFeedback(true);
    setScore(prev => prev + scoreIncrement);

    // In a real implementation, we'd submit to the backend
    // For now, we'll just track locally
  };

  const handleNext = () => {
    if (currentExerciseIndex < exercises.length - 1) {
      setCurrentExerciseIndex(prev => prev + 1);
      setSelectedOption(null);
      setIsCorrect(null);
      setShowFeedback(false);
    } else {
      // Complete the session
      completeSession();
    }
  };

  const completeSession = async () => {
    try {
      const durationMinutes = startTime ? Math.ceil((Date.now() - startTime) / 60000) : 1;
      
      const response = await fetch(`${API_BASE}/sessions/complete`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          duration_minutes: durationMinutes,
          listening_score: score,
        }),
        credentials: 'include',
      });

      if (response.ok) {
        const data = await response.json();
        setSessionId(data.id);
        setCompleted(true);
      }
    } catch (err) {
      console.error('Failed to complete session:', err);
    }
  };

  const handlePlayAudio = async (exercise: Exercise) => {
    try {
      setIsPlaying(true);
      
      // In a real implementation, we'd use the TTS API or cached audio
      // For now, we'll simulate it
      if (exercise.audio_url) {
        await playAudio(exercise.audio_url);
      } else {
        // Use the backend to generate TTS
        try {
          const ttsResponse = await fetch(`${API_BASE}/tts`, {
            method: 'POST',
            headers: {
              'Content-Type': 'application/json',
            },
            body: JSON.stringify({ text: exercise.text }),
            credentials: 'include',
          });
          
          if (ttsResponse.ok) {
            const audioData = await ttsResponse.blob();
            const audioUrl = URL.createObjectURL(audioData);
            await playAudio(audioUrl);
          }
        } catch (ttsErr) {
          console.error('TTS failed:', ttsErr);
          // Fallback: just show the text
          alert(`Audio not available. Text: ${exercise.text}`);
        }
      }
      
      // Simulate audio playback
      setTimeout(() => {
        setIsPlaying(false);
      }, 3000);
    } catch (err) {
      console.error('Failed to play audio:', err);
      setIsPlaying(false);
    }
  };

  const startNewSession = () => {
    setCompleted(false);
    fetchExercises();
  };

  if (loading) {
    return (
      <div className="min-vh-100 d-flex align-items-center justify-content-center">
        <div className="spinner-border text-primary" role="status">
          <span className="visually-hidden">Loading...</span>
        </div>
      </div>
    );
  }

  if (completed) {
    return (
      <div className="py-5">
        <div className="text-center mb-5">
          <h2>Session Complete!</h2>
          <p className="lead">Great job on your French listening practice.</p>
          <div className="mt-4">
            <div className="card inline-block mx-auto" style={{ maxWidth: '400px' }}>
              <div className="card-body">
                <h4 className="card-title">Your Score</h4>
                <div className="display-4 fw-bold text-primary">{score}</div>
                <div className="progress mt-3">
                  <div 
                    className="progress-bar bg-primary"
                    style={{ width: `${score}%` }}
                  >
                    {score}%
                  </div>
                </div>
                <p className="mt-3">
                  {score >= 80 ? 'Excellent work! You\'re ready for B1 level.' : 
                   score >= 60 ? 'Good job! Keep practicing to reach B1.' : 
                   'Keep practicing! You\'ll improve with more sessions.'}
                </p>
              </div>
            </div>
          </div>
          <button 
            className="btn btn-primary btn-lg mt-4"
            onClick={startNewSession}
          >
            Start New Session
          </button>
        </div>
      </div>
    );
  }

  if (exercises.length === 0) {
    return (
      <div className="py-5 text-center">
        <h3>No exercises available</h3>
        <p>Please try again later.</p>
      </div>
    );
  }

  const currentExercise = exercises[currentExerciseIndex];

  return (
    <div className="py-4">
      <div className="d-flex justify-content-between align-items-center mb-4">
        <div>
          <h2>Listening Practice</h2>
          <div className="mt-2">
            <select
              className="form-select form-select-sm d-inline-block w-auto"
              value={level}
              onChange={(e) => setLevel(e.target.value as 'A2' | 'B1')}
              disabled={currentExerciseIndex > 0}
            >
              <option value="A2">A2 Level</option>
              <option value="B1">B1 Level</option>
            </select>
          </div>
        </div>
        <div className="text-end">
          <div className="h4 mb-0">Score: {score}</div>
          <div className="text-muted">
            Question {currentExerciseIndex + 1} of {exercises.length}
          </div>
        </div>
      </div>

      <div className="card shadow mb-4">
        <div className="card-body">
          <div className="d-flex justify-content-between align-items-center mb-4">
            <h4 className="mb-0">Exercise {currentExerciseIndex + 1}</h4>
            <button
              className="btn btn-primary"
              onClick={() => handlePlayAudio(currentExercise)}
              disabled={isPlaying}
            >
              {isPlaying ? (
                <>
                  <span className="spinner-border spinner-border-sm me-2" role="status" />
                  Playing...
                </>
              ) : (
                <>
                  <i className="bi bi-play-circle me-2" />
                  Play Audio
                </>
              )}
            </button>
          </div>

          <div className="mb-4 p-3 bg-light rounded">
            <h5 className="mb-3">Question:</h5>
            <p className="lead">{currentExercise.question}</p>
          </div>

          <div className="mb-4">
            <h5 className="mb-3">Options:</h5>
            <div className="d-grid gap-2">
              {currentExercise.options.map((option, index) => (
                <button
                  key={index}
                  className={`btn btn-lg btn-outline-primary ${selectedOption === index ? 'active' : ''}`}
                  onClick={() => handleOptionSelect(index)}
                  disabled={showFeedback}
                >
                  {String.fromCharCode(65 + index)}. {option}
                </button>
              ))}
            </div>
          </div>

          {showFeedback && (
            <div className={`alert ${isCorrect ? 'alert-success' : 'alert-danger'} mb-4`}>
              <h5 className="alert-heading">
                {isCorrect ? 'Correct!' : 'Incorrect'}
              </h5>
              <p>
                {isCorrect 
                  ? 'Great job! That was the right answer.'
                  : `The correct answer was: ${String.fromCharCode(65 + currentExercise.correct_answer)}. ${currentExercise.options[currentExercise.correct_answer]}`}
              </p>
              <hr />
              <p className="mb-0">
                <strong>Transcript:</strong> {currentExercise.text}
              </p>
            </div>
          )}

          {!showFeedback ? (
            <button
              className="btn btn-primary btn-lg w-100"
              onClick={handleSubmit}
              disabled={selectedOption === null}
            >
              Submit Answer
            </button>
          ) : (
            <button
              className="btn btn-primary btn-lg w-100"
              onClick={handleNext}
            >
              {currentExerciseIndex < exercises.length - 1 ? 'Next Question' : 'Complete Session'}
            </button>
          )}
        </div>
      </div>

      <div className="progress mb-4">
        <div 
          className="progress-bar bg-primary"
          style={{ width: `${((currentExerciseIndex + (showFeedback ? 1 : 0)) / exercises.length) * 100}%` }}
        >
          {Math.round(((currentExerciseIndex + (showFeedback ? 1 : 0)) / exercises.length) * 100)}%
        </div>
      </div>
    </div>
  );
};

export default PracticePage;
