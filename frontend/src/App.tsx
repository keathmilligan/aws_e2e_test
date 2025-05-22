import React from 'react';
import { Routes, Route, Navigate } from 'react-router-dom';
import './App.css';
import './auth/Auth.css';

// Auth components
import { useAuth } from './auth/AuthContext';
import Login from './auth/Login';
import Signup from './auth/Signup';
import ConfirmSignup from './auth/ConfirmSignup';
import ForgotPassword from './auth/ForgotPassword';
import ResetPassword from './auth/ResetPassword';
import ProtectedRoute from './auth/ProtectedRoute';

// App components
import Home from './components/Home';

const App: React.FC = () => {
  const { isLoading } = useAuth();

  if (isLoading) {
    return (
      <div className="loading-container">
        <div className="loading-spinner"></div>
        <p>Loading...</p>
      </div>
    );
  }

  return (
    <Routes>
      {/* Public routes */}
      <Route path="/login" element={<Login />} />
      <Route path="/signup" element={<Signup />} />
      <Route path="/confirm-signup" element={<ConfirmSignup />} />
      <Route path="/forgot-password" element={<ForgotPassword />} />
      <Route path="/reset-password" element={<ResetPassword />} />
      
      {/* Protected routes */}
      <Route element={<ProtectedRoute />}>
        <Route path="/" element={<Home />} />
      </Route>
      
      {/* Redirect any unknown routes to home */}
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  );
};

export default App;