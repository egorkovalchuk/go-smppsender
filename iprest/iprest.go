package iprest

import (
	"net/http"

	structdata "github.com/egorkovalchuk/go-smppsender/StructData"
)

// IPRestCheck Проверка по шаблонам IP
func IPRestCheck(IpAllow []string, IpType int, IP string) (allowed bool, err error) {
	if IpType == 0 {
		return true, nil
	}
	return true, nil
}

// AuthCheck проверка авторизации
func AuthCheck(w http.ResponseWriter, r *http.Request, cfg *structdata.Config) {
	if cfg.AuthType == 0 {
		return
	}
	return
}
