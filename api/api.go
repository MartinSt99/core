package api

import (
	"net/http"
	"strconv"

	// "image/jpeg"
	// "image/png"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	log "github.com/sirupsen/logrus"
	"github.com/u-speak/core/config"
	"github.com/u-speak/core/node"
	"github.com/u-speak/logrusmiddleware"
)

// API is used as a container, allowing the REST API to access the node
type API struct {
	ListenInterface string
	Message         string
	node            *node.Node
	certfile        string
	keyfile         string
	adminEnabled    bool
	user            string
	password        string
}

// Error is returned when something has gone wrong
type Error struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

type jsonBlock struct {
	Nonce        uint32 `json:"nonce"`
	PrevHash     string `json:"previous_hash"`
	Hash         string `json:"hash"`
	Content      string `json:"content"`
	Signature    string `json:"signature"`
	Type         string `json:"type"`
	PubKey       string `json:"public_key"`
	Date         int64  `json:"date"`
	BubbleBabble string `json:"bubblebabble"`
}

// New returns a configured instance of the API server
func New(c config.Configuration, n *node.Node) *API {
	a := &API{
		node:         n,
		keyfile:      c.Global.SSLKey,
		certfile:     c.Global.SSLCert,
		Message:      c.Global.Message,
		adminEnabled: c.Web.API.AdminEnabled,
		user:         c.Web.API.AdminUser,
		password:     c.Web.API.AdminPassword,
	}
	a.ListenInterface = c.Web.API.Interface + ":" + strconv.Itoa(c.Web.API.Port)
	return a
}

// Run starts the API server as specified in the configuration
func (a *API) Run() error {
	e := echo.New()
	e.HideBanner = true
	e.Logger = logrusmiddleware.Logger{log.StandardLogger()}
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		Skipper:       middleware.DefaultSkipper,
		AllowOrigins:  []string{"*"},
		AllowMethods:  []string{echo.GET, echo.HEAD, echo.PUT, echo.PATCH, echo.POST, echo.DELETE},
		ExposeHeaders: []string{"X-Server-Message"},
	}))

	serverMessage := func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Response().Header().Set("X-Server-Message", a.Message)
			return next(c)
		}
	}

	e.Use(serverMessage)

	apiV1 := e.Group("/api/v1")
	apiV1.GET("/status", a.getStatus)
	log.Infof("Starting API Server on interface %s", a.ListenInterface)
	return e.StartTLS(a.ListenInterface, a.certfile, a.keyfile)
}

func (a *API) getStatus(c echo.Context) error {
	return c.JSON(http.StatusOK, a.node.Status())
}

func (a *API) uploadImage(c echo.Context) error {
	// if !a.node.ImageChain.Valid() {
	// 	return c.JSON(http.StatusInternalServerError, Error{Code: http.StatusInternalServerError, Message: chain.ErrInvalidChain.Error()})
	// }
	// nonce := c.FormValue("nonce")
	// prevHash, err := decodeHash(c.FormValue("prevHash"))
	// if err != nil {
	// 	return c.JSON(http.StatusBadRequest, Error{Message: "Invalid field: PrevHash", Code: http.StatusBadRequest})
	// }
	// ts, err := strconv.ParseInt(c.FormValue("timestamp"), 10, 64)
	// if err != nil {
	// 	return c.JSON(http.StatusBadRequest, Error{Message: "Invalid field: timestamp", Code: http.StatusBadRequest})
	// }
	// rh, err := decodeHash(c.FormValue("hash"))
	// if err != nil {
	// 	return c.JSON(http.StatusBadRequest, Error{Message: "Invalid field: Hash", Code: http.StatusBadRequest})
	// }
	// n, err := strconv.ParseUint(nonce, 10, 32)
	// if err != nil {
	// 	return c.JSON(http.StatusBadRequest, Error{Message: "Invalid field: Nonce", Code: http.StatusBadRequest})
	// }
	// d := time.Unix(ts, 0)
	// bl := chain.Block{
	// 	Nonce:    uint32(n),
	// 	PrevHash: prevHash,
	// 	Type:     "image",
	// 	Date:     d,
	// }
	// file, err := c.FormFile("image")
	// if err != nil {
	// 	return c.JSON(http.StatusBadRequest, Error{Message: "Could not find image", Code: http.StatusBadRequest})
	// }
	// src, err := file.Open()
	// if err != nil {
	// 	return c.JSON(http.StatusBadRequest, Error{Message: "Could not process image", Code: http.StatusBadRequest})
	// }
	// defer src.Close()

	// buff := bytes.NewBuffer([]byte{})
	// io.Copy(buff, src)
	// if buff.Len() >= node.MaxMsgSize {
	// 	return c.JSON(http.StatusBadRequest, Error{Message: "Image to large, please compress it further or crop it", Code: http.StatusBadRequest})
	// }
	// bl.Content = base64.URLEncoding.EncodeToString(buff.Bytes())
	// if bl.Hash() != rh {
	// 	return c.JSON(http.StatusBadRequest, Error{Message: "Invalid hash. Please recalculate the nonce", Code: http.StatusBadRequest})
	// }
	// a.node.Push(&bl)
	return c.NoContent(http.StatusCreated)
}

