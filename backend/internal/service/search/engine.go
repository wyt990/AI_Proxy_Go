package search

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	m "AI_Proxy_Go/backend/internal/model"

	"github.com/go-redis/redis/v8"
	"github.com/yanyiwu/gojieba"
	"golang.org/x/sync/semaphore"
	"gorm.io/gorm"
)

// SearchEngine 实现 SearchService 接口
type SearchEngine struct {
	baseURL      string
	client       *http.Client
	logger       *log.Logger
	db           *gorm.DB
	redis        *redis.Client
	sem          *semaphore.Weighted // 用于并发控制
	jieba        *gojieba.Jieba      // 添加分词器
	cacheManager *CacheManager       // 添加缓存管理器
}

// NewSearchEngine 创建新的搜索引擎实例
func NewSearchEngine(db *gorm.DB) SearchService {
	engine := &SearchEngine{
		client: &http.Client{},
		logger: log.Default(),
		db:     db,
		jieba:  gojieba.NewJieba(), // 初始化分词器
	}

	// 从数据库获取配置
	var settings []m.SystemSettings
	if err := db.Where("config_key IN ?", []string{
		"search.engine_url",
		"search.concurrency",
		"redis.host",
		"redis.port",
		"redis.password",
		"redis.db",
	}).Find(&settings).Error; err == nil {
		config := make(map[string]string)
		for _, setting := range settings {
			config[setting.ConfigKey] = setting.Value
		}

		// 设置搜索引擎URL
		if url := config["search.engine_url"]; url != "" {
			engine.baseURL = url
		} else {
			engine.baseURL = "http://127.0.0.1:9999/search?format=json&q=" // 默认URL
		}

		// 设置并发控制
		concurrency, _ := strconv.ParseInt(config["search.concurrency"], 10, 64)
		if concurrency <= 0 {
			concurrency = 5 // 默认并发数
		}
		engine.sem = semaphore.NewWeighted(concurrency)

		// 初始化Redis客户端
		redisDB, _ := strconv.Atoi(config["redis.db"])
		redisPort, _ := strconv.Atoi(config["redis.port"])
		if redisPort == 0 {
			redisPort = 6379 // 默认端口
		}

		engine.redis = redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%d", config["redis.host"], redisPort),
			Password: config["redis.password"],
			DB:       redisDB,
		})

		// 测试Redis连接
		ctx := context.Background()
		if err := engine.redis.Ping(ctx).Err(); err != nil {
			engine.logger.Printf("Redis连接失败: %v", err)
		}
	} else {
		engine.logger.Printf("获取配置失败: %v", err)
		// 使用默认值
		engine.baseURL = "http://127.0.0.1:9999/search?format=json&q="
		engine.sem = semaphore.NewWeighted(5)
	}

	// 初始化缓存管理器
	if engine.redis != nil {
		engine.cacheManager = NewCacheManager(engine.redis, engine.logger, db)
	}

	return engine
}

// Search 执行搜索
func (e *SearchEngine) Search(query string) (*m.SearchResult, error) {
	// 检查速率限制
	if exceeded, waitTime := e.CheckRateLimit(); exceeded {
		return nil, &SearchError{
			Code:    ErrRateLimitExceeded,
			Message: fmt.Sprintf("请求过于频繁，请等待 %.0f 秒后重试", waitTime.Seconds()),
		}
	}

	// 检查屏蔽关键词
	if blocked, keyword := e.CheckBlockedKeywords(query); blocked {
		return nil, &SearchError{
			Code:    ErrContentFiltered,
			Message: fmt.Sprintf("搜索内容包含屏蔽关键词: %s", keyword),
		}
	}

	// 直接使用配置的URL，仅附加查询参数
	searchURL := e.baseURL + url.QueryEscape(query)

	// 创建请求
	req, err := http.NewRequest("GET", searchURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	// 设置请求头
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "AI_Proxy_Go/1.0")

	// 创建客户端
	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			DisableKeepAlives: true, // 禁用 keep-alive
		},
	}

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 记录响应状态和内容
	body, _ := io.ReadAll(resp.Body)

	// 先格式化 JSON 便于阅读
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, body, "", "  "); err == nil {
		// 直接将 JSON 字符串转换为结构体
		var jsonMap map[string]interface{}
		if err := json.Unmarshal(prettyJSON.Bytes(), &jsonMap); err == nil {
			// 重新编码为格式化的 JSON，并将 Unicode 转为中文
			//prettyResult, _ := json.MarshalIndent(jsonMap, "", "  ")
			//log.Printf("搜索响应原始内容:\n%s", string(prettyResult))
		}
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("搜索服务返回错误状态码: %d", resp.StatusCode)
	}

	// 解析响应到结构体
	var result m.SearchResult
	if err := json.NewDecoder(bytes.NewReader(body)).Decode(&result); err != nil {
		//log.Printf("解析响应失败: %v", err) // 添加错误日志
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	// 输出解析后的结果
	//resultJSON, _ := json.MarshalIndent(result, "", "  ")
	//log.Printf("解析后的搜索结果:\n%s", string(resultJSON))

	return &result, nil
}

