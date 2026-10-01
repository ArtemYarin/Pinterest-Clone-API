import { useState } from 'react'
import LoginModal from './LoginModal'

export default function LoginButton({ onSuccess }) {
  const [showModal, setShowModal] = useState(false)
  return (
    <>
      <button
        className='bg-card text-text font-semibold p-button rounded-button hover:bg-card-focus'
        type='button'
        onClick={() => setShowModal(true)}
      >
        Login
      </button>
      {showModal && (
        <LoginModal onClose={() => setShowModal(false)} onSuccess={onSuccess} />
      )}
    </>
  )
}
