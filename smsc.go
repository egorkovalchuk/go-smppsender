package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/linxGnu/gosmpp"
	"github.com/linxGnu/gosmpp/data"
	"github.com/linxGnu/gosmpp/pdu"
)

//Power by  Egor Kovalchuk

//listnumber is список номеров из конфига
type listnumber struct {
	Name   string   //имя списка
	Msisdn []string //массив из строк =- список номеров
}

//Сonfig is стурктура конфига
type Config struct {
	SMSC                 string //Строка коннекта с портом
	SystemID             string // логин
	Password             string // пароль
	MsisdnFrom           string // отправитель по умолчанию
	SrcSetTon            byte
	SrcSetNpi            byte
	DestSetTon           byte
	DestSetNpi           byte
	ProtocolID           byte
	RegisteredDelivery   byte
	ReplaceIfPresentFlag byte
	EsmClass             byte
	HTTPport             int
	AuthType             int
	UserAuth             string
	PassAuth             string
	IPRestrictionType    int
	IPRestriction        []string
	Listnumbers          []listnumber //список рассылок
}

//sms is структура sms
type sms struct {
	msisdnto   string // кому
	msisdnfrom string // от кого
	message    string // тельце
}

//Ответ хттп сервиса
type Response struct {
	Status string
	Error  string
}

// логи
const logFileName = "smsc.log"
const pidFileName = "smsc.pid"

//конфиг
var cfg Config

//режим работы сервиса(дебаг мод)
var debugm bool
var emul bool

//Переменная для работы с смсц
var trans *gosmpp.TransceiverSession

//ошибки
var err error

//режим работы сервиса
var startdaemon bool

//запрос версии
var version bool

/*
Vesion 0.1
Add HTTP sevice
Verson 0.2
Add authorization (Login / Pass)
Verson 0.3
Add restriction IP
*/
const versionutil = "0.1"

func main() {

	//start program
	var argument string
	var progName string

	progName = os.Args[0]

	//load conf
	readconf(&cfg, "smsc.ini")

	if os.Args != nil && len(os.Args) > 1 {
		argument = os.Args[1]
	} else {
		helpstart()
		return
	}

	if argument == "-h" {
		helpstart()
		return
	}

	flag.BoolVar(&debugm, "t", false, "a bool")
	flag.BoolVar(&startdaemon, "d", false, "a bool")
	flag.BoolVar(&version, "v", false, "a bool")
	flag.BoolVar(&emul, "e", false, "a bool")
	// for Linux compile
	stdaemon := flag.Bool("s", false, "a bool") // для передачи
	// --for Linux compile
	var list_name string
	flag.StringVar(&list_name, "l", "", "Name list is not empty")
	var message string
	flag.StringVar(&message, "m", "", "Messge is not empty")
	flag.Parse()

	if version {
		fmt.Println("Version utils " + versionutil)
		return
	}

	if startdaemon || *stdaemon {

		filer, err := os.OpenFile(logFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			log.Fatal(err)
		}

		log.SetOutput(filer)

		log.Println("- - - - - - - - - - - - - - -")
		log.Println("Start daemon mode")
		if debugm {
			log.Println("Start with debug mode")
		}

		fmt.Println("Start daemon mode")

		processinghttp(&cfg, debugm)

		log.Println("daemon terminated")

	} else {

		if list_name == "" {
			fmt.Println("Could not start, list name is empty")
			return
		}

		if message == "" {
			fmt.Println(progName + " could not start, message is empty")
			return
		}

		fmt.Println("Start")
		if debugm {
			fmt.Println("List Name:" + list_name)
			fmt.Println("message:" + message)
		}
		StartShellMode(message, list_name)

	}
	fmt.Println("Done")
	return

}

func processError(err error) {
	fmt.Println(err)
	os.Exit(2)
}

func readconf(cfg *Config, confname string) {
	file, err := os.Open(confname)
	if err != nil {
		processError(err)
	}

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&cfg)
	if err != nil {
		processError(err)
	}

	file.Close()

}

//StartShellMode запуск в режиме скрипта
func StartShellMode(message string, listname string) {
	var cnt bool
	cnt = false
	var preloadcf listnumber

	fmt.Println("Check conf")
	for _, cf := range cfg.Listnumbers {
		if cf.Name == listname {
			cnt = true
			preloadcf = cf
			break
		}
	}

	if !cnt {
		fmt.Println("List " + listname + " does not exist")
		return
	}

	if debugm {
		fmt.Println(preloadcf.Name)
		fmt.Println(preloadcf.Msisdn)
	}

	processing(&preloadcf, message)
}

