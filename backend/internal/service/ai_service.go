package service

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	m "AI_Proxy_Go/backend/internal/model"
	"AI_Proxy_Go/backend/internal/service/search"

	"gorm.io/gorm"
)

// AIService AI服务接口
type AIService struct {
	DB            *gorm.DB
	searchService search.SearchService
	logger        *log.Logger
}

// NewAIService 创建新的AI服务实例
func NewAIService(db *gorm.DB) *AIService {
	logger := log.New(os.Stdout, "[AIService] ", log.LstdFlags)
	logger.Printf("正在初始化 AIService...")

	searchEngine := search.NewSearchEngine(db)
	if searchEngine == nil {
		logger.Printf("搜索引擎初始化失败")
		return nil
	}

	service := &AIService{
		DB:            db,
		searchService: searchEngine,
		logger:        logger,
	}

	logger.Printf("AIService 初始化完成")
	return service
}

// ChatRequest 对话请求结构
type ChatRequest struct {
	SessionID  uint
	ProviderID int64
	ModelID    int64
	KeyID      int64
	Content    string
	UserID     int64
	Parameters map[string]interface{} `json:"parameters"`
}

// ChatResponse 对话响应结构
type ChatResponse struct {
	Content          string
	PromptTokens     int64
	CompletionTokens int64
	TotalTokens      int64
}

// buildRequestBody 根据不同服务商构建请求体
func (s *AIService) buildRequestBody(provider *m.AI_Provider, model *m.AI_Model, content string, userID int64, userParams map[string]interface{}) ([]byte, error) {
	// 解析模型默认参数
	params := make(map[string]interface{})
	if model.Parameters != "" {
		if err := json.Unmarshal([]byte(model.Parameters), &params); err != nil {
			return nil, fmt.Errorf("解析模型参数失败: %v", err)
		}
	}

	// 合并用户设置的参数（用户参数优先级更高）
	for k, v := range userParams {
		params[k] = v
	}

	var messages []m.ChatMessage
	// 检查是否启用了上下文
	useContext := false
	if useContextVal, ok := userParams["use_context"]; ok {
		if boolVal, ok := useContextVal.(bool); ok {
			useContext = boolVal
		}
	}

	// 只有在启用上下文时才获取历史消息
	if useContext {
		sessionID, ok := userParams["session_id"].(float64)
		if !ok {
			s.logger.Printf("未找到会话ID，跳过获取历史消息")
			return nil, fmt.Errorf("未找到会话ID")
		}

		// 获取上下文长度设置
		contextLength := 3 // 默认值
		var setting m.SystemSettings
		if err := s.DB.Where("config_key = ?", "session.context_length").First(&setting).Error; err == nil {
			if length, err := strconv.Atoi(setting.Value); err == nil {
				contextLength = length
			}
		}

		// 使用配置的上下文长度获取历史消息
		if err := s.DB.Where("session_id = ? AND user_id = ?", uint(sessionID), userID).
			Order("created_at desc").
			Limit(contextLength).
			Find(&messages).Error; err != nil {
			return nil, fmt.Errorf("获取历史消息失败: %v", err)
		}

		// 反转消息顺序，使其按时间正序排列
		for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
			messages[i], messages[j] = messages[j], messages[i]
		}

		//s.logger.Printf("已获取 %d 条历史消息作为上下文", len(messages))
	} else {
		//s.logger.Printf("未启用上下文，跳过获取历史消息")
	}

	// 构建消息历史（只有在启用上下文时才会有历史消息）
	var historyMessages []map[string]string
	var lastRole string
	for _, msg := range messages {
		// 如果当前消息与上一条消息的角色相同，跳过
		if msg.Role == lastRole {
			continue
		}
		historyMessages = append(historyMessages, map[string]string{
			"role":    msg.Role,
			"content": msg.Content,
		})
		lastRole = msg.Role
	}

	// 如果最后一条历史消息与当前要发送的消息角色相同，移除最后一条历史消息
	if len(historyMessages) > 0 && historyMessages[len(historyMessages)-1]["role"] == "user" {
		historyMessages = historyMessages[:len(historyMessages)-1]
	}

	// 添加当前消息
	historyMessages = append(historyMessages, map[string]string{
		"role":    "user",
		"content": content,
	})

	switch provider.Type {
	case "OPENAI", "OPENAI_COMPATIBLE":
		// 如果是 deepseek-reasoner 模型，使用特殊格式
		if model.ModelID == "deepseek-reasoner" {
			// 确保消息格式正确
			var messages []map[string]string
			if len(historyMessages) > 0 {
				messages = historyMessages
			} else {
				messages = []map[string]string{
					{
						"role":    "user",
						"content": content,
					},
				}
			}

			requestBody := map[string]interface{}{
				"model":      "deepseek-reasoner", // 固定使用这个模型ID
				"messages":   messages,
				"max_tokens": 4096,
			}
			return json.Marshal(requestBody)
		}

		// 其他 OpenAI 兼容模型使用标准格式
		requestBody := map[string]interface{}{
			"model":    model.ModelID,
			"messages": historyMessages,
		}
		// 添加模型参数
		for k, v := range params {
			requestBody[k] = v
		}
		return json.Marshal(requestBody)

	case "ANTHROPIC":
		// Anthropic 需要特殊处理历史消息
		var prompt string
		for _, msg := range historyMessages {
			if msg["role"] == "user" {
				prompt += "\n\nHuman: " + msg["content"]
			} else if msg["role"] == "assistant" {
				prompt += "\n\nAssistant: " + msg["content"]
			}
		}

		requestBody := map[string]interface{}{
			"model":                model.ModelID,
			"prompt":               prompt,
			"max_tokens_to_sample": model.MaxTokens,
		}
		for k, v := range params {
			requestBody[k] = v
		}
		return json.Marshal(requestBody)

	case "GoogleGemini":
		return json.Marshal(map[string]interface{}{
			"contents": []map[string]interface{}{
				{
					"role": "user",
					"parts": []map[string]interface{}{
						{"text": content},
					},
				},
			},
			"generationConfig": params, // 添加生成配置
		})

	case "BAIDU":
		requestBody := map[string]interface{}{
			"messages": []map[string]string{
				{"role": "user", "content": content},
			},
		}
		// 添加模型参数
		for k, v := range params {
			requestBody[k] = v
		}
		return json.Marshal(requestBody)

	default:
		return nil, fmt.Errorf("不支持的服务商类型: %s", provider.Type)
	}
}

