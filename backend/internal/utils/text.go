package utils

// TruncateContent 截断内容到指定长度
func TruncateContent(content string, maxLength int) string {
    if len(content) <= maxLength {
        return content
    }
    return content[:maxLength] + "..."
} 