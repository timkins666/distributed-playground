import { Outlet } from "react-router";

import NavBar from "../components/topNav";

export default function RootLayout() {
  return (
    <>
      <h1>The Bank of Tim</h1>
      <NavBar />
      <Outlet />
    </>
  );
}
