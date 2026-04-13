import type { Router } from 'vue-router'
import { useUserStore } from '@/store/modules/user'

export function setupPageGuard(router: Router) {
  router.beforeEach(async (to, from, next) => {
    const userStore = useUserStore()
    // 非管理员路由拦截 - 添加安全检查，防止 userInfo 或 role 为 undefined
    if (userStore.userInfo && userStore.userInfo.role !== undefined && userStore.userInfo.role !== 1 && to.path.includes('admin'))
      next({ name: '404' })
    else
      next()
  })
}
