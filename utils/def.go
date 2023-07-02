package utils

const (
	// 空格分隔符
	SpaceSplit = " "
	// 分辨率分隔符
	ResolutionSplit = "x"

	// 转码拉源流后stderr必须包含的元素
	VideoInfo = "video:"
	// 转码推流成功后的stderr必须包含的元素(空格分割)
	TranscodePushInfo = "frame= fps= q= size= bitrate= speed="

	// 分辨率提取正则
	ResolutionPattern = `[1-9]+\d*x[1-9]+\d*`
	// 分辨率对应转码档位
	TotalPixel1080P = 1920 * 1080
	TotalPixel720P  = 1280 * 720
	TotalPixel540P  = 960 * 540
	// 监控上报档位
	Mode1080pPlus = "1080p+"
	Mode1080p     = "1080p"
	Mode720p      = "720p"
	Mode540p      = "540p"
	ModeDefault   = "null"

	// codec提取正则
	CodecPattern = VideoInfo + `\s*(.*?)\s`
	// codec类别
	CodecHevcAll = "h265 hevc"
	CodecH264All = "h264 avc"
	// 监控上报codec类型
	CodecHevc    = "h265"
	CodecH264    = "h264"
	CodecDefault = "null"
)
