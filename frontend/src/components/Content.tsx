import React from 'react';

interface ContentProps {
  count: number;
  incrementCount: () => void;
}

const Content: React.FC<ContentProps> = ({ count, incrementCount }) => {
  return (
    <main className="content">
      <div className="card">
        <h2>Welcome to your CloudFront-ready React App!</h2>
        <p>
          This simple application demonstrates a React TypeScript app that can be
          deployed to Amazon CloudFront.
        </p>
        <p>
          You've clicked the button <strong>{count}</strong> times.
        </p>
        <button className="button" onClick={incrementCount}>
          Click me!
        </button>
      </div>
    </main>
  );
};

export default Content;
