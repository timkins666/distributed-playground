import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { BrowserRouter, Routes, Route } from 'react-router'
// import './index.css'
import { Provider } from 'react-redux'
import store from './store'

import Login from './pages/Login'
import Admin from './pages/Admin'
import { RequireAuth } from './features/auth/auth'

createRoot(document.getElementById('root')).render(
  <StrictMode>
    <Provider store={store}>
      <BrowserRouter>
        <Routes>
          {/* <Route path='/' element={<RootLayout />}> */}
          <Route index path='/login' element={<Login />} />
          <Route path='/admin' element={<RequireAuth><Admin /></RequireAuth>} />
          {/* </Route> */}
        </Routes>
      </BrowserRouter>
    </Provider>
  </StrictMode>
)
