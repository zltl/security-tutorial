package main

import (
	"net/http"
	"st001/res"
	"sync"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"github.com/sirupsen/logrus"
)

type UserID int
type UserSecret string
type UserCookieI string

var (
	startTime = time.Now()

	tokenExpireTime = time.Hour * 120

	id2Secret map[UserID]UserSecret
	maxID     = 1007
	mu        sync.Mutex

	comments []Comment = []Comment{}

	cookieIKey   = "I"
	contextKeyID = "ID"

	jwtSecret = "SECRET"
)

func newUserID(c *gin.Context) {
	mu.Lock()
	newID := UserID(maxID + 1)
	maxID = int(newID)
	uid, _ := uuid.NewV4()
	secret := UserSecret(uid.String())
	id2Secret[newID] = secret
	mu.Unlock()

	claims := jwt.MapClaims{}
	claims["user_id"] = newID
	claims["exp"] = time.Now().Add(tokenExpireTime)
	claims["create_at"] = time.Now()
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	I, err := at.SignedString([]byte(jwtSecret))
	if err != nil {
		logrus.Error(err)
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	logrus.Tracef("no cookieI, create one user for this request: id=%d, secret=%s", newID, secret)
	c.SetCookie(cookieIKey, I, int(tokenExpireTime/time.Second), "/", "", true, false)
	c.Set(contextKeyID, newID)
}

// middleware, to ensure all request have an user id.
func handleEnsureUserID(c *gin.Context) {
	ck, err := c.Request.Cookie(cookieIKey)
	if err != nil {
		newUserID(c)
		return
	}

	claim, err := jwt.Parse(ck.Value, func(token *jwt.Token) (interface{}, error) {
		return []byte(jwtSecret), nil
	})
	if err != nil {
		// logrus.Error(err)
		newUserID(c)
		return
	}

	mapClaim := claim.Claims.(jwt.MapClaims)
	id := UserID(int(mapClaim["user_id"].(float64)))
	c.Set(contextKeyID, id)
}

type Comment struct {
	ID  UserID
	Msg string
}

type MainPageArgs struct {
	ID       UserID
	Secret   UserSecret
	Comments []Comment
}

func handleMainPage(c *gin.Context) {
	idi, ok := c.Get(contextKeyID)
	if !ok {
		logrus.Errorf("id not found")
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	id := idi.(UserID)
	mu.Lock()
	secret := id2Secret[id]
	mu.Unlock()

	logrus.Infof("id=%d requesting", id)

	mainPage := res.TMPL.Lookup("main_page.tmpl")

	mu.Lock() // protect slice 'comments'
	args := MainPageArgs{
		ID:       id,
		Secret:   secret,
		Comments: comments,
	}
	mainPage.ExecuteTemplate(c.Writer, "main_page", &args)
	mu.Unlock()
}

func handlePostComment(c *gin.Context) {
	idi, ok := c.Get(contextKeyID)
	if !ok {
		logrus.Errorf("id not found")
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	id := idi.(UserID)

	if err := c.Request.ParseForm(); err != nil {
		logrus.Errorf("parse form error: %s", err)
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	commentMsg := c.Request.PostFormValue("input_comments")

	mu.Lock()
	comments = append(comments, Comment{
		ID:  id,
		Msg: commentMsg,
	})
	mu.Unlock()

	c.Redirect(http.StatusMovedPermanently, "/")
}

func main() {
	id2Secret = make(map[UserID]UserSecret)
	randstr, _ := uuid.NewV4()
	jwtSecret = randstr.String()

	logrus.SetLevel(logrus.TraceLevel)
	logrus.SetReportCaller(true)

	gin.ForceConsoleColor()
	router := gin.Default()
	router.SetTrustedProxies(nil)

	router.Use(handleEnsureUserID)
	router.GET("/", handleMainPage)
	router.POST("/new_comment", handlePostComment)
	router.Run(":8080")
}
