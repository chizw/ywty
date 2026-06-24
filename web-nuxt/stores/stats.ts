// 全局统计 store：图片数、相册数、容量等
// 跨页面共享：上传/删除时调用 refresh() 即可让所有引用它的页面同步
import { defineStore } from 'pinia'

export const useStatsStore = defineStore('stats', {
  state: () => ({
    photos: 0,
    albums: 0,
    usedBytes: 0,
    capacityBytes: 0,
    lastUpdatedAt: 0,
    loading: false,
  }),
  actions: {
    async refresh() {
      const api = useApi()
      this.loading = true
      try {
        // raw: true 让 useApi 返回完整响应 {code, message, data, meta}
        const photosRes = await api.get<any>('/api/v1/photos', { query: { page: 1, per_page: 1 }, raw: true })
        this.photos = Number(photosRes?.meta?.total ?? (Array.isArray(photosRes?.data) ? photosRes.data.length : 0))

        const albumsRes = await api.get<any>('/api/v1/albums', { raw: true }).catch(() => null)
        const list: any[] = Array.isArray(albumsRes?.data) ? albumsRes.data : (Array.isArray(albumsRes) ? albumsRes : [])
        this.albums = list.length

        const cap = await api.get<any>('/api/v1/capacity').catch(() => null)
        if (cap) {
          this.usedBytes = Number(cap.used ?? 0)
          this.capacityBytes = Number(cap.capacity ?? 0)
        }
        this.lastUpdatedAt = Date.now()
      } catch {
        // ignore
      } finally {
        this.loading = false
      }
    },
    // 轻量更新：上传/删除图片后通过 offset 调整数字，避免重复请求
    bumpPhotos(delta: number, deltaBytes = 0) {
      this.photos = Math.max(0, this.photos + delta)
      this.usedBytes = Math.max(0, this.usedBytes + deltaBytes)
      this.lastUpdatedAt = Date.now()
    },
  },
})
