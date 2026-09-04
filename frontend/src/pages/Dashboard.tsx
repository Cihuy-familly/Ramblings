import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { Plus, Edit3, Trash2, Eye, EyeOff, AlertTriangle, Calendar, FileText } from 'lucide-react'
import { getMyPosts, deletePost, getStudioStats } from '../api/client'
import type { StudioStats } from '../api/client'
import type { Post } from '../types'

export default function Dashboard() {
  const [posts, setPosts] = useState<Post[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [deleteConfirm, setDeleteConfirm] = useState<string | null>(null)
  const [deleting, setDeleting] = useState(false)
  const [stats, setStats] = useState<StudioStats | null>(null)
  const [statsLoading, setStatsLoading] = useState(true)

  async function fetchData() {
    setLoading(true)
    setError(null)
    try {
      const [postsData, statsData] = await Promise.all([
        getMyPosts(),
        getStudioStats(),
      ])
      setPosts(postsData.posts)
      setStats(statsData)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load data')
    } finally {
      setLoading(false)
      setStatsLoading(false)
    }
  }

  useEffect(() => {
    fetchData()
  }, [])

  async function handleDelete(id: string) {
    setDeleting(true)
    try {
      await deletePost(id)
      setPosts((prev) => prev.filter((p) => p.id !== id))
      setDeleteConfirm(null)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to delete post')
    } finally {
      setDeleting(false)
    }
  }

  function formatDate(dateStr: string) {
    return new Date(dateStr).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
    })
  }

  return (
    <div className="min-h-screen pt-24 pb-16">
      <div className="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8">
        {/* Header */}
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 mb-8">
          <div>
            <h1 className="text-3xl font-bold text-gray-900">Dashboard</h1>
            <p className="mt-1 text-gray-600">Manage your blog posts</p>
          </div>
          <Link
            to="/dashboard/new"
            className="inline-flex items-center gap-2 px-4 py-2.5 text-sm font-medium text-white bg-brand-600 rounded-lg hover:bg-brand-700 transition-colors shadow-sm"
          >
            <Plus className="w-4 h-4" />
            New Post
          </Link>
        </div>

        {/* Stats cards */}
        {stats && (
          <div className="grid grid-cols-2 sm:grid-cols-4 gap-4 mb-8">
            <div className="bg-white rounded-xl border border-gray-200 p-4">
              <p className="text-2xl font-bold text-gray-900">{stats.stats.total_posts}</p>
              <p className="text-sm text-gray-500 mt-1">Total Posts</p>
            </div>
            <div className="bg-white rounded-xl border border-gray-200 p-4">
              <p className="text-2xl font-bold text-gray-900">{stats.stats.follower_count}</p>
              <p className="text-sm text-gray-500 mt-1">Followers</p>
            </div>
            <div className="bg-white rounded-xl border border-gray-200 p-4">
              <p className="text-2xl font-bold text-gray-900">{stats.stats.published_posts}</p>
              <p className="text-sm text-gray-500 mt-1">Published</p>
            </div>
            <div className="bg-white rounded-xl border border-gray-200 p-4">
              <p className="text-2xl font-bold text-gray-900">{stats.stats.draft_posts}</p>
              <p className="text-sm text-gray-500 mt-1">Drafts</p>
            </div>
          </div>
        )}

        {/* Stats loading skeleton */}
        {statsLoading && (
          <div className="grid grid-cols-2 sm:grid-cols-4 gap-4 mb-8">
            {[...Array(4)].map((_, i) => (
              <div key={i} className="bg-white rounded-xl border border-gray-200 p-4">
                <div className="h-8 w-16 bg-gray-100 rounded animate-pulse mb-2" />
                <div className="h-4 w-20 bg-gray-100 rounded animate-pulse" />
              </div>
            ))}
          </div>
        )}

        {/* Error message */}
        {error && (
          <div className="flex items-start gap-3 p-4 mb-6 bg-red-50 border border-red-200 rounded-lg">
            <AlertTriangle className="w-5 h-5 text-red-500 flex-shrink-0 mt-0.5" />
            <div className="flex-1">
              <p className="text-sm text-red-700">{error}</p>
              <button
                onClick={() => setError(null)}
                className="text-sm text-red-600 underline mt-1 hover:text-red-800"
              >
                Dismiss
              </button>
            </div>
          </div>
        )}

        {/* Loading state */}
        {loading && (
          <div className="space-y-4">
            {[...Array(5)].map((_, i) => (
              <div key={i} className="bg-white rounded-xl border border-gray-200 p-6">
                <div className="space-y-3">
                  <div className="h-5 bg-gray-100 rounded w-3/4 animate-pulse" />
                  <div className="h-4 bg-gray-100 rounded w-1/2 animate-pulse" />
                  <div className="flex gap-4">
                    <div className="h-4 w-16 bg-gray-100 rounded-full animate-pulse" />
                    <div className="h-4 w-24 bg-gray-100 rounded animate-pulse" />
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}

        {/* Empty state */}
        {!loading && !error && posts.length === 0 && (
          <div className="text-center py-20">
            <div className="inline-flex items-center justify-center w-16 h-16 rounded-full bg-gray-100 mb-4">
              <FileText className="w-8 h-8 text-gray-400" />
            </div>
            <h3 className="text-xl font-semibold text-gray-900 mb-2">No posts yet</h3>
            <p className="text-gray-500 mb-6">You haven&apos;t created any posts yet. Create your first post!</p>
            <Link
              to="/dashboard/new"
              className="inline-flex items-center gap-2 px-4 py-2.5 text-sm font-medium text-white bg-brand-600 rounded-lg hover:bg-brand-700 transition-colors"
            >
              <Plus className="w-4 h-4" />
              Create Your First Post
            </Link>
          </div>
        )}

        {/* Posts list */}
        {!loading && !error && posts.length > 0 && (
          <div className="space-y-4">
            {/* Desktop table header */}
            <div className="hidden sm:grid grid-cols-12 gap-4 px-6 py-3 text-xs font-medium text-gray-500 uppercase tracking-wider bg-gray-50 rounded-lg border border-gray-200">
              <div className="col-span-5">Title</div>
              <div className="col-span-2">Status</div>
              <div className="col-span-2">Category</div>
              <div className="col-span-2">Date</div>
              <div className="col-span-1 text-right">Actions</div>
            </div>

            {posts.map((post) => (
              <div
                key={post.id}
                className="bg-white rounded-xl border border-gray-200 p-4 sm:grid sm:grid-cols-12 sm:gap-4 sm:items-center sm:px-6 sm:py-4 hover:shadow-sm transition-shadow"
              >
                {/* Title (mobile: full width, desktop: col-span-5) */}
                <div className="sm:col-span-5 mb-2 sm:mb-0">
                  <Link
                    to={`/posts/${post.slug}`}
                    className="text-sm font-medium text-gray-900 hover:text-brand-600 transition-colors line-clamp-1"
                  >
                    {post.title}
                  </Link>
                </div>

                {/* Status */}
                <div className="sm:col-span-2 mb-1 sm:mb-0">
                  <span
                    className={`inline-flex items-center gap-1.5 px-2.5 py-1 text-xs font-medium rounded-full ${
                      post.published
                        ? 'bg-green-50 text-green-700'
                        : 'bg-yellow-50 text-yellow-700'
                    }`}
                  >
                    {post.published ? (
                      <Eye className="w-3 h-3" />
                    ) : (
                      <EyeOff className="w-3 h-3" />
                    )}
                    {post.published ? 'Published' : 'Draft'}
                  </span>
                </div>

                {/* Category */}
                <div className="sm:col-span-2 mb-1 sm:mb-0">
                  <span className="text-sm text-gray-600">
                    {post.categories && post.categories.length > 0
                      ? post.categories.map(c => c.name).join(', ')
                      : <span className="text-gray-400 italic">None</span>}
                  </span>
                </div>

                {/* Date */}
                <div className="sm:col-span-2 mb-3 sm:mb-0">
                  <span className="text-sm text-gray-500 flex items-center gap-1">
                    <Calendar className="w-3.5 h-3.5" />
                    {formatDate(post.created_at)}
                  </span>
                </div>

                {/* Actions */}
                <div className="sm:col-span-1 flex items-center gap-2 sm:justify-end">
                  <Link
                    to={`/dashboard/edit/${post.id}`}
                    className="p-2 text-gray-400 hover:text-brand-600 hover:bg-brand-50 rounded-lg transition-colors"
                    title="Edit post"
                  >
                    <Edit3 className="w-4 h-4" />
                  </Link>
                  <button
                    onClick={() => setDeleteConfirm(post.id)}
                    className="p-2 text-gray-400 hover:text-red-600 hover:bg-red-50 rounded-lg transition-colors"
                    title="Delete post"
                  >
                    <Trash2 className="w-4 h-4" />
                  </button>
                </div>
              </div>
            ))}
          </div>
        )}

        {/* Delete confirmation modal */}
        {deleteConfirm && (
          <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/50">
            <div className="bg-white rounded-xl shadow-xl max-w-sm w-full p-6">
              <div className="flex items-start gap-3 mb-4">
                <div className="flex-shrink-0 w-10 h-10 rounded-full bg-red-100 flex items-center justify-center">
                  <AlertTriangle className="w-5 h-5 text-red-600" />
                </div>
                <div>
                  <h3 className="text-lg font-semibold text-gray-900">Delete Post</h3>
                  <p className="text-sm text-gray-500 mt-1">
                    Are you sure you want to delete this post? This action cannot be undone.
                  </p>
                </div>
              </div>
              <div className="flex gap-3 justify-end">
                <button
                  onClick={() => setDeleteConfirm(null)}
                  disabled={deleting}
                  className="px-4 py-2 text-sm font-medium text-gray-700 bg-gray-100 rounded-lg hover:bg-gray-200 disabled:opacity-50 transition-colors"
                >
                  Cancel
                </button>
                <button
                  onClick={() => handleDelete(deleteConfirm)}
                  disabled={deleting}
                  className="px-4 py-2 text-sm font-medium text-white bg-red-600 rounded-lg hover:bg-red-700 disabled:opacity-50 transition-colors flex items-center gap-2"
                >
                  {deleting ? (
                    <>
                      <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white" />
                      Deleting...
                    </>
                  ) : (
                    'Delete'
                  )}
                </button>
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}