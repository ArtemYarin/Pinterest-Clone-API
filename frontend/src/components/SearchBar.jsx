export default function SearchBar({ value, onChange, placeholder = 'Search' }) {
  return (
    <input
      type='text'
      value={value}
      onChange={(e) => onChange(e.target.value)}
      placeholder={placeholder}
      className='w-full p-4'
    />
  )
}
