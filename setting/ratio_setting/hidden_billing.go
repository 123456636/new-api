package ratio_setting

import (
	"context"
	"fmt"

	"github.com/QuantumNous/new-api/dto"
	"github.com/QuantumNous/new-api/logger"
	"github.com/QuantumNous/new-api/types"
)

var hiddenInputTokenRatioMap = types.NewRWMap[string, float64]()
var hiddenCacheTokenRatioMap = types.NewRWMap[string, float64]()

func GetHiddenInputTokenRatio(modelName string) float64 {
	modelName = FormatMatchingModelName(modelName)
	if ratio, ok := hiddenInputTokenRatioMap.Get(modelName); ok && ratio > 0 {
		return ratio
	}
	return 1.0
}

func GetHiddenCacheTokenRatio(modelName string) float64 {
	modelName = FormatMatchingModelName(modelName)
	if ratio, ok := hiddenCacheTokenRatioMap.Get(modelName); ok && ratio > 0 {
		return ratio
	}
	return 1.0
}

// ApplyHiddenTokenMultiplier modifies usage in-place with hidden multipliers.
// Logs original values to server log. Returns true if any modification was made.
func ApplyHiddenTokenMultiplier(ctx context.Context, usage *dto.Usage, modelName string) bool {
	if usage == nil {
		return false
	}
	inputRatio := GetHiddenInputTokenRatio(modelName)
	cacheRatio := GetHiddenCacheTokenRatio(modelName)

	if inputRatio == 1.0 && cacheRatio == 1.0 {
		return false
	}

	// Log original values to server log (only visible to admin on server)
	logger.LogInfo(ctx, fmt.Sprintf(
		"[HiddenBilling] model=%s original: prompt=%d, cached=%d, cache_creation=%d | ratios: input=%.2f, cache=%.2f",
		modelName,
		usage.PromptTokens,
		usage.PromptTokensDetails.CachedTokens,
		usage.PromptTokensDetails.CachedCreationTokens,
		inputRatio, cacheRatio,
	))

	// Apply multipliers
	if inputRatio != 1.0 {
		usage.PromptTokens = int(float64(usage.PromptTokens) * inputRatio)
	}
	if cacheRatio != 1.0 {
		usage.PromptTokensDetails.CachedTokens = int(float64(usage.PromptTokensDetails.CachedTokens) * cacheRatio)
		usage.PromptTokensDetails.CachedCreationTokens = int(float64(usage.PromptTokensDetails.CachedCreationTokens) * cacheRatio)
		if usage.ClaudeCacheCreation5mTokens > 0 {
			usage.ClaudeCacheCreation5mTokens = int(float64(usage.ClaudeCacheCreation5mTokens) * cacheRatio)
		}
		if usage.ClaudeCacheCreation1hTokens > 0 {
			usage.ClaudeCacheCreation1hTokens = int(float64(usage.ClaudeCacheCreation1hTokens) * cacheRatio)
		}
	}

	// Recalculate total
	usage.TotalTokens = usage.PromptTokens + usage.CompletionTokens

	return true
}

func UpdateHiddenInputTokenRatioByJSONString(jsonStr string) error {
	return types.LoadFromJsonString(hiddenInputTokenRatioMap, jsonStr)
}

func UpdateHiddenCacheTokenRatioByJSONString(jsonStr string) error {
	return types.LoadFromJsonString(hiddenCacheTokenRatioMap, jsonStr)
}

func HiddenInputTokenRatio2JSONString() string {
	return hiddenInputTokenRatioMap.MarshalJSONString()
}

func HiddenCacheTokenRatio2JSONString() string {
	return hiddenCacheTokenRatioMap.MarshalJSONString()
}
