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
    }

    interface ItemIconGroup extends Common.InfoBase {
        icon?: string
        title?: string
        sort?:number
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

