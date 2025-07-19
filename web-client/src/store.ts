import { configureStore } from '@reduxjs/toolkit'
import authReducer from './components/auth/authSlice'
import banksReducer from './components/banks/banksSlice'
import userAccountsReducer from './components/accounts/accountsSlice'
import { Bank } from './components/banks/bank'
import { Account } from './components/accounts/account'

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

export interface AuthState {
  loggedIn: boolean
  username: string | null
  roles: string[] | null
  token: string | null
  loginTime?: number | null
}

export interface BankState {
  triedLoad: boolean
  banks: Bank[]
}

export interface AccountsState {
  triedLoad: boolean
  accounts: Account[]
}
