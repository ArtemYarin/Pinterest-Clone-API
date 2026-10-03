import AuthModal from './AuthModal'
import { login } from '../../api/login'

// Wrong password returns 401, unknown email 404 — show the same message for both.
const STATUS_MESSAGES = {
  401: 'Incorrect email or password.',
  404: 'Incorrect email or password.',
}

export default function LoginModal({ onClose }) {
  return (
    <AuthModal
      title='Log in'
      submitLabel='Log in'
      pendingLabel='Logging in…'
      passwordAutoComplete='current-password'
      submit={login}
      statusMessages={STATUS_MESSAGES}
      fallbackError='Login failed. Please try again.'
      onClose={onClose}
    />
  )
}
