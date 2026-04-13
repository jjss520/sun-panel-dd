import { get, post } from '@/utils/request'

export interface NotepadInfo {
    id: number
    userId: number
    title: string
    content: string
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

export function saveNotepadContent(data: { id: number, title: string, content: string }) {
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
