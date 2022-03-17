package v2

const (
	CheckStatusOk       = uint32(0)
	CheckStatusWarning  = uint32(1)
	CheckStatusCritical = uint32(2)

	CheckStatusCaptionOk       = "ok"
	CheckStatusCaptionWarning  = "warning"
	CheckStatusCaptionCritical = "critical"
	CheckStatusCaptionUnknown  = "unknown"
)

var (
	statusToCaptionMap = map[uint32]string{
		CheckStatusOk:       CheckStatusCaptionOk,
		CheckStatusWarning:  CheckStatusCaptionWarning,
		CheckStatusCritical: CheckStatusCaptionCritical,
	}
)

func CheckStatusToCaption(status uint32) string {
	caption, ok := statusToCaptionMap[status]
	if ok {
		return caption
	}
	return CheckStatusCaptionUnknown
}
