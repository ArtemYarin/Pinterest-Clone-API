import { Routes, Route } from 'react-router'
import { AuthProvider } from './context/authContext'
import PinsPage from './pages/PinsPage'
import PinPage from './pages/PinPage'
import ProfilePage from './pages/ProfilePage'

function App() {
  return (
    <div className='font-body text-base font-medium text-text bg-bg'>
      <AuthProvider>
        <Routes>
          <Route path='/' element={<PinsPage />} />
          <Route path='/users/:userId' element={<ProfilePage />} />
          <Route path='/:pinId' element={<PinPage />} />
        </Routes>
      </AuthProvider>
    </div>
  )
}

export default App
