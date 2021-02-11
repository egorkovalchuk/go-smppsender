package iprest

import (
	"net/http"

	"github.com/egorkovalchuk/go-smppsender/pdata"
)

// IPRestCheck Проверка по шаблонам IP
func IPRestCheck(IpAllow []string, IpType int, IP string) (allowed bool, err error) {
	if IpType == 0 {
		return true, nil
	}
	return true, nil
}

// AuthCheck проверка авторизации
func AuthCheck(w http.ResponseWriter, r *http.Request, cfg *pdata.Config) {
	if cfg.AuthType == 0 {
		return
	}
	return
}
