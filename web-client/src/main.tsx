import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { BrowserRouter, Routes, Route } from 'react-router'
// import './index.css'
import { Provider } from 'react-redux'
import store from './store'

import Login from './pages/Login'
import Admin from './pages/Admin'
import { RequireAuth } from './features/auth/auth'
import RootLayout from './layouts/RootLayout'
import Index from './pages/Index'
import Accounts from './pages/Accounts'

createRoot(document.getElementById('root')).render(
  <StrictMode>
    <Provider store={store}>
      <BrowserRouter>
        <Routes>
          <Route path='/' element={<RootLayout />}>
            <Route index path='/' element={<Index />} />
            <Route path='/login' element={<Login />} />
            <Route path='/accounts' element={<RequireAuth><Accounts /></RequireAuth>} />
            <Route path='/admin' element={<RequireAuth requiredRoles={["admin"]}><Admin /></RequireAuth>} />
          </Route>
        </Routes>
      </BrowserRouter>
    </Provider>
  </StrictMode>
)
