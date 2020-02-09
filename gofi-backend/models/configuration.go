package models

import "time"

type Configuration struct {
	Id                 int64     `json:"-"`
	ThemeStyle         string    `json:"themeStyle" validate:"required,oneof=dark light"` // 可选 dark,light
	NavMode            string    `json:"navMode" validate:"required,oneof=side top"`      // 导航模式,side,top
	CustomStoragePath  string    `json:"customStoragePath"`                   // 自定义文件仓库路径
	Initialized        bool      `json:"initialized"`                                     // 是否初始化
	Created            time.Time `json:"-" xorm:"created"`                                // 创建时间
	Updated            time.Time `json:"-" xorm:"updated"`                                // 更新时间
	LogDirectoryPath   string    `json:"logDirectoryPath" xorm:"-"`                       // 默认日志目录路径
	DatabaseFilePath   string    `json:"databaseFilePath" xorm:"-"`                       // 默认数据库文件路径
	DefaultStoragePath string    `json:"defaultStoragePath"  xorm:"-"`                    // 默认文件仓库路径,动态字段,无需持久化
	Version            string    `json:"version"  xorm:"-"`                               // 应用版本,动态字段,无需持久化
	AppPath            string    `json:"appPath" xorm:"-"`                                // 应用程序所在目录路径,动态字段,无需持久化
}
