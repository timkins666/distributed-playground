import { configureStore } from '@reduxjs/toolkit'
import authReducer, { Auth } from './features/auth/authSlice'

export default configureStore({
  reducer: {
    auth: authReducer
  },
})

export interface Store {
  auth: Auth
}