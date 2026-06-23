package service

import (
	"encoding/json"
	"sync"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/logger"
)

// N1nGroupPrice holds upstream prices for one model+group combination.
type N1nGroupPrice struct {
	InputPerM      float64 `json:"input"`
	OutputPerM     float64 `json:"output"`
	CacheReadPerM  float64 `json:"cache_read,omitempty"`
}

// n1nGroupPricingMap: model_id -> group_name -> price
// Populated once at startup and refreshed via UpdateN1nGroupPricing.
var (
	n1nGroupPricingMu  sync.RWMutex
	n1nGroupPricingMap map[string]map[string]N1nGroupPrice
)

const n1nGroupPricingOptionKey = "N1nGroupPricing"

// LoadN1nGroupPricing loads the pricing table from the options store into memory.
// Call this during service initialisation alongside other option loads.
func LoadN1nGroupPricing() {
	raw := common.OptionMap[n1nGroupPricingOptionKey]
	if raw == "" {
		setDefaultN1nGroupPricing()
		return
	}
	var m map[string]map[string]N1nGroupPrice
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		logger.SysError("LoadN1nGroupPricing: failed to parse option: " + err.Error())
		setDefaultN1nGroupPricing()
		return
	}
	n1nGroupPricingMu.Lock()
	n1nGroupPricingMap = m
	n1nGroupPricingMu.Unlock()
}

// setDefaultN1nGroupPricing seeds the in-memory map with the prices confirmed
// by real n1n API tests on 2026-06-23. These are upstream costs in USD/1M tokens.
// sale price = upstream * 1.25 is applied at quota-calculation time via ModelRatio.
func setDefaultN1nGroupPricing() {
	m := map[string]map[string]N1nGroupPrice{
		"gpt-4o-mini": {
			"default":        {InputPerM: 0.15, OutputPerM: 0.60, CacheReadPerM: 0.075},
			"官转":            {InputPerM: 0.45, OutputPerM: 1.80, CacheReadPerM: 0.225},
			"官转OpenAI":      {InputPerM: 0.90, OutputPerM: 3.60, CacheReadPerM: 0.450},
			"优质官转OpenAI":   {InputPerM: 1.20, OutputPerM: 4.80, CacheReadPerM: 0.600},
			"纯AZ":            {InputPerM: 0.225, OutputPerM: 0.90, CacheReadPerM: 0.1125},
			"限时体验":          {InputPerM: 0.21, OutputPerM: 0.84, CacheReadPerM: 0.105},
			"限时特价":          {InputPerM: 0.09, OutputPerM: 0.36, CacheReadPerM: 0.045},
			// auto_max fallback → 优质官转OpenAI
			"__auto_max__":   {InputPerM: 1.20, OutputPerM: 4.80, CacheReadPerM: 0.600},
		},
		"gpt-4o": {
			"default":        {InputPerM: 2.475, OutputPerM: 9.90, CacheReadPerM: 1.238},
			"官转":            {InputPerM: 7.50, OutputPerM: 30.00, CacheReadPerM: 3.75},
			"官转OpenAI":      {InputPerM: 15.00, OutputPerM: 60.00, CacheReadPerM: 7.50},
			"优质官转OpenAI":   {InputPerM: 20.00, OutputPerM: 80.00, CacheReadPerM: 10.00},
			"纯AZ":            {InputPerM: 3.75, OutputPerM: 15.00, CacheReadPerM: 1.875},
			"限时体验":          {InputPerM: 3.50, OutputPerM: 14.00, CacheReadPerM: 1.75},
			"限时特价":          {InputPerM: 1.50, OutputPerM: 6.00, CacheReadPerM: 0.75},
			// auto chain: 纯AZ → 官转 → 限时特价 → 限时体验 → default → 官转OpenAI; auto_max = 优质官转OpenAI
			"__auto_max__":   {InputPerM: 20.00, OutputPerM: 80.00, CacheReadPerM: 10.00},
		},
		"deepseek-v3": {
			"default":        {InputPerM: 2.00, OutputPerM: 8.00},
			"企业级高可用大模型":   {InputPerM: 2.00, OutputPerM: 8.00},
			"官转":            {InputPerM: 6.00, OutputPerM: 24.00},
			"纯AZ":            {InputPerM: 3.00, OutputPerM: 12.00},
			// auto chain: 纯AZ → 官转 → default; auto_max = 官转
			"__auto_max__":   {InputPerM: 6.00, OutputPerM: 24.00},
		},
		"grok-4.3": {
			"default":      {InputPerM: 1.25, OutputPerM: 2.50},
			"优质grok":      {InputPerM: 7.50, OutputPerM: 15.00},
			"纯AZ":          {InputPerM: 1.875, OutputPerM: 3.75},
			"限时体验":        {InputPerM: 1.75, OutputPerM: 3.50},
			// auto chain: 纯AZ → 限时体验 → default; auto_max = 纯AZ
			"__auto_max__": {InputPerM: 1.875, OutputPerM: 3.75},
		},
	}
	n1nGroupPricingMu.Lock()
	n1nGroupPricingMap = m
	n1nGroupPricingMu.Unlock()
}

// LookupN1nGroupPrice returns the upstream price for a model+group pair.
// Falls back to __auto_max__ if the exact group is not found.
// Returns (price, "exact_header_group") or (price, "auto_max_fallback") or (zero, "").
func LookupN1nGroupPrice(modelID, routingGroup string) (N1nGroupPrice, string) {
	n1nGroupPricingMu.RLock()
	defer n1nGroupPricingMu.RUnlock()

	if n1nGroupPricingMap == nil {
		return N1nGroupPrice{}, ""
	}

	groups, ok := n1nGroupPricingMap[modelID]
	if !ok {
		return N1nGroupPrice{}, ""
	}

	// Exact match
	if routingGroup != "" {
		if p, found := groups[routingGroup]; found {
			return p, "exact_header_group"
		}
	}

	// Fallback to auto_max
	if p, found := groups["__auto_max__"]; found {
		return p, "auto_max_fallback"
	}

	return N1nGroupPrice{}, ""
}

// N1nGroupPricingToModelRatio converts an upstream price (USD/1M tokens) to
// New API ModelRatio. Formula: ModelRatio = (upstream * markup) / 2
// where the /2 is the New API internal unit convention.
func N1nGroupPricingToModelRatio(upstreamInputPerM, markup float64) float64 {
	return (upstreamInputPerM * markup) / 2.0
}

// N1nGroupPricingToCompletionRatio returns CompletionRatio = output / input.
// Returns 1.0 if input is zero to avoid division by zero.
func N1nGroupPricingToCompletionRatio(upstreamInputPerM, upstreamOutputPerM float64) float64 {
	if upstreamInputPerM == 0 {
		return 1.0
	}
	return upstreamOutputPerM / upstreamInputPerM
}
