package infrastructureutilitylogger

import domainmodelslogger "github.com/MateWorkspace/lockmate/internal/domain/models/logger"

func normalizeMeta(meta domainmodelslogger.Meta) map[string]any {
	normalized := make(map[string]any, len(meta))
	for key, value := range meta {
		if err, ok := value.(error); ok && err != nil {
			normalized[key] = err.Error()
			continue
		}

		normalized[key] = value
	}

	return normalized
}