func processing(preloadcf *listnumber, message string) {
	smsrow := []sms{}

	for _, cf := range preloadcf.Msisdn {

		p := sms{cf, cfg.MsisdnFrom, message}
		if debugm {
			fmt.Println("prepared sms ")
			fmt.Println(p)
		}
		smsrow = append(smsrow, p)
	}

	fmt.Println("start send")

	auth := gosmpp.Auth{
		SMSC:       cfg.SMSC,
		SystemID:   cfg.SystemID,
		Password:   cfg.Password,
		SystemType: "",
	}

	trans, err := gosmpp.NewTransceiverSession(gosmpp.NonTLSDialer, auth, gosmpp.TransceiveSettings{
		EnquireLink: 5 * time.Second,

		WriteTimeout: time.Second,

		// this setting is very important to detect broken conn.
		// After timeout, there is no read packet, then we decide it's connection broken.
		ReadTimeout: 60 * time.Second,

		OnSubmitError: func(p pdu.PDU, err error) {
			log.Fatal("SubmitPDU error:", err)
		},

		OnReceivingError: func(err error) {
			fmt.Println("Receiving PDU/Network error:", err)
		},

		OnRebindingError: func(err error) {
			fmt.Println("Rebinding but error:", err)
		},

		OnPDU: handlePDU(),

		OnClosed: func(state gosmpp.State) {
			fmt.Println(state)
		},
	}, 5*time.Second)
	if err != nil {
		log.Println(err)
		fmt.Println("Ошибка открытия сессии к smsc")
		return
	}
	defer func() {
		_ = trans.Close()
	}()

	for _, p := range smsrow {

		fmt.Println("Start proccess SMS: " + p.msisdnto)
		if len(p.message) < 256 {
			err = trans.Transceiver().Submit(newSubmitSM(p.msisdnfrom, p.msisdnto, p.message))
		} else {
			newSubmitLongSM(p.msisdnfrom, p.msisdnto, p.message)
		}
		if err != nil {
			fmt.Println(err)
			fmt.Println("Ошибка отправки")
		}

		time.Sleep(time.Second)
		fmt.Println("End proccess SMS: " + p.msisdnto)
	}

	//trans.Transceiver().Close()

}

func handlePDU() func(pdu.PDU, bool) {
	concatenated := map[uint8][]string{}
	return func(p pdu.PDU, responded bool) {
		switch pd := p.(type) {
		case *pdu.SubmitSMResp:
			if startdaemon {
				log.Printf("SubmitSMResp:%+v\n", pd)
			} else {
				fmt.Printf("SubmitSMResp:%+v\n", pd)
			}

		case *pdu.GenericNack:
			if startdaemon {
				log.Println("GenericNack Received")
			} else {
				fmt.Println("GenericNack Received")
			}

		case *pdu.EnquireLinkResp:
			if startdaemon {
				log.Println("EnquireLinkResp Received")
			} else {
				fmt.Println("EnquireLinkResp Received")
			}

		case *pdu.DataSM:
			if startdaemon {
				log.Printf("DataSM:%+v\n", pd)
			} else {
				fmt.Printf("DataSM:%+v\n", pd)
			}

		case *pdu.DeliverSM:
			if startdaemon {
				log.Printf("DeliverSM:%+v\n", pd)
			} else {
				fmt.Printf("DeliverSM:%+v\n", pd)
			}
			log.Println(pd.Message.GetMessage())
			// region concatenated sms (sample code)
			message, err := pd.Message.GetMessage()
			if err != nil {
				log.Fatal(err)
			}
			totalParts, sequence, reference, found := pd.Message.UDH().GetConcatInfo()
			if found {
				if _, ok := concatenated[reference]; !ok {
					concatenated[reference] = make([]string, totalParts)
				}
				concatenated[reference][sequence-1] = message
			}
			if !found {
				log.Println(message)
			} else if parts, ok := concatenated[reference]; ok && isConcatenatedDone(parts, totalParts) {
				log.Println(strings.Join(parts, ""))
				delete(concatenated, reference)
			}
			// endregion
		}
	}
}

func newSubmitSM(srcaddr string, destaddr string, message string) *pdu.SubmitSM {
	// build up submitSM
	srcAddr := pdu.NewAddress()
	srcAddr.SetTon(cfg.SrcSetTon)
	srcAddr.SetNpi(cfg.SrcSetNpi)
	_ = srcAddr.SetAddress(srcaddr)

	destAddr := pdu.NewAddress()
	destAddr.SetTon(cfg.DestSetTon)
	destAddr.SetNpi(cfg.DestSetTon)
	_ = destAddr.SetAddress(destaddr)

	submitSM := pdu.NewSubmitSM().(*pdu.SubmitSM)
	submitSM.SourceAddr = srcAddr
	submitSM.DestAddr = destAddr
	err = submitSM.Message.SetMessageWithEncoding(message, data.UCS2)
	if err != nil {
		logwrite(err)
	}
	submitSM.ProtocolID = cfg.ProtocolID
	submitSM.RegisteredDelivery = cfg.RegisteredDelivery
	submitSM.ReplaceIfPresentFlag = cfg.ReplaceIfPresentFlag
	submitSM.EsmClass = cfg.EsmClass

	return submitSM
}

