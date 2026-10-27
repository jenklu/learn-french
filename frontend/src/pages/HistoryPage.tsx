import React, { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';

interface Session {
  id: number;
  user_id: number;
  date: string;
  duration_minutes: number;
  listening_score: number;
  completed_at: string;
}

interface ExerciseResult {
  id: number;
  session_id: number;
  exercise_index: number;
  is_correct: boolean;
  score: number;
  created_at: string;
}

const HistoryPage: React.FC = () => {
  const { user } = useAuth();
  const [sessions, setSessions] = useState<Session[]>([]);
  const [selectedSession, setSelectedSession] = useState<Session | null>(null);
  const [results, setResults] = useState<ExerciseResult[]>([]);
  const [loading, setLoading] = useState(true);
  const [stats, setStats] = useState<{
    totalSessions: number;
    avgScore: number;
    bestScore: number;
    totalDuration: number;
  } | null>(null);

  const API_BASE = import.meta.env.VITE_API_BASE_URL || '/api';

  useEffect(() => {
    if (user) {
      fetchSessions();
      fetchStats();
    }
  }, [user]);

  const fetchSessions = async () => {
    try {
      setLoading(true);
      const response = await fetch(`${API_BASE}/sessions`, {
        credentials: 'include',
      });
      
      if (response.ok) {
        const data = await response.json();
        setSessions(data);
      }
    } catch (err) {
      console.error('Failed to fetch sessions:', err);
    } finally {
      setLoading(false);
    }
  };

  const fetchStats = async () => {
    try {
      const response = await fetch(`${API_BASE}/stats/daily`, {
        credentials: 'include',
      });
      
      if (response.ok) {
        const data = await response.json();
        // Calculate stats from sessions
        const totalSessions = sessions.length;
        const scores = sessions.map(s => s.listening_score).filter(s => s > 0);
        const avgScore = scores.length > 0 ? scores.reduce((a, b) => a + b, 0) / scores.length : 0;
        const bestScore = scores.length > 0 ? Math.max(...scores) : 0;
        const totalDuration = sessions.reduce((a, b) => a + b.duration_minutes, 0);
        
        setStats({ totalSessions, avgScore, bestScore, totalDuration });
      }
    } catch (err) {
      console.error('Failed to fetch stats:', err);
    }
  };

  const fetchSessionResults = async (sessionId: number) => {
    try {
      const response = await fetch(`${API_BASE}/sessions/results?session_id=${sessionId}`, {
        credentials: 'include',
      });
      
      if (response.ok) {
        const data = await response.json();
        setResults(data);
      }
    } catch (err) {
      console.error('Failed to fetch session results:', err);
    }
  };

  const handleSessionClick = (session: Session) => {
    setSelectedSession(session);
    fetchSessionResults(session.id);
  };

  if (loading && sessions.length === 0) {
    return (
      <div className="min-vh-100 d-flex align-items-center justify-content-center">
        <div className="spinner-border text-primary" role="status">
          <span className="visually-hidden">Loading...</span>
        </div>
      </div>
    );
  }

  return (
    <div className="py-4">
      <div className="d-flex justify-content-between align-items-center mb-4">
        <h2>Practice History</h2>
        <Link to="/practice" className="btn btn-primary">
          <i className="bi bi-plus me-2" />
          New Session
        </Link>
      </div>

      {sessions.length === 0 ? (
        <div className="text-center py-5">
          <h3>No practice sessions yet</h3>
          <p className="text-muted">
            Start your first practice session to see your progress here.
          </p>
          <Link to="/practice" className="btn btn-primary mt-3">
            Start Practicing
          </Link>
        </div>
      ) : (
        <>
          {/* Stats Overview */}
          {stats && (
            <div className="row mb-4">
              <div className="col-md-3 col-6 mb-3">
                <div className="card h-100">
                  <div className="card-body text-center">
                    <h6 className="text-muted mb-2">Total Sessions</h6>
                    <div className="display-6 fw-bold">{stats.totalSessions}</div>
                  </div>
                </div>
              </div>
              <div className="col-md-3 col-6 mb-3">
                <div className="card h-100">
                  <div className="card-body text-center">
                    <h6 className="text-muted mb-2">Average Score</h6>
                    <div className="display-6 fw-bold text-primary">{Math.round(stats.avgScore)}%</div>
                  </div>
                </div>
              </div>
              <div className="col-md-3 col-6 mb-3">
                <div className="card h-100">
                  <div className="card-body text-center">
                    <h6 className="text-muted mb-2">Best Score</h6>
                    <div className="display-6 fw-bold text-success">{stats.bestScore}%</div>
                  </div>
                </div>
              </div>
              <div className="col-md-3 col-6 mb-3">
                <div className="card h-100">
                  <div className="card-body text-center">
                    <h6 className="text-muted mb-2">Total Time</h6>
                    <div className="display-6 fw-bold">{stats.totalDuration} min</div>
                  </div>
                </div>
              </div>
            </div>
          )}

          <div className="row">
            {/* Sessions List */}
            <div className="col-lg-6">
              <div className="card shadow mb-4">
                <div className="card-header">
                  <h5 className="mb-0">Recent Sessions</h5>
                </div>
                <div className="card-body">
                  <div className="list-group list-group-flush">
                    {sessions.map((session) => (
                      <button
                        key={session.id}
                        className={`list-group-item list-group-item-action ${selectedSession?.id === session.id ? 'active' : ''}`}
                        onClick={() => handleSessionClick(session)}
                      >
                        <div className="d-flex justify-content-between align-items-center">
                          <div>
                            <h6 className="mb-1">
                              Session #{session.id} - {new Date(session.date).toLocaleDateString()}
                            </h6>
                            <small className="text-muted">
                              {session.duration_minutes} min - {session.listening_score}% score
                            </small>
                          </div>
                          <span className={`badge ${session.listening_score >= 80 ? 'bg-success' : session.listening_score >= 60 ? 'bg-warning' : 'bg-danger'}`}>
                            {session.listening_score}%
                          </span>
                        </div>
                      </button>
                    ))}
                  </div>
                </div>
              </div>
            </div>

            {/* Session Details */}
            <div className="col-lg-6">
              <div className="card shadow">
                <div className="card-header">
                  <h5 className="mb-0">Session Details</h5>
                </div>
                <div className="card-body">
                  {selectedSession ? (
                    <>
                      <div className="mb-4">
                        <h6>Session #{selectedSession.id}</h6>
                        <p className="text-muted">
                          {new Date(selectedSession.date).toLocaleDateString()} - {selectedSession.duration_minutes} minutes
                        </p>
                        <div className="d-flex align-items-center">
                          <div className="progress flex-grow-1 me-3">
                            <div 
                              className="progress-bar bg-primary"
                              style={{ width: `${selectedSession.listening_score}%` }}
                            >
                              {selectedSession.listening_score}%
                            </div>
                          </div>
                          <span className="badge bg-primary">
                            {selectedSession.listening_score >= 80 ? 'Excellent' : 
                             selectedSession.listening_score >= 60 ? 'Good' : 'Needs Improvement'}
                          </span>
                        </div>
                      </div>

                      <h6 className="mb-3">Exercise Results</h6>
                      <div className="table-responsive">
                        <table className="table table-hover">
                          <thead>
                            <tr>
                              <th>#</th>
                              <th>Status</th>
                              <th>Score</th>
                            </tr>
                          </thead>
                          <tbody>
                            {results.length > 0 ? (
                              results.map((result) => (
                                <tr key={result.id}>
                                  <td>{result.exercise_index + 1}</td>
                                  <td>
                                    <span className={`badge ${result.is_correct ? 'bg-success' : 'bg-danger'}`}>
                                      {result.is_correct ? 'Correct' : 'Incorrect'}
                                    </span>
                                  </td>
                                  <td>{result.score}</td>
                                </tr>
                              ))
                            ) : (
                              <tr>
                                <td colSpan={3} className="text-center text-muted">
                                  No results available
                                </td>
                              </tr>
                            )}
                          </tbody>
                        </table>
                      </div>

                      {results.length > 0 && (
                        <div className="mt-3 pt-3 border-top">
                          <p className="mb-2">
                            <strong>Summary:</strong> {results.filter(r => r.is_correct).length} correct out of {results.length} questions
                          </p>
                          <p className="mb-0 text-muted">
                            Total score: {results.reduce((sum, r) => sum + r.score, 0)} points
                          </p>
                        </div>
                      )}
                    </>
                  ) : (
                    <div className="text-center py-4 text-muted">
                      <p>Select a session to view details</p>
                    </div>
                  )}
                </div>
              </div>
            </div>
          </div>
        </>
      )}
    </div>
  );
};

export default HistoryPage;