// ProcessQuery 处理搜索查询
func (e *SearchEngine) ProcessQuery(query string) (string, error) {
	// 获取配置
	var settings []m.SystemSettings
	if err := e.db.Where("config_key IN ?", []string{
		"search.max_keywords",
		"search.cache_duration",
	}).Find(&settings).Error; err != nil {
		return "", fmt.Errorf("获取配置失败: %v", err)
	}

	config := make(map[string]int)
	for _, setting := range settings {
		value, _ := strconv.Atoi(setting.Value)
		config[setting.ConfigKey] = value
	}

	maxKeywords := config["search.max_keywords"]
	if maxKeywords == 0 {
		maxKeywords = 50
	}

	// 查询缓存
	cacheKey := fmt.Sprintf("query_process:%s", query)
	if e.redis != nil {
		if cached, err := e.redis.Get(context.Background(), cacheKey).Result(); err == nil {
			return cached, nil
		}
	}

	// 基本清理
	query = strings.TrimSpace(query)
	if query == "" {
		return "", fmt.Errorf("查询内容为空")
	}

	// 使用正则清理特殊字符
	reg := regexp.MustCompile(`[^\p{Han}\w\s]`)
	query = reg.ReplaceAllString(query, " ")

	// 使用结巴分词
	words := e.jieba.Cut(query, true)

	// 词语权重计算和过滤
	type WordWeight struct {
		word   string
		weight float64
	}
	weightedWords := make([]WordWeight, 0, len(words))

	for _, word := range words {
		if !isStopWord(word) {
			weight := calculateWordWeight(word)
			if weight > 0 {
				weightedWords = append(weightedWords, WordWeight{word, weight})
			}
		}
	}

	// 按权重排序
	sort.Slice(weightedWords, func(i, j int) bool {
		return weightedWords[i].weight > weightedWords[j].weight
	})

	// 长度控制
	totalLen := 0
	resultWords := make([]string, 0)
	for _, ww := range weightedWords {
		wordLen := utf8.RuneCountInString(ww.word)
		if totalLen+wordLen > maxKeywords {
			break
		}
		resultWords = append(resultWords, ww.word)
		totalLen += wordLen
	}

	// 至少保留一个词
	if len(resultWords) == 0 && len(weightedWords) > 0 {
		resultWords = []string{weightedWords[0].word}
	}

	// 合并结果
	result := strings.Join(resultWords, " ")

	// 缓存处理后的查询
	if e.redis != nil {
		cacheDuration := time.Duration(config["search.cache_duration"]) * time.Second
		if cacheDuration == 0 {
			cacheDuration = time.Hour
		}
		e.redis.Set(context.Background(), cacheKey, result, cacheDuration)
	}

	return result, nil
}

// 停用词列表
var stopWords = map[string]bool{
	// 中文停用词
	"的": true, "了": true, "和": true, "与": true, "或": true,
	"这": true, "那": true, "是": true, "在": true, "有": true,
	"个": true, "好": true, "来": true, "去": true, "到": true,
	"想": true, "要": true, "会": true, "对": true, "能": true,

	// 英文停用词
	"the": true, "a": true, "an": true, "and": true, "or": true,
	"in": true, "on": true, "at": true, "to": true, "for": true,
	"of": true, "with": true, "by": true, "from": true, "up": true,
	"about": true, "into": true, "over": true, "after": true,
}

