/**
 * 分组图标列表缓存
 *
 * 为什么不用 localStorage 单独扛：
 * 图标（尤其是上传后转成 base64 的图片）体积很大，一个分组动辄上百 KB ~ 数 MB，
 * 所有分组加起来很容易撞上浏览器 localStorage 5MB 的配额。一旦 setItem 抛
 * QuotaExceededError，这个分组的缓存就永远写不进去，于是每次刷新主页都要重新
 * 走一遍网络请求，表现为"别的分组秒出，这个分组慢好几秒"。
 *
 * 所以这里做两层缓存：
 * 1. localStorage：同步可读，命中后首屏无需等待，保持分组秒出；
 * 2. IndexedDB：容量大几个数量级，localStorage 写不进去时兜底，异步读取（几十毫秒）。
 * 另外加一层内存缓存，避免同一次会话内反复反序列化。
 */
import { ss } from './local'

const CACHE_KEY_PREFIX = 'itemIconList_'
const DB_NAME = 'sun-panel-cache'
const DB_VERSION = 1
const STORE_NAME = 'itemIconList'

// 会话内内存缓存，命中后完全同步返回
const memoryCache = new Map<number, Panel.ItemInfo[]>()

let dbPromise: Promise<IDBDatabase | null> | null = null
// IndexedDB 不可用（隐私模式 / 老浏览器）时只走 localStorage
let idbUnavailable = false

function cacheKey(groupId: number) {
  return `${CACHE_KEY_PREFIX}${groupId}`
}

// IndexedDB 在极少数场景（版本被其它标签页占用等）会一直不回调，加超时兜底，
// 避免一个卡住的事务把整页图标拖住
function withTimeout<T>(promise: Promise<T>, ms: number, fallback: T): Promise<T> {
  return new Promise<T>((resolve) => {
    let settled = false
    const timer = setTimeout(() => {
      if (!settled) {
        settled = true
        resolve(fallback)
      }
    }, ms)

    promise.then((value) => {
      if (!settled) {
        settled = true
        clearTimeout(timer)
        resolve(value)
      }
    })
  })
}

function createDb(): Promise<IDBDatabase | null> {
  return new Promise<IDBDatabase | null>((resolve) => {
    if (typeof indexedDB === 'undefined') {
      idbUnavailable = true
      resolve(null)
      return
    }

    try {
      const request = indexedDB.open(DB_NAME, DB_VERSION)
      request.onupgradeneeded = () => {
        const db = request.result
        if (!db.objectStoreNames.contains(STORE_NAME))
          db.createObjectStore(STORE_NAME)
      }
      request.onsuccess = () => resolve(request.result)
      request.onerror = () => {
        idbUnavailable = true
        resolve(null)
      }
      request.onblocked = () => resolve(null)
    }
    catch {
      idbUnavailable = true
      resolve(null)
    }
  })
}

function openDb(): Promise<IDBDatabase | null> {
  if (idbUnavailable)
    return Promise.resolve(null)

  if (!dbPromise)
    dbPromise = createDb()

  // 每次调用都套一层超时，避免首次打开被 blocked 时后续调用一起挂住
  return withTimeout(dbPromise, 2000, null)
}

function idbGet<T>(key: number): Promise<T | null> {
  return openDb().then(db => new Promise<T | null>((resolve) => {
    if (!db)
      return resolve(null)

    try {
      const request = db.transaction(STORE_NAME, 'readonly').objectStore(STORE_NAME).get(key)
      request.onsuccess = () => resolve((request.result as T) ?? null)
      request.onerror = () => resolve(null)
    }
    catch {
      resolve(null)
    }
  }))
}

function idbPut(key: number, value: unknown): Promise<void> {
  return openDb().then(db => new Promise<void>((resolve) => {
    if (!db)
      return resolve()

    try {
      const tx = db.transaction(STORE_NAME, 'readwrite')
      tx.objectStore(STORE_NAME).put(value, key)
      tx.oncomplete = () => resolve()
      tx.onerror = () => resolve()
      tx.onabort = () => resolve()
    }
    catch {
      resolve()
    }
  }))
}

function idbDelete(key: number): Promise<void> {
  return openDb().then(db => new Promise<void>((resolve) => {
    if (!db)
      return resolve()

    try {
      const tx = db.transaction(STORE_NAME, 'readwrite')
      tx.objectStore(STORE_NAME).delete(key)
      tx.oncomplete = () => resolve()
      tx.onerror = () => resolve()
      tx.onabort = () => resolve()
    }
    catch {
      resolve()
    }
  }))
}

function idbClear(): Promise<void> {
  return openDb().then(db => new Promise<void>((resolve) => {
    if (!db)
      return resolve()

    try {
      const tx = db.transaction(STORE_NAME, 'readwrite')
      tx.objectStore(STORE_NAME).clear()
      tx.oncomplete = () => resolve()
      tx.onerror = () => resolve()
      tx.onabort = () => resolve()
    }
    catch {
      resolve()
    }
  }))
}

/** 同步读取（只查内存 + localStorage），用于首屏需要"立刻出内容"的场景 */
export function getIconListCacheSync(groupId: number): Panel.ItemInfo[] | null {
  if (!groupId)
    return null

  const inMemory = memoryCache.get(groupId)
  if (inMemory)
    return inMemory

  const cached = ss.get(cacheKey(groupId)) as Panel.ItemInfo[] | null
  if (cached && Array.isArray(cached)) {
    memoryCache.set(groupId, cached)
    return cached
  }

  return null
}

/** 异步读取：先同步缓存，再查 IndexedDB（localStorage 配额写满时的兜底） */
export async function getIconListCache(groupId: number): Promise<Panel.ItemInfo[] | null> {
  if (!groupId)
    return null

  const syncHit = getIconListCacheSync(groupId)
  if (syncHit)
    return syncHit

  const fromIdb = await withTimeout<Panel.ItemInfo[] | null>(idbGet<Panel.ItemInfo[]>(groupId), 1500, null)
  if (fromIdb && Array.isArray(fromIdb)) {
    memoryCache.set(groupId, fromIdb)
    // 顺手回填 localStorage（这次能写上就下次更快，写不上也不影响索引）
    try {
      ss.set(cacheKey(groupId), fromIdb)
    }
    catch {}
    return fromIdb
  }

  return null
}

/** 写入缓存；localStorage 写失败不抛错，交给 IndexedDB */
export async function setIconListCache(groupId: number, list: Panel.ItemInfo[]): Promise<void> {
  if (!groupId || !Array.isArray(list))
    return

  memoryCache.set(groupId, list)

  try {
    ss.set(cacheKey(groupId), list)
  }
  catch (error) {
    // 配额不足：把这份数据交给 IndexedDB，避免这个分组永远命中不了缓存
    try {
      ss.remove(cacheKey(groupId))
    }
    catch {}
    console.warn(`[iconCache] 分组 ${groupId} 的 localStorage 缓存写入失败，改用 IndexedDB 存储`, error)
  }

  await idbPut(groupId, list)
}

/** 删除单个分组的缓存（两条链路都要删，否则 IDB 里还留着旧数据） */
export function removeIconListCache(groupId: number): void {
  if (!groupId)
    return

  memoryCache.delete(groupId)
  try {
    ss.remove(cacheKey(groupId))
  }
  catch {}
  idbDelete(groupId)
}

/** 清空全部图标缓存 */
export function clearIconListCache(): void {
  memoryCache.clear()

  try {
    Object.keys(localStorage).forEach((key) => {
      if (key.startsWith(CACHE_KEY_PREFIX))
        localStorage.removeItem(key)
    })
  }
  catch {}

  idbClear()
}
