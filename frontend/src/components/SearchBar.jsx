export default function SearchBar({ value, onChange, placeholder = 'Search' }) {
  return (
    <input
      id='search'
      type='text'
      value={value}
      onChange={(e) => onChange(e.target.value)}
      placeholder={placeholder}
      className='w-full text-big py-2 px-4 rounded-card border-none bg-bg hover:bg-bg-focus focus:border-none'
    />
  )
}
