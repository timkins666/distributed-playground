import { createSlice } from "@reduxjs/toolkit";
import { AccountsState, State } from "../../store";
import { Account } from "./account";

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
      // MIBL
      newValue.payload.forEach((upd: Account) => {
        const idx = state.accounts.findIndex(acc => acc.accountId === upd.accountId)
        if (idx === -1) {
          state.accounts.push(upd)
        } else {
          state.accounts[idx].balance = upd.balance
        }
      });
      state.accounts = state.accounts.sort((a, b) => a.accountId - b.accountId);
    },
  },
});

export const { setTriedLoadAccounts, setAccounts, updateAccounts } =
  accountsSlice.actions;

export const userAccountsSelector = (store: State) => store.userAccounts;

export default accountsSlice.reducer;
