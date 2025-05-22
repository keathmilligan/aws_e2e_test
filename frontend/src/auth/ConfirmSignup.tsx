import React, { useState } from 'react';
import { useAuth } from './AuthContext';
import { Link, useLocation, useNavigate } from 'react-router-dom';

interface LocationState {
  email?: string;
}

const ConfirmSignup: React.FC = () => {
  const location = useLocation();
  const navigate = useNavigate();
  const { confirmSignup, resendConfirmationCode } = useAuth();
  
  const { email: locationEmail } = (location.state as LocationState) || {};
  
  const [email, setEmail] = useState(locationEmail || '');
  const [code, setCode] = useState('');
  const [isConfirming, setIsConfirming] = useState(false);
  const [isResending, setIsResending] = useState(false);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [successMessage, setSuccessMessage] = useState<string | null>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setErrorMessage(null);
    setSuccessMessage(null);
    setIsConfirming(true);

    try {
      await confirmSignup(email, code);
      setSuccessMessage('Your account has been verified successfully!');
      // Redirect to login page after a short delay
      setTimeout(() => {
        navigate('/login');
      }, 2000);
    } catch (error: any) {
      setErrorMessage(error.message || 'Failed to confirm signup. Please try again.');
    } finally {
      setIsConfirming(false);
    }
  };

  const handleResendCode = async () => {
    setErrorMessage(null);
    setSuccessMessage(null);
    setIsResending(true);

    try {
      await resendConfirmationCode(email);
      setSuccessMessage('A new verification code has been sent to your email.');
    } catch (error: any) {
      setErrorMessage(error.message || 'Failed to resend code. Please try again.');
    } finally {
      setIsResending(false);
    }
  };

  return (
    <div className="auth-container">
      <div className="auth-card">
        <h2>Verify Your Account</h2>
        {errorMessage && <div className="error-message">{errorMessage}</div>}
        {successMessage && <div className="success-message">{successMessage}</div>}
        <form onSubmit={handleSubmit}>
          <div className="form-group">
            <label htmlFor="email">Email</label>
            <input
              type="email"
              id="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              required
              disabled={isConfirming || isResending}
            />
          </div>
          <div className="form-group">
            <label htmlFor="code">Verification Code</label>
            <input
              type="text"
              id="code"
              value={code}
              onChange={(e) => setCode(e.target.value)}
              required
              disabled={isConfirming || isResending}
              placeholder="Enter the code from your email"
            />
          </div>
          <button type="submit" disabled={isConfirming || isResending} className="auth-button">
            {isConfirming ? 'Verifying...' : 'Verify Account'}
          </button>
        </form>
        <div className="auth-links">
          <button
            onClick={handleResendCode}
            disabled={isConfirming || isResending}
            className="text-button"
          >
            {isResending ? 'Sending...' : 'Resend verification code'}
          </button>
          <Link to="/login">Back to Login</Link>
        </div>
      </div>
    </div>
  );
};

export default ConfirmSignup;