// isStopWord 判断是否为停用词
func isStopWord(word string) bool {
	word = strings.ToLower(word)
	return stopWords[word]
}

// FilterResults 过滤和排序搜索结果
func (e *SearchEngine) FilterResults(results *m.SearchResult) (*m.SearchResult, error) {
	if results == nil || len(results.Results) == 0 {
		return results, nil
	}

	// 获取配置
	var settings []m.SystemSettings
	if err := e.db.Where("config_key IN ?", []string{
		"search.min_title_length",
		"search.max_summary_length",
		"search.source_weights",
		"search.time_weights",
		"search.min_relevance_score",
	}).Find(&settings).Error; err != nil {
		return nil, fmt.Errorf("获取配置失败: %v", err)
	}

	// 解析配置
	config := make(map[string]float64)
	filterDomains := make(map[string]bool)

	for _, setting := range settings {
		switch setting.ConfigKey {
		case "search.filter_domains":
			// 解析过滤域名列表
			domains := strings.Split(setting.Value, ",")
			for _, domain := range domains {
				domain = strings.TrimSpace(domain)
				if domain != "" {
					filterDomains[domain] = true
				}
			}
		default:
			value, _ := strconv.ParseFloat(setting.Value, 64)
			config[setting.ConfigKey] = value
		}
	}

	// 设置默认权重
	weights := map[string]float64{
		"official": config["search.weight.official"],
		"edu":      config["search.weight.edu"],
		"news":     config["search.weight.news"],
		"portal":   config["search.weight.portal"],
	}

	timeWeights := map[string]float64{
		"day":   config["search.time_weight.day"],
		"week":  config["search.time_weight.week"],
		"month": config["search.time_weight.month"],
		"year":  config["search.time_weight.year"],
	}

	minContentLength := int(config["search.min_content_length"])
	if minContentLength == 0 {
		minContentLength = 50 // 默认最小内容长度
	}

	maxSummaryLength := int(config["search.max_summary_length"])
	if maxSummaryLength == 0 {
		maxSummaryLength = 500 // 默认最大摘要长度
	}

	// 添加新的过滤配置
	minTitleLength := int(config["search.min_title_length"])
	if minTitleLength == 0 {
		minTitleLength = 10 // 默认最小标题长度
	}

	// 设置默认的最小相关度分数
	minRelevanceScore := 0.3 // 降低默认的最小相关度分数

	// 计算每个来源的权重并过滤
	type weightedResult struct {
		result m.Result
		weight float64
	}

	weighted := make([]weightedResult, 0)
	seen := make(map[string]bool) // 用于去重

	// 在过滤前输出原始结果
	//log.Printf("过滤前的搜索结果数量: %d", len(results.Results))

	for _, result := range results.Results {
		// 计算相关度分数
		result.RelevanceScore = calculateRelevanceScore(result, results.Query)

		// 输出每个结果的相关度分数
		//log.Printf("标题: %s, 相关度分数: %.2f", result.Title, result.RelevanceScore)

		// 相关度分数检查
		if result.RelevanceScore < minRelevanceScore {
			//log.Printf("过滤相关度较低的结果: %s (分数: %.2f)", result.Title, result.RelevanceScore)
			continue
		}

		// 生成内容指纹用于去重
		contentHash := generateContentHash(result.Title, result.Content)
		if seen[contentHash] {
			//log.Printf("过滤重复内容: %s", result.Title)
			continue
		}
		seen[contentHash] = true

		// 计算权重
		weight := calculateSourceWeight(result, weights, timeWeights)
		weighted = append(weighted, weightedResult{result, weight})
	}

	// 按权重排序
	sort.Slice(weighted, func(i, j int) bool {
		return weighted[i].weight > weighted[j].weight
	})

	// 构建结果
	filteredResult := &m.SearchResult{
		NumberOfResults: len(weighted),
		Query:           results.Query,
		Results:         make([]m.Result, len(weighted)),
		Suggestions:     results.Suggestions,
	}

	// 添加引用编号
	for i, w := range weighted {
		w.result.ID = i + 1 // 设置ID从1开始
		filteredResult.Results[i] = w.result
	}

	// 在返回前输出过滤后的结果
	//resultJSON, _ := json.MarshalIndent(filteredResult, "", "  ")
	//log.Printf("过滤后的最终搜索结果:\n%s", string(resultJSON))

	return filteredResult, nil
}

