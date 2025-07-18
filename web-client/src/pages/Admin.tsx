import { useSelector } from 'react-redux'
import { authSelector } from '../features/auth/authSlice';

export default function Admin() {
    const authStatus = useSelector(authSelector)

    const pingAdmin = async () => {
        const res = await fetch('http://localhost:8081/admin', {
            method: 'POST',
            headers: { 
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${authStatus.token}`
            },
            body: JSON.stringify({ username: authStatus.username})
        });
        // const data = await res.json();
    };

    return (
        <div>
            <h2>Admin</h2>

            <p>token: {authStatus.token}</p>

            <>
                <button onClick={pingAdmin}>Ping</button>
            </>
        </div>
    )
}