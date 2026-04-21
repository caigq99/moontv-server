package filter

var yellowWords = []string{
	"伦理片", "福利", "里番动漫", "门事件", "萝莉少女", "制服诱惑",
	"国产传媒", "cosplay", "黑丝诱惑", "无码", "日本无码", "有码",
	"日本有码", "SWAG", "网红主播", "色情片", "同性片", "福利视频",
	"福利片", "写真热舞", "倫理片", "理论片", "韩国伦理", "港台三级",
	"电影解说", "伦理", "日本伦理",
}

func IsYellow(typeName string) bool {
	for _, w := range yellowWords {
		if contains(typeName, w) {
			return true
		}
	}
	return false
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
