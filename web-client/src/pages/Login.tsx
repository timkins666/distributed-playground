import ErrorOutlineIcon from "@mui/icons-material/ErrorOutline";
import { Button, Input } from "@mui/material";
import { useEffect, useState } from "react";
import { useDispatch, useSelector } from "react-redux";
import { Navigate } from "react-router";
import { gatewayHost } from "../conf";
import { authSelector, login } from "../store/authSlice";
import { getLocalStorageLogin, setLocalStorageLogin } from "../utils/auth";

export default function Login() {
  const [loginFailed, setLoginFailed] = useState(false);
  const [username, setUsername] = useState("");
  const loggedIn = useSelector(authSelector).loggedIn;
  const dispatch = useDispatch();


  useEffect(() => {
    const sessionCreds = getLocalStorageLogin();
    if (!loggedIn && sessionCreds) {
      dispatch(login(sessionCreds));
    }
  }, [])


  if (loggedIn) {
    return <Navigate to='/accounts'></Navigate>;
  }
  
  const doLogin = async () => {
    const res = await fetch(`${gatewayHost}/login`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ username, password: "test" }),
    });

    if (!res.ok) {
      setLoginFailed(true);
      return;
    }

    const creds = await res.json()
    setLocalStorageLogin(creds)
    dispatch(login(creds));
  };

  return (
    <>
      <Input value={username} onChange={(e) => setUsername(e.target.value)} />
      <Button onClick={doLogin}>Login</Button>
      {loginFailed ? <ErrorOutlineIcon /> : null}
    </>
  );
}
