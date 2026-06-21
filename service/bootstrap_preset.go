package service

import (
	"os"
	"strings"

	"github.com/QuantumNous/new-api/model"
)

func ApplyDeploymentPreset() error {
	values := map[string]string{}

	if systemName := strings.TrimSpace(os.Getenv("BOLUO_SYSTEM_NAME")); systemName != "" {
		values["SystemName"] = systemName
	}
	if logo := strings.TrimSpace(os.Getenv("BOLUO_LOGO_URL")); logo != "" {
		values["Logo"] = logo
	}
	if footer := strings.TrimSpace(os.Getenv("BOLUO_FOOTER_HTML")); footer != "" {
		values["Footer"] = footer
	}
	if docsLink := strings.TrimSpace(os.Getenv("BOLUO_DOCS_LINK")); docsLink != "" {
		values["general_setting.docs_link"] = docsLink
	}
	if retryTimes := strings.TrimSpace(os.Getenv("BOLUO_RETRY_TIMES")); retryTimes != "" {
		values["RetryTimes"] = retryTimes
	}
	if disableStatusCodes := strings.TrimSpace(os.Getenv("BOLUO_AUTO_DISABLE_STATUS_CODES")); disableStatusCodes != "" {
		values["AutomaticDisableStatusCodes"] = disableStatusCodes
	}
	if retryStatusCodes := strings.TrimSpace(os.Getenv("BOLUO_AUTO_RETRY_STATUS_CODES")); retryStatusCodes != "" {
		values["AutomaticRetryStatusCodes"] = retryStatusCodes
	}

	if len(values) == 0 {
		return nil
	}

	return model.UpdateOptionsBulk(values)
}
