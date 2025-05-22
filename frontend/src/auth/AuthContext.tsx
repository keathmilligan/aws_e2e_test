import React, { createContext, useState, useEffect, useContext } from 'react';
import {
  CognitoUser,
  AuthenticationDetails,
  CognitoUserPool,
  CognitoUserSession,
  ISignUpResult,
  CognitoUserAttribute
} from 'amazon-cognito-identity-js';

// Define the authentication state interface
interface AuthState {
  isAuthenticated: boolean;
  user: any | null;
  accessToken: string | null;
  idToken: string | null;
  refreshToken: string | null;
}

// Define the authentication context interface
interface AuthContextType {
  authState: AuthState;
  login: (email: string, password: string) => Promise<void>;
  signup: (email: string, password: string, firstName: string, lastName: string) => Promise<void>;
  confirmSignup: (email: string, code: string) => Promise<void>;
  resendConfirmationCode: (email: string) => Promise<void>;
  logout: () => void;
  forgotPassword: (email: string) => Promise<void>;
  confirmForgotPassword: (email: string, code: string, newPassword: string) => Promise<void>;
  isLoading: boolean;
  error: string | null;
}

// Create the authentication context
const AuthContext = createContext<AuthContextType | undefined>(undefined);

// Get the Cognito configuration from environment variables
const userPoolId = process.env.REACT_APP_USER_POOL_ID || '';
const clientId = process.env.REACT_APP_USER_POOL_CLIENT_ID || '';

// Create the Cognito user pool
const userPool = new CognitoUserPool({
  UserPoolId: userPoolId,
  ClientId: clientId,
});

