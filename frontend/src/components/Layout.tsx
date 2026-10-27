import React from 'react';
import { Link, useLocation } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';

const Layout: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const { user, logout } = useAuth();
  const location = useLocation();

  const isActive = (path: string) => {
    return location.pathname === path;
  };

  return (
    <div className="min-h-screen d-flex flex-column">
      <nav className="navbar navbar-expand-lg navbar-dark bg-primary">
        <div className="container">
          <Link to="/" className="navbar-brand">
            <i className="bi bi-globe" style={{ marginRight: '8px' }} />
            Learn French
          </Link>
          
          <button
            className="navbar-toggler"
            type="button"
            data-bs-toggle="collapse"
            data-bs-target="#navbarNav"
          >
            <span className="navbar-toggler-icon"></span>
          </button>

          <div className="collapse navbar-collapse" id="navbarNav">
            <ul className="navbar-nav me-auto">
              <li className="nav-item">
                <Link to="/" className={`nav-link ${isActive('/') ? 'active' : ''}`}>
                  Home
                </Link>
              </li>
              
              {user && (
                <>
                  <li className="nav-item">
                    <Link to="/practice" className={`nav-link ${isActive('/practice') ? 'active' : ''}`}>
                      Practice
                    </Link>
                  </li>
                  <li className="nav-item">
                    <Link to="/history" className={`nav-link ${isActive('/history') ? 'active' : ''}`}>
                      History
                    </Link>
                  </li>
                  <li className="nav-item">
                    <Link to="/settings" className={`nav-link ${isActive('/settings') ? 'active' : ''}`}>
                      Settings
                    </Link>
                  </li>
                  {user.is_admin && (
                    <li className="nav-item">
                      <Link to="/admin" className={`nav-link ${isActive('/admin') ? 'active' : ''}`}>
                        Admin
                      </Link>
                    </li>
                  )}
                </>
              )}
            </ul>

            <ul className="navbar-nav">
              {user ? (
                <>
                  <li className="nav-item">
                    <span className="navbar-text me-3">
                      Welcome, {user.username}!
                    </span>
                  </li>
                  <li className="nav-item">
                    <button
                      className="btn btn-outline-light"
                      onClick={logout}
                    >
                      Logout
                    </button>
                  </li>
                </>
              ) : (
                <>
                  <li className="nav-item">
                    <Link to="/login" className="nav-link">Login</Link>
                  </li>
                  <li className="nav-item">
                    <Link to="/register" className="btn btn-light">Register</Link>
                  </li>
                </>
              )}
            </ul>
          </div>
        </div>
      </nav>

      <main className="flex-grow-1 py-4">
        <div className="container">
          {children}
        </div>
      </main>

      <footer className="bg-light py-3 mt-5">
        <div className="container text-center">
          <p className="mb-0 text-muted">
            &copy; {new Date().getFullYear()} Learn French App. All rights reserved.
          </p>
        </div>
      </footer>
    </div>
  );
};

export default Layout;
