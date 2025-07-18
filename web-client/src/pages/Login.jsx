import React, { useState } from 'react';
import { useDispatch } from 'react-redux';
import { useSelector } from 'react-redux'
import { Link } from 'react-router';

import { setLogin } from '../features/auth/authSlice';

export default function Login() {
    const [username, setUsername] = useState('');
    const token = useSelector((state) => state.auth?.token)
    const dispatch = useDispatch()

    const doLogin = async () => {
      const res = await fetch('http://localhost:8081/login', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ username, password:"test" })
      });
      const data = await res.json();
      dispatch(setLogin({user:username, token:data.token}));
  };

    return (
        <div>
            <h1>Banking Local</h1>
            {token ? (
                <>
                <p>Logged in! Token: {token}</p>
                <p><Link to='/admin'>Admin</Link></p>
                </>
            ) : (
                <>
                    <input value={username} onChange={e => setUsername(e.target.value)} />
                    <button onClick={doLogin}>Login</button>
                </>
            )}
        </div>
    )
}