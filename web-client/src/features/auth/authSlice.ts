import { createSlice } from '@reduxjs/toolkit'

export interface Auth {
  user: string | null
  token: string | null
  loginTime?: number | null
}

export const authSlice = createSlice({
  name: 'store',
  initialState: {
    user: null,
    token: null,
    loginTime: null,
  } as Auth,
  reducers: {
    setLogin: (state, newValue) => {
      state.user = newValue.payload.user
      state.token = newValue.payload.token
      state.loginTime = Date.now()
    }
  },
})

export const { setLogin } = authSlice.actions

export default authSlice.reducer
