/// <reference types="astro/client" />
import type { TokenPair } from './lib/auth'

declare global {
  namespace App {
    interface Locals {
      /** SSR 端认证信息（由 middleware 从 cookie 解析注入） */
      auth: TokenPair | null
    }
  }
}

interface ImportMetaEnv {
  readonly PUBLIC_API_BASE?: string
  readonly SITE_URL?: string
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}

export {}
