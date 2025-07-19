import { createSlice } from "@reduxjs/toolkit";
import { AuthState, State } from "../../store";

export const authSlice = createSlice({
  name: "auth",
  initialState: {
    username: null,
    roles: null,
    token: null,
    loginTime: null,
  } as AuthState,
  reducers: {
    setLogin: (state, newValue) => {
      state.username = newValue.payload.username;
      state.roles = newValue.payload.roles;
      state.token = newValue.payload.token;
      state.loginTime = Date.now();
      state.loggedIn = true;
    },
    setLogout: (state) => {
      state.username = null 
      state.roles = null
      state.token = null
      state.loginTime = null;
      state.loggedIn = false;
    },
  },
});

export const { setLogin, setLogout } = authSlice.actions;

export const authSelector = (store: State) => store.auth;

export default authSlice.reducer;
