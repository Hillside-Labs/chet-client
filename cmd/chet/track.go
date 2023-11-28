package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/hillside-labs/chet-client/lib"
	"github.com/hillside-labs/chet-client/models"
)

func Track(cmd *exec.Cmd, total time.Duration, config models.LocalConfig) error {
	label, err := lib.CreateLabelFromCommand(cmd, lib.NewCmdMatcher(lib.DefaultMatches))
	if err != nil {
		return err
	}

	env, err := lib.NewCommandEnv()
	if err != nil {
		log.Println("Error loading env: ", err)
	}

	// Track the call in the local sqlite db for the user.
	rec := &models.Record{
		Label:     label,
		Cmd:       cmd.String(),
		Duration:  total,
		Repo:      env.Repo,
		Branch:    env.Branch,
		Username:  env.User,
		OS:        env.OS,
		Container: env.Container,
	}

	//Connect to the db & save the cmd record
	db, err := models.Connect(defaultDB())
	if err != nil {
		fmt.Println("chet: couldn't open ~/.chet.db")
	}

	result := db.Create(rec)
	if result.Error != nil {
		return result.Error
	}

	err = db.First(&config).Error
	if err != nil {
		log.Println(err)
	}

	if config.DisableRemote || config.ClientID == "" || config.ClientSecret == "" {
		log.Println("the result is not being sent to the server")
		return nil
	}

	//Send the result to the server:
	token, err := GetTokenFromConfig(&config)
	if err != nil {
		fmt.Println("Error getting token")
		return err
	}
	if token != config.Token { //if we refreshed the token, then save it to the local db
		config.Token = token
		err := db.Save(config).Error
		if err != nil {
			log.Println("failed to save the token")
		}
	}

	out, err := json.Marshal(rec)
	if err != nil {
		log.Println("Error creating JSON body: ", err)
		return err
	}
	
	url := config.ServerAddress + CreateRecordEndpoint
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(out))
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

	if err != nil {
		log.Println("Error creating request: ", err)
		return err
	}

	_, err = http.DefaultClient.Do(req)

	return err

}

func GetTokenFromConfig(config *models.LocalConfig) (string, error) {
	isExpired := IsTokenExpired(config.Token)
	if !isExpired {
		return config.Token, nil
	}

	data := map[string]string{"client_id": config.ClientID, "client_secret": config.ClientSecret}
	payload, err := json.Marshal(data)
	if err != nil {
		log.Println("Error marshaling payload JSON:", err)
		return "", err
	}

	url := config.ServerAddress + GetJwtEndpoint
	resp, err := http.Post(url, "application/json", bytes.NewReader(payload))
	if err != nil {
		log.Println("Error creating request:", err)
		return "", err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var responseMap map[string]interface{}
	err = json.Unmarshal(body, &responseMap)
	if err != nil {
		return "", err
	}

	token, exists := responseMap["token"].(string)
	if !exists {
		return "", errors.New("token not found or is not a string")
	}

	return token, nil
}


func IsTokenExpired(token string) bool {
	// Split the token into its three parts: header, payload, signature
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return true // Invalid token format
	}

	// Decode the payload (second part)
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return true // Unable to decode payload
	}

	// Unmarshal the payload into a map
	var claims map[string]interface{}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return true // Unable to unmarshal payload
	}

	// Check if the "exp" claim exists and is in the future
	expiration, ok := claims["exp"].(float64)
	if !ok {
		return true // "exp" claim is missing or not a number
	}

	return time.Now().Unix() > int64(expiration)
}
