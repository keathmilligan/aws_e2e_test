import React, { useState } from 'react';
import { useAuth } from './AuthContext';
import { Link, useNavigate } from 'react-router-dom';

const ForgotPassword: React.FC = () => {
  const [email, setEmail] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [isCodeSent, setIsCodeSent] = useState(false);
  const { forgotPassword } = useAuth();
  const navigate = useNavigate();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setErrorMessage(null);
    setIsSubmitting(true);

    try {
      await forgotPassword(email);
      setIsCodeSent(true);
    } catch (error: any) {
      setErrorMessage(error.message || 'Failed to initiate password reset. Please try again.');
    } finally {
      setIsSubmitting(false);
    }
  };

  if (isCodeSent) {
    return (
      <div className="auth-container">
        <div className="auth-card">
          <h2>Reset Code Sent</h2>
          <p>
            A password reset code has been sent to your email address. Please check your email and
            proceed to reset your password.
          </p>
          <button
            onClick={() => navigate('/reset-password', { state: { email } })}
            className="auth-button"
          >
            Proceed to Reset Password
          </button>
          <div className="auth-links">
            <Link to="/login">Back to Login</Link>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="auth-container">
      <div className="auth-card">
        <h2>Forgot Password</h2>
        <p>Enter your email address to receive a password reset code.</p>
        {errorMessage && <div className="error-message">{errorMessage}</div>}
        <form onSubmit={handleSubmit}>
          <div className="form-group">
            <label htmlFor="email">Email</label>
            <input
              type="email"
              id="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              required
              disabled={isSubmitting}
            />
          </div>
          <button type="submit" disabled={isSubmitting} className="auth-button">
            {isSubmitting ? 'Sending...' : 'Send Reset Code'}
          </button>
        </form>
        <div className="auth-links">
          <Link to="/login">Back to Login</Link>
        </div>
      </div>
    </div>
  );
};

export default ForgotPassword;