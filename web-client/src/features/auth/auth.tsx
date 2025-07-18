import { JSX } from "react";
import { useDispatch, useSelector } from "react-redux";

import { Auth, setLogin } from './authSlice'
import { Store } from "../../store";
import { Navigate, useLocation } from "react-router";

export function RequireAuth({ children }: { children: JSX.Element }) {
  const authStatus = useSelector<Store, Auth>((state) => state.auth)

  if (!authStatus.user) {
    // Redirect them to the /login page, but save the current location they were
    // trying to go to when they were redirected. This allows us to send them
    // along to that page after they login, which is a nicer user experience
    // than dropping them off on the home page.
    const location = useLocation()
    return <Navigate to="/login" state={{ from: location }} replace />;
  }

  return children;
}
