import request from '@/utils/request'

const api = {
    getList: '/panel/searchEngine/getList',
    add: '/panel/searchEngine/add',
    update: '/panel/searchEngine/update',
    delete: '/panel/searchEngine/delete',
    updateSort: '/panel/searchEngine/updateSort',
}

export const getList = () => {
    return request({
        url: api.getList,
        method: 'post',
    })
}

export const add = (data: { title: string; url: string; iconSrc: string }) => {
    return request({
        url: api.add,
        method: 'post',
        data,
    })
}

export const update = (data: { id: number; title: string; url: string; iconSrc: string; sort?: number }) => {
    return request({
        url: api.update,
        method: 'post',
        data,
    })
}

export const deletes = (data: { id: number }) => {
    return request({
        url: api.delete,
        method: 'post',
        data,
    })
}

export const updateSort = (data: { items: { id: number; sort: number }[] }) => {
    return request({
        url: api.updateSort,
        method: 'post',
        data,
    })
}
