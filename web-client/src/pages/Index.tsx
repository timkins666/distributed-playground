import { useSelector } from "react-redux";
import { Navigate } from "react-router";
import { authSelector } from "../components/auth/authSlice";

export default function Index() {
  const authStatus = useSelector(authSelector);

  if (!authStatus?.loggedIn) {
    return <Navigate to="/login"></Navigate>;
  }

  return <Navigate to="/accounts"></Navigate>;
}
