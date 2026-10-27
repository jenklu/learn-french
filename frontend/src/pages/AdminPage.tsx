import React, { useState, useEffect } from 'react';
import { useAuth } from '../contexts/AuthContext';

interface User {
  id: number;
  username: string;
  email: string;
  timezone: string;
  is_admin: boolean;
  created_at: string;
}

const AdminPage: React.FC = () => {
  const { user } = useAuth();
  const [users, setUsers] = useState<User[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [editingUser, setEditingUser] = useState<User | null>(null);
  const [formData, setFormData] = useState({
    username: '',
    email: '',
    timezone: '',
    is_admin: false,
    password: '',
  });
  const [showCreateForm, setShowCreateForm] = useState(false);
  const [saving, setSaving] = useState(false);

  const API_BASE = import.meta.env.VITE_API_BASE_URL || '/api';

  useEffect(() => {
    if (user?.is_admin) {
      fetchUsers();
    }
  }, [user]);

  const fetchUsers = async () => {
    try {
      setLoading(true);
      const response = await fetch(`${API_BASE}/admin/users`, {
        credentials: 'include',
      });
      
      if (response.ok) {
        const data = await response.json();
        setUsers(data);
      } else {
        setError('Failed to fetch users');
      }
    } catch (err) {
      setError('Failed to fetch users');
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  const handleEdit = (user: User) => {
    setEditingUser(user);
    setFormData({
      username: user.username,
      email: user.email,
      timezone: user.timezone,
      is_admin: user.is_admin,
      password: '',
    });
  };

  const handleCancelEdit = () => {
    setEditingUser(null);
    setFormData({
      username: '',
      email: '',
      timezone: '',
      is_admin: false,
      password: '',
    });
  };

  const handleCreate = () => {
    setShowCreateForm(true);
    setFormData({
      username: '',
      email: '',
      timezone: Intl.DateTimeFormat().resolvedOptions().timeZone,
      is_admin: false,
      password: '',
    });
  };

  const handleCancelCreate = () => {
    setShowCreateForm(false);
    setFormData({
      username: '',
      email: '',
      timezone: '',
      is_admin: false,
      password: '',
    });
  };

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) => {
    const { name, value, type } = e.target;
    const checked = type === 'checkbox' ? (e.target as HTMLInputElement).checked : undefined;
    
    setFormData(prev => ({
      ...prev,
      [name]: type === 'checkbox' ? checked : value,
    }));
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!formData.username || !formData.email) {
      setError('Username and email are required');
      return;
    }

    try {
      setSaving(true);
      setError(null);

      if (editingUser) {
        // Update user
        const response = await fetch(`${API_BASE}/admin/users?id=${editingUser.id}`, {
          method: 'PUT',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({
            username: formData.username,
            email: formData.email,
            timezone: formData.timezone,
            is_admin: formData.is_admin,
          }),
          credentials: 'include',
        });

        if (response.ok) {
          await fetchUsers();
          handleCancelEdit();
        } else {
          setError('Failed to update user');
        }
      } else if (showCreateForm) {
        // Create user
        const response = await fetch(`${API_BASE}/admin/users`, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({
            username: formData.username,
            email: formData.email,
            password: formData.password || 'password', // In real app, this would be set by admin
            timezone: formData.timezone,
            is_admin: formData.is_admin,
          }),
          credentials: 'include',
        });

        if (response.ok) {
          await fetchUsers();
          handleCancelCreate();
        } else {
          setError('Failed to create user');
        }
      }
    } catch (err) {
      setError('Operation failed');
      console.error(err);
    } finally {
      setSaving(false);
    }
  };

  const handleDelete = async (userId: number) => {
    if (!window.confirm('Are you sure you want to delete this user?')) return;

    try {
      setSaving(true);
      const response = await fetch(`${API_BASE}/admin/users?id=${userId}`, {
        method: 'DELETE',
        credentials: 'include',
      });

      if (response.ok) {
        await fetchUsers();
      } else {
        setError('Failed to delete user');
      }
    } catch (err) {
      setError('Failed to delete user');
      console.error(err);
    } finally {
      setSaving(false);
    }
  };

  const handleUpdatePassword = async (userId: number) => {
    const newPassword = prompt('Enter new password:');
    if (!newPassword) return;

    try {
      setSaving(true);
      const response = await fetch(`${API_BASE}/admin/users/password?id=${userId}`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ password: newPassword }),
        credentials: 'include',
      });

      if (response.ok) {
        alert('Password updated successfully');
      } else {
        setError('Failed to update password');
      }
    } catch (err) {
      setError('Failed to update password');
      console.error(err);
    } finally {
      setSaving(false);
    }
  };

  if (!user?.is_admin) {
    return (
      <div className="py-5 text-center">
        <h3>Access Denied</h3>
        <p className="text-muted">
          You do not have permission to access this page.
        </p>
      </div>
    );
  }

  if (loading && users.length === 0) {
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
        <h2>User Management</h2>
        <button
          className="btn btn-primary"
          onClick={handleCreate}
          disabled={saving}
        >
          <i className="bi bi-plus me-2" />
          Create User
        </button>
      </div>

      {error && (
        <div className="alert alert-danger mb-4">
          {error}
          <button 
            className="btn btn-sm btn-outline-danger ms-3"
            onClick={() => setError(null)}
          >
            Dismiss
          </button>
        </div>
      )}

      {/* Create/Edit Form */}
      {(showCreateForm || editingUser) && (
        <div className="card shadow mb-4">
          <div className="card-header">
            <h5 className="mb-0">
              {editingUser ? `Edit User #${editingUser.id}` : 'Create New User'}
            </h5>
          </div>
          <div className="card-body">
            <form onSubmit={handleSubmit}>
              <div className="row">
                <div className="col-md-6 mb-3">
                  <label htmlFor="username" className="form-label">
                    Username *
                  </label>
                  <input
                    type="text"
                    id="username"
                    name="username"
                    className="form-control"
                    value={formData.username}
                    onChange={handleInputChange}
                    required
                    disabled={saving}
                  />
                </div>
                <div className="col-md-6 mb-3">
                  <label htmlFor="email" className="form-label">
                    Email *
                  </label>
                  <input
                    type="email"
                    id="email"
                    name="email"
                    className="form-control"
                    value={formData.email}
                    onChange={handleInputChange}
                    required
                    disabled={saving}
                  />
                </div>
              </div>
              
              <div className="row">
                <div className="col-md-6 mb-3">
                  <label htmlFor="timezone" className="form-label">
                    Timezone
                  </label>
                  <select
                    id="timezone"
                    name="timezone"
                    className="form-select"
                    value={formData.timezone}
                    onChange={handleInputChange}
                    disabled={saving}
                  >
                    {Intl.supportedValuesOf('timeZone').map(tz => (
                      <option key={tz} value={tz}>{tz}</option>
                    ))}
                  </select>
                </div>
                <div className="col-md-6 mb-3">
                  <label htmlFor="is_admin" className="form-label">
                    Admin Privileges
                  </label>
                  <div className="form-check form-switch">
                    <input
                      type="checkbox"
                      id="is_admin"
                      name="is_admin"
                      className="form-check-input"
                      checked={formData.is_admin}
                      onChange={handleInputChange}
                      disabled={saving}
                    />
                    <label className="form-check-label" htmlFor="is_admin">
                      Is Admin
                    </label>
                  </div>
                </div>
              </div>

              {showCreateForm && (
                <div className="mb-3">
                  <label htmlFor="password" className="form-label">
                    Password *
                  </label>
                  <input
                    type="password"
                    id="password"
                    name="password"
                    className="form-control"
                    value={formData.password}
                    onChange={handleInputChange}
                    required
                    disabled={saving}
                  />
                  <div className="form-text">
                    Set a password for the new user
                  </div>
                </div>
              )}

              <div className="d-flex gap-2">
                <button
                  type="submit"
                  className="btn btn-primary"
                  disabled={saving}
                >
                  {saving ? (
                    <>
                      <span className="spinner-border spinner-border-sm me-2" />
                      Saving...
                    </>
                  ) : (
                    editingUser ? 'Update User' : 'Create User'
                  )}
                </button>
                <button
                  type="button"
                  className="btn btn-outline-secondary"
                  onClick={editingUser ? handleCancelEdit : handleCancelCreate}
                  disabled={saving}
                >
                  Cancel
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Users Table */}
      <div className="card shadow">
        <div className="card-header">
          <h5 className="mb-0">Users ({users.length})</h5>
        </div>
        <div className="card-body">
          <div className="table-responsive">
            <table className="table table-hover">
              <thead>
                <tr>
                  <th>ID</th>
                  <th>Username</th>
                  <th>Email</th>
                  <th>Timezone</th>
                  <th>Role</th>
                  <th>Created</th>
                  <th>Actions</th>
                </tr>
              </thead>
              <tbody>
                {users.map((u) => (
                  <tr key={u.id}>
                    <td>{u.id}</td>
                    <td>{u.username}</td>
                    <td>{u.email}</td>
                    <td>{u.timezone}</td>
                    <td>
                      <span className={`badge ${u.is_admin ? 'bg-primary' : 'bg-secondary'}`}>
                        {u.is_admin ? 'Admin' : 'User'}
                      </span>
                    </td>
                    <td>{new Date(u.created_at).toLocaleDateString()}</td>
                    <td>
                      <div className="d-flex gap-1">
                        <button
                          className="btn btn-sm btn-outline-primary"
                          onClick={() => handleEdit(u)}
                          disabled={saving}
                        >
                          <i className="bi bi-pencil" />
                        </button>
                        <button
                          className="btn btn-sm btn-outline-success"
                          onClick={() => handleUpdatePassword(u.id)}
                          disabled={saving}
                        >
                          <i className="bi bi-key" />
                        </button>
                        <button
                          className="btn btn-sm btn-outline-danger"
                          onClick={() => handleDelete(u.id)}
                          disabled={saving || u.id === user?.id}
                        >
                          <i className="bi bi-trash" />
                        </button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </div>
  );
};

export default AdminPage;
