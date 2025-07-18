import { createSlice } from '@reduxjs/toolkit'
import { Auth, State } from '../../store'


export const authSlice = createSlice({
  name: 'store',
  initialState: {
    username: null,
    roles: null,
    token: null,
    loginTime: null,
  } as Auth,
  reducers: {
    setLogin: (state, newValue) => {
      state.username = newValue.payload.username
      state.roles = newValue.payload.roles
      state.token = newValue.payload.token
      state.loginTime = Date.now()
      state.loggedIn = true
    }
  },
})

export const { setLogin } = authSlice.actions

export const authSelector = (store: State) => store.auth

export default authSlice.reducer
