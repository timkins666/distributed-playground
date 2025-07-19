import { Outlet } from "react-router";

import NavBar from "../components/topNav";

export default function RootLayout() {
  return (
    <>
      <h1>Banking Local</h1>
      <NavBar />
      <Outlet />
    </>
  );
}
