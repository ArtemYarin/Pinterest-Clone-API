import { useState } from 'react'
import SearchBar from '../components/SearchBar'
import Avatar from '../components/Avatar'
import { useAuth } from '../context/authContext'
import LoginModal from '../components/auth/LoginModal'
import SignUpModal from '../components/auth/SignUpModal'

export default function Header({ value, onChange }) {
  const { isAuthenticated, isInitializing, logout } = useAuth()
  // 'login' | 'signup' | null
  const [openModal, setOpenModal] = useState(null)
  const closeModal = () => setOpenModal(null)

  return (
    <header className='sticky top-0 z-40 flex items-center gap-4 bg-bg py-4 sm:gap-16'>
      <SearchBar value={value} onChange={onChange} />

      {/* Render nothing while the session is restored to avoid flashing the auth buttons. */}
      {isInitializing ? null : isAuthenticated ? (
        <div className='flex items-center gap-4'>
          <button
            className='whitespace-nowrap rounded-button p-button font-semibold hover:bg-card-focus'
            type='button'
            onClick={logout}
          >
            Log out
          </button>
          <Avatar />
        </div>
      ) : (
        <div className='flex gap-4'>
          <button
            className='whitespace-nowrap rounded-button bg-card p-button font-semibold hover:bg-card-focus'
            type='button'
            onClick={() => setOpenModal('login')}
          >
            Log in
          </button>
          <button
            className='whitespace-nowrap rounded-button bg-accent p-button font-semibold text-bg hover:bg-accent-focus'
            type='button'
            onClick={() => setOpenModal('signup')}
          >
            Sign up
          </button>
        </div>
      )}

      {openModal === 'login' && <LoginModal onClose={closeModal} />}
      {openModal === 'signup' && <SignUpModal onClose={closeModal} />}
    </header>
  )
}
