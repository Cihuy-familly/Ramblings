import { Link } from 'react-router-dom'
import { Calendar } from 'lucide-react'
import type { Post } from '../types'

interface PostCardProps {
  post: Post
}

export default function PostCard({ post }: PostCardProps) {
  const formattedDate = new Date(post.created_at).toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  })

  return (
    <Link
      to={`/posts/${post.slug}`}
      className="group block bg-white rounded-xl border border-gray-200 overflow-hidden shadow-sm hover:shadow-lg hover:-translate-y-1 transition-all duration-300"
    >
      {/* Thumbnail */}
      <div className="aspect-video bg-gray-100 overflow-hidden">
        {post.thumbnail_url ? (
          <img
            src={post.thumbnail_url}
            alt={post.title}
            className="w-full h-full object-cover group-hover:scale-105 transition-transform duration-300"
            loading="lazy"
          />
        ) : (
          <div className="w-full h-full bg-gradient-to-br from-brand-400 to-brand-600 flex items-center justify-center">
            <span className="text-white/60 text-4xl font-bold">
              {post.title.charAt(0).toUpperCase()}
            </span>
          </div>
        )}
      </div>

      {/* Content */}
      <div className="p-5">
        {/* Category badge */}
        {post.category && (
          <span className="inline-block px-2.5 py-1 text-xs font-medium text-brand-700 bg-brand-50 rounded-full mb-3">
            {post.category.name}
          </span>
        )}

        {/* Title */}
        <h3 className="text-lg font-semibold text-gray-900 line-clamp-2 group-hover:text-brand-600 transition-colors mb-2">
          {post.title}
        </h3>

        {/* Excerpt */}
        <p className="text-sm text-gray-600 line-clamp-2 mb-4">
          {post.excerpt}
        </p>

        {/* Meta */}
        <div className="flex items-center justify-between text-xs text-gray-500">
          <span className="font-medium text-gray-700">{post.creator.display_name}</span>
          <span className="flex items-center gap-1">
            <Calendar className="w-3.5 h-3.5" />
            {formattedDate}
          </span>
        </div>
      </div>
    </Link>
  )
}