// buildHeaders 构建请求头
func (s *AIService) buildHeaders(provider *m.AI_Provider, key *m.AI_APIKey) (map[string]string, error) {
	headers := make(map[string]string)

	// 解析服务商配置的请求头
	if provider.Headers != "" {
		if err := json.Unmarshal([]byte(provider.Headers), &headers); err != nil {
			return nil, fmt.Errorf("解析请求头失败: %v", err)
		}
	}

	// 根据服务商类型设置认证头
	switch provider.Type {
	case "OPENAI", "OPENAI_COMPATIBLE":
		headers["Authorization"] = "Bearer " + key.KeyValue
	case "ANTHROPIC":
		headers["x-api-key"] = key.KeyValue
	case "GoogleGemini":
		headers["x-goog-api-key"] = key.KeyValue
	case "BAIDU":
		headers["X-Bce-Authorization"] = key.KeyValue
	}

	headers["Content-Type"] = "application/json"
	return headers, nil
}

// formatSearchResultForPrompt 将搜索结果格式化为提示词
func (s *AIService) formatSearchResultForPrompt(result *m.SearchResult) string {
	var prompt strings.Builder

	// 添加搜索结果摘要
	prompt.WriteString("我找到了以下相关信息，按照时间顺序和相关度排序：\n\n")

	// 按时间分组
	today := time.Now()
	var (
		todayRefs []m.Result
		weekRefs  []m.Result
		monthRefs []m.Result
		olderRefs []m.Result
	)

	for _, ref := range result.Results {
		age := today.Sub(ref.PublishTime)
		switch {
		case age < 24*time.Hour:
			todayRefs = append(todayRefs, ref)
		case age < 7*24*time.Hour:
			weekRefs = append(weekRefs, ref)
		case age < 30*24*time.Hour:
			monthRefs = append(monthRefs, ref)
		default:
			olderRefs = append(olderRefs, ref)
		}
	}

	// 添加时间分组信息
	if len(todayRefs) > 0 {
		prompt.WriteString("24小时内的信息：\n")
		for _, ref := range todayRefs {
			prompt.WriteString(formatReference(ref))
		}
		prompt.WriteString("\n")
	}

	if len(weekRefs) > 0 {
		prompt.WriteString("最近一周的信息：\n")
		for _, ref := range weekRefs {
			prompt.WriteString(formatReference(ref))
		}
		prompt.WriteString("\n")
	}

	if len(monthRefs) > 0 {
		prompt.WriteString("最近一月的信息：\n")
		for _, ref := range monthRefs {
			prompt.WriteString(formatReference(ref))
		}
		prompt.WriteString("\n")
	}

	if len(olderRefs) > 0 {
		prompt.WriteString("更早的相关信息：\n")
		for _, ref := range olderRefs {
			prompt.WriteString(formatReference(ref))
		}
		prompt.WriteString("\n")
	}

	// 添加引用指南
	prompt.WriteString("\n请在回答时：\n")
	prompt.WriteString("1. 优先使用最新的信息\n")
	prompt.WriteString("2. 通过[数字]引用信息来源\n")
	prompt.WriteString("3. 如果信息有冲突，请说明并分析原因\n")
	prompt.WriteString("4. 如果信息不足，请明确指出\n\n")
	prompt.WriteString("根据以上信息，")

	return prompt.String()
}

