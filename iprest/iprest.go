package iprest

import (
	"errors"
	"net"
	"net/http"
	"regexp"
	"strconv"
)

// IPRest reload config []string to []net.IPNet
func IPRest(p string) (nets net.IPNet, err error) {

	re := regexp.MustCompile(`(?P<IP>^\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})(\/(?P<mask>[0-9]|[1-2][0-9]|3[0-2]))?$`)
	re.MatchString(p)
	if re.MatchString(p) {

		n := net.ParseIP(re.ReplaceAllString(p, "${IP}"))
		maskstring := re.ReplaceAllString(p, "${mask}")

		var m net.IPMask
		if maskstring == "" {
			m = net.CIDRMask(32, 32)
		} else {
			maskint, _ := strconv.Atoi(maskstring)
			m = net.CIDRMask(maskint, 32)
		}
		nets = net.IPNet{n, m}
	} else {
		return nets, errors.New("Error parse IP " + p)
	}

	return nets, nil
}

// IPRestCheck Проверка по шаблонам IP
func IPRestCheck(RemoteAddr string, IPRestrictionType int, Nets []net.IPNet) (allowed bool, err error) {
	if IPRestrictionType == 1 {

		remoteipstring, _, _ := net.SplitHostPort(RemoteAddr)
		ok := false
		ip := net.ParseIP(remoteipstring)
		for _, n := range Nets {
			if n.Contains(ip) {
				ok = true
				break
			}
		}
		if !ok {
			return false, errors.New("Access denied")
		}
		return true, nil
	}

	return true, nil
}

// AuthCheck проверка авторизации
func AuthCheck(w http.ResponseWriter, r *http.Request, AuthType int, UserAuth, PassAuth string) (allowed bool, err error) {
	if AuthType == 0 {
		return true, nil
	}

	w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)

	if AuthType == 1 {

		username, password, authOK := r.BasicAuth()
		if authOK == false {
			http.Error(w, "Not authorized", 401)
			return false, nil
		}

		if username != UserAuth || password != PassAuth {
			http.Error(w, "Not authorized", 401)
			return false, nil
		}
	}
	return true, nil
}
