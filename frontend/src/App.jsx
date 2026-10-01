import { AuthProvider } from './context/authContext'
import PinPage from './pages/PinPage'

function App() {
  return (
    <div className='font-body text-base font-medium text-text bg-bg'>
      <AuthProvider>
        <PinPage />
      </AuthProvider>
    </div>
  )
}

export default App
