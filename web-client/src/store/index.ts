import { configureStore } from '@reduxjs/toolkit'
import { Account, Bank, UserCreds } from '../types'
import userAccountsReducer from './accountsSlice'
import authReducer from './authSlice'
import banksReducer from './banksSlice'

export default configureStore({
  reducer: {
    auth: authReducer,
    banks: banksReducer,
    userAccounts: userAccountsReducer,
  },
})

export interface State {
  auth: AuthState
  banks: BankState
  userAccounts: AccountsState
}

export interface AuthState extends UserCreds {
  loggedIn: boolean
}

export interface BankState {
  triedLoad: boolean
  banks: Bank[]
}

export interface AccountsState {
  triedLoad: boolean
  accounts: Account[]
}