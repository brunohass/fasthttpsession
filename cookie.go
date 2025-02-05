package fasthttpsession

import (
	"time"

	"github.com/valyala/fasthttp"
)

func NewCookie() *Cookie {
	return &Cookie{}
}

type Cookie struct {
}

// get cookie by name
func (c *Cookie) Get(ctx *fasthttp.RequestCtx, name string) (value string) {
	cookieByte := ctx.Request.Header.Cookie(name)
	if len(cookieByte) > 0 {
		value = string(cookieByte)
	}
	return
}

// response set cookie
func (c *Cookie) Set(ctx *fasthttp.RequestCtx, name string, value string, domain string, expires time.Duration, secure bool, sameSite fasthttp.CookieSameSite, httpOnly bool) {

	cookie := fasthttp.AcquireCookie()
	defer fasthttp.ReleaseCookie(cookie)

	cookie.SetKey(name)
	cookie.SetPath("/")
	cookie.SetHTTPOnly(true)
	cookie.SetDomain(domain)
	cookie.SetSameSite(sameSite)
	cookie.SetHTTPOnly(httpOnly)
	if expires >= 0 {
		// = 0 unlimited life
		var expiredTime time.Time = fasthttp.CookieExpireUnlimited
		if expires > 0 {
			// > 0
			expiredTime = time.Now().Add(expires)
		}
		cookie.SetExpire(expiredTime)
	}
	if ctx.IsTLS() && secure {
		cookie.SetSecure(true)
	}

	cookie.SetValue(value)
	ctx.Response.Header.SetCookie(cookie)
}

// delete cookie by cookie name
func (c *Cookie) Delete(ctx *fasthttp.RequestCtx, name string) {

	// delete response cookie
	ctx.Response.Header.DelCookie(name)

	// reset response cookie
	cookie := fasthttp.AcquireCookie()
	defer fasthttp.ReleaseCookie(cookie)
	cookie.SetKey(name)
	cookie.SetValue("")
	cookie.SetPath("/")
	cookie.SetHTTPOnly(true)
	//RFC says 1 second, but let's do it 1 minute to make sure is working...
	exp := time.Now().Add(-time.Duration(1) * time.Minute)
	cookie.SetExpire(exp)
	ctx.Response.Header.SetCookie(cookie)

	// delete request's cookie also
	ctx.Request.Header.DelCookie(name)
}
