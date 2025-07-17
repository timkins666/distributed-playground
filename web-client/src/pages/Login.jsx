import React, { useState } from 'react';

export default function Login() {
    const [username, setUsername] = useState('');
    const [token, setToken] = useState(null);

    const doLogin = async () => {
        const res = await fetch('http://localhost:8081/login', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ username, password: 'test' })
        });
        const data = await res.json();
        setToken(data.token);
    };

    return (
        <div>
            <h1>Banking Local</h1>
            {token ? (
                <p>Logged in! Token: {token}</p>
            ) : (
                <>
                    <input value={username} onChange={e => setUsername(e.target.value)} />
                    <button onClick={doLogin}>Login</button>
                </>
            )}
        </div>
    )
}