// formatReference 格式化单个引用
func formatReference(ref m.Result) string {
	// 使用 Positions[0] 作为引用编号，如果没有则使用 0
	id := 0
	if len(ref.Positions) > 0 {
		id = ref.Positions[0]
	}

	return fmt.Sprintf("[%d] %s\n%s\n来源：%s\n\n",
		id,
		ref.Title,
		ref.Content,
		ref.URL)
}

// SendChatMessage 发送对话消息
func (s *AIService) SendChatMessage(req *ChatRequest) (interface{}, error) {
	//s.logger.Printf("开始处理聊天消息请求")

	// 获取模型信息
	var model m.AI_Model
	if err := s.DB.First(&model, req.ModelID).Error; err != nil {
		s.logger.Printf("获取模型信息失败: %v", err)
		return "", fmt.Errorf("获取模型信息失败: %v", err)
	}

	// 解析模型参数，检查是否使用流式传输
	var params map[string]interface{}
	if err := json.Unmarshal([]byte(model.Parameters), &params); err != nil {
		s.logger.Printf("解析模型参数失败: %v", err)
		return "", fmt.Errorf("解析模型参数失败: %v", err)
	}

	// 检查是否使用流式传输
	useStream := false
	if streamParam, ok := params["stream"].(bool); ok {
		useStream = streamParam
	}
	// 用户参数可以覆盖模型参数
	if streamParam, ok := req.Parameters["stream"].(bool); ok {
		useStream = streamParam
	}

	//s.logger.Printf("使用流式传输: %v", useStream)

	// 检查是否启用联网搜索
	if enableInternet, ok := req.Parameters["enable_internet"].(bool); ok && enableInternet {
		//s.logger.Printf("联网搜索已启用，开始搜索: %s", req.Content)

		// 尝试搜索
		searchSuccess := false
		var newContent string

		// 进行搜索尝试
		if err := func() error {
			// 处理搜索查询
			query, err := s.searchService.ProcessQuery(req.Content)
			if err != nil {
				s.logger.Printf("处理搜索查询失败: %v", err)
				return err
			}

			// 执行搜索
			searchResult, err := s.searchService.Search(query)
			if err != nil {
				s.logger.Printf("搜索失败: %v", err)
				return err
			}

			// 过滤结果
			filteredResult, err := s.searchService.FilterResults(searchResult)
			if err != nil {
				s.logger.Printf("过滤搜索结果失败: %v", err)
				return err
			}

			// 将搜索结果添加到提示词
			searchPrompt := s.formatSearchResultForPrompt(filteredResult)
			newContent = searchPrompt + "请回答以下问题：" + req.Content
			searchSuccess = true
			return nil
		}(); err != nil {
			s.logger.Printf("搜索处理失败，将使用原始内容继续处理: %v", err)
		}

		// 如果搜索成功，使用增强的内容
		if searchSuccess {
			req.Content = newContent
			//s.logger.Printf("已添加搜索结果到提示词")
		} else {
			s.logger.Printf("搜索失败，使用原始内容继续处理")
		}
	}

	// 根据模式选择处理方式
	if useStream {
		//s.logger.Printf("使用流式传输模式处理请求")
		return s.sendStreamChatMessage(req, &model)
	}
	//s.logger.Printf("使用普通模式处理请求")
	return s.sendNormalChatMessage(req, &model)
}

