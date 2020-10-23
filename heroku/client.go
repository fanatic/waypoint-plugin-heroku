package heroku

import (
	"fmt"
	"log"
	"os/user"
	"path/filepath"

	heroku "github.com/heroku/heroku-go/v5"
	"github.com/jdxcode/netrc"
)

func New() (*heroku.Service, error) {
	p, err := fetchPassword()
	if err != nil {
		return nil, fmt.Errorf("Please login first with `heroku login`.")
	}

	heroku.DefaultTransport.Password = p
	h := heroku.NewService(heroku.DefaultClient)

	return h, nil
}

func fetchPassword() (string, error) {
	usr, err := user.Current()
	if err != nil {
		log.Fatalf("Error finding current user: %s", err.Error())
	}
	n, err := netrc.Parse(filepath.Join(usr.HomeDir, ".netrc"))
	if err != nil {
		log.Fatalf("Error finding current user: %s", err.Error())
	}
	p := n.Machine("api.heroku.com").Get("password")
	if p == "" {
		return "", fmt.Errorf("No saved password found")
	}
	return p, nil
}
