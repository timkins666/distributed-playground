import { createSlice } from "@reduxjs/toolkit";
import { BankState, State } from "../../store";

export const banksSlice = createSlice({
  name: "banks",
  initialState: {
    triedLoad: false,
    banks: [],
  } as BankState,
  reducers: {
    setTriedLoad: (state) => {
      state.triedLoad = true;
    },
    setBanks: (state, newValue) => {
      state.banks = newValue.payload
    },
  },
});

export const { setTriedLoad, setBanks } = banksSlice.actions;

export const banksSelector = (store: State) => store.banks;

export default banksSlice.reducer;
