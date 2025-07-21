import { createSlice } from "@reduxjs/toolkit";
import { AccountsState, State } from "../../store";

export const accountsSlice = createSlice({
  name: "userAccounts",
  initialState: {
    triedLoad: false,
    accounts: [],
  } as AccountsState,
  reducers: {
    setTriedLoadAccounts: (state) => {
      state.triedLoad = true;
    },
    setAccounts: (state, newValue) => {
      state.accounts = newValue.payload;
    },
    updateAccounts: (state, newValue) => {
      const updateIds = newValue.payload.map((acc) => acc.accountId);
      state.accounts = [
        ...state.accounts.filter((acc) => !updateIds.includes(acc.accountId)),
        ...newValue.payload,
      ].sort((a, b) => a.accountId - b.accountId);
    },
  },
});

export const { setTriedLoadAccounts, setAccounts, updateAccounts } =
  accountsSlice.actions;

export const userAccountsSelector = (store: State) => store.userAccounts;

export default accountsSlice.reducer;
