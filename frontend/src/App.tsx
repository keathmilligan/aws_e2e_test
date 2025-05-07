import React, { useState, useEffect } from 'react';
import { getMessages, createMessage, Message } from './api';
import './App.css';

const App: React.FC = () => {
  const [messages, setMessages] = useState<Message[]>([]);
  const [newMessage, setNewMessage] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Fetch messages on component mount
  useEffect(() => {
    fetchMessages();
  }, []);

  const fetchMessages = async () => {
    try {
      setLoading(true);
      setError(null);
      const data = await getMessages();
      setMessages(data);
    } catch (err) {
      console.error('Error fetching messages:', err);
      setError('Failed to load messages. Please try again later.');
    } finally {
      setLoading(false);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newMessage.trim()) return;

    try {
      setLoading(true);
      setError(null);
      const createdMessage = await createMessage(newMessage);
      setMessages([...messages, createdMessage]);
      setNewMessage('');
    } catch (err) {
      console.error('Error creating message:', err);
      setError('Failed to send message. Please try again later.');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="app">
      <header className="app-header">
        <h1>AWS End-to-End Example</h1>
        <p>React + TypeScript frontend with Go + Gin backend</p>
      </header>

      <main className="app-main">
        <div className="message-list">
          <h2>Messages</h2>
          {loading && <p className="loading">Loading...</p>}
          {error && <p className="error">{error}</p>}
          
          {messages.length === 0 && !loading ? (
            <p className="empty">No messages yet. Be the first to post!</p>
          ) : (
            <ul>
              {messages.map((message) => (
                <li key={message.id} className="message-item">
                  <p className="message-text">{message.text}</p>
                  <span className="message-timestamp">
                    {new Date(message.timestamp).toLocaleString()}
                  </span>
                </li>
              ))}
            </ul>
          )}
        </div>

        <form className="message-form" onSubmit={handleSubmit}>
          <h2>Post a Message</h2>
          <div className="form-group">
            <input
              type="text"
              value={newMessage}
              onChange={(e) => setNewMessage(e.target.value)}
              placeholder="Type your message here..."
              disabled={loading}
            />
            <button type="submit" disabled={loading || !newMessage.trim()}>
              Send
            </button>
          </div>
        </form>
      </main>

      <footer className="app-footer">
        <p>
          Deployed on AWS using CloudFront, ECS Fargate, and CloudFormation
        </p>
      </footer>
    </div>
  );
};

export default App;