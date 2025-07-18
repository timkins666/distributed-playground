import { useSelector } from "react-redux"
import { authSelector } from "./auth/authSlice"
import { Link } from "react-router"


export const navBar = () => {
    const auth = useSelector(authSelector)

    if (!auth.loggedIn) {
        return null
    }

    return (
        <div style={{ width: '100%' }}>
            <Link to="/accounts">accounts</Link>
            
            {auth.roles?.includes("admin") ? <Link to="/admin">admin</Link> : ""}
            
        </div>
    )
}

export default navBar