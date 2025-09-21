package dialects

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

func GetDbTimeZone(driver string, conn *sql.DB) (*time.Location, error) {
	switch driver {
	case "postgres", "kingbase", "kingbase8", "opengauss":
		return GetDbTimeZoneForPG(conn)
	case "mysql":
		return GetDbTimeZoneForMysql(conn)
	case "oracle", "ora":
		return GetDbTimeZoneForOracle(conn)
	case "sqlserver", "mssql":
		return GetDbTimeZoneForMssql(conn)
	}
	return nil, errors.New("driver '" + driver + "' is unsupported")
}

// GetDbTimeZoneForPG 通过查询 pg_timezone_names 获取当前数据库时区的UTC偏移量，并构造对应的 time.Location
func GetDbTimeZoneForPG(conn *sql.DB) (*time.Location, error) {
	// 首先获取当前会话的时区名称
	var currentTzName string
	err := conn.QueryRow("SHOW timezone;").Scan(&currentTzName)
	if err != nil {
		return nil, fmt.Errorf("查询当前时区名称失败: %w", err)
	}

	// 如果已经是UTC，直接返回
	if strings.ToUpper(currentTzName) == "UTC" {
		return time.UTC, nil
	}

	// 查询 pg_timezone_names 获取该时区的详细信息
	var (
		utcOffset    string
		isDst        bool
		actualTzName string // 实际在系统视图中的名称
	)

	// 注意：时区名称在 pg_timezone_names 中可能区分大小写和特殊字符（如空格）
	// 使用 ILIKE 进行不区分大小写的匹配，更可靠
	query := `
        SELECT name, utc_offset, is_dst 
        FROM pg_timezone_names 
        WHERE name ILIKE $1 OR abbrev ILIKE $1
        ORDER BY name = $1 DESC, name ILIKE $1 DESC -- 优先匹配完全相同的名称
        LIMIT 1
    `
	err = conn.QueryRow(query, currentTzName).Scan(&actualTzName, &utcOffset, &isDst)
	if err != nil {
		// 如果查询失败，尝试回退到简单映射或固定偏移量方法
		return loadLocationWithFallback(currentTzName)
	}

	// 解析UTC偏移量字符串（PostgreSQL格式: HH:MM:SS, 可能有负号）
	// 例如: 08:00:00, -05:00:00, 00:00:00
	var hours, minutes, seconds int
	negMultiplier := 1
	if strings.HasPrefix(utcOffset, "-") {
		negMultiplier = -1
		utcOffset = strings.TrimPrefix(utcOffset, "-")
	}
	_, err = fmt.Sscanf(utcOffset, "%d:%d:%d", &hours, &minutes, &seconds)
	if err != nil {
		return nil, fmt.Errorf("解析UTC偏移量 '%s' 失败: %w", utcOffset, err)
	}

	// 计算总秒数
	totalSeconds := negMultiplier * (hours*3600 + minutes*60 + seconds)

	// 使用 time.FixedZone 创建固定偏移量的时区
	// 名称使用从 pg_timezone_names 查询到的实际名称
	loc := time.FixedZone(actualTzName, totalSeconds)

	return loc, nil
}

// loadLocationWithFallback 辅助函数：尝试加载时区，失败时使用固定偏移量回退
func loadLocationWithFallback(tzName string) (*time.Location, error) {
	// 先尝试用标准名称加载（如 "Asia/Shanghai"）
	loc, err := time.LoadLocation(tzName)
	if err == nil {
		return loc, nil
	}

	// 如果失败，尝试一些常见缩写或别名的映射
	commonMappings := map[string]string{
		"CST": "Asia/Shanghai", // 中国标准时间
		"EST": "America/New_York",
		"PST": "America/Los_Angeles",
		"UTC": "UTC",
		"PRC": "Asia/Shanghai",
		// 可根据需要添加更多映射
	}
	if mappedName, ok := commonMappings[strings.ToUpper(tzName)]; ok {
		loc, err := time.LoadLocation(mappedName)
		if err == nil {
			return loc, nil
		}
	}

	// 最后手段：如果数据库返回的时区名称很特殊（如自定义偏移），我们无法从pg_timezone_names获取，
	// 可以尝试解析类似 "UTC+8" 这样的格式
	if strings.HasPrefix(strings.ToUpper(tzName), "UTC") {
		sign := 1
		offsetStr := strings.ToUpper(tzName)
		if strings.Contains(offsetStr, "UTC+") {
			offsetStr = strings.TrimPrefix(offsetStr, "UTC+")
		} else if strings.Contains(offsetStr, "UTC-") {
			sign = -1
			offsetStr = strings.TrimPrefix(offsetStr, "UTC-")
		} else {
			offsetStr = strings.TrimPrefix(offsetStr, "UTC")
		}

		var hours int
		if _, err := fmt.Sscanf(offsetStr, "%d", &hours); err == nil {
			totalSeconds := sign * hours * 3600
			return time.FixedZone(tzName, totalSeconds), nil
		}
	}

	return nil, fmt.Errorf("无法解析时区 '%s' 且无合适回退方案", tzName)
}

