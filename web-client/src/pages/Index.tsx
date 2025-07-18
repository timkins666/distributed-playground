import React from 'react';
import { useSelector } from 'react-redux'
import { authSelector } from '../features/auth/authSlice';
import { Navigate } from 'react-router';

export default function Index() {
    const authStatus = useSelector(authSelector)

    if (!authStatus?.loggedIn){
        return <Navigate to="/login"></Navigate>
    }

    return <Navigate to="/accounts"></Navigate>
}