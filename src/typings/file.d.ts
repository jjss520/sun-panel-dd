declare namespace File {

	interface Info extends Common.InfoBase {
		src: string
		userId: number
		fileName: string
		method: number
		ext: string
		fileType?: string // 文件类型: notepad/wallpaper
	}


}