package main

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "net/http"

    "github.com/gin-gonic/gin"
    "golang.org/x/oauth2"
    "golang.org/x/oauth2/strava"
)

var (
    oauth2Config = &oauth2.Config{
        ClientID:     "YOUR_CLIENT_ID",
		ClientSecret: "YOUR_CLIENT_SECRET",
        RedirectURL:  "http://localhost:8080/callback",
        Endpoint:     strava.Endpoint,
        Scopes:       []string{"read", "activity:read"},
    }
    oauth2State = "random_state_string"
)

func main() {
    r := gin.Default()

    r.GET("/login", loginHandler)
    r.GET("/callback", callbackHandler)
    r.GET("/activities", activitiesHandler)

    r.Run(":8080")
}

func loginHandler(c *gin.Context) {
    url := oauth2Config.AuthCodeURL(oauth2State)
    c.Redirect(http.StatusTemporaryRedirect, url)
}

func callbackHandler(c *gin.Context) {
    state := c.Query("state")
    if state != oauth2State {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid state"})
        return
    }

    code := c.Query("code")
    token, err := oauth2Config.Exchange(context.Background(), code)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to exchange token"})
        return
    }

    http.SetCookie(c.Writer, &http.Cookie{
        Name:  "access_token",
        Value: token.AccessToken,
        Path:  "/",
    })

    c.Redirect(http.StatusTemporaryRedirect, "/activities")
}

func activitiesHandler(c *gin.Context) {
    cookie, err := c.Request.Cookie("access_token")
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "missing access token"})
        return
    }

    client := oauth2Config.Client(context.Background(), &oauth2.Token{AccessToken: cookie.Value})
    resp, err := client.Get("https://www.strava.com/api/v3/athlete/activities")
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch activities"})
        return
    }
    defer resp.Body.Close()

    var activities []map[string]interface{}
    if err := json.NewDecoder(resp.Body).Decode(&activities); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to decode activities"})
        return
    }

    c.JSON(http.StatusOK, activities)
}