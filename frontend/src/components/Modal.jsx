import { useEffect, useRef } from 'react'

// Native <dialog> shown as a modal: gives Esc-to-close, focus trapping and
// inert background for free. Closes on Esc, backdrop click and the X button.
export default function Modal({ onClose, labelledBy, children }) {
  const dialogRef = useRef(null)

  useEffect(() => {
    const dialog = dialogRef.current
    dialog.showModal()
    const { overflow } = document.body.style
    document.body.style.overflow = 'hidden'
    return () => {
      document.body.style.overflow = overflow
      dialog.close()
    }
  }, [])

  return (
    <dialog
      ref={dialogRef}
      aria-labelledby={labelledBy}
      className='m-auto bg-transparent p-0 text-text backdrop:bg-bg/40 backdrop:backdrop-blur-[2px]'
      onCancel={(e) => {
        // Let React own the open state instead of the browser closing the dialog.
        e.preventDefault()
        onClose()
      }}
      onClick={(e) => {
        // Clicks on the backdrop land on the <dialog> itself, not on the card.
        if (e.target === e.currentTarget) onClose()
      }}
    >
      <div className='relative rounded-card bg-card p-card'>
        <button
          type='button'
          onClick={onClose}
          className='absolute top-4 right-4 rounded-full p-button hover:bg-card-focus'
        >
          Close
        </button>
        {children}
      </div>
    </dialog>
  )
}