// GetDbTimeZoneForMysql 获取 MySQL 数据库的当前会话时区，并转换为 *time.Location
func GetDbTimeZoneForMysql(conn *sql.DB) (*time.Location, error) {
	var sessionTz, systemTz string

	// 1. 查询当前会话的时区设置
	err := conn.QueryRow("SELECT @@session.time_zone;").Scan(&sessionTz)
	if err != nil {
		return nil, fmt.Errorf("查询会话时区失败: %w", err)
	}

	// 2. 如果会话时区设置为 'SYSTEM'，则需要获取系统时区
	if sessionTz == "SYSTEM" {
		err := conn.QueryRow("SELECT @@global.system_time_zone;").Scan(&systemTz)
		if err != nil {
			return nil, fmt.Errorf("查询系统时区失败: %w", err)
		}
		// 系统时区通常是缩写（如 CST、EST），需要映射或处理
		return convertMySQLTimeZone(systemTz)
	}

	// 3. 如果会话时区是具体的偏移量或时区名（如 '+08:00', 'Asia/Shanghai'）
	return convertMySQLTimeZone(sessionTz)
}

// convertMySQLTimeZone 将 MySQL 返回的时区字符串转换为 Go 的 *time.Location
func convertMySQLTimeZone(mysqlTz string) (*time.Location, error) {
	// 处理空值默认情况
	if mysqlTz == "" {
		mysqlTz = "SYSTEM" // 或者根据你的环境设置为一个默认值，如 "+00:00"
	}

	// 情况 1: 已经是 UTC
	if mysqlTz == "UTC" || mysqlTz == "+00:00" {
		return time.UTC, nil
	}

	// 情况 2: 时区偏移量格式 (e.g., '+08:00', '-05:00')
	if strings.HasPrefix(mysqlTz, "+") || strings.HasPrefix(mysqlTz, "-") {
		// 解析偏移量字符串，格式为 ±HH:MM
		var sign string
		var hours, minutes int
		if strings.HasPrefix(mysqlTz, "+") {
			sign = "+"
		} else {
			sign = "-"
		}
		_, err := fmt.Sscanf(mysqlTz, "%s%d:%d", &sign, &hours, &minutes)
		if err != nil {
			return nil, fmt.Errorf("解析时区偏移量 '%s' 失败: %w", mysqlTz, err)
		}

		// 计算总秒数
		totalSeconds := (hours*3600 + minutes*60)
		if sign == "-" {
			totalSeconds = -totalSeconds
		}

		// 使用 FixedZone 创建时区，名称使用偏移量字符串
		return time.FixedZone(fmt.Sprintf("UTC%s", mysqlTz), totalSeconds), nil
	}

	// 情况 3: 时区名称 (e.g., 'Asia/Shanghai', 'CST')
	// 先尝试直接加载
	loc, err := time.LoadLocation(mysqlTz)
	if err == nil {
		return loc, nil
	}

	// 如果直接加载失败，尝试常见缩写映射
	commonMappings := map[string]string{
		"CST": "Asia/Shanghai", // 中国标准时间，注意 CST 也可能代表美国中部时间，存在歧义
		"EST": "America/New_York",
		"PST": "America/Los_Angeles",
		"UTC": "UTC",
		// 可根据需要添加更多映射
	}
	if mappedName, ok := commonMappings[mysqlTz]; ok {
		loc, err := time.LoadLocation(mappedName)
		if err == nil {
			return loc, nil
		}
	}

	// 情况 4: 如果以上都不行，尝试处理 SYSTEM 时区返回的常见缩写（如 CST、EST、PST 等）
	// 对于一些已知的缩写，可以尝试赋予一个固定的偏移量（存在风险，因为缩写有歧义）
	// 例如，假设我们认定 CST 就是中国标准时间 (UTC+8)
	if mysqlTz == "CST" {
		return time.FixedZone("CST", 8*3600), nil // 风险：CST 也可能指 UTC-6 或 UTC-5 (夏令时)
	}

	return nil, fmt.Errorf("无法识别或转换 MySQL 时区字符串: %s", mysqlTz)
}

