import React, { useState } from 'react';

function App() {
  const [username, setUsername] = useState('');
  const [token, setToken] = useState(null);

  const login = async () => {
    const res = await fetch('http://localhost:8080/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ username, password: 'test' })
    });
    const data = await res.json();
    setToken(data.token);
  };

  return (
    <div>
      <h1>Monzo Local</h1>
      {token ? (
        <p>Logged in! Token: {token}</p>
      ) : (
        <>
          <input value={username} onChange={e => setUsername(e.target.value)} />
          <button onClick={login}>Login</button>
        </>
      )}
    </div>
  );
}

export default App;
