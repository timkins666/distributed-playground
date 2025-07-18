import { configureStore } from '@reduxjs/toolkit'
import authReducer from './features/auth/authSlice'

export default configureStore({
  reducer: {
    auth: authReducer
  },
})

export interface State {
  auth: Auth
}

export interface Auth {
  loggedIn: boolean
  username: string | null
  roles: [string?] | null
  token: string | null
  loginTime?: number | null
}