import { useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { Navigate } from 'react-router';
import { setLogin, authSelector } from '../features/auth/authSlice';
import { Button, Icon, Input } from '@mui/material';
import ErrorOutlineIcon from '@mui/icons-material/ErrorOutline';

export default function Login() {    
    const [loginFailed, setLoginFailed] = useState(false);
    const [username, setUsername] = useState('');
    const dispatch = useDispatch()
    const loggedIn = useSelector(authSelector).loggedIn
    
    if (loggedIn) {
        return <Navigate to="/accounts"></Navigate>
    }

    const doLogin = async () => {
        const res = await fetch('http://localhost:8081/login', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ username, password: "test" })
        });

        if (res.status !== 200) {
            setLoginFailed(true)
            return
        }

        const data = await res.json();
        dispatch(setLogin({ username: data.username, roles: data.roles, token: data.token }));
    };

    return (
        <>
            <Input value={username} onChange={e => setUsername(e.target.value)} />
            <Button onClick={doLogin}>Login</Button>
            {loginFailed ? <ErrorOutlineIcon /> : null }
        </>
    )
}