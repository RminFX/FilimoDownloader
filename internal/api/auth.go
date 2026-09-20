package api

import (
	"encoding/json"
	"fmt"
	"strings"

	"FilimoDownloader-GholamTaksir/internal/helper"
)

type Auth struct {
	Data AuthData `json:"data"`
}

type AuthData struct {
	User AuthUser `json:"user"`
}

type AuthUser struct {
	Profile AuthProfile `json:"selectedProfile"`
}

type AuthProfile struct {
	Name string `json:"name"`
}

func GetUserName(client helper.HttpClient) string {
	name, err := GetUserNameErr(client)
	if err != nil {
		helper.ShowErrorAndExit(err.Error())
	}
	return name
}

func GetUserNameErr(client helper.HttpClient) (string, error) {
	var auth Auth
	response, err := client.Get("https://api.filimo.com/api/fa/v1/web/config/uxEvent")
	if err != nil {
		return "", fmt.Errorf("خواندن حساب فیلیمو ممکن نشد: %w", err)
	}
	if err = json.Unmarshal([]byte(response), &auth); err != nil {
		return "", fmt.Errorf("پاسخ حساب فیلیمو نامعتبر است")
	}
	name := strings.TrimSpace(auth.Data.User.Profile.Name)
	if name == "" {
		return "", fmt.Errorf("توکن نامعتبر است")
	}
	return name, nil
}