// Create the authentication provider component
export const AuthProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  // Initialize the authentication state
  const [authState, setAuthState] = useState<AuthState>({
    isAuthenticated: false,
    user: null,
    accessToken: null,
    idToken: null,
    refreshToken: null,
  });
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);

  // Check if the user is authenticated on component mount
  useEffect(() => {
    const checkAuth = async () => {
      try {
        // Check for tokens in localStorage first
        const idToken = localStorage.getItem('idToken');
        const accessToken = localStorage.getItem('accessToken');
        const refreshToken = localStorage.getItem('refreshToken');
        
        // If we have tokens in localStorage, try to get the current user
        const currentUser = userPool.getCurrentUser();
        
        if (currentUser && idToken && accessToken && refreshToken) {
          currentUser.getSession((err: Error | null, session: any) => {
            if (err) {
              console.error('Error getting session:', err);
              // Clear invalid tokens
              localStorage.removeItem('accessToken');
              localStorage.removeItem('idToken');
              localStorage.removeItem('refreshToken');
              setIsLoading(false);
              return;
            }

            if (session.isValid()) {
              setAuthState({
                isAuthenticated: true,
                user: currentUser,
                accessToken,
                idToken,
                refreshToken,
              });
            } else {
              // Session is invalid, clear tokens
              localStorage.removeItem('accessToken');
              localStorage.removeItem('idToken');
              localStorage.removeItem('refreshToken');
            }
            setIsLoading(false);
          });
        } else {
          // No user or tokens found
          setIsLoading(false);
        }
      } catch (error) {
        console.error('Error checking authentication:', error);
        setIsLoading(false);
      }
    };

    checkAuth();
  }, []);

  // Login function
  const login = async (email: string, password: string) => {
    setIsLoading(true);
    setError(null);

    try {
      const authenticationDetails = new AuthenticationDetails({
        Username: email,
        Password: password,
      });

      const cognitoUser = new CognitoUser({
        Username: email,
        Pool: userPool,
      });

      return new Promise<void>((resolve, reject) => {
        cognitoUser.authenticateUser(authenticationDetails, {
          onSuccess: (session: CognitoUserSession) => {
            const accessToken = session.getAccessToken().getJwtToken();
            const idToken = session.getIdToken().getJwtToken();
            const refreshToken = session.getRefreshToken().getToken();
            
            // Store tokens in localStorage
            localStorage.setItem('accessToken', accessToken);
            localStorage.setItem('idToken', idToken);
            localStorage.setItem('refreshToken', refreshToken);
            
            setAuthState({
              isAuthenticated: true,
              user: cognitoUser,
              accessToken,
              idToken,
              refreshToken,
            });
            setIsLoading(false);
            resolve();
          },
          onFailure: (err: Error) => {
            console.error('Error logging in:', err);
            setError(err.message || 'Failed to login');
            setIsLoading(false);
            reject(err);
          },
        });
      });
    } catch (error: any) {
      console.error('Error logging in:', error);
      setError(error.message || 'Failed to login');
      setIsLoading(false);
      throw error;
    }
  };

  // Signup function
  const signup = async (email: string, password: string, firstName: string, lastName: string) => {
    setIsLoading(true);
    setError(null);

    try {
      return new Promise<void>((resolve, reject) => {
        const attributeList = [
          new CognitoUserAttribute({ Name: 'email', Value: email }),
          new CognitoUserAttribute({ Name: 'given_name', Value: firstName }),
          new CognitoUserAttribute({ Name: 'family_name', Value: lastName }),
        ];
        
        userPool.signUp(
          email,
          password,
          attributeList,
          [],
          function(err: Error | undefined, result: ISignUpResult | undefined) {
            if (err) {
              console.error('Error signing up:', err);
              setError(err.message || 'Failed to sign up');
              setIsLoading(false);
              reject(err);
              return;
            }
            setIsLoading(false);
            resolve();
          }
        );
      });
    } catch (error: any) {
      console.error('Error signing up:', error);
      setError(error.message || 'Failed to sign up');
      setIsLoading(false);
      throw error;
    }
  };

  // Confirm signup function
  const confirmSignup = async (email: string, code: string) => {
    setIsLoading(true);
    setError(null);

    try {
      const cognitoUser = new CognitoUser({
        Username: email,
        Pool: userPool,
      });

      return new Promise<void>((resolve, reject) => {
        cognitoUser.confirmRegistration(code, true, function(err: Error | undefined, result: string | undefined) {
          if (err) {
            console.error('Error confirming signup:', err);
            setError(err.message || 'Failed to confirm signup');
            setIsLoading(false);
            reject(err);
            return;
          }
          setIsLoading(false);
          resolve();
        });
      });
    } catch (error: any) {
      console.error('Error confirming signup:', error);
      setError(error.message || 'Failed to confirm signup');
      setIsLoading(false);
      throw error;
    }
  };

  // Resend confirmation code function
  const resendConfirmationCode = async (email: string) => {
    setIsLoading(true);
    setError(null);

    try {
      const cognitoUser = new CognitoUser({
        Username: email,
        Pool: userPool,
      });

      return new Promise<void>((resolve, reject) => {
        cognitoUser.resendConfirmationCode(function(err: Error | undefined, result: any) {
          if (err) {
            console.error('Error resending confirmation code:', err);
            setError(err.message || 'Failed to resend confirmation code');
            setIsLoading(false);
            reject(err);
            return;
          }
          setIsLoading(false);
          resolve();
        });
      });
    } catch (error: any) {
      console.error('Error resending confirmation code:', error);
      setError(error.message || 'Failed to resend confirmation code');
      setIsLoading(false);
      throw error;
    }
  };

  // Logout function
  const logout = () => {
    const currentUser = userPool.getCurrentUser();
    if (currentUser) {
      currentUser.signOut();
      
      // Clear tokens from localStorage
      localStorage.removeItem('accessToken');
      localStorage.removeItem('idToken');
      localStorage.removeItem('refreshToken');
      
      setAuthState({
        isAuthenticated: false,
        user: null,
        accessToken: null,
        idToken: null,
        refreshToken: null,
      });
    }
  };

  // Forgot password function
  const forgotPassword = async (email: string) => {
    setIsLoading(true);
    setError(null);

    try {
      const cognitoUser = new CognitoUser({
        Username: email,
        Pool: userPool,
      });

      return new Promise<void>((resolve, reject) => {
        cognitoUser.forgotPassword({
          onSuccess: () => {
            setIsLoading(false);
            resolve();
          },
          onFailure: (err: Error) => {
            console.error('Error initiating forgot password:', err);
            setError(err.message || 'Failed to initiate password reset');
            setIsLoading(false);
            reject(err);
          },
        });
      });
    } catch (error: any) {
      console.error('Error initiating forgot password:', error);
      setError(error.message || 'Failed to initiate password reset');
      setIsLoading(false);
      throw error;
    }
  };

  // Confirm forgot password function
  const confirmForgotPassword = async (email: string, code: string, newPassword: string) => {
    setIsLoading(true);
    setError(null);

    try {
      const cognitoUser = new CognitoUser({
        Username: email,
        Pool: userPool,
      });

      return new Promise<void>((resolve, reject) => {
        cognitoUser.confirmPassword(code, newPassword, {
          onSuccess: () => {
            setIsLoading(false);
            resolve();
          },
          onFailure: (err: Error) => {
            console.error('Error confirming new password:', err);
            setError(err.message || 'Failed to reset password');
            setIsLoading(false);
            reject(err);
          },
        });
      });
    } catch (error: any) {
      console.error('Error confirming new password:', error);
      setError(error.message || 'Failed to reset password');
      setIsLoading(false);
      throw error;
    }
  };

  // Create the context value
  const contextValue: AuthContextType = {
    authState,
    login,
    signup,
    confirmSignup,
    resendConfirmationCode,
    logout,
    forgotPassword,
    confirmForgotPassword,
    isLoading,
    error,
  };

  // Return the provider component
  return <AuthContext.Provider value={contextValue}>{children}</AuthContext.Provider>;
};

// Create a hook to use the authentication context
export const useAuth = () => {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
};