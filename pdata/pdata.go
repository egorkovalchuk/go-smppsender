package pdata

// Listnumber is список номеров из конфига
type Listnumber struct {
	// Name list
	Name string
	// Msisdn list msisdns
	Msisdn []string
}

// Config configuration stucture
type Config struct {

	// SMSC Строка коннекта с портом
	SMSC string
	// SystemID Login connect to SMSC
	SystemID string
	// Password Password connect to SMSC
	Password string
	// MsisdnFrom Отправитель
	MsisdnFrom string
	// SrcSetTon Source Ton
	SrcSetTon byte
	// SrcSetNPI Source Ton
	SrcSetNpi byte
	// DestSetTon DestSetTon
	DestSetTon byte
	// DestSetNpi DestSetNpi
	DestSetNpi byte
	// ProtocolID version Protocol
	ProtocolID byte
	// RegisteredDelivery report delivey
	RegisteredDelivery byte
	// ReplaceIfPresentFlag ReplaceIfPresentFlag
	ReplaceIfPresentFlag byte
	// EsmClass EsmClass
	EsmClass byte
	// HTTPPort Port for http service
	HTTPport int
	// AuthType 0-disable 1-enable check IP Restriction
	AuthType int
	// UserAuth User Basic auth
	UserAuth string
	// PassAuth User Basic auth
	PassAuth string
	// IPRestrictionType 0-disable 1-enable check IP Restriction
	IPRestrictionType int
	// IPRestriction List IP mask for allow access
	IPRestriction []string
	// Listnumbers List destination
	Listnumbers []Listnumber
}

// Sms is stuct sms
type Sms struct {
	// MsisdnTo destination Msisdn
	MsisdnTo string
	// MsisdnFrom source Msisdn
	MsisdnFrom string
	// Message text sms
	Message string
}

// Response Ответ хттп сервиса
type Response struct {
	// Status processing success report
	Status string
	// Error text
	Error string
}
