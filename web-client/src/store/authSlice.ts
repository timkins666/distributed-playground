import { createSlice } from "@reduxjs/toolkit";
import { AuthState, State } from "./index";

export const authSlice = createSlice({
  name: "auth",
  initialState: {
    loggedIn: false,
    username: null,
    roles: null,
    token: null,
  } as AuthState,
  reducers: {
    login: (state, action) => {
      state.loggedIn = true;
      state.username = action.payload.username;
      state.roles = action.payload.roles;
      state.token = action.payload.token;
    },
    logout: (state) => {
      state.loggedIn = false;
      state.username = null;
      state.roles = null;
      state.token = null;
    },
  },
});

export const { login, logout } = authSlice.actions;

export const authSelector = (store: State) => store.auth;

export default authSlice.reducer;