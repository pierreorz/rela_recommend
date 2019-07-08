package routers

import (
	"encoding/json"
	"encoding/xml"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"rela_recommend/utils/binding"
	"rela_recommend/utils/signature"
)

// MIME types
const (
	MIMEApplicationJSON                  = "application/json"
	MIMEApplicationJSONCharsetUTF8       = MIMEApplicationJSON + "; " + charsetUTF8
	MIMEApplicationJavaScript            = "application/javascript"
	MIMEApplicationJavaScriptCharsetUTF8 = MIMEApplicationJavaScript + "; " + charsetUTF8
	MIMEApplicationXML                   = "application/xml"
	MIMEApplicationXMLCharsetUTF8        = MIMEApplicationXML + "; " + charsetUTF8
	MIMEApplicationForm                  = "application/x-www-form-urlencoded"
	MIMEApplicationProtobuf              = "application/protobuf"
	MIMEApplicationMsgpack               = "application/msgpack"
	MIMETextHTML                         = "text/html"
	MIMETextHTMLCharsetUTF8              = MIMETextHTML + "; " + charsetUTF8
	MIMETextPlain                        = "text/plain"
	MIMETextPlainCharsetUTF8             = MIMETextPlain + "; " + charsetUTF8
	MIMEMultipartForm                    = "multipart/form-data"
	MIMEOctetStream                      = "application/octet-stream"
)
const (
	charsetUTF8 = "charset=utf-8"
)

// Headers
const (
	HeaderAcceptEncoding                = "Accept-Encoding"
	HeaderAllow                         = "Allow"
	HeaderAuthorization                 = "Authorization"
	HeaderContentDisposition            = "Content-Disposition"
	HeaderContentEncoding               = "Content-Encoding"
	HeaderContentLength                 = "Content-Length"
	HeaderContentType                   = "Content-Type"
	HeaderCookie                        = "Cookie"
	HeaderSetCookie                     = "Set-Cookie"
	HeaderIfModifiedSince               = "If-Modified-Since"
	HeaderLastModified                  = "Last-Modified"
	HeaderLocation                      = "Location"
	HeaderUpgrade                       = "Upgrade"
	HeaderVary                          = "Vary"
	HeaderWWWAuthenticate               = "WWW-Authenticate"
	HeaderXForwardedProto               = "X-Forwarded-Proto"
	HeaderXHTTPMethodOverride           = "X-HTTP-Method-Override"
	HeaderXForwardedFor                 = "X-Forwarded-For"
	HeaderXRealIP                       = "X-Real-IP"
	HeaderUserAgent                     = "User-Agent"
	HeaderServer                        = "Server"
	HeaderOrigin                        = "Origin"
	HeaderAccessControlRequestMethod    = "Access-Control-Request-Method"
	HeaderAccessControlRequestHeaders   = "Access-Control-Request-Headers"
	HeaderAccessControlAllowOrigin      = "Access-Control-Allow-Origin"
	HeaderAccessControlAllowMethods     = "Access-Control-Allow-Methods"
	HeaderAccessControlAllowHeaders     = "Access-Control-Allow-Headers"
	HeaderAccessControlAllowCredentials = "Access-Control-Allow-Credentials"
	HeaderAccessControlExposeHeaders    = "Access-Control-Expose-Headers"
	HeaderAccessControlMaxAge           = "Access-Control-Max-Age"

	// Security
	HeaderStrictTransportSecurity = "Strict-Transport-Security"
	HeaderXContentTypeOptions     = "X-Content-Type-Options"
	HeaderXXSSProtection          = "X-XSS-Protection"
	HeaderXFrameOptions           = "X-Frame-Options"
	HeaderContentSecurityPolicy   = "Content-Security-Policy"
	HeaderXCSRFToken              = "X-CSRF-Token"
)

type Context struct {
	Writer  http.ResponseWriter
	Request *http.Request
	Params  Params
	//Index   int8
}

// Bind checks the Content-Type to select a binding engine automatically,
// Depending the "Content-Type" header different bindings are used:
// 		"application/json" --> JSON binding
// 		"application/xml"  --> XML binding
// otherwise --> returns an error
// It parses the request's body as JSON if Content-Type == "application/json" using JSON or XML as a JSON input.
// It decodes the json payload into the struct specified as a pointer.
// Like ParseBody() but this method also writes a 400 error if the json is not valid.
func (c *Context) Bind(obj interface{}) error {
	b := binding.Default(c.Request.Method, c.ContentType())
	return c.BindWith(obj, b)
}

func (c *Context) BindJSON(obj interface{}) error {
	return binding.JSON.Bind(c.Request, obj)
}

func (c *Context) BindForm(obj interface{}) error {
	return binding.Form.Bind(c.Request, obj)
}

func (c *Context) BindAndSingnature(obj interface{}) error {
	if err := signature.Signature(c.Request); err != nil {
		return err
	}
	b := binding.Default(c.Request.Method, c.ContentType())
	return c.BindWith(obj, b)
}

func (c *Context) BindFormAndSingnature(obj interface{}) error {
	if err := signature.Signature(c.Request); err != nil {
		return err
	}
	return binding.Form.Bind(c.Request, obj)
}

func (c *Context) BindJsonAndSingnature(obj interface{}) error {
	if err := signature.Signature(c.Request); err != nil {
		return err
	}
	return binding.JSON.Bind(c.Request, obj)
}

