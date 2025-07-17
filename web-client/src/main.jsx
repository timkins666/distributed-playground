import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { BrowserRouter, Routes, Route } from 'react-router'
// import './index.css'
import Login from './pages/Login'

createRoot(document.getElementById('root')).render(
  <StrictMode>
    <BrowserRouter>
      <Routes>
        {/* <Route path='/' element={<RootLayout />}> */}
          <Route index path='/' element={<Login />} />
        {/* </Route> */}
      </Routes>
    </BrowserRouter>
  </StrictMode>
)
