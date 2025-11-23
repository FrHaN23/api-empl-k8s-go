package global

var (
	Apiv1        string
	MaxBodyBytes int64
)

func init() {
	Apiv1 = "/api/v1"
	MaxBodyBytes = 1 << 20 // 1MB
}
