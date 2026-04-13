export enum PanelStateNetworkModeEnum {
  // 原 WAN 模式（默认，公开内容）
  'wan_old' = 0,
  // 新 WAN 模式（待修改）
  'wan_new' = 1,
  // 新 LAN 模式（待修改）
  'lan_new' = 2,
  // 原 LAN/私密模式（需要密码，全部内容）
  'lan_old' = 3,
}

export enum PanelPanelConfigStyleEnum {
  'icon' = 1, // 图标风格
  'info' = 0, // 详情风格
  'small' = 1, // 同icon
}