// calculateRelevanceScore 计算相关度分数
func calculateRelevanceScore(result m.Result, query string) float64 {
	// 将查询词分割成关键词
	keywords := strings.Fields(query)
	score := 0.0

	// 检查标题中的关键词匹配
	titleLower := strings.ToLower(result.Title)
	for _, keyword := range keywords {
		if strings.Contains(titleLower, strings.ToLower(keyword)) {
			score += 0.3 // 标题匹配权重更高
		}
	}

	// 检查内容中的关键词匹配
	contentLower := strings.ToLower(result.Content)
	for _, keyword := range keywords {
		if strings.Contains(contentLower, strings.ToLower(keyword)) {
			score += 0.2
		}
	}

	// 根据来源类型调整分数
	if strings.Contains(result.URL, "edu.cn") {
		score *= 1.2 // 教育网站加权
	}

	return score
}

// truncateContent 截断内容到指定长度
//
//lint:ignore U1000 This function is reserved for future use
func truncateContent(content string, maxLength int) string {
	if len(content) <= maxLength {
		return content
	}
	return content[:maxLength] + "..."
}

// generateContentHash 生成内容哈希用于去重
func generateContentHash(title, content string) string {
	// 简单的去重策略：标题+内容的前100个字符
	text := title + content
	if len(text) > 100 {
		text = text[:100]
	}
	return strings.ToLower(strings.TrimSpace(text))
}

// 在引擎关闭时释放分词器资源
func (e *SearchEngine) Close() {
	if e.jieba != nil {
		e.jieba.Free()
	}
	if e.cacheManager != nil {
		e.cacheManager.Close()
	}
}

// calculateWordWeight 计算词语权重
func calculateWordWeight(word string) float64 {
	weight := 1.0

	// 长度权重：词语越长，权重越大
	length := utf8.RuneCountInString(word)
	if length > 1 {
		weight *= float64(length) * 0.5
	}

	// 中文词权重
	if regexp.MustCompile(`\p{Han}`).MatchString(word) {
		weight *= 1.5
	}

	// 专业词权重（可以维护一个专业词库）
	if isProfessionalTerm(word) {
		weight *= 2.0
	}

	// 时效性词权重
	if isTimeSensitiveTerm(word) {
		weight *= 1.8
	}

	return weight
}

// isProfessionalTerm 判断是否为专业术语
func isProfessionalTerm(word string) bool {
	professionalTerms := map[string]bool{
		"人工智能": true,
		"机器学习": true,
		"深度学习": true,
		"区块链":  true,
		"云计算":  true,
		"大数据":  true,
		"物联网":  true,
		"网络安全": true,
		"教案":   true,
		"课件":   true,
		"试卷":   true,
		"习题":   true,
		"答案":   true,
		"解析":   true,
		"复习":   true,
		"预习":   true,
		"作业":   true,
		"测试":   true,
		"练习":   true,
		"参考":   true,
		"语文":   true,
		"数学":   true,
		"英语":   true,
		"物理":   true,
		"化学":   true,
		"生物":   true,
		"地理":   true,
		"历史":   true,
		"政治":   true,
		"音乐":   true,
		"美术":   true,
		"体育":   true,
		"信息技术": true,
		"通用技术": true,
		"小学":   true,
		"初中":   true,
		"高中":   true,
		"大学":   true,
		"研究生":  true,
		"博士":   true,
		"硕士":   true,
		"小升初":  true,
		"中考":   true,
		"高考":   true,
		"考研":   true,
		"四级":   true,
		"六级":   true,
		"托福":   true,
		// 可以添加更多专业术语
	}
	return professionalTerms[word]
}