// sendNormalChatMessage 发送普通对话消息
func (s *AIService) sendNormalChatMessage(req *ChatRequest, model *m.AI_Model) (interface{}, error) {
	// 获取服务商信息
	var provider m.AI_Provider
	if err := s.DB.First(&provider, req.ProviderID).Error; err != nil {
		return "", fmt.Errorf("获取服务商信息失败: %v", err)
	}

	// 获取密钥信息
	var key m.AI_APIKey
	if err := s.DB.First(&key, req.KeyID).Error; err != nil {
		return "", fmt.Errorf("获取密钥信息失败: %v", err)
	}

	// 构建请求头
	headers, err := s.buildHeaders(&provider, &key)
	if err != nil {
		return "", err
	}
	//fmt.Printf("请求头: %+v\n", headers)

	// 构建请求体
	if req.Parameters == nil {
		req.Parameters = make(map[string]interface{})
	}
	req.Parameters["session_id"] = float64(req.SessionID)
	body, err := s.buildRequestBody(&provider, model, req.Content, req.UserID, req.Parameters)
	if err != nil {
		return "", err
	}
	//fmt.Printf("【ai_service】请求体: %s\n", string(body))

	// 获取超时时间设置
	timeoutStr := "240" // 默认值
	var setting m.SystemSettings
	if err := s.DB.Where("config_key = ?", "ai.request_timeout").First(&setting).Error; err == nil {
		timeoutStr = setting.Value
	}
	timeout, _ := strconv.Atoi(timeoutStr)

	// 创建带超时的客户端
	client := &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}

	// 创建 token 计数器和变量
	tokenCounter := GetTokenCounter(provider.Type, model.ModelID)
	s.logger.Printf("创建 token 计数器: %s", provider.Type)

	var requestTokens int  // 请求的 tokens
	var responseTokens int // 响应的 tokens
	var totalTokens int    // 总 tokens

	// 解析请求体内容并计算请求 tokens
	var requestBody map[string]interface{}
	if err := json.NewDecoder(bytes.NewBuffer(body)).Decode(&requestBody); err != nil {
		s.logger.Printf("解析请求体失败: %v", err)
	} else {
		if messages, ok := requestBody["messages"].([]interface{}); ok {
			for _, msg := range messages {
				if msgMap, ok := msg.(map[string]interface{}); ok {
					if content, ok := msgMap["content"].(string); ok {
						tokens, err := tokenCounter.CountTokens(content)
						if err != nil {
							s.logger.Printf("计算消息普通响应 tokens 失败: %v", err)
							continue
						}
						requestTokens += tokens
					}
				}
			}
			s.logger.Printf("请求普通响应 tokens 数量: %d", requestTokens)
		}
	}

	// 创建HTTP请求
	httpReq, err := http.NewRequest("POST", provider.BaseURL, bytes.NewBuffer(body))
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %v", err)
	}
	//fmt.Printf("请求URL: %s\n", provider.BaseURL)

	// 设置请求头
	for k, v := range headers {
		httpReq.Header.Set(k, v)
	}

	// 发送请求
	//fmt.Printf("开始发送请求...\n")
	resp, err := client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("发送请求失败: %v", err)
	}
	defer resp.Body.Close()
	//fmt.Printf("收到响应状态码: %d\n", resp.StatusCode)

	// 读取响应体
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %v", err)
	}
	//fmt.Printf("响应内容: %s\n", string(respBody))

	// 解析响应
	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("解析响应失败: %v", err)
	}

	// 尝试获取 usage 信息
	if usage, ok := result["usage"].(map[string]interface{}); ok {
		// 检查是否有 properties 层
		if properties, ok := usage["properties"].(map[string]interface{}); ok {
			if pt, ok := properties["prompt_tokens"].(float64); ok {
				requestTokens = int(pt)
			}
			if ct, ok := properties["completion_tokens"].(float64); ok {
				responseTokens = int(ct)
			}
			if tt, ok := properties["total_tokens"].(float64); ok {
				totalTokens = int(tt)
			}
		} else {
			// 直接从 usage 中获取数据
			if pt, ok := usage["prompt_tokens"].(float64); ok {
				requestTokens = int(pt)
			}
			if ct, ok := usage["completion_tokens"].(float64); ok {
				responseTokens = int(ct)
			}
			if tt, ok := usage["total_tokens"].(float64); ok {
				totalTokens = int(tt)
			}
		}
	}

	// 根据服务商类型解析响应内容
	var content string
	switch provider.Type {
	case "OPENAI", "OPENAI_COMPATIBLE":
		if choices, ok := result["choices"].([]interface{}); ok && len(choices) > 0 {
			if choice, ok := choices[0].(map[string]interface{}); ok {
				// DeepSeek 的响应格式可能直接包含 content
				if content, ok := choice["content"].(string); ok {
					return content, nil
				}
				// 标准 OpenAI 格式
				if message, ok := choice["message"].(map[string]interface{}); ok {
					// 检查是否有 reasoning_content
					if reasoningContent, hasReasoning := message["reasoning_content"]; hasReasoning {
						content = fmt.Sprintf("思维过程：\n%s\n\n最终回答：\n%s",
							reasoningContent.(string),
							message["content"].(string))
					} else {
						content = message["content"].(string)
					}
				}
			}
		}

		// 如果上面的解析都失败了，尝试直接从错误信息中获取
		if errMsg, ok := result["message"].(string); ok && content == "" {
			return "", fmt.Errorf("API 错误: %s", errMsg)
		}
	case "ANTHROPIC":
		if completion, ok := result["completion"].(string); ok {
			content = completion
		}
	case "GoogleGemini":
		if choices, ok := result["choices"].([]interface{}); ok && len(choices) > 0 {
			if message, ok := choices[0].(map[string]interface{})["message"].(map[string]interface{}); ok {
				content = message["content"].(string)
			}
		}
	case "BAIDU":
		if choices, ok := result["choices"].([]interface{}); ok && len(choices) > 0 {
			if message, ok := choices[0].(map[string]interface{})["message"].(map[string]interface{}); ok {
				content = message["content"].(string)
			}
		}
	}

	if content == "" {
		return "", fmt.Errorf("无法获取响应内容，原始响应：%s", string(respBody))
	}

	// 如果没有从 usage 获取到 tokens 信息，使用我们自己的计算
	if totalTokens == 0 {
		// 计算响应 tokens
		tokens, err := tokenCounter.CountTokens(content)
		if err != nil {
			s.logger.Printf("计算普通响应 tokens 失败: %v", err)
		} else {
			responseTokens = tokens
			totalTokens = requestTokens + responseTokens
		}
	}

	// 输出最终的 tokens 统计
	s.logger.Printf("最终普通响应 tokens 统计 - 请求: %d, 响应: %d, 总计: %d",
		requestTokens, responseTokens, totalTokens)

	// 保存完整响应，包含 tokens 信息
	assistantMessage := m.ChatMessage{
		SessionID:        req.SessionID,
		UserID:           req.UserID,
		ProviderID:       req.ProviderID,
		ModelID:          req.ModelID,
		KeyID:            req.KeyID,
		Role:             "assistant",
		Content:          content,
		PromptTokens:     int64(requestTokens),
		CompletionTokens: int64(responseTokens),
		TotalTokens:      int64(totalTokens),
	}

	if err := s.DB.Create(&assistantMessage).Error; err != nil {
		s.logger.Printf("保存AI普通响应失败: %v", err)
	}

	// 添加一个新的方法用于更新 tokens 统计
	if err := s.updateTokensStatistics(req, requestTokens, responseTokens, totalTokens); err != nil {
		s.logger.Printf("更新 tokens 统计失败: %v", err)
	}

	// 返回包含 tokens 信息的响应
	return &ChatResponse{
		Content:          content,
		PromptTokens:     int64(requestTokens),
		CompletionTokens: int64(responseTokens),
		TotalTokens:      int64(totalTokens),
	}, nil
}

