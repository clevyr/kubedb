package cli

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"
)

var UserDeclinedErr = errors.New("user declined")

func Confirm(r io.Reader, defaultVal bool) (err error) {
	fmt.Print("Continue? ")
	if defaultVal {
		fmt.Print("[Y/n]: ")
	} else {
		fmt.Print("[y/N]: ")
	}

	buf := bufio.NewReader(r)
	var response string
	response, err = buf.ReadString('\n')
	if err != nil {
		return err
	}

	response = strings.ToLower(strings.TrimSpace(response))
	switch response {
	case "yes", "y":
		return nil
	case "no", "n":
		return UserDeclinedErr
	}
	if defaultVal {
		return nil
	} else {
		return UserDeclinedErr
	}
}
