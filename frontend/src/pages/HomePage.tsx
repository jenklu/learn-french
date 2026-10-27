import React from 'react';
import { Link } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';

const HomePage: React.FC = () => {
  const { user } = useAuth();

  return (
    <div className="py-5">
      <div className="row align-items-center">
        <div className="col-lg-6">
          <h1 className="display-4 fw-bold mb-4">
            Learn French with Listening Practice
          </h1>
          <p className="lead mb-4">
            Improve your French listening skills with our interactive exercises.
            Practice at your own pace and track your progress.
          </p>
          
          {user ? (
            <div className="d-flex gap-3">
              <Link to="/practice" className="btn btn-primary btn-lg">
                Start Practicing
              </Link>
              <Link to="/history" className="btn btn-outline-primary btn-lg">
                View History
              </Link>
            </div>
          ) : (
            <div className="d-flex gap-3">
              <Link to="/register" className="btn btn-primary btn-lg">
                Get Started
              </Link>
              <Link to="/login" className="btn btn-outline-primary btn-lg">
                Login
              </Link>
            </div>
          )}
        </div>
        
        <div className="col-lg-6">
          <div className="card shadow">
            <div className="card-body">
              <h3 className="card-title mb-4">Features</h3>
              <ul className="list-group list-group-flush">
                <li className="list-group-item">
                  <i className="bi bi-headphones me-2" />
                  A2 and B1 level listening exercises
                </li>
                <li className="list-group-item">
                  <i className="bi bi-graph-up me-2" />
                  Progress tracking and analytics
                </li>
                <li className="list-group-item">
                  <i className="bi bi-bell me-2" />
                  Daily reminders to practice
                </li>
                <li className="list-group-item">
                  <i className="bi bi-trophy me-2" />
                  Gamification with scores and levels
                </li>
                <li className="list-group-item">
                  <i className="bi bi-globe me-2" />
                  Timezone-aware notifications
                </li>
              </ul>
            </div>
          </div>
        </div>
      </div>

      <div className="mt-5">
        <h2 className="mb-4">How It Works</h2>
        <div className="row">
          <div className="col-md-4 mb-4">
            <div className="card h-100">
              <div className="card-body text-center">
                <div className="mb-3">
                  <span className="badge bg-primary bg-opacity-10 text-primary p-3">
                    <i className="bi bi-1-circle" style={{ fontSize: '2rem' }} />
                  </span>
                </div>
                <h4 className="card-title">Listen</h4>
                <p className="card-text">
                  Listen to authentic French audio clips at your level.
                </p>
              </div>
            </div>
          </div>
          
          <div className="col-md-4 mb-4">
            <div className="card h-100">
              <div className="card-body text-center">
                <div className="mb-3">
                  <span className="badge bg-primary bg-opacity-10 text-primary p-3">
                    <i className="bi bi-2-circle" style={{ fontSize: '2rem' }} />
                  </span>
                </div>
                <h4 className="card-title">Answer</h4>
                <p className="card-text">
                  Answer multiple-choice questions to test your understanding.
                </p>
              </div>
            </div>
          </div>
          
          <div className="col-md-4 mb-4">
            <div className="card h-100">
              <div className="card-body text-center">
                <div className="mb-3">
                  <span className="badge bg-primary bg-opacity-10 text-primary p-3">
                    <i className="bi bi-3-circle" style={{ fontSize: '2rem' }} />
                  </span>
                </div>
                <h4 className="card-title">Improve</h4>
                <p className="card-text">
                  Track your progress and see your improvement over time.
                </p>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default HomePage;
