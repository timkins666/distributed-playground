import { Outlet } from "react-router";

import NavBar from "../features/topNav";

export default function RootLayout() {
    return <>
        <h1>Banking Local</h1>
        <NavBar />
        <Outlet />
    </>
}