// isTimeSensitiveTerm 判断是否为时效性词语
func isTimeSensitiveTerm(word string) bool {
	timeTerms := map[string]bool{
		"最新":     true,
		"最近":     true,
		"今天":     true,
		"本周":     true,
		"本月":     true,
		"latest": true,
		"new":    true,
		"today":  true,
		// 可以添加更多时效性词语
	}
	return timeTerms[word]
}

// GetRedisClient 获取Redis客户端
func (e *SearchEngine) GetRedisClient() *redis.Client {
	return e.redis
}

// 添加自定义错误类型
type SearchError struct {
	Code    int
	Message string
	Cause   error
}

const (
	ErrConfigNotFound    = 1001
	ErrConnectionFailed  = 1002
	ErrResponseInvalid   = 1003
	ErrCacheFailure      = 1004
	ErrRateLimitExceeded = 1005
	ErrContentFiltered   = 1006
)

func (e *SearchError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%d] %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

// CheckRateLimit 检查请求速率限制
func (e *SearchEngine) CheckRateLimit() (bool, time.Duration) {
	// 获取速率限制配置
	var rateSetting m.SystemSettings
	rateLimit := 60 // 默认每分钟60次
	if err := e.db.Where("config_key = ?", m.KeySearchRateLimit).First(&rateSetting).Error; err == nil {
		if limit, err := strconv.Atoi(rateSetting.Value); err == nil && limit > 0 {
			rateLimit = limit
		}
	}

	// 使用Redis实现速率限制
	if e.redis != nil {
		ctx := context.Background()
		now := time.Now()
		key := "search_rate_limit"

		// 获取当前分钟的请求次数
		count, err := e.redis.ZCount(ctx, key, strconv.FormatInt(now.Add(-time.Minute).Unix(), 10), strconv.FormatInt(now.Unix(), 10)).Result()
		if err != nil {
			e.logger.Printf("检查速率限制失败: %v", err)
			return false, 0
		}

		if count >= int64(rateLimit) {
			// 获取最早的请求时间
			oldest, err := e.redis.ZRange(ctx, key, 0, 0).Result()
			if err != nil || len(oldest) == 0 {
				return true, time.Minute
			}
			oldestTime, _ := strconv.ParseInt(oldest[0], 10, 64)
			return true, time.Until(time.Unix(oldestTime, 0).Add(time.Minute))
		}

		// 记录本次请求
		e.redis.ZAdd(ctx, key, &redis.Z{
			Score:  float64(now.Unix()),
			Member: strconv.FormatInt(now.Unix(), 10),
		})

		// 清理旧数据
		e.redis.ZRemRangeByScore(ctx, key, "0", strconv.FormatInt(now.Add(-time.Minute).Unix(), 10))
	}

	return false, 0
}

// checkBlockedKeywords 检查是否包含被屏蔽的关键词
func (e *SearchEngine) CheckBlockedKeywords(query string) (blocked bool, keyword string) {
	// 获取屏蔽关键词配置
	var keywordSetting m.SystemSettings
	if err := e.db.Where("config_key = ?", m.KeySearchBlockedKeywords).First(&keywordSetting).Error; err != nil {
		return false, ""
	}

	// 分割关键词列表
	blockedKeywords := strings.Split(keywordSetting.Value, ",")
	for _, kw := range blockedKeywords {
		kw = strings.TrimSpace(kw)
		if kw != "" && strings.Contains(strings.ToLower(query), strings.ToLower(kw)) {
			return true, kw
		}
	}

	return false, ""
}

// calculateSourceWeight 计算来源的综合权重
func calculateSourceWeight(source m.Result, weights map[string]float64, timeWeights map[string]float64) float64 {
	weight := weights[source.Type] // 来源类型权重

	// 计算时效性权重
	age := time.Since(source.PublishTime)
	switch {
	case age <= 24*time.Hour:
		weight *= timeWeights["day"]
	case age <= 7*24*time.Hour:
		weight *= timeWeights["week"]
	case age <= 30*24*time.Hour:
		weight *= timeWeights["month"]
	case age <= 365*24*time.Hour:
		weight *= timeWeights["year"]
	default:
		weight *= 0.5 // 超过一年的内容权重降低
	}

	// 相关度分数影响权重
	weight *= (1 + source.RelevanceScore)

	return weight
}