// GetDbTimeZoneForOracle 获取 Oracle 数据库的当前会话时区，并转换为 *time.Location
func GetDbTimeZoneForOracle(conn *sql.DB) (*time.Location, error) {
	var tzStr string

	// 查询当前会话的时区设置 (SESSIONTIMEZONE)
	// 也可以查询 DBTIMEZONE 获取数据库时区，但通常会话时区更相关
	err := conn.QueryRow("SELECT SESSIONTIMEZONE FROM DUAL").Scan(&tzStr)
	if err != nil {
		return nil, fmt.Errorf("查询会话时区失败: %w", err)
	}

	// 处理 Oracle 返回的时区字符串
	return parseOracleTimeZone(tzStr)
}

// parseOracleTimeZone 解析 Oracle 返回的时区字符串
func parseOracleTimeZone(oracleTz string) (*time.Location, error) {
	// 处理空值或未知值
	if oracleTz == "" {
		return time.UTC, nil // 或根据你的需求返回默认时区
	}

	// 情况 1: 时区是 UTC
	if oracleTz == "UTC" || strings.ToUpper(oracleTz) == "+00:00" {
		return time.UTC, nil
	}

	// 情况 2: 时区是偏移量格式 (例如: "+08:00", "-05:00")
	if strings.Contains(oracleTz, ":") && (strings.HasPrefix(oracleTz, "+") || strings.HasPrefix(oracleTz, "-")) {
		// 解析偏移量字符串，格式为 ±HH:MI
		parts := strings.Split(oracleTz, ":")
		if len(parts) != 2 {
			return nil, fmt.Errorf("无效的时区偏移量格式: %s", oracleTz)
		}

		signStr := parts[0][:1] // "+" 或 "-"
		hoursStr := parts[0][1:]
		minutesStr := parts[1]

		hours, err := strconv.Atoi(hoursStr)
		if err != nil {
			return nil, fmt.Errorf("解析时区小时失败: %w", err)
		}

		minutes, err := strconv.Atoi(minutesStr)
		if err != nil {
			return nil, fmt.Errorf("解析时区分钟失败: %w", err)
		}

		// 计算总秒数
		totalSeconds := hours*3600 + minutes*60
		if signStr == "-" {
			totalSeconds = -totalSeconds
		}

		// 使用 FixedZone 创建固定偏移量的时区
		zoneName := fmt.Sprintf("UTC%s", oracleTz)
		return time.FixedZone(zoneName, totalSeconds), nil
	}

	// 情况 3: 时区是地区名称 (例如: "Asia/Shanghai", "America/New_York")
	// 先尝试直接加载
	loc, err := time.LoadLocation(oracleTz)
	if err == nil {
		return loc, nil
	}

	// 情况 4: 尝试常见 Oracle 时区名称映射
	commonMappings := map[string]string{
		"CST": "Asia/Shanghai", // 中国标准时间 (注意: CST 可能有歧义)
		"EST": "America/New_York",
		"PST": "America/Los_Angeles",
		"UTC": "UTC",
		// 可根据需要添加更多映射
	}
	if mappedName, ok := commonMappings[oracleTz]; ok {
		loc, err := time.LoadLocation(mappedName)
		if err == nil {
			return loc, nil
		}
	}

	// 情况 5: 如果以上方法都失败，尝试将 Oracle 时区名称转换为 IANA 格式
	// 例如: "US/Pacific" -> "America/Los_Angeles"
	// 这里可以添加特定的转换逻辑

	return nil, fmt.Errorf("无法识别或转换 Oracle 时区: %s", oracleTz)
}

