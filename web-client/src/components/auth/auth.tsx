import { JSX } from "react";
import { useSelector } from "react-redux";

import { Navigate, useLocation } from "react-router";
import { authSelector } from "../../store/authSlice";

function RequireAuth({
  requiredRoles = null,
  children,
}: {
  requiredRoles?: [string] | null;
  children: JSX.Element;
}) {
  const authStatus = useSelector(authSelector);

  if (!authStatus.username) {
    // Redirect them to the /login page, but save the current location they were
    // trying to go to when they were redirected. This allows us to send them
    // along to that page after they login, which is a nicer user experience
    // than dropping them off on the home page.
    const location = useLocation();
    return <Navigate to="/login" state={{ from: location }} replace />;
  }

  if (requiredRoles !== null) {
    const hasRoles = requiredRoles.filter((role) =>
      authStatus.roles?.includes(role)
    );
    if (!hasRoles.length) {
      return <p>Unauthorised</p>;
    }
  }

  return children;
}

export { RequireAuth };
