export interface Creator {
  id: string
  email: string
  display_name: string
  avatar_url: string
  bio?: string
}

export interface Category {
  id: number
  name: string
  slug: string
  post_count?: number
}

export interface Post {
  id: string
  title: string
  slug: string
  content: string
  excerpt: string
  thumbnail_url: string
  published: boolean
  category: Category | null
  creator: Creator
  created_at: string
  updated_at: string
}

export interface PaginatedResponse {
  posts: Post[]
  total: number
  page: number
  totalPages: number
}

export interface AuthResponse {
  token: string
  creator: Creator
}

export interface UploadResponse {
  url: string
}

export interface PostFormData {
  title: string
  content: string
  excerpt: string
  category_id: number | null
  thumbnail_url: string
  published: boolean
}