import SearchBar from '../components/SearchBar'
import Avatar from '../components/Avatar'
import { useAuth } from '../context/authContext'
import LoginModal from '../components/auth/LoginModal'
import SignupButton from '../components/auth/SignUpButton'
import LoginButton from '../components/auth/LoginButton'

export default function Header({ onAuthSuccess, value, onChange, avaUrl }) {
  const { isAuthenticated } = useAuth()

  return (
    <header className='flex fixed shrink-0 top-0 right-0 left-0 z-50 bg-bg p-4 gap-16'>
      <SearchBar value={value} onChange={onChange} />

      {isAuthenticated ? (
        <Avatar url={avaUrl} />
      ) : (
        <div className='flex gap-4'>
          <LoginButton onSuccess={onAuthSuccess} />
          <SignupButton onSuccess={onAuthSuccess} />
        </div>
      )}
    </header>
  )
}
