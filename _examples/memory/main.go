package main

// fasthttpsession memory provider example

import (
	"log"
	"os"

	"github.com/brunohass/fasthttpsession"
	"github.com/brunohass/fasthttpsession/memory"
	"github.com/valyala/fasthttp"
)

// default config
var session = fasthttpsession.NewSession(fasthttpsession.NewDefaultConfig())

// custom config
//var session = fasthttpsession.NewSession(&fasthttpsession.Config{
//	CookieName: "ssid",
//	Domain: "",
//	Expires: time.Hour * 2,
//	GCLifetime: 3,
//	SessionLifetime: 60,
//	Secure: true,
//	SessionIdInURLQuery: false,
//	SessionNameInUrlQuery: "",
//	SessionIdInHttpHeader: false,
//	SessionNameInHttpHeader: "",
//	SessionIdGeneratorFunc: func() string {return ""},
//	EncodeFunc: func(cookieValue string) (string, error) {return "", nil},
//	DecodeFunc: func(cookieValue string) (string, error) {return "", nil},
//})

func main() {

	// You must set up provider before use
	err := session.SetProvider("memory", &memory.Config{})
	if err != nil {
		log.Println(err.Error())
		os.Exit(1)
	}
	addr := ":8086"
	log.Println("fasthttpsession memory example server listen: " + addr)
	// Fasthttp start listen serve
	err = fasthttp.ListenAndServe(addr, requestRouter)
	if err != nil {
		log.Println("listen server error :" + err.Error())
	}
}