// StreamResponse 流式响应结构
type StreamResponse struct {
	Stream           chan string
	Done             chan bool
	MessageID        uint  // 保存消息ID
	PromptTokens     int64 // 添加 tokens 统计
	CompletionTokens int64
	TotalTokens      int64
}

// sendStreamChatMessage 处理流式对话消息
func (s *AIService) sendStreamChatMessage(req *ChatRequest, model *m.AI_Model) (interface{}, error) {
	//s.logger.Printf("开始流式传输请求 - 用户ID: %d, 模型: %s", req.UserID, model.Name)

	// 获取服务商信息
	var provider m.AI_Provider
	if err := s.DB.First(&provider, req.ProviderID).Error; err != nil {
		s.logger.Printf("获取服务商信息失败: %v", err)
		return nil, fmt.Errorf("获取服务商信息失败: %v", err)
	}

	// 获取密钥信息
	var key m.AI_APIKey
	if err := s.DB.First(&key, req.KeyID).Error; err != nil {
		s.logger.Printf("获取密钥信息失败: %v", err)
		return nil, fmt.Errorf("获取密钥信息失败: %v", err)
	}

	// 构建请求头
	headers, err := s.buildHeaders(&provider, &key)
	if err != nil {
		return nil, err
	}

	// 确保参数中包含 stream=true
	if req.Parameters == nil {
		req.Parameters = make(map[string]interface{})
	}
	req.Parameters["stream"] = true

	// 构建请求体
	if req.Parameters == nil {
		req.Parameters = make(map[string]interface{})
	}
	req.Parameters["session_id"] = float64(req.SessionID)
	body, err := s.buildRequestBody(&provider, model, req.Content, req.UserID, req.Parameters)
	if err != nil {
		return nil, err
	}

	// 创建HTTP请求
	httpReq, err := http.NewRequest("POST", provider.BaseURL, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	// 设置请求头
	for k, v := range headers {
		httpReq.Header.Set(k, v)
	}

	// 设置流式传输相关的header
	httpReq.Header.Set("Accept", "text/event-stream")
	httpReq.Header.Set("Cache-Control", "no-cache")
	httpReq.Header.Set("Connection", "keep-alive")

	// 创建响应通道
	stream := make(chan string, 100)
	done := make(chan bool)

	// 创建 StreamResponse 实例，返回指针
	streamResp := &StreamResponse{
		Stream: stream,
		Done:   done,
	}

	// 启动goroutine处理响应
	go func() {
		defer func() {
			//s.logger.Printf("流式处理完成，关闭通道")
			close(stream)
			close(done)
		}()

		// 获取超时时间设置
		timeoutStr := "240" // 默认值
		var setting m.SystemSettings
		if err := s.DB.Where("config_key = ?", "ai.request_timeout").First(&setting).Error; err == nil {
			timeoutStr = setting.Value
		}
		timeout, _ := strconv.Atoi(timeoutStr)

		// 创建带超时的客户端
		client := &http.Client{
			Timeout: time.Duration(timeout) * time.Second,
		}

		// 构建请求体
		if req.Parameters == nil {
			req.Parameters = make(map[string]interface{})
		}
		req.Parameters["session_id"] = float64(req.SessionID)
		body, err := s.buildRequestBody(&provider, model, req.Content, req.UserID, req.Parameters)
		if err != nil {
			s.logger.Printf("构建请求体失败: %v", err)
			return
		}

		// 创建 token 计数器
		tokenCounter := GetTokenCounter(provider.Type, model.ModelID)
		s.logger.Printf("创建 token 计数器: %s", provider.Type)

		var requestTokens int  // 请求的 tokens
		var responseTokens int // 响应的 tokens
		var totalTokens int    // 总 tokens

		// 解析请求体内容
		var requestBody map[string]interface{}
		if err := json.NewDecoder(bytes.NewBuffer(body)).Decode(&requestBody); err != nil {
			s.logger.Printf("解析请求体失败: %v", err)
		} else {
			// 计算请求的 tokens
			if messages, ok := requestBody["messages"].([]interface{}); ok {
				for _, msg := range messages {
					if msgMap, ok := msg.(map[string]interface{}); ok {
						if content, ok := msgMap["content"].(string); ok {
							tokens, err := tokenCounter.CountTokens(content)
							if err != nil {
								s.logger.Printf("计算消息流式响应 tokens 失败: %v", err)
								continue
							}
							requestTokens += tokens
						}
					}
				}
				s.logger.Printf("流式响应请求 tokens 数量: %d", requestTokens)
			}
		}

		// 创建HTTP请求
		httpReq, err := http.NewRequest("POST", provider.BaseURL, bytes.NewBuffer(body))
		if err != nil {
			s.logger.Printf("创建流式响应请求失败: %v", err)
			return
		}

		// 设置请求头
		for k, v := range headers {
			httpReq.Header.Set(k, v)
		}

		resp, err := client.Do(httpReq)
		if err != nil {
			//s.logger.Printf("发送请求失败: %v", err)
			return
		}
		defer resp.Body.Close()

		//s.logger.Printf("收到响应，状态码: %d", resp.StatusCode)
		if resp.StatusCode != http.StatusOK {
			//body, _ := io.ReadAll(resp.Body)
			//s.logger.Printf("服务器返回错误: %d, body: %s", resp.StatusCode, string(body))
			return
		}

		reader := bufio.NewReader(resp.Body)
		var fullResponse strings.Builder

		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				if err != io.EOF {
					//s.logger.Printf("读取响应出错: %v", err)
				} else {
					//s.logger.Printf("读取响应完成(EOF)")
				}
				break
			}

			line = strings.TrimSpace(line)
			//s.logger.Printf("收到原始数据: %s", line)

			if line == "" {
				continue
			}

			if line == "data: [DONE]" {
				//s.logger.Printf("收到结束标记")
				break
			}

			if strings.HasPrefix(line, "data: ") {
				data := strings.TrimPrefix(line, "data: ")
				var result map[string]interface{}
				if err := json.Unmarshal([]byte(data), &result); err != nil {
					s.logger.Printf("解析JSON数据失败: %v", err)
					continue
				}

				// 提取内容并计算 tokens
				content := s.extractStreamContent(data, provider.Type)
				if content != "" {
					stream <- content
					fullResponse.WriteString(content)

					// 计算响应 tokens
					tokens, err := tokenCounter.CountTokens(content)
					if err != nil {
						s.logger.Printf("计算流式响应 tokens 失败: %v", err)
					} else {
						responseTokens += tokens // 使用 += 而不是赋值
						totalTokens = requestTokens + responseTokens

						// 输出当前的 tokens 统计
						//s.logger.Printf("当前 Token 统计 - 请求: %d, 响应: %d, 总计: %d",
						//	requestTokens, responseTokens, totalTokens)
					}
				}

				// 尝试获取 usage 信息
				if usage, ok := result["usage"].(map[string]interface{}); ok {
					// 检查是否有 properties 层
					if properties, ok := usage["properties"].(map[string]interface{}); ok {
						if pt, ok := properties["prompt_tokens"].(float64); ok {
							requestTokens = int(pt)
						}
						if ct, ok := properties["completion_tokens"].(float64); ok {
							responseTokens = int(ct)
						}
						if tt, ok := properties["total_tokens"].(float64); ok {
							totalTokens = int(tt)
						}
					} else {
						// 直接从 usage 中获取数据
						if pt, ok := usage["prompt_tokens"].(float64); ok {
							requestTokens = int(pt)
						}
						if ct, ok := usage["completion_tokens"].(float64); ok {
							responseTokens = int(ct)
						}
						if tt, ok := usage["total_tokens"].(float64); ok {
							totalTokens = int(tt)
						}
					}
				}
			}
		}

		// 如果服务器没有返回 total_tokens，自己计算
		if totalTokens == 0 {
			totalTokens = requestTokens + responseTokens
		}

		// 输出最终的 tokens 统计
		s.logger.Printf("最终流式响应 tokens 统计 - 请求: %d, 响应: %d, 总计: %d",
			requestTokens, responseTokens, totalTokens)

		// 保存完整响应，包含 tokens 信息
		assistantMessage := m.ChatMessage{
			SessionID:        req.SessionID,
			UserID:           req.UserID,
			ProviderID:       req.ProviderID,
			ModelID:          req.ModelID,
			KeyID:            req.KeyID,
			Role:             "assistant",
			Content:          fullResponse.String(),
			PromptTokens:     int64(requestTokens),
			CompletionTokens: int64(responseTokens),
			TotalTokens:      int64(totalTokens),
		}

		//s.logger.Printf("准备保存AI响应 - 消息内容: %s", assistantMessage.Content)

		if err := s.DB.Create(&assistantMessage).Error; err != nil {
			s.logger.Printf("保存AI流式响应失败: %v", err)
		} else {
			//s.logger.Printf("保存AI流式响应成功，ID: %d", assistantMessage.ID)
			// 设置消息ID和tokens信息
			streamResp.MessageID = assistantMessage.ID
			streamResp.PromptTokens = int64(requestTokens)
			streamResp.CompletionTokens = int64(responseTokens)
			streamResp.TotalTokens = int64(totalTokens)
		}

		// 添加一个新的方法用于更新 tokens 统计
		if err := s.updateTokensStatistics(req, requestTokens, responseTokens, totalTokens); err != nil {
			s.logger.Printf("更新 tokens 统计失败: %v", err)
		}

		done <- true
		//s.logger.Printf("已发送完成信号，当前 StreamResp: %+v", streamResp)
	}()
	if err := s.DB.Model(&m.ChatSession{}).
		Where("id = ?", req.SessionID).
		Update("updated_at", time.Now()).Error; err != nil {
		s.logger.Printf("更新会话时间失败: %v", err)
	}

	//s.logger.Printf("返回流式响应通道")
	return streamResp, nil
}

