declare namespace Panel {

    interface Info extends ItemInfo {

    }

	interface bookmarkInfo extends ItemInfo {
		parentUrl?: string
		parentId?: number
	}

    interface ItemInfo extends Common.InfoBase {

        icon: ItemIcon |null
        title: string
        url: string
        sort?: number
        lanUrl?: string
        description?: string
        openMethod: number
        itemIconGroupId ?:number
        // 仅内网展示 0:关闭 1:开启
        lanOnly?: number
        // 所属文件夹ID，0 表示直接挂在分组下
        parentId?: number
        // 是否为文件夹 0:否 1:是
        isFolder?: number
        // 仅前端使用：文件夹下的图标
        children?: ItemInfo[]
    }

    interface ItemIconGroup extends Common.InfoBase {
        icon?: string
        title?: string
        sort?:number
        pageId?: number  // 所属页面ID
    }

    interface ItemPage extends Common.InfoBase {
        icon?: string
        title?: string
        sort?: number
    }

    interface ItemIcon {
        itemType: number
        src ?: string
        text ?: string
        // bgColor ?: string
        backgroundColor ?: string
    }

    interface State {
        rightSiderCollapsed: boolean
        leftSiderCollapsed: boolean
        networkMode:PanelStateNetworkModeEnum | null
        panelConfig:panelConfig
    }

    interface panelConfig{
        backgroundImageSrc?:string
        backgroundBlur?:number
        backgroundMaskNumber?:number
        iconStyle?:PanelPanelConfigStyleEnum
        iconTextColor?:string
        iconTextInfoHideDescription?:boolean
        iconTextIconHideTitle?:boolean
        logoText?:string
        logoImageSrc?:string
        clockShowSecond?:boolean
        clockColor?:string
        searchBoxShow?:boolean
        searchBoxSearchIcon?:boolean
        marginTop?:number
        marginBottom?:number
        maxWidth?:number
        maxWidthUnit:string
        marginX?:number
        footerHtml?:string
        systemMonitorShow?:boolean
        systemMonitorShowTitle?:boolean
        systemMonitorPublicVisitModeShow?:boolean
        netModeChangeButtonShow?:boolean
        secretModeButtonShow?:boolean
        autoNetworkWallpaper?:boolean // 是否启用自动获取网络壁纸
        autoNetworkWallpaperApi?:string // 自动获取网络壁纸API地址
        useBingWallpaper?:boolean // 是否使用 BING 每日壁纸
    }

    interface userConfig{
        panel:panelConfig
        searchEngine?:any
    }

    interface ItemIconSortRequest{
        sortItems:Common.SortItemRequest[]
        itemIconGroupId:number
    }
}

