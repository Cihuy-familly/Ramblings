import type { Category } from '../types'

interface CategoryFilterProps {
  categories: Category[]
  activeCategory: string | null
  onCategoryChange: (slug: string | null) => void
}

export default function CategoryFilter({ categories, activeCategory, onCategoryChange }: CategoryFilterProps) {
  return (
    <div className="overflow-x-auto pb-2">
      <div className="flex gap-2 min-w-max">
        <button
          onClick={() => onCategoryChange(null)}
          className={`px-4 py-2 text-sm font-medium rounded-full transition-colors ${
            activeCategory === null
              ? 'bg-brand-600 text-white shadow-sm'
              : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
          }`}
        >
          All
        </button>
        {categories.map((category) => (
          <button
            key={category.id}
            onClick={() => onCategoryChange(category.slug)}
            className={`px-4 py-2 text-sm font-medium rounded-full transition-colors whitespace-nowrap ${
              activeCategory === category.slug
                ? 'bg-brand-600 text-white shadow-sm'
                : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
            }`}
          >
            {category.name}
            {category.post_count !== undefined && (
              <span className="ml-1.5 text-xs opacity-75">({category.post_count})</span>
            )}
          </button>
        ))}
        {categories.length === 0 && (
          <span className="px-4 py-2 text-sm text-gray-400">No categories</span>
        )}
      </div>
    </div>
  )
}