package dedupe

import (
	"fmt"
)

func OvertureId(id string) string {
	return idWithPrefix(OVERTURE_PREFIX, id)
}

func WhosOnFirstId(id string) string {
	return idWithPrefix(WHOSONFIRST_PREFIX, id)
}

func AllThePlacesId(id string) string {
	return idWithPrefix(ALLTHEPLACES_PREFIX, id)
}

func ILMSId(id string) string {
	return idWithPrefix(ILMS_PREFIX, id)
}

func idWithPrefix(prefix string, id string) string {
	return fmt.Sprintf("%s:id=%s", prefix, id)
}