// extractStreamContent 从不同服务商的响应中提取内容
func (s *AIService) extractStreamContent(data string, providerType string) string {
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(data), &result); err != nil {
		return ""
	}

	switch providerType {
	case "OPENAI", "OPENAI_COMPATIBLE":
		if choices, ok := result["choices"].([]interface{}); ok && len(choices) > 0 {
			if choice, ok := choices[0].(map[string]interface{}); ok {
				if delta, ok := choice["delta"].(map[string]interface{}); ok {
					if content, ok := delta["content"].(string); ok {
						return content
					}
				}
			}
		}
	case "ANTHROPIC":
		if completion, ok := result["completion"].(string); ok {
			return completion
		}
	case "GoogleGemini", "BAIDU":
		if choices, ok := result["choices"].([]interface{}); ok && len(choices) > 0 {
			if message, ok := choices[0].(map[string]interface{})["message"].(map[string]interface{}); ok {
				if content, ok := message["content"].(string); ok {
					return content
				}
			}
		}
	}
	return ""
}

// 修改使用 References 的代码段
func (s *AIService) processSearchResults(result *m.SearchResult) {
	if result == nil || len(result.Results) == 0 {
		return
	}

	today := time.Now()

	// 遍历 Results 而不是 References
	for _, ref := range result.Results {
		age := today.Sub(ref.PublishTime)
		switch {
		case age < 24*time.Hour:
			s.logger.Printf("最近24小时的结果: %s", ref.Title)
		case age < 7*24*time.Hour:
			s.logger.Printf("最近一周的结果: %s", ref.Title)
		case age < 30*24*time.Hour:
			s.logger.Printf("最近一月的结果: %s", ref.Title)
		default:
			s.logger.Printf("较早的结果: %s", ref.Title)
		}
	}
}

