import React, { useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { Link } from 'react-router';
import { setLogin, authSelector } from '../features/auth/authSlice';
import { State } from '../store';

export default function Login() {
    const [username, setUsername] = useState('');
    const loggedIn = useSelector((state: State) => authSelector(state).loggedIn)
    const dispatch = useDispatch()

    const doLogin = async () => {
        const res = await fetch('http://localhost:8081/login', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ username, password: "test" })
        });
        const data = await res.json();
        dispatch(setLogin({ username: data.username, roles: data.roles, token: data.token }));
    };

    return (
        <div>
            {loggedIn ? (
                <p>Logged in!</p>
            ) : (
                <>
                    <input value={username} onChange={e => setUsername(e.target.value)} />
                    <button onClick={doLogin}>Login</button>
                </>
            )}
        </div>
    )
}