package util

import (
	"encoding/json"
	"os"
	"path"

	"github.com/golang/glog"
	"golang.org/x/oauth2"
)

type tokenCache struct {
	source    oauth2.TokenSource
	cacheFile string
}

func (t *tokenCache) Token() (*oauth2.Token, error) {
	if tok, err := parseTokenFromFile(t.cacheFile); err == nil && tok.Valid() {
		return tok, nil
	}
	tok, err := t.source.Token()
	if err != nil {
		return nil, err
	}
	if err := saveTokenToFile(tok, t.cacheFile); err != nil {
		glog.Warningf("Failed to save token to file: %v", err)
	}
	return tok, nil
}

func parseTokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	var t oauth2.Token
	if err := json.NewDecoder(f).Decode(&t); err != nil {
		return nil, err
	}
	return &t, nil
}

func saveTokenToFile(token *oauth2.Token, file string) error {
	tok := *token
	tok.RefreshToken = ""
	dir := path.Dir(file)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}
	f, err := os.Create(file)
	if err != nil {
		return err
	}
	if err := json.NewEncoder(f).Encode(&tok); err != nil {
		return err
	}
	return nil
}

func NewCachedTokenSource(source oauth2.TokenSource, cacheFile string) oauth2.TokenSource {
	return &tokenCache{source, cacheFile}
}
