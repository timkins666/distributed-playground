import { useSelector } from "react-redux";
import { Link } from "react-router";
import { authSelector } from "../store/authSlice";

import { Button, ButtonGroup } from "@mui/material";

export const navBar = () => {
  const auth = useSelector(authSelector);

  if (!auth.loggedIn) {
    return null;
  }

  return (
    <ButtonGroup variant="contained" aria-label="Basic button group">
      <Button variant="outlined">
        <Link to="/accounts">accounts</Link>
      </Button>
      {auth.roles?.includes("admin") ? (
        <Button variant="outlined">
          <Link to="/admin">admin</Link>
        </Button>
      ) : null}
      {auth.loggedIn ? (
        <Button variant="outlined">
          <Link to="/logout">log out</Link>
        </Button>
      ) : null}
    </ButtonGroup>
  );
};

export default navBar;
