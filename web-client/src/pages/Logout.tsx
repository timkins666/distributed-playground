import { Button } from "@mui/material";
import { useEffect } from "react";
import { useDispatch, useSelector } from "react-redux";
import { Link } from "react-router";
import { authSelector, setLogout } from "../components/auth/authSlice";

export default function Logout() {
  const authStatus = useSelector(authSelector);
  const dispatch = useDispatch();

  useEffect(() => {
    const doLogout = () => {
      const timer = setTimeout(() => dispatch(setLogout()), 5 * 1000);

      return () => {
        clearTimeout(timer);
      };
    };
    doLogout();
  }, []);

  if (authStatus.loggedIn) {
    return <>logging out in 5 seconds...</>;
  } else {
    return (
      <>
        <p>logged out</p>
        <Button>
          <Link to="/login">Return to login</Link>
        </Button>
      </>
    );
  }
}
