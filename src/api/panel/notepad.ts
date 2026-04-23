import { get, post } from '@/utils/request'

export interface NotepadInfo {
    id: number
    userId: number
    title: string
    content: string
    remindTime?: string  // 提醒时间 ISO 格式
    remindStatus?: number  // 0=等待触发, 1=待确认, 2=已结束
    remindRepeat?: string  // 重复类型: none/daily/weekly/monthly/yearly
    remindForce?: number  // 强制提醒: 0=关闭, 1=开启
    remindAdvanceDays?: number  // 提前提醒天数（0=不提前）
    createdAt: string
    updatedAt: string
}

export function getNotepadContent(id?: number) {
    return get<NotepadInfo | null>({
        url: '/panel/notepad/get',
        data: id ? { id } : {},
    })
}

export function getNotepadList() {
    return get<NotepadInfo[]>({
        url: '/panel/notepad/getList',
    })
}

export function saveNotepadContent(data: { id: number, title: string, content: string, remindTime?: string | null, remindStatus?: number, remindRepeat?: string, remindForce?: number, remindAdvanceDays?: number }) {
    return post<NotepadInfo>({
        url: '/panel/notepad/save',
        data,
    })
}

export function deleteNotepad(data: { id: number }) {
    return post<any>({
        url: '/panel/notepad/delete',
        data,
    })
}

export function uploadNotepadFile(data: FormData) {
    return post<{ url: string, name: string, type: string }>({
        url: '/panel/notepad/upload',
        data,
        headers: {
            'Content-Type': 'multipart/form-data',
        },
    })
}

// 确认提醒（用户点击“我知道”）
export function acknowledgeReminder(data: { id: number }) {
    return post<any>({
        url: '/panel/notepad/acknowledge',
        data,
    })
}
