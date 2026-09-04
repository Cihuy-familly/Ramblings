import { useEffect, useState } from 'react'
import { useParams, Link } from 'react-router-dom'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import { ArrowLeft, Calendar, User, Clock } from 'lucide-react'
import { getPost } from '../api/client'
import type { Post } from '../types'

export default function PostView() {
  const { slug } = useParams<{ slug: string }>()
  const [avatarError, setAvatarError] = useState(false)
  const [post, setPost] = useState<Post | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (!slug) return

    async function fetchPost(slugValue: string) {
      setLoading(true)
      setError(null)
      try {
        const { post: data } = await getPost(slugValue)
        setPost(data)
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Post not found')
      } finally {
        setLoading(false)
      }
    }

    if (slug) {
      fetchPost(slug)
    }
  }, [slug])

  // Loading skeleton
  if (loading) {
    return (
      <div className="min-h-screen pt-24 pb-16">
        <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="animate-pulse space-y-6">
            <div className="h-8 w-24 skeleton rounded"></div>
            <div className="aspect-video skeleton rounded-xl"></div>
            <div className="h-10 skeleton rounded w-3/4"></div>
            <div className="flex gap-4">
              <div className="h-5 w-16 skeleton rounded-full"></div>
              <div className="h-5 w-32 skeleton rounded"></div>
            </div>
            <div className="space-y-3">
              <div className="h-4 skeleton rounded w-full"></div>
              <div className="h-4 skeleton rounded w-full"></div>
              <div className="h-4 skeleton rounded w-5/6"></div>
              <div className="h-4 skeleton rounded w-full"></div>
              <div className="h-4 skeleton rounded w-4/5"></div>
            </div>
          </div>
        </div>
      </div>
    )
  }

  // Error state
  if (error || !post) {
    return (
      <div className="min-h-screen pt-24 pb-16">
        <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 text-center py-20">
          <div className="text-6xl mb-4">&#128533;</div>
          <h2 className="text-2xl font-bold text-gray-900 mb-2">Post not found</h2>
          <p className="text-gray-500 mb-6">{error || 'The post you are looking for does not exist or has been removed.'}</p>
          <Link
            to="/posts"
            className="inline-flex items-center gap-2 px-4 py-2 text-sm font-medium text-white bg-brand-600 rounded-lg hover:bg-brand-700 transition-colors"
          >
            <ArrowLeft className="w-4 h-4" />
            Back to posts
          </Link>
        </div>
      </div>
    )
  }

  const formattedDate = new Date(post.created_at).toLocaleDateString('en-US', {
    weekday: 'long',
    year: 'numeric',
    month: 'long',
    day: 'numeric',
  })

  const readingTime = Math.max(1, Math.ceil(post.content.split(/\s+/).length / 200))

  return (
    <div className="min-h-screen pt-24 pb-16">
      <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8">
        {/* Back button */}
        <Link
          to="/posts"
          className="inline-flex items-center gap-1.5 text-sm font-medium text-gray-600 hover:text-gray-900 transition-colors mb-8"
        >
          <ArrowLeft className="w-4 h-4" />
          Back to posts
        </Link>

        {/* Thumbnail */}
        {post.thumbnail_url && (
          <div className="rounded-xl overflow-hidden mb-8 shadow-lg">
            <img
              src={post.thumbnail_url}
              alt={post.title}
              className="w-full h-auto max-h-96 object-cover"
            />
          </div>
        )}

        {/* Category badges */}
        {post.categories && post.categories.length > 0 && (
          <div className="flex flex-wrap gap-2 mb-6">
            {post.categories.map((cat) => (
              <span key={cat.id} className="inline-block px-3 py-1 text-sm font-medium text-brand-700 bg-brand-50 rounded-full">
                {cat.name}
              </span>
            ))}
          </div>
        )}

        {/* Title */}
        <h1 className="text-3xl sm:text-4xl lg:text-5xl font-extrabold text-gray-900 leading-tight mb-6">
          {post.title}
        </h1>

        {/* Meta */}
        <div className="flex flex-wrap items-center gap-4 text-sm text-gray-500 mb-8 pb-8 border-b border-gray-200">
          <div className="flex items-center gap-2">
            <div className="w-8 h-8 rounded-full bg-brand-100 flex items-center justify-center overflow-hidden">
              {post.creator.avatar_url && !avatarError ? (
                <img
                  src={post.creator.avatar_url}
                  alt={post.creator.display_name}
                  className="w-full h-full object-cover"
                  onError={() => setAvatarError(true)}
                />
              ) : (
                <User className="w-4 h-4 text-brand-600" />
              )}
            </div>
            <span className="font-medium text-gray-700">{post.creator.display_name}</span>
          </div>
          <span className="flex items-center gap-1.5">
            <Calendar className="w-4 h-4" />
            {formattedDate}
          {post.updated_at !== post.created_at && (
            <span className="text-xs text-gray-400 bg-gray-100 px-2 py-0.5 rounded-full">
              Edited
            </span>
          )}
          </span>
          <span className="flex items-center gap-1.5">
            <Clock className="w-4 h-4" />
            {readingTime} min read
          </span>
        </div>

        {/* Content */}
        <article className="markdown-content text-gray-800 leading-relaxed">
          <ReactMarkdown remarkPlugins={[remarkGfm]}>
            {post.content}
          </ReactMarkdown>
        </article>
      </div>
    </div>
  )
}