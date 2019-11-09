package handler

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/auyer/massmoverbot/web/service"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"golang.org/x/oauth2"
)

// Claims struct stores the user ID and the Guilds its able to access
type Claims struct {
	ID     string   `json:"id"`
	Guilds []string `json:"srv"`
	jwt.StandardClaims
}

type key int

const (
	// JwtKeyID represents the JWT value in Contexts
	JwtKeyID key = iota
	// ClaimsKeyID represents the Calims value in Contexts
	ClaimsKeyID
)

var mySigningKey []byte

// Handler struct stores the necessary information for the handlers
type Handler struct {
	OauthConfig      *oauth2.Config
	OauthStateString string
	Service          service.Bot
}

// GenerateRandomBytes returns securely generated random bytes.
// It will return an error if the system's secure random
// number generator fails to function correctly, in which
// case the caller should not continue.
func GenerateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	// Note that err == nil only if we read len(b) bytes.
	if err != nil {
		return nil, err
	}

	return b, nil
}

// GenerateRandomString returns a URL-safe, base64 encoded
// securely generated random string.
func GenerateRandomString(s int) (string, error) {
	b, err := GenerateRandomBytes(s)
	return base64.URLEncoding.EncodeToString(b), err
}

// NewHandler function initializes the Handler struct with the provided configs, and also a random state string
func NewHandler(oauthConfig *oauth2.Config, service service.Bot) (*Handler, error) {
	// Note : to scale this server into multiple instances, the State must be retrieved from the environment

	// State
	state, err := GenerateRandomString(20)
	if err != nil {
		return nil, err
	}

	// Signing Key
	c := 256
	mySigningKey, err = GenerateRandomBytes(c)
	if err != nil {
		return nil, err
	}

	return &Handler{
		Service:          service,
		OauthConfig:      oauthConfig,
		OauthStateString: state,
	}, nil
}

// Login handler redirects the user to the Login URL
func (h Handler) Login(w http.ResponseWriter, r *http.Request) {
	url := h.OauthConfig.AuthCodeURL(h.OauthStateString)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)

}

// Callback handler is responsible for receiving the Discord's callback with the necessary code and state for the Token exchange
func (h Handler) Callback(w http.ResponseWriter, r *http.Request) {
	id, guilds, err := h.getUserInfo(r.FormValue("state"), r.FormValue("code"))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	expiration := time.Now().Add(5 * time.Minute)
	authtoken, err := createJWT(id, guilds, expiration.Unix())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err != nil {
		fmt.Println(err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	cookie := &http.Cookie{
		Name:    "MMB_auth",
		Value:   authtoken,
		Expires: expiration,
	}
	http.SetCookie(w, cookie)
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

// Refresh handler is responsible for refreshing the JWT token
func (h Handler) Refresh() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		claims := ctx.Value(JwtKeyID).(*Claims)

		// We ensure that a new token is not issued until enough time has elapsed
		// In this case, a new token will only be issued if the old token is within
		// 30 seconds of expiry. Otherwise, return a bad request status
		if time.Unix(claims.ExpiresAt, 0).Sub(time.Now()) > 30*time.Second {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Now, create a new token for the current use, with a renewed expiration time
		expiration := time.Now().Add(5 * time.Minute)
		tokenString, err := createJWT(claims.ID, claims.Guilds, expiration.Unix())
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Set the new token as the users `session_token` cookie
		http.SetCookie(w, &http.Cookie{
			Name:    "session_token",
			Value:   tokenString,
			Expires: expiration,
		})
	})
}

// Auth handler authenticates the JWT token and sends the claims and token for the next handler via context
func (h Handler) Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("MMB_auth")
		if err != nil {
			log.Println("No cookie")
			log.Println(err)
			// h.Login(w, r) REDIRECT TO LOGIN
			return
		}
		claims := &Claims{}
		token, err := jwt.ParseWithClaims(cookie.Value, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
			}
			return mySigningKey, nil
		})
		if err != nil {
			log.Println(err)
			// h.Login(w, r) REDIRECT TO LOGIN
			return
		}
		ctx := context.WithValue(r.Context(), ClaimsKeyID, claims)
		ctx = context.WithValue(ctx, JwtKeyID, token)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

// Guilds handler returns a list of SimpleGuild objects that the user is capable of accessing
func (h Handler) Guilds() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		claims := ctx.Value(ClaimsKeyID).(*Claims)

		guilds := h.Service.GetGuilds(claims.Guilds...)

		response, err := json.Marshal(guilds)
		if err != nil {
			w.Header().Set("Content-Type", "application/text")
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(response)
	})
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

// GuildByID handler retrieves a complete Guild by its ID
func (h Handler) GuildByID() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		claims := ctx.Value(ClaimsKeyID).(*Claims)

		vars := mux.Vars(r)
		id := vars["id"]
		if !stringInSlice(id, claims.Guilds) {
			// w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error":"either you or the bot dont have permissions to see this guild.}`))
			return
		}

		guild := h.Service.GetGuild(id)

		response, err := json.Marshal(guild)
		if err != nil {
			w.Header().Set("Content-Type", "application/text")
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(response)
	})
}

// getUserInfo function exchanges the state and code by the user Token for verification, and returns the users ID, and Guilds where the bot is available
func (h Handler) getUserInfo(state string, code string) (string, []string, error) {
	ctx := context.Background()
	if state != h.OauthStateString {
		return "", nil, fmt.Errorf("invalid oauth state")
	}

	token, err := h.OauthConfig.Exchange(oauth2.NoContext, code)
	if err != nil {
		return "", nil, fmt.Errorf("code exchange failed: %s", err.Error())
	}
	// ctxTimeout, cancel := context.WithTimeout(ctx, 10)
	user, err := service.UserRequest(ctx, token)
	if err != nil {
		return "", nil, err
	}

	userGuilds, err := service.GuildsRequest(ctx, token)
	if err != nil {
		return "", nil, err
	}
	botGuilds := h.Service.GetGuildIDs()

	hash := make(map[string]bool)
	set := []string{}

	for _, uGuild := range userGuilds {
		hash[uGuild.ID] = true
	}
	for _, sGuild := range botGuilds {
		if hash[sGuild] {
			set = append(set, sGuild)
		}
	}

	return user.ID, set, nil
}

// createJWT helper function creates a signed JWT with the provided informations
func createJWT(id string, guilds []string, expiration int64) (string, error) {
	// Create the Claims
	claims := Claims{
		id,
		guilds,
		jwt.StandardClaims{
			ExpiresAt: expiration,
			Issuer:    "MMB",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(mySigningKey)
}
