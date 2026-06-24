// 全局共享类型
export interface UserInfo {
  id: number
  username: string
  name: string
  email: string
  avatar: string
  is_admin: boolean
  status: string
}

export interface TokenPair {
  access_token: string
  refresh_token: string
  token_type: string
  expires_at: string
  user: UserInfo
}

export interface ApiResponse<T = unknown> {
  code: number
  message: string
  data?: T
  meta?: {
    current_page: number
    per_page: number
    total: number
    last_page: number
  }
}