func newSubmitLongSM(srcaddr string, destaddr string, message string) {
	// build up submitSM
	srcAddr := pdu.NewAddress()
	srcAddr.SetTon(cfg.SrcSetTon)
	srcAddr.SetNpi(cfg.SrcSetNpi)
	_ = srcAddr.SetAddress(srcaddr)

	destAddr := pdu.NewAddress()
	destAddr.SetTon(cfg.DestSetTon)
	destAddr.SetNpi(cfg.DestSetTon)
	_ = destAddr.SetAddress(destaddr)

	//submitSM := pdu.NewSubmitMulti().(*pdu.SubmitMulti)
	submitSM := pdu.NewSubmitSM().(*pdu.SubmitSM)
	submitSM.SourceAddr = srcAddr
	submitSM.DestAddr = destAddr
	err = submitSM.Message.SetLongMessageWithEnc(message, data.UCS2)
	if err != nil {
		logwrite(err)
	}
	var multimessage []*pdu.SubmitSM
	multimessage, err := submitSM.Split()
	if err != nil {
		logwrite(err)
	}

	submitSM.ProtocolID = cfg.ProtocolID
	submitSM.RegisteredDelivery = cfg.RegisteredDelivery
	submitSM.ReplaceIfPresentFlag = cfg.ReplaceIfPresentFlag
	submitSM.EsmClass = cfg.EsmClass

	for _, p := range multimessage {
		err = trans.Transceiver().Submit(p)
		if err != nil {
			logwrite(err)
		}
	}

	return
}

func isConcatenatedDone(parts []string, total byte) bool {
	for _, part := range parts {
		if part != "" {
			total--
		}
	}
	return total == 0
}

