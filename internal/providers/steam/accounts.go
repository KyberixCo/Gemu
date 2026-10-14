package steam

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"os"
	"path/filepath"
	"strconv"
)

const maxAvatarBytes = 512 * 1024

type Account struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Avatar string `json:"avatar,omitempty"`
}

func AvailableAccounts() ([]Account, error) {
	root, err := DetectRoot()
	if err != nil {
		return nil, err
	}
	return LocalAccounts(root)
}

// LocalAccounts reads Steam's non-secret local profile list and cached avatars.
func LocalAccounts(root string) ([]Account, error) {
	profiles, err := LocalProfiles(root)
	if err != nil {
		return nil, err
	}
	accounts := make([]Account, 0, len(profiles))
	for _, profile := range profiles {
		accounts = append(accounts, Account{
			ID:     profile.ID,
			Name:   profile.Name,
			Avatar: cachedAvatar(root, profile.ID),
		})
	}
	return accounts, nil
}

func cachedAvatar(root, steamID string) string {
	paths := []string{
		filepath.Join(root, "config", "avatarcache", steamID+".jpg"),
		filepath.Join(root, "config", "avatarcache", steamID+".png"),
		filepath.Join(root, "config", "avatarcache", steamID+"_medium.jpg"),
		filepath.Join(root, "config", "avatarcache", steamID+"_full.jpg"),
	}
	if id, err := strconv.ParseUint(steamID, 10, 64); err == nil && id >= 76561197960265728 {
		accountID := strconv.FormatUint(id-76561197960265728, 10)
		paths = append(paths, filepath.Join(root, "userdata", accountID, "config", "avatar.jpg"))
	}
	for _, path := range paths {
		data, err := readAvatar(path)
		if err != nil {
			continue
		}
		config, format, err := image.DecodeConfig(bytes.NewReader(data))
		if err != nil || config.Width < 1 || config.Height < 1 || config.Width > 1024 || config.Height > 1024 {
			continue
		}
		mime := ""
		switch format {
		case "jpeg":
			mime = "image/jpeg"
		case "png":
			mime = "image/png"
		default:
			continue
		}
		return "data:" + mime + ";base64," + base64.StdEncoding.EncodeToString(data)
	}
	return ""
}

func readAvatar(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, maxAvatarBytes+1))
	if err != nil {
		return nil, err
	}
	if len(data) > maxAvatarBytes {
		return nil, fmt.Errorf("avatar cache file is too large")
	}
	return data, nil
}
