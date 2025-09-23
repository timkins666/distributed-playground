import { createSlice } from "@reduxjs/toolkit";
import { BankState, State } from "./index";

export const banksSlice = createSlice({
  name: "banks",
  initialState: {
    triedLoad: false,
    banks: [],
  } as BankState,
  reducers: {
    setTriedLoadBanks: (state) => {
      state.triedLoad = true;
    },
    setBanks: (state, newValue) => {
      state.banks = newValue.payload
    },
  },
});

export const { setTriedLoadBanks, setBanks } = banksSlice.actions;

export const banksSelector = (store: State) => store.banks;

export default banksSlice.reducer;