func processinghttp(cfg *Config, debugm bool) {

	auth := gosmpp.Auth{
		SMSC:       cfg.SMSC,
		SystemID:   cfg.SystemID,
		Password:   cfg.Password,
		SystemType: "",
	}

	if !emul {
		trans, err = gosmpp.NewTransceiverSession(gosmpp.NonTLSDialer, auth, gosmpp.TransceiveSettings{
			EnquireLink: 5 * time.Second,

			WriteTimeout: time.Second,

			// this setting is very important to detect broken conn.
			// After timeout, there is no read packet, then we decide it's connection broken.
			ReadTimeout: 60 * time.Second,

			OnSubmitError: func(p pdu.PDU, err error) {
				log.Fatal("SubmitPDU error:", err)
			},

			OnReceivingError: func(err error) {
				log.Println("Receiving PDU/Network error:", err)
			},

			OnRebindingError: func(err error) {
				log.Println("Rebinding but error:", err)
			},

			OnPDU: handlePDU(),

			OnClosed: func(state gosmpp.State) {
				log.Println(state)
			},
		}, 5*time.Second)

		if err != nil {
			log.Println(err)
			log.Println("Ошибка открытия сессии к smsc")
			return
		}
		defer func() {
			_ = trans.Close()
		}()

	}
	s := &http.Server{
		Addr:           ":" + strconv.Itoa(cfg.HTTPport),
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	http.HandleFunc("/", httpHandler)
	http.HandleFunc("/list", httpHandlerlist)
	http.HandleFunc("/conf", httpHandlerconf)

	log.Fatal(s.ListenAndServe())

}

func httpHandlerconf(w http.ResponseWriter, r *http.Request) {
	log.Printf("request from %s: %s %q", r.RemoteAddr, r.Method, r.URL)
	var reloadconf string
	var errortext string
	var st Response

	if cfg.AuthType == 1 {

		username, password, authOK := r.BasicAuth()
		if authOK == false {
			http.Error(w, "Not authorized", 401)
			return
		}

		if username != cfg.UserAuth || password != cfg.PassAuth {
			http.Error(w, "Not authorized", 401)
			return
		}
	}

	reloadconf = r.FormValue("reloadconf")

	//перегрузка конфига
	if reloadconf == "1" {
		//load conf
		readconf(&cfg, "smsc.ini")
		if errortext == "" {
			st = Response{"OK", errortext}
		} else {
			st = Response{"ERROR", errortext}
		}

		js, err := json.Marshal(st)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(js)
		return
	}

	return

}

func httpHandlerlist(w http.ResponseWriter, r *http.Request) {
	log.Printf("request from %s: %s %q", r.RemoteAddr, r.Method, r.URL)

	w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)

	if cfg.AuthType == 1 {

		username, password, authOK := r.BasicAuth()
		if authOK == false {
			http.Error(w, "Not authorized", 401)
			return
		}

		if username != cfg.UserAuth || password != cfg.PassAuth {
			http.Error(w, "Not authorized", 401)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")

	type Response struct {
		Status      string
		Error       string
		Listnumbers []listnumber
	}

	var st Response
	st = Response{"OK", "", cfg.Listnumbers}

	js, err := json.Marshal(st)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(js)

	return

}

func httpHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("request from %s: %s %q", r.RemoteAddr, r.Method, r.URL)

	if cfg.AuthType == 1 {

		username, password, authOK := r.BasicAuth()
		if authOK == false {
			http.Error(w, "Not authorized", 401)
			return
		}

		if username != cfg.UserAuth || password != cfg.PassAuth {
			http.Error(w, "Not authorized", 401)
			return
		}
	}

	var srcmsisdn string
	var dstmsisdn string
	var lst string
	var message string
	var errortext string
	var errortype int
	var preloadcf listnumber
	var st Response

	srcmsisdn = r.FormValue("src")
	dstmsisdn = r.FormValue("dst")
	lst = r.FormValue("lst")
	message = r.FormValue("text")

	if srcmsisdn == "" {
		errortext = "Empty sender "
	}
	if (dstmsisdn == "") && (lst == "") {
		errortext += "Empty destination "
	}
	if message == "" {
		errortext += "Empty text "
	}

	smsrow := []sms{}
	errortype = 0

	if errortext == "" {
		if lst != "" {
			var cnt bool
			cnt = false

			if debugm {
				log.Println("Check list")
			}

			for _, cf := range cfg.Listnumbers {
				if cf.Name == lst {
					cnt = true
					preloadcf = cf
					break
				}
			}

			if !cnt {
				log.Println("List " + lst + " does not exist")
				errortype = 1
			}

			if debugm && errortype != 1 {
				log.Println(preloadcf.Name)
				log.Println(preloadcf.Msisdn)
			}
		}

		if errortype != 1 {
			for _, cf := range preloadcf.Msisdn {

				p := sms{cf, srcmsisdn, message}
				if debugm {
					log.Println("prepared sms ")
					log.Println(p)
				}
				smsrow = append(smsrow, p)
			}
		}

		if dstmsisdn != "" {
			p := sms{dstmsisdn, srcmsisdn, message}
			smsrow = append(smsrow, p)
			if debugm {
				log.Println("prepared sms ")
				log.Println(p)
			}
		}

		if errortype == 1 && dstmsisdn == "" {
			errortext = "List " + lst + " does not exist"
		}

	}

	for _, p := range smsrow {

		log.Println("Start proccess SMS: " + p.msisdnto)

		if !emul {
			if len(p.message) < 256 {
				err = trans.Transceiver().Submit(newSubmitSM(p.msisdnfrom, p.msisdnto, p.message))

			} else {
				//err = trans.Transceiver().Submit(newSubmitLongSM(p.msisdnfrom, p.msisdnto, p.message))
				newSubmitLongSM(p.msisdnfrom, p.msisdnto, p.message)
			}
			if err != nil {
				log.Println(err)
				log.Println("Ошибка отправки")
			}
		}

		time.Sleep(time.Second)
		log.Println("End proccess SMS: " + p.msisdnto)
	}

	if errortext == "" {
		st = Response{"OK", errortext}
	} else {
		st = Response{"ERROR", errortext}
	}

	js, err := json.Marshal(st)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)

}

func helpstart() {
	fmt.Println("Use -l Name list -m \"Text message\"")
	fmt.Println("Use -d start deamon mode(HTTP service)")
	fmt.Println("Example 1 curl localhost:8080 -X GET -F src=IT -F lst=rss_1 -F text=hello")
	fmt.Println("Example 2 curl localhost:8080 -X GET -F src=IT -F dst=79XXXXXXXX -F text=hello)")
	fmt.Println("Example 3 curl localhost:8080/conf -X GET -F reloadconf=1")
	fmt.Println("Example 4 curl localhost:8080/list -X GET ")
	fmt.Println("Use -s stop deamon mode(HTTP service)")
	fmt.Println("Use -t start with debug mode")
}

func logwrite(err error) {
	if startdaemon {
		log.Println(err)
	} else {
		fmt.Println(err)
	}
}