// It parses the request's body as JSON if Content-Type == "application/json" using JSON or XML as a JSON input.
// It decodes the json payload into the struct specified as a pointer.
// Like ParseBody() but this method also writes a 400 error if the json is not valid.
func (c *Context) DefaultBind(obj interface{}) error {
	b := binding.Default(c.Request.Method, c.ContentType())
	return c.BindWith(obj, b)
}

// BindWith binds the passed struct pointer using the specified binding engine.
// See the binding package.
func (c *Context) BindWith(obj interface{}, b binding.Binding) error {
	if err := b.Bind(c.Request, obj); err != nil {
		return err
	}
	return nil
}

func (c *Context) String(code int, s string) error {
	c.Header(HeaderContentType, MIMETextPlainCharsetUTF8)
	c.Status(code)
	return c.Write([]byte(s))
}

func (c *Context) Byte(code int, b []byte) error {
	c.Header(HeaderContentType, MIMETextPlainCharsetUTF8)
	c.Status(code)
	return c.Write(b)
}

func (c *Context) JSON(code int, i interface{}) error {
	b, err := json.Marshal(i)
	//b, err := json.MarshalIndent(i, "", "  ")
	if err != nil {
		return err
	}
	//c.Header(HeaderContentType, MIMEApplicationJSON)
	c.Header(HeaderContentType, MIMEApplicationJSONCharsetUTF8)
	c.Status(code)
	return c.Write(b)
}

func (c *Context) XML(code int, i interface{}) error {
	//b, err := xml.Marshal(i)
	b, err := xml.MarshalIndent(i, "", "  ")
	if err != nil {
		return err
	}
	//c.Header(HeaderContentType, MIMEApplicationXML)
	c.Header(HeaderContentType, MIMEApplicationXMLCharsetUTF8)
	c.Status(code)
	return c.Write(b)
}

func (c *Context) REDIRECT(location string) error {
	c.Header(HeaderContentType, MIMETextPlainCharsetUTF8)
	c.Header(HeaderLocation, location)
	c.Status(http.StatusFound)
	return c.Write(nil)
}

// ClientIP implements a best effort algorithm to return the real client IP, it parses
// X-Real-IP and X-Forwarded-For in order to work properly with reverse-proxies such us: nginx or haproxy.
func (c *Context) ClientIP() string {
	// clientIP := strings.TrimSpace(c.requestHeader(HeaderXRealIP))
	// if len(clientIP) > 0 {
	// 	return clientIP
	// }
	clientIP := c.requestHeader(HeaderXForwardedFor)
	if index := strings.IndexByte(clientIP, ','); index >= 0 {
		clientIP = clientIP[0:index]
	}
	clientIP = strings.TrimSpace(clientIP)
	if len(clientIP) > 0 {
		return clientIP
	}
	// if ip, _, err := net.SplitHostPort(strings.TrimSpace(c.Request.RemoteAddr)); err == nil {
	// 	return ip
	// }
	return ""
}

// ContentType returns the Content-Type header of the request.
func (c *Context) ContentType() string {
	return filterFlags(c.requestHeader(HeaderContentType))
}

// IsWebsocket returns true if the request headers indicate that a websocket
// handshake is being initiated by the client.
func (c *Context) IsWebsocket() bool {
	if strings.Contains(strings.ToLower(c.requestHeader("Connection")), "upgrade") &&
		strings.ToLower(c.requestHeader("Upgrade")) == "websocket" {
		return true
	}
	return false
}

// UserAgent returns the User-Agent header of the request.
func (c *Context) UserAgent() string {
	return c.requestHeader(HeaderUserAgent)
}

func (c *Context) requestHeader(key string) string {
	return c.Request.Header.Get(key)
}

/************************************/
/******** RESPONSE RENDERING ********/
/************************************/

func (c *Context) Status(code int) {
	c.Writer.WriteHeader(code)
}

// GetRawData return stream data.
func (c *Context) GetRawData() ([]byte, error) {
	return ioutil.ReadAll(c.Request.Body)
}

// SetCookie adds a Set-Cookie header to the ResponseWriter's headers.
// The provided cookie must have a valid Name. Invalid cookies may be
// silently dropped.
func (c *Context) SetCookie(name, value string, maxAge int, path, domain string, secure, httpOnly bool) {
	if path == "" {
		path = "/"
	}
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     name,
		Value:    url.QueryEscape(value),
		MaxAge:   maxAge,
		Path:     path,
		Domain:   domain,
		Secure:   secure,
		HttpOnly: httpOnly,
	})
}

func (c *Context) Cookie(name string) (string, error) {
	cookie, err := c.Request.Cookie(name)
	if err != nil {
		return "", err
	}
	val, _ := url.QueryUnescape(cookie.Value)
	return val, nil
}

// Header is a intelligent shortcut for c.Writer.Header().Set(key, value)
// It writes a header in the response.
// If value == "", this method removes the header `c.Writer.Header().Del(key)`
func (c *Context) Header(key, value string) {
	if len(value) == 0 {
		c.Writer.Header().Del(key)
	} else {
		c.Writer.Header().Set(key, value)
	}
}

func (c *Context) Write(data []byte) error {
	_, err := c.Writer.Write(data)
	return err
}
