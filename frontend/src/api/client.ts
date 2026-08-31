import type { Post, PaginatedResponse, Category, AuthResponse, Creator, PostFormData, UploadResponse } from '../types'

const BASE_URL = '/api/v1'

function getToken(): string | null {
  return localStorage.getItem('token')
}

async function fetchWithAuth(url: string, options: RequestInit = {}): Promise<Response> {
  const token = getToken()
  const headers: Record<string, string> = {
    ...(options.headers as Record<string, string>),
  }

  if (token) {
    headers['Authorization'] = `Bearer ${token}`
  }

  return fetch(url, { ...options, headers })
}

async function handleResponse<T>(response: Response): Promise<T> {
  if (!response.ok) {
    const error = await response.json().catch(() => ({ message: 'An error occurred' }))
    throw new Error(error.message || `HTTP ${response.status}`)
  }
  return response.json()
}

export async function getPosts(page = 1, limit = 12, category?: string): Promise<PaginatedResponse> {
  const params = new URLSearchParams({ page: String(page), limit: String(limit) })
  if (category) {
    params.set('category', category)
  }
  const res = await fetch(`${BASE_URL}/posts?${params}`)
  return handleResponse<PaginatedResponse>(res)
}

export async function getPost(slug: string): Promise<{ post: Post }> {
  const res = await fetch(`${BASE_URL}/posts/${slug}`)
  return handleResponse<{ post: Post }>(res)
}

export async function getCategories(): Promise<{ categories: Category[] }> {
  const res = await fetch(`${BASE_URL}/categories`)
  return handleResponse<{ categories: Category[] }>(res)
}

export async function login(email: string, password: string): Promise<AuthResponse> {
  const res = await fetch(`${BASE_URL}/auth/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, password }),
  })
  return handleResponse<AuthResponse>(res)
}

export async function getMe(): Promise<{ creator: Creator }> {
  const res = await fetchWithAuth(`${BASE_URL}/auth/me`)
  return handleResponse<{ creator: Creator }>(res)
}

export async function getMyPosts(): Promise<{ posts: Post[] }> {
  const res = await fetchWithAuth(`${BASE_URL}/creator/posts`)
  return handleResponse<{ posts: Post[] }>(res)
}

export async function createPost(data: PostFormData): Promise<{ post: Post }> {
  const res = await fetchWithAuth(`${BASE_URL}/creator/posts`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data),
  })
  return handleResponse<{ post: Post }>(res)
}

export async function updatePost(id: string, data: PostFormData): Promise<{ post: Post }> {
  const res = await fetchWithAuth(`${BASE_URL}/creator/posts/${id}`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data),
  })
  return handleResponse<{ post: Post }>(res)
}

export async function deletePost(id: string): Promise<void> {
  const res = await fetchWithAuth(`${BASE_URL}/creator/posts/${id}`, {
    method: 'DELETE',
  })
  if (!res.ok) {
    const error = await res.json().catch(() => ({ message: 'Failed to delete post' }))
    throw new Error(error.message || `HTTP ${res.status}`)
  }
}

export async function uploadImage(file: File): Promise<string> {
  const formData = new FormData()
  formData.append('file', file)
  const res = await fetchWithAuth(`${BASE_URL}/creator/upload`, {
    method: 'POST',
    body: formData,
  })
  const data = await handleResponse<UploadResponse>(res)
  return data.url
}