// 添加一个新的方法用于更新 tokens 统计
func (s *AIService) updateTokensStatistics(req *ChatRequest, promptTokens, completionTokens, totalTokens int) error {
	// 开启事务
	tx := s.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 更新系统级别的 tokens 统计
	systemSettings := []struct {
		config_key  string
		value       int
		description string
	}{
		{m.PromptTokens, promptTokens, "提示词已使用的token总数"},
		{m.CompletionTokens, completionTokens, "回复已使用的token总数"},
		{m.TotalTokens, totalTokens, "已使用的token总数"},
	}

	for _, setting := range systemSettings {
		var existingSetting m.SystemSettings
		result := tx.Where("config_key = ?", setting.config_key).First(&existingSetting)

		if result.Error != nil {
			if result.Error == gorm.ErrRecordNotFound {
				// 记录不存在，创建新记录
				newSetting := m.SystemSettings{
					ConfigKey:   setting.config_key,
					Value:       strconv.Itoa(setting.value),
					Description: setting.description,
				}
				if err := tx.Create(&newSetting).Error; err != nil {
					tx.Rollback()
					return fmt.Errorf("创建系统 tokens 统计记录失败: %v", err)
				}
			} else {
				tx.Rollback()
				return fmt.Errorf("查询系统 tokens 统计记录失败: %v", result.Error)
			}
		} else {
			// 记录存在，更新值
			currentValue, _ := strconv.Atoi(existingSetting.Value)
			newValue := currentValue + setting.value
			if err := tx.Model(&m.SystemSettings{}).
				Where("config_key = ?", setting.config_key).
				Update("value", strconv.Itoa(newValue)).Error; err != nil {
				tx.Rollback()
				return fmt.Errorf("更新系统 tokens 统计记录失败: %v", err)
			}
		}
	}

	// 更新服务商统计
	if err := tx.Model(&m.AI_Provider{}).
		Where("id = ?", req.ProviderID).
		Updates(map[string]interface{}{
			"prompt_tokens":     gorm.Expr("prompt_tokens + ?", promptTokens),
			"completion_tokens": gorm.Expr("completion_tokens + ?", completionTokens),
			"total_tokens":      gorm.Expr("total_tokens + ?", totalTokens),
		}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("更新服务商 tokens 统计失败: %v", err)
	}

	// 更新AI模型统计
	if err := tx.Model(&m.AI_Model{}).
		Where("id = ?", req.ModelID).
		Updates(map[string]interface{}{
			"prompt_tokens":     gorm.Expr("prompt_tokens + ?", promptTokens),
			"completion_tokens": gorm.Expr("completion_tokens + ?", completionTokens),
			"total_tokens":      gorm.Expr("total_tokens + ?", totalTokens),
		}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("更新AI模型 tokens 统计失败: %v", err)
	}

	// 更新API密钥统计
	if err := tx.Model(&m.AI_APIKey{}).
		Where("id = ?", req.KeyID).
		Updates(map[string]interface{}{
			"prompt_tokens":     gorm.Expr("prompt_tokens + ?", promptTokens),
			"completion_tokens": gorm.Expr("completion_tokens + ?", completionTokens),
			"total_tokens":      gorm.Expr("total_tokens + ?", totalTokens),
		}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("更新API密钥 tokens 统计失败: %v", err)
	}

	// 更新用户统计
	if err := tx.Model(&m.User{}).
		Where("id = ?", req.UserID).
		Updates(map[string]interface{}{
			"prompt_tokens":     gorm.Expr("prompt_tokens + ?", promptTokens),
			"completion_tokens": gorm.Expr("completion_tokens + ?", completionTokens),
			"total_tokens":      gorm.Expr("total_tokens + ?", totalTokens),
		}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("更新用户 tokens 统计失败: %v", err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("提交事务失败: %v", err)
	}

	return nil
}
