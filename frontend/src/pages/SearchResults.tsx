import { useEffect, useState, useCallback } from 'react'
import { useSearchParams, Link } from 'react-router-dom'
import { Search, FileText, User, RefreshCw, ChevronLeft, ChevronRight } from 'lucide-react'
import { search } from '../api/client'
import type { SearchResult, SearchCreatorResult } from '../api/client'
import PostCard from '../components/PostCard'

export default function SearchResults() {
  const [searchParams, setSearchParams] = useSearchParams()
  const query = searchParams.get('q') || ''
  const currentPage = Number(searchParams.get('page')) || 1

  const [posts, setPosts] = useState<SearchResult[]>([])
  const [creators, setCreators] = useState<SearchCreatorResult[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const fetchResults = useCallback(async () => {
    if (!query.trim()) return

    setLoading(true)
    setError(null)
    try {
      const data = await search(query, currentPage, 12)
      setPosts(data.posts)
      setCreators(data.creators)
      setTotal(data.total)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Search failed')
      setPosts([])
      setCreators([])
    } finally {
      setLoading(false)
    }
  }, [query, currentPage])

  useEffect(() => {
    fetchResults()
  }, [fetchResults])

  function handlePageChange(page: number) {
    const params = new URLSearchParams(searchParams)
    params.set('page', String(page))
    setSearchParams(params)
    window.scrollTo({ top: 0, behavior: 'smooth' })
  }

  const totalPages = Math.max(1, Math.ceil(total / 12))

  return (
    <div className="min-h-screen pt-24 pb-16">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        {/* Header */}
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-gray-900">Search</h1>
          {query && (
            <p className="mt-2 text-gray-600">
              {loading ? 'Searching...' : `Results for "${query}"`}
            </p>
          )}
        </div>

        {/* Empty query */}
        {!query.trim() && (
          <div className="text-center py-20">
            <div className="inline-flex items-center justify-center w-16 h-16 rounded-full bg-gray-100 mb-4">
              <Search className="w-8 h-8 text-gray-400" />
            </div>
            <h3 className="text-lg font-semibold text-gray-900 mb-2">Enter a search term</h3>
            <p className="text-gray-500">Type something in the search bar to find posts and creators.</p>
          </div>
        )}

        {/* Loading */}
        {loading && query.trim() && (
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
            {[...Array(6)].map((_, i) => (
              <div key={i} className="bg-white rounded-xl border border-gray-200 overflow-hidden">
                <div className="aspect-video skeleton"></div>
                <div className="p-5 space-y-3">
                  <div className="h-4 w-16 skeleton rounded-full"></div>
                  <div className="h-5 skeleton rounded w-3/4"></div>
                  <div className="h-4 skeleton rounded w-full"></div>
                  <div className="h-4 skeleton rounded w-1/2"></div>
                </div>
              </div>
            ))}
          </div>
        )}

        {/* Error */}
        {!loading && error && (
          <div className="text-center py-20">
            <div className="inline-flex items-center justify-center w-16 h-16 rounded-full bg-red-100 mb-4">
              <RefreshCw className="w-8 h-8 text-red-500" />
            </div>
            <h3 className="text-lg font-semibold text-gray-900 mb-2">Search failed</h3>
            <p className="text-gray-500 mb-6">{error}</p>
            <button
              onClick={fetchResults}
              className="inline-flex items-center gap-2 px-4 py-2 text-sm font-medium text-white bg-brand-600 rounded-lg hover:bg-brand-700 transition-colors"
            >
              <RefreshCw className="w-4 h-4" />
              Try Again
            </button>
          </div>
        )}

        {/* Results */}
        {!loading && !error && query.trim() && posts.length === 0 && creators.length === 0 && (
          <div className="text-center py-20">
            <div className="inline-flex items-center justify-center w-16 h-16 rounded-full bg-gray-100 mb-4">
              <FileText className="w-8 h-8 text-gray-400" />
            </div>
            <h3 className="text-lg font-semibold text-gray-900 mb-2">No results found</h3>
            <p className="text-gray-500">
              No results for "{query}". Try a different search term.
            </p>
          </div>
        )}

        {!loading && !error && query.trim() && (
          <>
            {/* Posts Section */}
            {posts.length > 0 && (
              <div className="mb-10">
                <h2 className="text-xl font-semibold text-gray-800 mb-4">Posts</h2>
                <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
                  {posts.map((post) => (
                    <PostCard
                      key={post.id}
                      post={{
                        id: post.id,
                        title: post.title,
                        slug: post.slug,
                        content: '',
                        excerpt: post.excerpt,
                        thumbnail_url: post.thumbnail_url,
                        published: true,
                        categories: post.category_name
                          ? post.category_name.split(', ').filter(Boolean).map((name, i) => ({ id: i, name, slug: '' }))
                          : [],
                        creator: {
                          id: post.creator_id,
                          display_name: post.creator_name,
                          avatar_url: '',
                          slug: '',
                          email: '',
                        },
                        created_at: new Date(post.created_at).toISOString(),
                        updated_at: new Date(post.created_at).toISOString(),
                      }}
                    />
                  ))}
                </div>
              </div>
            )}

            {/* Creators Section */}
            {creators.length > 0 && (
              <div className="mb-10">
                <h2 className="text-xl font-semibold text-gray-800 mb-4">Creators</h2>
                <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
                  {creators.map((creator) => (
                    <Link
                      key={creator.id}
                      to={`/@${creator.slug}`}
                      className="flex items-center gap-4 p-4 rounded-xl border border-gray-200 hover:bg-gray-50 transition-colors"
                    >
                      <div className="w-12 h-12 rounded-full bg-gray-100 flex items-center justify-center flex-shrink-0 overflow-hidden">
                        {creator.avatar_url ? (
                          <img
                            src={creator.avatar_url}
                            alt={creator.display_name}
                            className="w-full h-full object-cover"
                            onError={(e) => {
                              const target = e.currentTarget
                              target.style.display = 'none'
                              target.parentElement!.classList.add('flex', 'items-center', 'justify-center')
                              const icon = document.createElement('div')
                              icon.innerHTML = '<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="w-6 h-6 text-gray-400"><path d="M19 21v-2a4 4 0 0 0-4-4H9a4 4 0 0 0-4 4v2"/><circle cx="12" cy="7" r="4"/></svg>'
                              target.parentElement!.appendChild(icon.firstElementChild!)
                            }}
                          />
                        ) : (
                          <User className="w-6 h-6 text-gray-400" />
                        )}
                      </div>
                      <div className="min-w-0">
                        <p className="font-medium text-gray-900 truncate">{creator.display_name}</p>
                        <p className="text-sm text-gray-500 truncate">@{creator.slug}</p>
                        {creator.bio && (
                          <p className="text-sm text-gray-400 truncate mt-0.5">{creator.bio}</p>
                        )}
                      </div>
                    </Link>
                  ))}
                </div>
              </div>
            )}

            {/* Pagination */}
            {totalPages > 1 && (
              <div className="mt-12 flex items-center justify-center gap-2">
                <button
                  onClick={() => handlePageChange(currentPage - 1)}
                  disabled={currentPage <= 1}
                  className="inline-flex items-center gap-1 px-3 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-lg hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
                >
                  <ChevronLeft className="w-4 h-4" />
                  Previous
                </button>
                <span className="text-sm text-gray-500">
                  Page {currentPage} of {totalPages}
                </span>
                <button
                  onClick={() => handlePageChange(currentPage + 1)}
                  disabled={currentPage >= totalPages}
                  className="inline-flex items-center gap-1 px-3 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-lg hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
                >
                  Next
                  <ChevronRight className="w-4 h-4" />
                </button>
              </div>
            )}
          </>
        )}
      </div>
    </div>
  )
}