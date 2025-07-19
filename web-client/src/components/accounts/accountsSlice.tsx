import { createSlice } from "@reduxjs/toolkit";
import { AccountsState, State } from "../../store";

export const accountsSlice = createSlice({
  name: "userAccounts",
  initialState: {
    triedLoad: false,
    accounts: [],
  } as AccountsState,
  reducers: {
    setTriedLoad: (state) => {
      state.triedLoad = true;
    },
    setAccounts: (state, newValue) => {
      state.accounts = newValue.payload;
    },
    addAccount: (state, newValue) => {
      state.accounts = [...state.accounts, newValue.payload];
    },
  },
});

export const { setTriedLoad, setAccounts, addAccount } = accountsSlice.actions;

export const userAccountsSelector = (store: State) => store.userAccounts;

export default accountsSlice.reducer;
