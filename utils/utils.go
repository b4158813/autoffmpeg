package utils

import (
	"regexp"
	"strconv"
	"strings"
)

// 以sep分割mustContains，判断line是否包含任意的分割元素
func ContainsInfoAny(line, anyContains, sep string) bool {
	var temp []bool
	for _, str := range strings.Split(anyContains, sep) {
		temp = append(temp, strings.Contains(line, str))
	}
	return ReduceBoolListAnyTrue(temp)
}

// 以sep分割mustContains，判断line是否包含全部的分割元素
func ContainsInfoAll(line, mustContains, sep string) bool {
	var temp []bool
	for _, str := range strings.Split(mustContains, sep) {
		temp = append(temp, strings.Contains(line, str))
	}
	return ReduceBoolListAllTrue(temp)
}

// 返回bool类型的list是否包含true
func ReduceBoolListAnyTrue(list []bool) bool {
	anyTrue := reduceBoolList(list, func(a, b bool) bool {
		return a || b
	})
	return anyTrue
}

// 返回bool类型的list是否全是true
func ReduceBoolListAllTrue(list []bool) bool {
	allTrue := reduceBoolList(list, func(a, b bool) bool {
		return a && b
	})
	return allTrue
}

// reduce操作bool类型的list
func reduceBoolList(list []bool, fn func(bool, bool) bool) bool {
	if len(list) == 0 {
		return false
	}

	result := list[0]
	for i := 1; i < len(list); i++ {
		result = fn(result, list[i])
	}

	return result
}

func ContainsTranscodePushInfoAll(line string) bool {
	return ContainsInfoAll(line, TranscodePushInfo, SpaceSplit)
}

func ContainsVideoInfoAll(line string) bool {
	return ContainsInfoAll(line, VideoInfo, SpaceSplit)
}

func ContainsH265CodecAny(line string) bool {
	return ContainsInfoAny(line, CodecHevcAll, SpaceSplit)
}

func ContainsH264CodecAny(line string) bool {
	return ContainsInfoAny(line, CodecH264All, SpaceSplit)
}

// description: 返回正则匹配满足的最后一个字符串
func PatternMatchesLastFromString(line string, pattern string) (ret string) {
	var matches []string
	re := regexp.MustCompile(pattern)
	matches = re.FindAllString(line, -1)
	if len(matches) > 0 {
		ret = matches[len(matches)-1]
	}
	return
}

func GetCodecFromString(line string) (codec string) {
	codec_str := PatternMatchesLastFromString(line, CodecPattern)
	codec_ls := strings.Split(codec_str, VideoInfo)
	if len(codec_ls) > 1 {
		codec = codec_ls[1]
	}
	return
}

// description: 从字符串中获取形如"123x123"的分辨率信息
// return: (width, height)
func GetResolutionFromString(line string) (width, height int) {
	resolution_str := PatternMatchesLastFromString(line, ResolutionPattern)
	resolution_wh := strings.Split(resolution_str, ResolutionSplit)
	if len(resolution_wh) > 1 {
		width, _ = strconv.Atoi(resolution_wh[0])
		height, _ = strconv.Atoi(resolution_wh[1])
	}
	return
}

func GetStdCodecFromCodec(codec_b string) string {
	switch {
	case ContainsH264CodecAny(codec_b):
		return CodecH264
	case ContainsH265CodecAny(codec_b):
		return CodecHevc
	}
	return CodecDefault
}

func GetStdModeFromResolution(width, height int) string {
	total_pixel := width * height
	switch {
	case total_pixel > 0 && total_pixel <= TotalPixel540P:
		return Mode540p
	case total_pixel > TotalPixel540P && total_pixel <= TotalPixel720P:
		return Mode720p
	case total_pixel > TotalPixel720P && total_pixel <= TotalPixel1080P:
		return Mode1080p
	case total_pixel > TotalPixel1080P:
		return Mode1080pPlus
	default:
		return ModeDefault
	}
}

func GetStdCodecFromString(line string) string {
	codec := GetCodecFromString(line)
	return GetStdCodecFromCodec(codec)
}

func GetStdModeFromString(line string) string {
	width, height := GetResolutionFromString(line)
	return GetStdModeFromResolution(width, height)
}

func ContainsFFmpeg(line string) bool {
	return strings.Contains(strings.ToLower(line), "ffmpeg")
}
