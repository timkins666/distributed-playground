import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { BrowserRouter, Route, Routes } from "react-router";
// import './index.css'
import { Provider } from "react-redux";
import store from "./store";

import { RequireAuth } from "./components/auth/auth";
import RootLayout from "./layouts/RootLayout";
import Admin from "./pages/Admin";
import Index from "./pages/Index";
import Login from "./pages/Login";

import CssBaseline from "@mui/material/CssBaseline";
import Logout from "./pages/Logout";

import { ThemeProvider, createTheme } from "@mui/material/styles";
import Accounts2 from "./pages/Accounts2";

const theme = createTheme({ palette: { mode: "light" } }); // or 'dark'

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <Provider store={store}>
      <CssBaseline>
        <BrowserRouter>
          <ThemeProvider theme={theme}>
            <Routes>
              <Route path="/" element={<RootLayout />}>
                <Route index path="/" element={<Index />} />
                <Route path="/login" element={<Login />} />
                <Route path="/logout" element={<Logout />} />
                <Route
                  path="/accounts"
                  element={
                    <RequireAuth>
                      <Accounts2 />
                    </RequireAuth>
                  }
                />
                <Route
                  path="/admin"
                  element={
                    <RequireAuth requiredRoles={["admin"]}>
                      <Admin />
                    </RequireAuth>
                  }
                />
              </Route>
            </Routes>
          </ThemeProvider>
        </BrowserRouter>
      </CssBaseline>
    </Provider>
  </StrictMode>
);
