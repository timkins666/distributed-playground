import { useSelector } from 'react-redux'
import { authSelector } from '../features/auth/authSlice';

export default function Accounts() {
    const authStatus = useSelector(authSelector)

    return (
        <div>
            <h2>Accounts</h2>
        </div>
    )
}