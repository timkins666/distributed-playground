import { useSelector } from "react-redux"
import { authSelector } from "./auth/authSlice"
import { Link } from "react-router"

import { ButtonGroup, Button } from '@mui/material';

export const navBar = () => {
    const auth = useSelector(authSelector)

    if (!auth.loggedIn) {
        return null
    }

    return (
        <ButtonGroup variant="contained" aria-label="Basic button group">
            <Button variant="outlined"><Link to="/accounts">accounts</Link></Button>
            {auth.roles?.includes("admin") ?<Button variant="outlined"><Link to="/admin">admin</Link></Button>: null}
        </ButtonGroup>

    )
}

export default navBar