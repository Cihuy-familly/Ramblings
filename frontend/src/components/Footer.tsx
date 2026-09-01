export default function Footer() {
  return (
    <footer className="bg-white border-t border-gray-200 mt-auto">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="flex flex-col sm:flex-row items-center justify-between gap-4">
          <p className="text-sm text-gray-500">&copy; 2026 Blog Platform</p>
          <p className="text-sm text-gray-500">
            Built with <span className="font-medium text-gray-700">Go</span> +{' '}
            <span className="font-medium text-gray-700">React</span>
          </p>
        </div>
      </div>
    </footer>
  )
}