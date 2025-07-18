// import React, { useState } from 'react';
import { useSelector } from 'react-redux'
import { useNavigate } from 'react-router';

export default function Admin() {
    const authStatus = useSelector((state) => state.auth)

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
            <h1>Banking Local</h1>
            <h2>Admin</h2>

            <p>token: {authStatus.token}</p>

            <>
                <button onClick={pingAdmin}>Ping</button>
            </>
        </div>
    )
}