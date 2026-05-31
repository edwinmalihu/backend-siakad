package license

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"siakad/backend/internal/response"
)

var licenseExemptPrefixes = []string{
	"/api/v1/auth/",
	"/api/v1/license/",
	"/health",
	"/docs/",
	"/openapi.yaml",
}

var licenseExemptExact = map[string]bool{
	"/api/v1":     true,
	"/openapi.yaml": true,
}

func LicenseMiddleware(repo *Repository, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		// Skip exempt paths
		if licenseExemptExact[path] {
			next.ServeHTTP(w, r)
			return
		}
		for _, prefix := range licenseExemptPrefixes {
			if strings.HasPrefix(path, prefix) {
				next.ServeHTTP(w, r)
				return
			}
		}

		license, err := repo.GetActive(r.Context())
		if err != nil {
			// DB error — block to be safe
			response.JSON(w, http.StatusForbidden, map[string]any{
				"success": false,
				"error":   "Unable to verify license.",
				"code":    "LICENSE_UNKNOWN",
			})
			return
		}

		// No license — block access
		if license == nil {
			response.JSON(w, http.StatusForbidden, map[string]any{
				"success": false,
				"error":   "No license installed. Please activate a license first.",
				"code":    "LICENSE_REQUIRED",
			})
			return
		}

		// License expired
		if time.Now().After(license.ExpiresAt) {
			response.JSON(w, http.StatusForbidden, map[string]any{
				"success": false,
				"error":   "License expired. Please activate a valid license.",
				"code":    "LICENSE_EXPIRED",
			})
			return
		}

		// License valid — add license info to response headers
		w.Header().Set("X-License-Tier", license.Tier)
		w.Header().Set("X-License-Expires", license.ExpiresAt.Format(time.RFC3339))

		// Check if expiring soon (H-30)
		daysRemaining := int(time.Until(license.ExpiresAt).Hours() / 24)
		if daysRemaining <= 30 && daysRemaining > 0 {
			w.Header().Set("X-License-Expiring-Soon", "true")
			w.Header().Set("X-License-Days-Remaining", fmt.Sprintf("%d", daysRemaining))
		}

		next.ServeHTTP(w, r)
	})
}
