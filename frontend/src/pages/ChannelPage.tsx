import { useEffect, useState } from 'react'
import { useLocation, Link } from 'react-router-dom'
import { Calendar, User, Users, Heart, ArrowLeft, ExternalLink } from 'lucide-react'
import { getChannel, followChannel, unfollowChannel, isFollowingChannel } from '../api/client'
import type { ChannelResponse } from '../api/client'
import { useAuth } from '../context/AuthContext'

export default function ChannelPage() {
  const location = useLocation()
  const match = location.pathname.match(/^\/@([^/]+)/)
  const slug = match ? match[1] : null
  const { token } = useAuth()
  const [channel, setChannel] = useState<ChannelResponse['channel'] | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [isFollowing, setIsFollowing] = useState(false)
  const [followerCount, setFollowerCount] = useState(0)
  const [followLoading, setFollowLoading] = useState(false)
  const [avatarError, setAvatarError] = useState(false)

  useEffect(() => {
    if (!slug) return
    setLoading(true)
    setError(null)

    Promise.all([
      getChannel(slug),
      token ? isFollowingChannel(slug) : Promise.resolve(null),
    ])
      .then(([channelData, followData]) => {
        setChannel(channelData.channel)
        setFollowerCount(channelData.channel.follower_count)
        if (followData) {
          setIsFollowing(followData.is_following)
        }
      })
      .catch((err) => setError(err instanceof Error ? err.message : 'Failed to load channel'))
      .finally(() => setLoading(false))
  }, [slug, token])

  // If the URL doesn't match /@slug, render nothing (let other catch-all routes handle it)
  if (!slug) return null

  async function handleFollow() {
    if (!slug || !token) return
    setFollowLoading(true)
    try {
      if (isFollowing) {
        const res = await unfollowChannel(slug)
        setIsFollowing(false)
        setFollowerCount(res.follower_count)
      } else {
        const res = await followChannel(slug)
        setIsFollowing(true)
        setFollowerCount(res.follower_count)
      }
    } catch (err) {
      console.error('Follow action failed:', err)
    } finally {
      setFollowLoading(false)
    }
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-[60vh]">
        <div className="w-8 h-8 border-4 border-brand-200 border-t-brand-600 rounded-full animate-spin" />
      </div>
    )
  }

  if (error || !channel) {
    return (
      <div className="max-w-3xl mx-auto px-4 py-20 text-center">
        <h1 className="text-2xl font-bold text-gray-800 mb-4">Channel not found</h1>
        <p className="text-gray-500 mb-8">{error || 'This channel does not exist.'}</p>
        <Link to="/" className="text-brand-600 hover:text-brand-700 font-medium">
          ← Back to home
        </Link>
      </div>
    )
  }

  return (
    <div className="max-w-4xl mx-auto px-4 py-10">
      {/* Back link */}
      <Link to="/" className="inline-flex items-center gap-1.5 text-sm text-gray-500 hover:text-gray-700 mb-8">
        <ArrowLeft className="w-4 h-4" />
        Back to home
      </Link>

      {/* Channel header */}
      <div className="flex flex-col sm:flex-row items-start sm:items-center gap-6 mb-10 pb-10 border-b border-gray-200">
        <div className="w-20 h-20 rounded-full bg-brand-100 flex items-center justify-center overflow-hidden flex-shrink-0">
          {channel.avatar_url && !avatarError ? (
            <img
              src={channel.avatar_url}
              alt={channel.display_name}
              className="w-full h-full object-cover"
              referrerPolicy="no-referrer"
              onError={() => setAvatarError(true)}
            />
          ) : (
            <User className="w-8 h-8 text-brand-600" />
          )}
        </div>

        <div className="flex-1 min-w-0">
          <h1 className="text-2xl font-bold text-gray-900 mb-1">@{channel.slug}</h1>
          <p className="text-lg text-gray-700 mb-1">{channel.display_name}</p>
          {channel.bio && <p className="text-sm text-gray-500 mb-3">{channel.bio}</p>}
          <div className="flex items-center gap-4 text-sm text-gray-500">
            <span className="flex items-center gap-1.5">
              <Users className="w-4 h-4" />
              {followerCount} {followerCount === 1 ? 'follower' : 'followers'}
            </span>
            <span className="flex items-center gap-1.5">
              <Calendar className="w-4 h-4" />
              Joined {new Date(channel.created_at).toLocaleDateString('en-US', { year: 'numeric', month: 'long' })}
            </span>
          </div>
        </div>

        <div className="flex-shrink-0">
          {token ? (
            <button
              onClick={handleFollow}
              disabled={followLoading}
              className={`inline-flex items-center gap-2 px-5 py-2.5 rounded-full text-sm font-medium transition-colors ${
                isFollowing
                  ? 'bg-gray-100 text-gray-700 hover:bg-gray-200 border border-gray-300'
                  : 'bg-brand-600 text-white hover:bg-brand-700'
              } disabled:opacity-50 disabled:cursor-not-allowed`}
            >
              <Heart className={`w-4 h-4 ${isFollowing ? 'fill-current' : ''}`} />
              {followLoading ? '...' : isFollowing ? 'Following' : 'Follow'}
            </button>
          ) : (
            <Link
              to="/login"
              className="inline-flex items-center gap-2 px-5 py-2.5 rounded-full text-sm font-medium bg-brand-600 text-white hover:bg-brand-700"
            >
              <Heart className="w-4 h-4" />
              Follow
            </Link>
          )}
        </div>
      </div>

      {/* Posts */}
      <div>
        <h2 className="text-lg font-semibold text-gray-800 mb-6">
          Posts
          {channel.posts.length > 0 && (
            <span className="text-gray-400 font-normal ml-1">({channel.posts.length})</span>
          )}
        </h2>

        {channel.posts.length === 0 ? (
          <div className="text-center py-16">
            <ExternalLink className="w-12 h-12 text-gray-300 mx-auto mb-4" />
            <p className="text-gray-500">No published posts yet.</p>
          </div>
        ) : (
          <div className="grid gap-6 sm:grid-cols-2">
            {channel.posts.map((post) => (
              <Link
                key={post.id}
                to={`/posts/${post.slug}`}
                className="group block p-5 rounded-xl border border-gray-200 hover:border-brand-300 hover:shadow-sm transition-all"
              >
                {post.thumbnail_url && (
                  <img
                    src={post.thumbnail_url}
                    alt={post.title}
                    className="w-full h-40 object-cover rounded-lg mb-4"
                  />
                )}
                <h3 className="font-semibold text-gray-800 group-hover:text-brand-600 mb-2 line-clamp-2">
                  {post.title}
                </h3>
                {post.excerpt && (
                  <p className="text-sm text-gray-500 line-clamp-2 mb-3">{post.excerpt}</p>
                )}
                <span className="text-xs text-gray-400">
                  {new Date(post.created_at).toLocaleDateString('en-US', {
                    year: 'numeric',
                    month: 'short',
                    day: 'numeric',
                  })}
                </span>
              </Link>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}