func (a *API) getImage(c echo.Context) error {
	// if !a.node.ImageChain.Valid() {
	// 	return c.JSON(http.StatusInternalServerError, Error{Code: http.StatusInternalServerError, Message: chain.ErrInvalidChain.Error()})
	// }
	// h, t := decodeImageHash(c.Param("hash"))
	// if h.Empty() {
	// 	return c.JSON(http.StatusBadRequest, Error{Code: http.StatusBadRequest, Message: "Could not decode hash"})
	// }
	// ib := a.node.ImageChain.Get(h)
	// if ib == nil {
	// 	return c.JSON(http.StatusNotFound, Error{Message: "Image not found", Code: http.StatusNotFound})
	// }
	// dr := base64.NewDecoder(base64.URLEncoding, strings.NewReader(ib.Content))
	// img, _, err := image.Decode(dr)
	// if err != nil {
	// 	log.Error(err)
	// 	return c.JSON(http.StatusBadRequest, Error{Message: "Could not process image", Code: http.StatusBadRequest})
	// }

	// if t == "" {
	// 	t = c.Request().Header.Get("Accept")
	// }

	// switch t {
	// case "image/jpeg":
	// 	c.Response().Header().Set("Content-Type", "image/jpeg")
	// 	jpeg.Encode(c.Response().Writer, img, &jpeg.Options{Quality: 80})
	// 	return nil
	// case "image/png":
	// 	c.Response().Header().Set("Content-Type", "image/png")
	// 	png.Encode(c.Response().Writer, img)
	// 	return nil
	// default:
	// 	return c.JSON(http.StatusBadRequest, Error{Message: "Please indicate the requested format with the Accept header or the file type", Code: http.StatusBadRequest})
	// }
	return c.NoContent(http.StatusOK)
}

func (a *API) addBlock(c echo.Context) error {
	// block := new(jsonBlock)
	// if err := c.Bind(block); err != nil {
	// 	return err
	// }
	// b := chain.Block{
	// 	Content:   block.Content,
	// 	Signature: block.Signature,
	// 	Nonce:     block.Nonce,
	// 	Type:      block.Type,
	// 	Date:      time.Unix(block.Date, 0),
	// 	PubKey:    block.PubKey,
	// }
	// hash, err := decodeHash(block.Hash)
	// if err != nil {
	// 	return c.JSON(http.StatusBadRequest, Error{
	// 		Code:    http.StatusBadRequest,
	// 		Message: err.Error(),
	// 	})
	// }

	// prevhash, err := decodeHash(block.PrevHash)
	// if err != nil {
	// 	return c.JSON(http.StatusBadRequest, Error{
	// 		Code:    http.StatusBadRequest,
	// 		Message: err.Error(),
	// 	})
	// }
	// b.PrevHash = prevhash
	// if b.Hash() != hash {
	// 	h := b.Hash()
	// 	log.Debugf("Should: %s, Was: %s", base64.URLEncoding.EncodeToString(h[:]), base64.URLEncoding.EncodeToString(hash[:]))
	// 	return c.JSON(http.StatusBadRequest, Error{Code: http.StatusBadRequest, Message: "Block hash did not match its contents"})
	// }
	// a.node.SubmitBlock(b)
	return c.NoContent(http.StatusCreated)
}

func (a *API) getSearch(c echo.Context) error {
	// results := []jsonBlock{}
	// bs := a.node.PostChain.Search(c.QueryParam("q"))
	// if len(bs) == 0 {
	// 	return c.JSON(http.StatusNotFound, Error{Message: "No results found", Code: http.StatusNotFound})
	// }
	// for _, b := range bs {
	// 	results = append(results, jsonize(b))
	// }
	// return c.JSON(http.StatusOK, struct {
	// 	Results []jsonBlock `json:"results"`
	// }{Results: results})
	return c.NoContent(http.StatusCreated)
}

func (a *API) getBlocks(c echo.Context) error {
	// results := []jsonBlock{}
	// ls := c.QueryParam("limit")
	// limit := 10
	// if ls != "" {
	// 	ln, err := strconv.Atoi(ls)
	// 	if err == nil {
	// 		limit = ln
	// 	}
	// }
	// os := c.QueryParam("offset")
	// offset := a.node.PostChain.LastHash()
	// if os != "" {
	// 	osr, err := decodeHash(os)
	// 	if err == nil {
	// 		offset = osr
	// 	}
	// }
	// switch c.Param("type") {
	// case "key", "image":
	// 	return c.JSON(http.StatusBadRequest, Error{Code: http.StatusBadRequest, Message: "This operation is only supported for type=post"})
	// case "post":
	// 	l, err := a.node.PostChain.Latest(limit, offset)
	// 	if err != nil {
	// 		return c.JSON(http.StatusInternalServerError, Error{Code: http.StatusInternalServerError, Message: err.Error()})
	// 	}
	// 	for _, b := range l {
	// 		results = append(results, jsonize(b))
	// 	}
	// }
	// return c.JSON(http.StatusOK, struct {
	// 	Results []jsonBlock `json:"results"`
	// }{Results: results})
	return c.NoContent(http.StatusCreated)
}
