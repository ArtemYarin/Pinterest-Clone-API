import AuthModal from './AuthModal'
import { signup } from '../../api/signup'

const STATUS_MESSAGES = {
  409: 'An account with this email already exists.',
}

export default function SignUpModal({ onClose }) {
  return (
    <AuthModal
      title='Sign up'
      submitLabel='Sign up'
      pendingLabel='Signing up…'
      passwordAutoComplete='new-password'
      submit={signup}
      statusMessages={STATUS_MESSAGES}
      fallbackError='Sign up failed. Please try again.'
      onClose={onClose}
    />
  )
}
