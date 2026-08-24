// 共享业务类型（与后端 API 契约对齐）

/** 公开图片（public/photos 列表项） */
export interface PublicPhoto {
  id: number
  url: string
  thumbnail_url?: string | null
  name?: string
  username?: string
  avatar_url?: string | null
  views?: number
  likes?: number
  width?: number | null
  height?: number | null
  created_at?: string
  // 允许额外字段（相册/分享等场景会带其他字段）
  [key: string]: unknown
}

/** 分享数据（/s/:slug 返回） */
export interface ShareData {
  type: 'photo' | 'album'
  share_id: number
  slug: string
  uuid: string
  url?: string
  thumbnail_url?: string | null
  size?: number
  width?: number | null
  height?: number | null
  name?: string
  description?: string | null
  photo_count?: number
  requires_password: boolean
}

/** 上传返回 */
export interface UploadResult {
  photo: {
    id: number
    name: string
    pathname: string
    mimetype: string
    size: number
    is_public: boolean
  }
  url: string
  markdown: string
  html: string
}

/** 我的图片（/photos 列表项，PhotoResponse） */
export interface MyPhoto {
  id: number
  uuid: string
  user_id: number
  album_id: number | null
  filename: string
  original_name: string
  url: string
  thumbnail_url: string | null
  size: number
  width: number | null
  height: number | null
  mime_type: string
  is_public: boolean
  views: number
  likes: number
  status: number
  created_at: string
  updated_at: string
}

/** 分页信封 */
export interface PageMeta {
  current_page: number
  per_page: number
  total: number
  last_page: number
}

export interface PagedData<T> {
  data: T[]
  meta: PageMeta
}