// GetDbTimeZoneForMssql 获取 MSSQL 数据库的当前时区偏移量，并转换为 *time.Location
func GetDbTimeZoneForMssql(conn *sql.DB) (*time.Location, error) {
	var timezoneStr string

	// 方法1: 优先尝试使用 SQL Server 2016+ 引入的 CURRENT_TIMEZONE() 函数
	// 注意：此函数返回的是时区名称（如 "Central European Standard Time"），但并非所有版本都支持
	err := conn.QueryRow("SELECT CURRENT_TIMEZONE();").Scan(&timezoneStr)
	if err == nil && timezoneStr != "" {
		// 如果成功获取到时区名称，尝试转换为 IANA 时区或处理偏移
		return parseMSSQLTimeZone(timezoneStr)
	}

	// 方法2: 如果 CURRENT_TIMEZONE() 不可用或失败，使用 SYSDATETIMEOFFSET() 获取偏移量
	// 此方法适用于所有支持版本的 SQL Server
	var offsetStr string
	err = conn.QueryRow("SELECT FORMAT(SYSDATETIMEOFFSET(), '%H:%M');").Scan(&offsetStr)
	if err != nil {
		return nil, fmt.Errorf("查询时区偏移量失败: %w", err)
	}

	// 解析偏移量字符串，格式为 ±HH:MM
	return parseOffsetToLocation(offsetStr)
}

// parseMSSQLTimeZone 解析 MSSQL 返回的时区字符串
func parseMSSQLTimeZone(mssqlTz string) (*time.Location, error) {
	// 尝试将 MSSQL 时区名称映射到 IANA 时区
	mssqlToIANA := map[string]string{
		"UTC":                            "UTC",
		"Central European Standard Time": "Europe/Paris",
		"Eastern Standard Time":          "America/New_York",
		"Pacific Standard Time":          "America/Los_Angeles",
		"China Standard Time":            "Asia/Shanghai",
		// 可根据需要添加更多映射
	}

	if ianaName, ok := mssqlToIANA[mssqlTz]; ok {
		loc, err := time.LoadLocation(ianaName)
		if err == nil {
			return loc, nil
		}
	}

	// 如果映射失败或无法加载，尝试从名称中解析出偏移量
	// 例如，有些版本可能返回类似 "(UTC+08:00) Beijing" 的格式
	if strings.Contains(mssqlTz, "UTC") {
		// 尝试提取偏移量部分
		start := strings.Index(mssqlTz, "UTC")
		if start != -1 && start+3 < len(mssqlTz) {
			offsetPart := mssqlTz[start+3:]
			// 尝试解析偏移量
			if loc, err := parseOffsetToLocation(offsetPart); err == nil {
				return loc, nil
			}
		}
	}

	// 如果所有方法都失败，返回错误
	return nil, fmt.Errorf("无法识别或转换 MSSQL 时区: %s", mssqlTz)
}

// parseOffsetToLocation 将偏移量字符串转换为 *time.Location
func parseOffsetToLocation(offsetStr string) (*time.Location, error) {
	// 处理空值或未知值
	if offsetStr == "" {
		return time.UTC, nil
	}

	// 确保字符串以 + 或 - 开头
	if !strings.HasPrefix(offsetStr, "+") && !strings.HasPrefix(offsetStr, "-") {
		offsetStr = "+" + offsetStr // 默认假设为正偏移
	}

	// 解析偏移量字符串，格式为 ±HH:MM
	parts := strings.Split(offsetStr, ":")
	if len(parts) != 2 {
		return nil, fmt.Errorf("无效的时区偏移量格式: %s", offsetStr)
	}

	signStr := parts[0][:1] // "+" 或 "-"
	hoursStr := parts[0][1:]
	minutesStr := parts[1]

	hours, err := strconv.Atoi(hoursStr)
	if err != nil {
		return nil, fmt.Errorf("解析时区小时失败: %w", err)
	}

	minutes, err := strconv.Atoi(minutesStr)
	if err != nil {
		return nil, fmt.Errorf("解析时区分钟失败: %w", err)
	}

	// 计算总秒数
	totalSeconds := hours*3600 + minutes*60
	if signStr == "-" {
		totalSeconds = -totalSeconds
	}

	// 使用 FixedZone 创建固定偏移量的时区
	zoneName := fmt.Sprintf("UTC%s", offsetStr)
	return time.FixedZone(zoneName, totalSeconds), nil
}
