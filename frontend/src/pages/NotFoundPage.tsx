import React from 'react';
import { Link } from 'react-router-dom';

const NotFoundPage: React.FC = () => {
  return (
    <div className="min-vh-100 d-flex align-items-center justify-content-center py-5">
      <div className="text-center">
        <div className="display-1 fw-bold text-primary mb-4">404</div>
        <h1 className="h2 mb-3">Page Not Found</h1>
        <p className="text-muted mb-4">
          The page you're looking for doesn't exist or has been moved.
        </p>
        <div className="d-flex gap-3 justify-content-center">
          <Link to="/" className="btn btn-primary">
            Go Home
          </Link>
          <Link to="/practice" className="btn btn-outline-primary">
            Start Practicing
          </Link>
        </div>
      </div>
    </div>
  );
};

export default NotFoundPage;
