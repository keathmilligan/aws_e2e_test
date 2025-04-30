import React, { useState } from 'react';
import Header from './components/Header';
import Content from './components/Content';
import Footer from './components/Footer';
import './App.css';

const App: React.FC = () => {
  const [count, setCount] = useState<number>(0);

  const incrementCount = (): void => {
    setCount(prevCount => prevCount + 1);
  };

  return (
    <div className="app">
      <Header />
      <Content count={count} incrementCount={incrementCount} />
      <Footer />
    </div>
  );
};

export default App;
