package service

import (
	"fmt"
	"strings"

	"github.com/pkoukk/tiktoken-go"
)

// TokenCounter 接口定义了计算tokens的方法
type TokenCounter interface {
	CountTokens(text string) (int, error)
	GetModelType() string
}

// OpenAITokenCounter OpenAI模型的token计数器
type OpenAITokenCounter struct {
	model string
}

// NewOpenAITokenCounter 创建一个新的OpenAI token计数器
func NewOpenAITokenCounter(model string) *OpenAITokenCounter {
	return &OpenAITokenCounter{
		model: model,
	}
}

// CountTokens 计算OpenAI模型的tokens
func (c *OpenAITokenCounter) CountTokens(text string) (int, error) {
	tkm, err := tiktoken.EncodingForModel(c.model)
	if err != nil {
		// 如果模型不支持，使用默认编码
		tkm, err = tiktoken.GetEncoding("cl100k_base")
		if err != nil {
			return 0, err
		}
	}
	tokens := tkm.Encode(text, nil, nil)
	return len(tokens), nil
}

func (c *OpenAITokenCounter) GetModelType() string {
	return "openai"
}

// SimpleTokenCounter 简单的token计数器
type SimpleTokenCounter struct{}

// NewSimpleTokenCounter 创建一个新的简单token计数器
func NewSimpleTokenCounter() *SimpleTokenCounter {
	return &SimpleTokenCounter{}
}

// CountTokens 使用简单规则计算tokens
func (c *SimpleTokenCounter) CountTokens(text string) (int, error) {
	// 1. 按空格分词
	words := strings.Fields(text)

	// 2. 初始化token计数
	tokens := 0

	// 3. 计算每个单词的tokens
	for _, word := range words {
		// 基本规则：
		// - 普通单词: 按平均4个字符1个token计算
		// - 数字和标点: 每个字符当作0.25个token
		// - 中文字符: 每个字符当作1个token
		// - 其他特殊字符: 每个字符当作0.5个token

		for _, char := range word {
			if char >= 0x4E00 && char <= 0x9FFF {
				// 中文字符
				tokens++
			} else if (char >= '0' && char <= '9') || char <= 0x7F {
				// 数字和ASCII字符
				tokens += 1
			} else {
				// 其他字符
				tokens++
			}
		}
	}

	// 4. 添加空格的token计数
	tokens += len(words) - 1

	return tokens, nil
}

func (c *SimpleTokenCounter) GetModelType() string {
	return "simple"
}

// GetTokenCounter 根据提供商类型和模型ID获取合适的token计数器
func GetTokenCounter(providerType string, modelID string) TokenCounter {
	switch providerType {
	case "OPENAI", "OPENAI_COMPATIBLE":
		return NewOpenAITokenCounter(modelID)
	default:
		return NewSimpleTokenCounter()
	}
}

// CountMessageTokens 计算单条消息的tokens
func CountMessageTokens(counter TokenCounter, role, content string) (int, error) {
	// 1. 基础消息格式的tokens
	// 每条消息的格式大约是: {"role": "xxx", "content": "xxx"}
	formatTokens := 4 // 包括 {}, role, content 等字段名

	// 2. 计算角色tokens (包括引号)
	roleTokens, err := counter.CountTokens(fmt.Sprintf(`"%s"`, role))
	if err != nil {
		return 0, err
	}

	// 3. 计算内容tokens (包括引号)
	contentTokens, err := counter.CountTokens(fmt.Sprintf(`"%s"`, content))
	if err != nil {
		return 0, err
	}

	return formatTokens + roleTokens + contentTokens, nil
}

// CountMessagesTokens 计算多条消息的总tokens
func CountMessagesTokens(counter TokenCounter, messages []map[string]string) (int, error) {
	// 1. 消息数组的基础格式tokens
	totalTokens := 2 // [] 的tokens

	// 2. 遍历所有消息
	for i, msg := range messages {
		// 获取角色和内容
		role := msg["role"]
		content := msg["content"]

		// 计算当前消息的tokens
		msgTokens, err := CountMessageTokens(counter, role, content)
		if err != nil {
			return 0, err
		}
		totalTokens += msgTokens

		// 如果不是最后一条消息，添加逗号的token
		if i < len(messages)-1 {
			totalTokens++
		}
	}

	return totalTokens, nil
}

// CountSystemPromptTokens 计算系统提示词的tokens
func CountSystemPromptTokens(counter TokenCounter, systemPrompt string) (int, error) {
	if systemPrompt == "" {
		return 0, nil
	}

	// 系统提示词的完整格式：{"role": "system", "content": "xxx"}
	return CountMessageTokens(counter, "system", systemPrompt)
}

// CountRequestTokens 计算完整请求的tokens
func CountRequestTokens(counter TokenCounter, systemPrompt string, messages []map[string]string) (int, error) {
	// 1. 计算系统提示词tokens
	systemTokens, err := CountSystemPromptTokens(counter, systemPrompt)
	if err != nil {
		return 0, err
	}

	// 2. 计算所有消息tokens
	messageTokens, err := CountMessagesTokens(counter, messages)
	if err != nil {
		return 0, err
	}

	// 3. 计算请求格式的基础tokens
	// 完整请求格式：{"model": "xxx", "messages": [...]}
	baseTokens := 3 // {}, messages 字段名

	return baseTokens + systemTokens + messageTokens, nil
}
