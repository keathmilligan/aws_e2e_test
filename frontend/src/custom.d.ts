declare module '*.css' {
  const content: { [className: string]: string };
  export default content;
}

declare module './components/Header' {
  import React from 'react';
  const Header: React.FC;
  export default Header;
}

declare module './components/Content' {
  import React from 'react';
  interface ContentProps {
    count: number;
    incrementCount: () => void;
  }
  const Content: React.FC<ContentProps>;
  export default Content;
}

declare module './components/Footer' {
  import React from 'react';
  const Footer: React.FC;
  export default Footer;
}

declare module './App' {
  import React from 'react';
  const App: React.FC;
  export default App;
}
