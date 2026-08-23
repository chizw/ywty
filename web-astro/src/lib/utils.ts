import { clsx, type ClassValue } from 'clsx'
import { twMerge } from 'tailwind-merge'

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

/** 文件大小格式化 */
export function formatBytes(bytes: number): string {
  if (!bytes || bytes < 0) return '0 B'
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  if (bytes < 1024 * 1024 * 1024) return `${(bytes / 1024 / 1024).toFixed(2)} MB`
  return `${(bytes / 1024 / 1024 / 1024).toFixed(2)} GB`
}

/** 日期格式化（本地） */
export function formatDate(input: string | number | Date | undefined | null): string {
  if (!input) return '-'
  const d = new Date(input)
  if (Number.isNaN(d.getTime())) return String(input)
  return d.toLocaleString('zh-CN', { hour12: false })
}

/** 相对时间：刚刚 / n分钟前 / n小时前 / n天前 */
export function timeAgo(input: string | number | Date | undefined | null): string {
  if (!input) return '-'
  const t = new Date(input).getTime()
  if (Number.isNaN(t)) return '-'
  const diff = Date.now() - t
  const min = Math.floor(diff / 60000)
  if (min < 1) return '刚刚'
  if (min < 60) return `${min} 分钟前`
  const hour = Math.floor(min / 60)
  if (hour < 24) return `${hour} 小时前`
  const day = Math.floor(hour / 24)
  if (day < 30) return `${day} 天前`
  return new Date(input).toLocaleDateString('zh-CN')
}

/** 安全取数组 */
export function asArray<T>(v: unknown): T[] {
  if (Array.isArray(v)) return v as T[]
  if (v && Array.isArray((v as any).data)) return (v as any).data as T[]
  return []
}
